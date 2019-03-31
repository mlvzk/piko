package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	shovel "github.com/mlvzk/shovel-go/services"
)

func main() {
	services := []shovel.Service{shovel.ImgurService{}}

	target := "https://imgur.com/t/article13/EfY6CxU"
	for _, service := range services {
		if !service.IsValidTarget(target) {
			break
		}

		fmt.Printf("Found valid service: %+v\n", service)
		items, err := service.FetchItems(target)
		if err != nil {
			log.Fatalln("Error: ", err)
		}

		for _, item := range items {
			reader, err := service.Download(item.Meta, nil)
			if err != nil {
				log.Printf("Download error: %v, item: %+v\n", err, item)
			}

			file, err := os.Create(item.DefaultName)
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

		fmt.Printf("items: %+v\n", items)
	}
}

func format(formatter string, meta map[string]string) string {
	for key, value := range meta {
		formatter = strings.ReplaceAll(formatter, fmt.Sprintf("%%[%s]", key), value)
	}

	return formatter
}
