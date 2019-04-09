package instagram

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mlvzk/shovel-go/service"
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

func (s Instagram) IsValidTarget(target string) bool {
	return strings.Contains(target, "instagram.com/")
}

func (s Instagram) FetchItems(target string) service.ServiceIterator {
	return &InstagramIterator{
		url: target,
	}
}

func (s Instagram) Download(meta, options map[string]string) (io.Reader, error) {
	resp, err := http.Get(meta["imgURL"])
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
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

		imgURL, imgExists := doc.Find(`meta[property="og:image"]`).Attr("content")
		if !imgExists {
			return nil, errors.New("Couldn't find the image url in meta tags")
		}

		schemaStr := doc.Find(`script[type="application/ld+json"]`).Text()
		sch := schema{}
		json.Unmarshal([]byte(schemaStr), &sch)

		return []service.Item{
			service.Item{
				Meta: map[string]string{
					"imgURL":  imgURL,
					"caption": sch.Caption,
					"author":  sch.Author.AlternateName[1:],
					"date":    sch.UploadDate,
					"id":      id,
				},
				DefaultName: "%[title].jpg",
			},
		}, nil
	}

	return []service.Item{}, nil
}

func (i InstagramIterator) HasEnded() bool {
	return i.end
}
