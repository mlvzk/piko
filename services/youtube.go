package shovel

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Youtube struct{}
type YoutubeIterator struct {
	url string
	end bool
}

// youtubeConfig is a partial structure for deserializing youtube's json config
type youtubeConfig struct {
	Args struct {
		PlayerResponseStr string `json:"player_response"`
	} `json:"args"`
}

type playerResponse struct {
	StreamingData struct {
		AdaptiveFormats []adaptiveFormat `json:"adaptiveFormats"`
	} `json:"streamingData"`
	VideoDetails struct {
		Title  string `json:"title"`
		Author string `json:"author"`
	} `json:"videoDetails"`
}

type adaptiveFormat struct {
	URL           string `json:"url"`
	ITag          int    `json:"itag"`
	MimeType      string `json:"mimeType"`
	ContentLength int    `json:"contentLength"`
}

func (s Youtube) IsValidTarget(target string) bool {
	return strings.Contains(target, "youtube.com/") || strings.Contains(target, "youtu.be/")
}

func (s Youtube) FetchItems(target string) ServiceIterator {
	return &YoutubeIterator{
		url: target,
		end: false,
	}
}

func (s Youtube) Download(meta, options map[string]string) (io.ReadCloser, error) {
	ytPlayerResponse := playerResponse{}
	json.Unmarshal([]byte(meta["_playerResponse"]), &ytPlayerResponse)

	selectedFormat, found := adaptiveFormat{}, false
	for _, format := range ytPlayerResponse.StreamingData.AdaptiveFormats {
		if options["itag"] == strconv.Itoa(format.ITag) {
			selectedFormat, found = format, true
			break
		}
	}

	if !found {
		return nil, errors.New("Couldn't find a format with given itag")
	}

	resp, err := http.Get(selectedFormat.URL)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

var ytConfigRegexp = regexp.MustCompile(`ytplayer\.config = (.*?);ytplayer\.load = function()`)

func (i *YoutubeIterator) Next() ([]Item, error) {
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

	// TODO: download all videos from playlists/channels
	// this only downloads the main video from the url
	ytMatches := ytConfigRegexp.FindStringSubmatch(doc.Find("script").Text())
	if len(ytMatches) < 2 {
		return nil, errors.New("Couldn't match youtube's json config")
	}
	ytConfigStr := ytMatches[1]
	ytConfig := youtubeConfig{}
	json.Unmarshal([]byte(ytConfigStr), &ytConfig)

	ytPlayer := playerResponse{}
	json.Unmarshal([]byte(ytConfig.Args.PlayerResponseStr), &ytPlayer)

	fmt.Println(ytConfig.Args.PlayerResponseStr)
	// fmt.Printf("ytPlayer: %+v\n", ytPlayer)

	item := Item{
		Meta: map[string]string{
			"title":           ytPlayer.VideoDetails.Title,
			"author":          ytPlayer.VideoDetails.Author,
			"_playerResponse": ytConfig.Args.PlayerResponseStr,
		},
		DefaultName: ytPlayer.VideoDetails.Title + ".mp4",
		AvailableOptions: map[string]([]string){
			"itag": make([]string, len(ytPlayer.StreamingData.AdaptiveFormats)),
		},
		DefaultOptions: map[string]string{
			"itag": strconv.Itoa(ytPlayer.StreamingData.AdaptiveFormats[0].ITag),
		},
	}

	for i, format := range ytPlayer.StreamingData.AdaptiveFormats {
		item.AvailableOptions["itag"][i] = strconv.Itoa(format.ITag)
	}

	return []Item{item}, nil
}

func (i YoutubeIterator) HasEnded() bool {
	return i.end
}
