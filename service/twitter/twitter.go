package twitter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mlvzk/piko/service"
)

type videoTweet struct {
	Track struct {
		ContentType       string `json:"contentType"`
		PublisherID       string `json:"publisherId"`
		ContentID         string `json:"contentId"`
		DurationMs        int    `json:"durationMs"`
		PlaybackURL       string `json:"playbackUrl"`
		PlaybackType      string `json:"playbackType"`
		ExpandedURL       string `json:"expandedUrl"`
		ShouldLoop        bool   `json:"shouldLoop"`
		ViewCount         string `json:"viewCount"`
		IsEventGeoblocked bool   `json:"isEventGeoblocked"`
		Is360             bool   `json:"is360"`
	} `json:"track"`
}

type output struct {
	io.ReadCloser
	length uint64
}

func (o output) Size() uint64 {
	return o.length
}

type Twitter struct {
	key string
}
type TwitterIterator struct {
	url string
	end bool
}

func NewTwitter(apiKey string) Twitter {
	return Twitter{
		key: apiKey,
	}
}

func (s Twitter) IsValidTarget(target string) bool {
	return strings.Contains(target, "twitter.com/")
}

func (s Twitter) FetchItems(target string) service.ServiceIterator {
	return &TwitterIterator{
		url: target,
	}
}

func (s Twitter) Download(meta, options map[string]string) (io.Reader, error) {
	downloadURL, hasDownloadURL := meta["downloadURL"]

	if !hasDownloadURL {
		return nil, errors.New("Missing downloadURL")
	}

	if meta["type"] == "image" {
		resp, err := http.Get(meta["downloadURL"])
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("GET %v returned a wrong status code - %v", downloadURL, resp.StatusCode)
		}

		return resp.Body, nil
	} else if meta["type"] == "video" {
		configURL := fmt.Sprintf("https://api.twitter.com/1.1/videos/tweet/config/%s.json", meta["id"])
		configReq, err := http.NewRequest("GET", configURL, nil)
		if err != nil {
			return nil, err
		}
		configReq.Header.Add("Authorization", "Bearer "+s.key)

		playbackURLStr := ""
		// retry 4 times, api calls sometimes fail
		for i := 0; i < 4; i++ {
			configRes, err := http.DefaultClient.Do(configReq)
			if err != nil {
				return nil, err
			}
			defer configRes.Body.Close()

			configBytes, err := ioutil.ReadAll(configRes.Body)
			if err != nil {
				return nil, err
			}

			config := videoTweet{}
			json.Unmarshal(configBytes, &config)

			if config.Track.PlaybackURL != "" {
				playbackURLStr = config.Track.PlaybackURL
				break
			}

			time.Sleep(time.Millisecond * 500)
		}

		if playbackURLStr == "" {
			return nil, errors.New("Couldn't get playbackURL")
		}

		playbackRes, err := http.Get(playbackURLStr)
		if err != nil {
			return nil, err
		}

		if strings.Contains(playbackURLStr, ".m3u8") {
			defer playbackRes.Body.Close()
			contentBytes, err := ioutil.ReadAll(playbackRes.Body)
			if err != nil {
				return nil, err
			}
			playbackURL, err := url.Parse(playbackURLStr)
			if err != nil {
				return nil, err
			}
			playbackBase := playbackURL.Scheme + "://" + playbackURL.Host
			bestContent, err := getBestM3u8(playbackBase, string(contentBytes))
			if err != nil {
				return nil, err
			}

			pipeReader, pipeWriter := io.Pipe()
			go m3u8ToMpeg(bestContent, pipeWriter)

			meta["ext"] = "mp4"
			return pipeReader, nil
		}

		if playbackRes.ContentLength == -1 {
			return playbackRes.Body, nil
		}

		return output{
			ReadCloser: playbackRes.Body,
			length:     uint64(playbackRes.ContentLength),
		}, nil
	}

	return nil, errors.New("Unsupported type")
}

func (i *TwitterIterator) Next() ([]service.Item, error) {
	i.end = true

	resp, err := http.Get(i.url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GET %v returned a wrong status code - %v", i.url, resp.StatusCode)
	}

	urlParsed, err := url.Parse(i.url)
	if err != nil {
		return nil, err
	}
	author := ""
	pathParts := strings.Split(urlParsed.Path, "/")
	if len(pathParts) >= 1 {
		author = pathParts[1]
	}
	id := pathParts[len(pathParts)-1]

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	description, hasDescription := doc.Find(`meta[property="og:description"]`).Attr("content")
	if hasDescription {
		runes := []rune(description)
		description = string(runes[1 : len(runes)-1])
	}

	items := []service.Item{}

	doc.Find(`meta[property="og:video:url"]`).Each(func(index int, videoSel *goquery.Selection) {
		videoURL, hasVideo := videoSel.Attr("content")
		if !hasVideo {
			return
		}

		items = append(items, service.Item{
			Meta: map[string]string{
				"index":       strconv.Itoa(index),
				"author":      author,
				"id":          id,
				"description": description,
				"ext":         "mp4",
				"type":        "video",
				"downloadURL": videoURL,
			},
			DefaultName: "%[author]-%[id]-%[index].%[ext]",
		})
	})

	doc.Find(`meta[property="og:image"]`).Each(func(index int, imageSel *goquery.Selection) {
		imageURL, hasImage := imageSel.Attr("content")
		if !hasImage || len(imageURL) == 0 {
			return
		}

		items = append(items, service.Item{
			Meta: map[string]string{
				"index":       strconv.Itoa(index),
				"author":      author,
				"id":          id,
				"description": description,
				"ext":         "jpg",
				"type":        "image",
				"downloadURL": imageURL,
			},
			DefaultName: "%[author]-%[id]-%[index].%[ext]",
		})
	})

	return items, nil
}

func (i TwitterIterator) HasEnded() bool {
	return i.end
}
