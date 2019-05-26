// Copyright 2019 mlvzk
// This file is part of the piko library.
//
// The piko library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The piko library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the piko library. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mlvzk/piko"
	"github.com/mlvzk/piko/service"
	"github.com/mlvzk/qtils/commandparser"
	"github.com/mlvzk/qtils/commandparser/commandhelper"
	"gopkg.in/cheggaaa/pb.v1"
)

var (
	discoveryMode bool
	stdoutMode    bool
	formatStr     string
	targets       []string
	userOptions   = map[string]string{}
)

func handleArgv(argv []string) {
	parser := commandparser.New()
	helper := commandhelper.New()

	helper.SetName("piko")
	helper.SetVersion("alpha")
	helper.AddAuthor("mlvzk")

	helper.AddUsage(
		"piko [urls...]",
		"piko 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'",
		"piko 'https://www.youtube.com/watch?v=dQw4w9WgXcQ' --stdout | mpv -",
	)

	parser.AddOption(helper.EatOption(
		commandhelper.NewOption("help").Alias("h").Boolean().Description("Prints this page"),
		commandhelper.
			NewOption("format").
			Alias("f").
			Description(`File path format, ex: --format %[id].%[ext]. "id" and "ext" are meta tags(see --discover).
Use %[default] to fill with default format, ex: downloads/%[default]`),
		commandhelper.
			NewOption("option").
			Alias("o").
			Arrayed().
			ValidateBind(commandhelper.ValidateKeyValue("=")).
			Description("Download options, ex: --option quality=best"),
		commandhelper.
			NewOption("discover").
			Alias("d").
			Boolean().
			Description("Discovery mode, doesn't download anything, only outputs information"),
		commandhelper.NewOption("stdout").Boolean().Description("Output download media to stdout"),
	)...)

	cmd, err := parser.Parse(argv)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if cmd.Booleans["help"] || len(cmd.Positionals) == 0 {
		fmt.Print(helper.Help())
		os.Exit(1)
	}

	cmd.Args = helper.FillDefaults(cmd.Args)
	errs := helper.Verify(cmd.Args, cmd.Arrayed)
	for _, err := range errs {
		log.Println(err)
	}
	if len(errs) != 0 {
		os.Exit(1)
	}

	formatStr = cmd.Args["format"]
	discoveryMode = cmd.Booleans["discover"]
	stdoutMode = cmd.Booleans["stdout"]

	for _, option := range cmd.Arrayed["option"] {
		keyValue := strings.Split(option, "=")
		key, value := keyValue[0], keyValue[1]

		userOptions[key] = value
	}

	targets = cmd.Positionals
}

func main() {
	// handleArgv can not be in init()
	// because it would be called(and errored)
	// if tests were run in main_test
	handleArgv(os.Args)

	services := piko.GetAllServices()

	// target = "https://boards.4channel.org/adv/thread/20765545/i-want-to-be-the-very-best-like-no-one-ever-was"
	// target = "https://imgur.com/t/article13/EfY6CxU"
	// target = "https://www.youtube.com/watch?v=Gs069dndIYk"
	// target = "https://www.youtube.com/watch?v=7IwYakbxmxo"
	// target = "https://www.instagram.com/p/Bv9MJCsAvZV/"
	// target = "https://soundcloud.com/musicpromouser/mac-miller-ok-ft-tyler-the-creator"
	// target = "https://twitter.com/deadprogram/status/1090554988768698368"
	// target = "https://www.facebook.com/groups/veryblessedimages/permalink/478153699389793/"

	for _, target := range targets {
		if target == "" {
			log.Println("Target can't be empty")
			continue
		}

		var foundAnyService bool
		for _, s := range services {
			if !s.IsValidTarget(target) {
				continue
			}

			foundAnyService = true
			log.Printf("Found valid service: %s\n", reflect.TypeOf(s).Name())
			iterator, err := s.FetchItems(target)
			if err != nil {
				log.Printf("failed to fetch items: %v; target: %v\n", err, target)
				break
			}

			for !iterator.HasEnded() {
				items, err := iterator.Next()
				if err != nil {
					log.Printf("Iteration error: %v; target: %v\n", err, target)
					continue
				}

				for _, item := range items {
					handleItem(s, item)
				}
			}

			break
		}
		if !foundAnyService {
			log.Printf("Error: Couldn't find a valid service for url '%s'. Your link is probably unsupported.\n", target)
		}
	}
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

	dir := filepath.Dir(name)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Printf("Error creating directory: %v; dir: '%v'\n", err, dir)
		return
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
