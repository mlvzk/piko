package shovel

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"regexp"
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
		AdaptiveFormats []format `json:"adaptiveFormats"`
		Formats         []format `json:"formats"`
	} `json:"streamingData"`
	VideoDetails struct {
		Title  string `json:"title"`
		Author string `json:"author"`
	} `json:"videoDetails"`
}

type format struct {
	URL           string `json:"url"`
	ITag          int    `json:"itag"`
	MimeType      string `json:"mimeType"`
	ContentLength string `json:"contentLength"`
	Bitrate       int    `json:"bitrate"`
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

func (s Youtube) Download(meta, options map[string]string) (io.Reader, error) {
	ytPlayerResponse := playerResponse{}
	json.Unmarshal([]byte(meta["_playerResponse"]), &ytPlayerResponse)

	if _, err := exec.LookPath("ffmpeg"); err != nil || options["useFfmpeg"] == "no" {
		// no ffmpeg, fallbacking to format with both audio and video
		video := findBestVideo(ytPlayerResponse.StreamingData.Formats)
		videoResp, err := http.Get(video.URL)
		if err != nil {
			return nil, err
		}
		videoStream := videoResp.Body

		mimeParts := strings.Split(video.MimeType, ";")
		ext := strings.Split(mimeParts[0], "/")[1]
		// this is bad, in order for this to work file name needs to be formatted after Download is called
		meta["ext"] = ext

		return videoStream, nil
	}

	audio := findBestAudio(ytPlayerResponse.StreamingData.AdaptiveFormats)
	video := findBestVideo(ytPlayerResponse.StreamingData.AdaptiveFormats)

	audioResp, err := http.Get(audio.URL)
	if err != nil {
		return nil, err
	}
	audioStream := audioResp.Body

	videoResp, err := http.Get(video.URL)
	if err != nil {
		return nil, err
	}
	videoStream := videoResp.Body

	// MimeType should be: audio/webm; codecs="opus"
	mimeParts := strings.Split(audio.MimeType, ";")
	ext := strings.Split(mimeParts[0], "/")[1]

	tmpAudioFile, err := ioutil.TempFile("", "audio*."+ext)
	io.Copy(tmpAudioFile, audioStream)
	tmpAudioFile.Close()
	audioStream.Close()

	cmd := exec.Command("ffmpeg", "-i", tmpAudioFile.Name(), "-i", "-", "-c", "copy", "-f", "matroska", "-")
	cmd.Stdin = videoStream
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	go cmd.Run()

	return stdout, nil
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

	item := Item{
		Meta: map[string]string{
			"title":           ytPlayer.VideoDetails.Title,
			"author":          ytPlayer.VideoDetails.Author,
			"ext":             "mkv",
			"_playerResponse": ytConfig.Args.PlayerResponseStr,
		},
		DefaultName: "%[title].%[ext]",
		AvailableOptions: map[string]([]string){
			"quality":   []string{"best", "medium", "worst"},
			"useFfmpeg": []string{"yes", "no"},
		},
		DefaultOptions: map[string]string{
			"quality":   "best",
			"useFfmpeg": "yes",
		},
	}

	return []Item{item}, nil
}

func (i YoutubeIterator) HasEnded() bool {
	return i.end
}

func findBest(formats []format, t string) format {
	var best format

	for _, f := range formats {
		if !strings.Contains(f.MimeType, t) {
			continue
		}

		if f.Bitrate > best.Bitrate {
			best = f
		}
	}

	return best
}

func findBestAudio(formats []format) format {
	return findBest(formats, "audio")
}

func findBestVideo(formats []format) format {
	return findBest(formats, "video")
}
