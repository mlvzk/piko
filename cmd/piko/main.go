package main

import (
	"flag"
	"io"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/mlvzk/piko"
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
	flag.StringVar(&formatStr, "format", "", "File path format, ex: -format %[id].%[ext]. Use %[default] to fill with default format, ex: downloads/%[default]")
	flag.StringVar(&optionsStr, "options", "", "Download options, ex: -options thumbnail=yes,quality=high")
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
	// target = "https://twitter.com/deadprogram/status/1090554988768698368"
	// target = "https://www.facebook.com/groups/veryblessedimages/permalink/478153699389793/"

	if target == "" {
		log.Println("Target can't be empty")
		return
	}

	for _, s := range services {
		if !s.IsValidTarget(target) {
			continue
		}

		log.Printf("Found valid service: %s\n", reflect.TypeOf(s).Name())
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

		return
	}
	log.Println("Couldn't find a valid service. Your link is probably unsupported.")
}

func handleItem(s service.Service, item service.Item) {
	if discoveryMode {
		log.Println("Item:\n" + prettyPrintItem(item))
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
		nameFormat = strings.Replace(formatStr, "%[default]", item.DefaultName, -1)
	}
	name := format(nameFormat, item.Meta)

	if sizedIO, ok := reader.(service.Sized); ok {
		bar := pb.New64(int64(sizedIO.Size())).SetUnits(pb.U_BYTES)
		bar.Prefix(truncateString(name, 25))
		bar.Start()
		defer bar.Finish()

		reader = bar.NewProxyReader(reader)
	}

	file, err := os.Create(name)
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
