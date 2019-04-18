package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	piko "github.com/mlvzk/piko"
	"github.com/mlvzk/piko/service"
	"gopkg.in/cheggaaa/pb.v1"
)

var (
	discoveryMode bool
	stdoutMode    bool
	formatStr     string
	userOptions   map[string]string
)

func init() {
	var optionsStr string
	flag.StringVar(&formatStr, "format", "", "File path format, ex: -format %[id].%[ext]")
	flag.StringVar(&optionsStr, "options", "", "Download options, ex: -o thumbnail=yes,quality=high")
	flag.BoolVar(&discoveryMode, "discover", false, "Discovery mode, doesn't download anything, only outputs information")
	flag.BoolVar(&stdoutMode, "stdout", false, "Output download media to stdout")
	flag.Parse()

	userOptions = parseOptions(optionsStr)
}

func main() {
	services := piko.GetAllServices()

	target := flag.Arg(0)

	// target = "https://boards.4channel.org/adv/thread/20765545/i-want-to-be-the-very-best-like-no-one-ever-was"
	// target = "https://imgur.com/t/article13/EfY6CxU"
	// target = "https://www.youtube.com/watch?v=Gs069dndIYk"
	// target = "https://www.instagram.com/p/Bv9MJCsAvZV/"
	// target = "https://soundcloud.com/musicpromouser/mac-miller-ok-ft-tyler-the-creator"

	if target == "" {
		log.Println("Target can't be empty")
		return
	}

	for _, s := range services {
		if !s.IsValidTarget(target) {
			continue
		}

		fmt.Printf("Found valid service: %+v\n", s)
		iterator := s.FetchItems(target)

		for !iterator.HasEnded() {
			items, err := iterator.Next()
			if err != nil {
				log.Printf("Iteration error: %v; target: %v", err, target)
				break
			}

			for _, item := range items {
				handleItem(s, item)
			}
		}
	}
}

func handleItem(s service.Service, item service.Item) {
	if discoveryMode {
		fmt.Println(prettyPrintItem(item))
		return
	}

	options := mergeStringMaps(item.DefaultOptions, userOptions)

	reader, err := s.Download(item.Meta, options)
	if err != nil {
		log.Printf("Download error: %v, item: %+v\n", err, item)
		return
	}
	defer tryClose(reader)

	if stdoutMode {
		io.Copy(os.Stdout, reader)
		return
	}

	nameFormat := item.DefaultName
	if formatStr != "" {
		nameFormat = formatStr
	}
	name := format(nameFormat, item.Meta)

	if sizedIO, ok := reader.(service.Sized); ok {
		bar := pb.New64(int64(sizedIO.Size())).SetUnits(pb.U_BYTES)
		bar.Prefix(truncateString(name, 15))
		bar.Start()
		defer bar.Finish()

		reader = bar.NewProxyReader(reader)
	}

	file, err := os.Create("downloads/" + name)
	if err != nil {
		log.Printf("Error creating file: %v, name: %v\n", err, name)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		log.Printf("Error copying from source to file: %v, item: %+v", err, item)
		return
	}
}

var formatRegexp = regexp.MustCompile(`%\[[[:alnum:]]*\]`)

func format(formatter string, meta map[string]string) string {
	return formatRegexp.ReplaceAllStringFunc(formatter, func(str string) string {
		key := str[2 : len(str)-1]

		v, ok := meta[key]
		if !ok {
			return key
		}

		return sanitizeFileName(v)
	})
}

func parseOptions(optionsStr string) map[string]string {
	options := map[string]string{}

	declarations := strings.Split(optionsStr, ",")
	for _, declaration := range declarations {
		if declaration == "" {
			continue
		}

		keyValue := strings.Split(declaration, "=")
		key, value := keyValue[0], keyValue[1]

		options[key] = value
	}

	return options
}

// order of args matters
func mergeStringMaps(maps ...map[string]string) map[string]string {
	merged := map[string]string{}

	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}

	return merged
}

func tryClose(reader interface{}) error {
	if closer, ok := reader.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}

var seps = regexp.MustCompile(`[\r\n &_=+:/]`)

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	name = seps.ReplaceAllString(name, "-")

	return name
}

func prettyPrintItem(item service.Item) string {
	builder := strings.Builder{}

	builder.WriteString("Default Name: " + item.DefaultName + "\n")

	builder.WriteString("Meta:\n")
	for k, v := range item.Meta {
		if k[0] == '_' {
			continue
		}

		builder.WriteString(fmt.Sprintf("\t%s=%s\n", k, v))
	}

	builder.WriteString("Available Options:\n")
	for key, values := range item.AvailableOptions {
		builder.WriteString("\t" + key + ":\n")
		for _, v := range values {
			builder.WriteString("\t\t- " + v + "\n")
		}
	}

	builder.WriteString("Default Options:\n")
	for k, v := range item.DefaultOptions {
		builder.WriteString(fmt.Sprintf("\t%s=%s\n", k, v))
	}

	return builder.String()
}

func truncateString(str string, limit int) string {
	if len(str)-3 <= limit {
		return str
	}

	return string([]rune(str)[:limit]) + "..."
}
