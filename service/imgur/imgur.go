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

func (s Imgur) IsValidTarget(target string) bool {
	return strings.Contains(target, "imgur.com/")
}

func (s Imgur) FetchItems(target string) service.ServiceIterator {
	return &ImgurIterator{
		url:  target,
		page: 1,
		end:  false,
	}
}

func (s Imgur) Download(meta, options map[string]string) (io.Reader, error) {
	resp, err := http.Get(fmt.Sprintf("https://i.imgur.com/%s.%s", meta["id"], meta["ext"]))
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
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
