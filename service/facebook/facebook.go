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

package facebook

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mlvzk/piko/service"
)

type output struct {
	io.ReadCloser
	length uint64
}

func (o output) Size() uint64 {
	return o.length
}

type Facebook struct{}
type FacebookIterator struct {
	url string
	end bool
}

func New() Facebook {
	return Facebook{}
}

func (s Facebook) IsValidTarget(target string) bool {
	return strings.Contains(target, "facebook.com/")
}

func (s Facebook) FetchItems(target string) (service.ServiceIterator, error) {
	return &FacebookIterator{
		url: target,
	}, nil
}

func (s Facebook) Download(meta, options map[string]string) (io.Reader, error) {
	downloadURL, hasDownloadURL := meta["downloadURL"]
	if !hasDownloadURL {
		return nil, errors.New("Missing meta downloadURL")
	}

	res, err := http.Get(downloadURL)
	if err != nil {
		return nil, err
	}

	if res.ContentLength == -1 {
		return res.Body, nil
	}

	return output{
		ReadCloser: res.Body,
		length:     uint64(res.ContentLength),
	}, nil
}

var srcRegexp = regexp.MustCompile(`(sd|hd)_src:"(.+?)"`)

func (i *FacebookIterator) Next() ([]service.Item, error) {
	i.end = true

	resp, err := http.Get(i.url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GET %v returned a wrong status code - %v", i.url, resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	title, _ := doc.Find(`meta[property="og:title"]`).Attr("content")
	description, _ := doc.Find(`meta[property="og:description"]`).Attr("content")
	image, hasImage := doc.Find(`meta[property="og:image"]`).Attr("content")

	var bestVideo, worstVideo string
	matches := srcRegexp.FindAllSubmatch(bodyBytes, -1)
	if matches != nil {
		for _, match := range matches {
			if len(match) < 3 {
				continue
			}

			if string(match[1]) == "hd" {
				bestVideo = string(match[len(match)-1])
			} else {
				worstVideo = string(match[len(match)-1])
			}
		}
	}

	items := []service.Item{}

	if hasImage && image != "" {
		imageURL, err := url.Parse(image)
		if err != nil {
			return nil, err
		}
		pathParts := strings.Split(imageURL.Path, "_")
		id := ""
		if len(pathParts) > 2 {
			id = pathParts[2]
		}

		items = append(items, service.Item{
			Meta: map[string]string{
				"id":          id,
				"author":      title,
				"description": description,
				"type":        "image",
				"ext":         "jpg",
				"downloadURL": image,
			},
			DefaultName: "%[author]-%[id].%[ext]",
		})
	}

	if bestVideo != "" || worstVideo != "" {
		video := bestVideo
		if video == "" {
			video = worstVideo
		}

		videoURL, err := url.Parse(video)
		if err != nil {
			return nil, err
		}

		id := ""
		if pathParts := strings.Split(videoURL.Path, "_"); len(pathParts) > 2 {
			id = pathParts[2]
		}

		items = append(items, service.Item{
			Meta: map[string]string{
				"id":          id,
				"author":      title,
				"description": description,
				"type":        "video",
				"ext":         "mp4",
				"downloadURL": video,
			},
			DefaultName: "%[author]-%[id].%[ext]",
		})
	}

	// don't care if this fails
	// this code fetches all images from album
	func() {
		hiddenElemText, found := "", false
		doc.Find("div.hidden_elem code").Each(func(_ int, sel *goquery.Selection) {
			selText, err := sel.Html()
			if err != nil {
				return
			}

			if !strings.Contains(selText, `rel="theater"`) {
				return
			}

			hiddenElemText, found = selText, true
		})

		if !found {
			return
		}

		uncommented := hiddenElemText[4 : len(hiddenElemText)-3]
		hiddenElem, err := goquery.NewDocumentFromReader(strings.NewReader(uncommented))
		if err != nil {
			return
		}

		hiddenElem.Find(`a[rel="theater"]`).Each(func(_ int, sel *goquery.Selection) {
			media, exists := sel.Attr("data-ploi")
			if !exists {
				return
			}

			mediaURL, err := url.Parse(media)
			if err != nil {
				return
			}

			id := ""
			if pathParts := strings.Split(mediaURL.Path, "_"); len(pathParts) > 2 {
				id = pathParts[2]
			}

			ext := filepath.Ext(mediaURL.Path)
			if len(ext) > 0 {
				// cut the dot
				ext = ext[1:]
			}

			items = append(items, service.Item{
				Meta: map[string]string{
					"id":          id,
					"author":      title,
					"description": description,
					"type":        "unknown",
					"ext":         ext,
					"downloadURL": media,
				},
				DefaultName: "%[author]-%[id].%[ext]",
			})
		})
	}()

	uniqueItems := []service.Item{}
itemIterator:
	for _, item := range items {
		for _, uniqueItem := range uniqueItems {
			if item.Meta["id"] == uniqueItem.Meta["id"] {
				continue itemIterator
			}
		}

		uniqueItems = append(uniqueItems, item)
	}

	return uniqueItems, nil
}

func (i FacebookIterator) HasEnded() bool {
	return i.end
}
