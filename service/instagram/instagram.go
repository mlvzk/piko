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

package instagram

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mlvzk/piko/service"
)

type schema struct {
	Context              string `json:"@context"`
	Type                 string `json:"@type"`
	Caption              string `json:"caption"`
	RepresentativeOfPage string `json:"representativeOfPage"`
	UploadDate           string `json:"uploadDate"`
	Author               struct {
		Type             string `json:"@type"`
		AlternateName    string `json:"alternateName"`
		MainEntityofPage struct {
			Type string `json:"@type"`
			ID   string `json:"@id"`
		} `json:"mainEntityofPage"`
	} `json:"author"`
	Comment []struct {
		Type   string `json:"@type"`
		Text   string `json:"text"`
		Author struct {
			Type             string `json:"@type"`
			AlternateName    string `json:"alternateName"`
			MainEntityofPage struct {
				Type string `json:"@type"`
				ID   string `json:"@id"`
			} `json:"mainEntityofPage"`
		} `json:"author"`
	} `json:"comment"`
	CommentCount         string `json:"commentCount"`
	InteractionStatistic struct {
		Type            string `json:"@type"`
		InteractionType struct {
			Type string `json:"@type"`
		} `json:"interactionType"`
		UserInteractionCount string `json:"userInteractionCount"`
	} `json:"interactionStatistic"`
	MainEntityofPage struct {
		Type string `json:"@type"`
		ID   string `json:"@id"`
	} `json:"mainEntityofPage"`
	Description string `json:"description"`
	Name        string `json:"name"`
}

type Instagram struct{}
type InstagramIterator struct {
	url string
	end bool
}

func New() Instagram {
	return Instagram{}
}

type output struct {
	io.ReadCloser
	length uint64
}

func (o output) Size() uint64 {
	return o.length
}

func (s Instagram) IsValidTarget(target string) bool {
	return strings.Contains(target, "instagram.com/")
}

func (s Instagram) FetchItems(target string) (service.ServiceIterator, error) {
	return &InstagramIterator{
		url: target,
	}, nil
}

func (s Instagram) Download(meta, options map[string]string) (io.Reader, error) {
	resp, err := http.Get(meta["imgURL"])
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

func (i *InstagramIterator) Next() ([]service.Item, error) {
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

	if strings.Contains(i.url, "/p/") {
		// single image
		urlParts := strings.Split(i.url, "/p/")
		id := urlParts[1][:len(urlParts[1])-1]

		title := doc.Find(`title`).Text()
		imgURL, imgExists := doc.Find(`meta[property="og:image"]`).Attr("content")
		if !imgExists {
			return nil, errors.New("Couldn't find the image url in meta tags")
		}

		var author string
		canonicalURL, hasCanonical := doc.Find(`link[rel="canonical"]`).Attr("href")
		if hasCanonical {
			canonicalParts := strings.Split(canonicalURL, "/")
			if len(canonicalParts) > 3 {
				author = canonicalParts[3]
			}
		}

		var caption string
		titleQuoteParts := strings.Split(title, "“")
		if len(titleQuoteParts) != 0 {
			caption = strings.Split(titleQuoteParts[len(titleQuoteParts)-1], "”")[0]
		}

		return []service.Item{
			{
				Meta: map[string]string{
					"imgURL":  imgURL,
					"caption": caption,
					"author":  author,
					"id":      id,
					"ext":     "jpg",
				},
				DefaultName: "%[author]_%[id].%[ext]",
			},
		}, nil
	}

	return []service.Item{}, nil
}

func (i InstagramIterator) HasEnded() bool {
	return i.end
}
