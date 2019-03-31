package shovel

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type ImgurService struct{}

func (s ImgurService) IsValidTarget(target string) bool {
	return strings.Contains(target, "imgur.com/")
}

func (s ImgurService) FetchItems(target string) ([]Item, error) {
	resp, err := http.Get(target)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GET %v returned a wrong status code - %v", target, resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	albumTitle := doc.Find("div.post-title-container h1").Text()
	println("albumTitle: ", albumTitle)

	items := []Item{}
	doc.Find("div.post-images div.post-image-container").Each(func(_ int, s *goquery.Selection) {
		itemType, itExists := s.Attr("itemtype")
		id, _ := s.Attr("id")

		ext := "png"
		if itExists && strings.Contains(itemType, "VideoObject") {
			ext = "mp4"
		}

		// TODO: take src from meta[@contentURL] if available
		items = append(items, Item{
			Meta: map[string]string{
				"id":         id,
				"ext":        ext,
				"itemType":   itemType,
				"albumTitle": albumTitle,
			},
			DefaultName: id + "." + ext,
		})
	})

	return items, nil
}

func (s ImgurService) Download(meta, options map[string]string) (io.Reader, error) {
	resp, err := http.Get(fmt.Sprintf("https://i.imgur.com/%s.%s", meta["id"], meta["ext"]))
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
