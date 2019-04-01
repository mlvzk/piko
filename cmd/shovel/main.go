package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"

	shovel "github.com/mlvzk/shovel-go/services"
)

func main() {
	services := []shovel.Service{shovel.Imgur{}, shovel.Fourchan{}}

	// target := "https://boards.4channel.org/g/thread/70361348/new-desktop-thread"
	target := "https://imgur.com/t/article13/EfY6CxU"
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
			fmt.Printf("items: %+v\n", items)

			for _, item := range items {
				reader, err := service.Download(item.Meta, item.DefaultOptions)
				if err != nil {
					log.Printf("Download error: %v, item: %+v\n", err, item)
				}

				file, err := os.Create("downloads/" + item.DefaultName)
				if err != nil {
					log.Printf("Error creating file: %v, name: %v\n", err, item.DefaultName)
				}

				_, err = io.Copy(file, reader)
				if err != nil {
					log.Printf("Error copying from source to file: %v, item: %+v", err, item)
				}

				file.Close()
				reader.Close()
			}
		}
	}
}

func format(formatter string, meta map[string]string) string {
	re := regexp.MustCompile(`%\[[[:alnum:]]*\]`)

	return re.ReplaceAllStringFunc(formatter, func(str string) string {
		fmt.Println("match: ", str)
		key := str[2 : len(str)-1]

		v, ok := meta[key]
		if !ok {
			return key
		}

		return v
	})
}
