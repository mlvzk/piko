package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/mlvzk/shovel-go"
)

func main() {
	var formatStr string
	var optionsStr string
	flag.StringVar(&formatStr, "format", "", "File path format, ex: -format %[id].%[ext]")
	flag.StringVar(&optionsStr, "options", "", "Download options, ex: -o thumbnail=yes,quality=high")
	flag.Parse()

	userOptions := parseOptions(optionsStr)

	services := shovel.GetAllServices()

	// target := "https://boards.4channel.org/g/thread/70361348/new-desktop-thread"
	// target := "https://imgur.com/t/article13/EfY6CxU"
	// target := "https://www.youtube.com/watch?v=HOK0uF-Z0xM"
	// target := "https://www.instagram.com/p/Bv9MJCsAvZV/"
	target := "https://soundcloud.com/musicpromouser/mac-miller-ok-ft-tyler-the-creator"
	for _, service := range services {
		if !service.IsValidTarget(target) {
			continue
		}

		fmt.Printf("Found valid service: %+v\n", service)
		iterator := service.FetchItems(target)

		for !iterator.HasEnded() {
			items, err := iterator.Next()
			if err != nil {
				log.Printf("Iteration error: %v; target: %v", err, target)
				break
			}

			for _, item := range items {
				options := mergeStringMaps(item.DefaultOptions, userOptions)

				reader, err := service.Download(item.Meta, options)
				if err != nil {
					log.Printf("Download error: %v, item: %+v\n", err, item)
					continue
				}

				nameFormat := item.DefaultName
				if formatStr != "" {
					nameFormat = formatStr
				}
				name := format(nameFormat, item.Meta)

				file, err := os.Create("downloads/" + name)
				if err != nil {
					log.Printf("Error creating file: %v, name: %v\n", err, name)
					tryClose(reader)
					continue
				}

				_, err = io.Copy(file, reader)
				if err != nil {
					log.Printf("Error copying from source to file: %v, item: %+v", err, item)
					tryClose(reader)
					file.Close()
					continue
				}

				file.Close()
				tryClose(reader)
			}
		}
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

var seps = regexp.MustCompile(`[\r\n &_=+:]`)

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	name = seps.ReplaceAllString(name, "-")

	return name
}
