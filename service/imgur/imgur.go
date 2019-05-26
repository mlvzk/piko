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

package imgur

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mlvzk/piko/service"
)

type Imgur struct{}
type ImgurIterator struct {
	url  string
	page int
	end  bool
}

func New() Imgur {
	return Imgur{}
}

type output struct {
	io.ReadCloser
	length uint64
}

func (o output) Size() uint64 {
	return o.length
}

func (s Imgur) IsValidTarget(target string) bool {
	return strings.Contains(target, "imgur.com/")
}

func (s Imgur) FetchItems(target string) (service.ServiceIterator, error) {
	return &ImgurIterator{
		url:  target,
		page: 1,
		end:  false,
	}, nil
}

func (s Imgur) Download(meta, options map[string]string) (io.Reader, error) {
	resp, err := http.Get(fmt.Sprintf("https://i.imgur.com/%s.%s", meta["id"], meta["ext"]))
	if err != nil {
		return nil, err
	}

	if resp.ContentLength == -1 {
		return resp.Body, nil
	}

	return output{
		ReadCloser: resp.Body,
		length:     uint64(resp.ContentLength),
	}, nil
}

func (i *ImgurIterator) Next() ([]service.Item, error) {
	i.end = true

	resp, err := http.Get(i.url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GET %v returned a wrong status code - %v", i.url, resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	albumTitle := doc.Find("div.post-title-container h1").Text()

	items := []service.Item{}
	doc.Find("div.post-images div.post-image-container").Each(func(_ int, sel *goquery.Selection) {
		itemType, itExists := sel.Attr("itemtype")
		id, _ := sel.Attr("id")

		ext := "png"
		if itExists && strings.Contains(itemType, "VideoObject") {
			ext = "mp4"
		}

		// TODO: take src from meta[@contentURL] if available
		items = append(items, service.Item{
			Meta: map[string]string{
				"id":         id,
				"ext":        ext,
				"itemType":   itemType,
				"albumTitle": albumTitle,
			},
			DefaultName: "%[id].%[ext]",
		})
	})

	return items, nil
}

func (i ImgurIterator) HasEnded() bool {
	return i.end
}
