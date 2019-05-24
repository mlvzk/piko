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

package youtube

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mlvzk/piko/service"
	"github.com/mlvzk/piko/service/youtube/ytdl"
)

type Youtube struct{}
type YoutubeIterator struct {
	url string
	end bool
}

func New() Youtube {
	return Youtube{}
}

// youtubeConfig is a partial structure for deserializing youtube's json config
type youtubeConfig struct {
	Args struct {
		PlayerResponseStr      string `json:"player_response"`
		URLEncodedFmtStreamMap string `json:"url_encoded_fmt_stream_map"`
		AdaptiveFmts           string `json:"adaptive_fmts"`
	} `json:"args"`
	Assets struct {
		JS string `json:"js"`
	} `json:"assets"`
}

type playerResponse struct {
	VideoDetails struct {
		Title  string `json:"title"`
		Author string `json:"author"`
	} `json:"videoDetails"`
}

type output struct {
	io.ReadCloser
	length uint64
}

func (o output) Size() uint64 {
	return o.length
}

func (s Youtube) IsValidTarget(target string) bool {
	return strings.Contains(target, "youtube.com/") || strings.Contains(target, "youtu.be/")
}

func (s Youtube) FetchItems(target string) service.ServiceIterator {
	return &YoutubeIterator{
		url: target,
		end: false,
	}
}

func (s Youtube) Download(meta, options map[string]string) (io.Reader, error) {
	quality := options["quality"]
	useFfmpeg := options["useFfmpeg"] == "yes"
	onlyAudio := options["onlyAudio"] == "yes"

	ytConfig := youtubeConfig{}
	json.Unmarshal([]byte(meta["_ytConfig"]), &ytConfig)

	formats := getFormats(ytConfig.Args.AdaptiveFmts, ytConfig.Args.URLEncodedFmtStreamMap)

	if _, err := exec.LookPath("ffmpeg"); (err != nil || !useFfmpeg) && !onlyAudio {
		// no ffmpeg, fallbacking to format with both audio and video
		video := findBestVideoAudio(formats)
		videoURL, err := ytdl.GetDownloadURL(video.Meta, ytConfig.Assets.JS)
		if err != nil {
			return nil, err
		}

		videoStream, videoStreamWriter := io.Pipe()
		go service.DownloadByChunks(videoURL.String(), 0xFFFFF, videoStreamWriter)

		// this is bad, in order for this to work file name needs to be formatted after Download is called
		meta["ext"] = video.Extension

		videoLengthMeta, hasLength := video.Meta["clen"]
		if hasLength {
			videoLength, err := strconv.ParseInt(videoLengthMeta.(string), 10, 64)
			if err != nil {
				return nil, err
			}

			return &output{
				ReadCloser: videoStream,
				length:     uint64(videoLength),
			}, nil
		}

		return videoStream, nil
	}

	audioFormat, audioFormatFound := findBestAudio(formats)
	var (
		videoFormat      ytdl.Format
		videoFormatFound bool
	)
	if quality == "best" {
		videoFormat, videoFormatFound = findBestVideo(formats)
	} else if quality == "medium" {
		videoFormat, videoFormatFound = findMediumVideo(formats)
		if !videoFormatFound {
			// fallback to best if medium not found
			videoFormat, videoFormatFound = findBestVideo(formats)
		}
	} else {
		videoFormat, videoFormatFound = findWorstVideo(formats)
	}

	if !audioFormatFound || !videoFormatFound {
		return nil, errors.New("Couldn't find either audio or video format")
	}

	// not actual length, but should be close
	audioLengthMeta, hasAudioLength := audioFormat.Meta["clen"]
	audioLength, _ := strconv.ParseInt(audioLengthMeta.(string), 10, 64)

	audioURL, err := ytdl.GetDownloadURL(audioFormat.Meta, ytConfig.Assets.JS)
	if err != nil {
		return nil, err
	}

	if onlyAudio {
		meta["ext"] = audioFormat.Extension
		audioStream, audioStreamWriter := io.Pipe()
		go service.DownloadByChunks(audioURL.String(), 0xFFFFF, audioStreamWriter)

		return output{
			ReadCloser: audioStream,
			length:     uint64(audioLength),
		}, nil
	}

	videoLengthMeta, hasVideoLength := videoFormat.Meta["clen"]
	videoLength, _ := strconv.ParseInt(videoLengthMeta.(string), 10, 64)

	tmpAudioFile, err := ioutil.TempFile("", "audio*."+audioFormat.Extension)
	if err != nil {
		return nil, err
	}

	audioStream, audioStreamWriter := io.Pipe()
	go io.Copy(tmpAudioFile, audioStream)
	err = service.DownloadByChunks(audioURL.String(), 0xFFFFF, audioStreamWriter)
	if err != nil {
		return nil, err
	}
	io.Copy(tmpAudioFile, audioStream)
	tmpAudioFile.Close()
	audioStream.Close()

	videoURL, err := ytdl.GetDownloadURL(videoFormat.Meta, ytConfig.Assets.JS)
	if err != nil {
		return nil, err
	}
	videoStream, videoStreamWriter := io.Pipe()
	// download by 10MB chunks to avoid throttling
	go service.DownloadByChunks(videoURL.String(), 0xFFFFF, videoStreamWriter)

	cmd := exec.Command("ffmpeg", "-i", tmpAudioFile.Name(), "-i", "-", "-c", "copy", "-f", "matroska", "-")
	cmd.Stdin = videoStream
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	go cmd.Run()

	if (hasAudioLength && hasVideoLength) || hasVideoLength {
		return output{
			ReadCloser: stdout,
			// length is not exact, but close enough
			length: uint64(audioLength + videoLength),
		}, nil
	}

	return stdout, nil
}

var ytConfigRegexp = regexp.MustCompile(`ytplayer\.config = (.*?);ytplayer\.load = function()`)

func (i *YoutubeIterator) Next() ([]service.Item, error) {
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

	item := service.Item{
		Meta: map[string]string{
			"title":     ytPlayer.VideoDetails.Title,
			"author":    ytPlayer.VideoDetails.Author,
			"ext":       "mkv",
			"_ytConfig": ytMatches[1],
		},
		DefaultName: "%[title].%[ext]",
		AvailableOptions: map[string]([]string){
			"quality":   []string{"best", "medium", "worst"},
			"useFfmpeg": []string{"yes", "no"},
			"onlyAudio": []string{"yes", "no"},
		},
		DefaultOptions: map[string]string{
			"quality":   "medium",
			"useFfmpeg": "yes",
			"onlyAudio": "no",
		},
	}

	return []service.Item{item}, nil
}

func (i YoutubeIterator) HasEnded() bool {
	return i.end
}

func findBestVideoAudio(formats []ytdl.Format) ytdl.Format {
	var best ytdl.Format

	for _, f := range formats {
		if f.AudioEncoding == "" || f.VideoEncoding == "" {
			continue
		}

		if f.AudioBitrate > best.AudioBitrate {
			best = f
		}
	}

	return best
}

func findBestAudio(formats []ytdl.Format) (audio ytdl.Format, found bool) {
	for _, f := range formats {
		if f.AudioEncoding == "" || f.VideoEncoding != "" {
			continue
		}

		if f.AudioBitrate > audio.AudioBitrate {
			audio = f
			found = true
		}
	}

	return
}

func _findVideo(formats []ytdl.Format, best bool) (video ytdl.Format, found bool) {
	var currBitrate int

	for _, f := range formats {
		if f.VideoEncoding == "" || f.AudioEncoding != "" {
			continue
		}

		bitrateI, exists := f.Meta["bitrate"]
		if !exists {
			continue
		}
		bitrate, err := strconv.Atoi(bitrateI.(string))
		if err != nil {
			continue
		}

		if (best && bitrate > currBitrate) ||
			(!best && (bitrate < currBitrate || !found)) {
			video = f
			currBitrate = bitrate
			found = true
		}
	}

	return
}

func findBestVideo(formats []ytdl.Format) (video ytdl.Format, found bool) {
	return _findVideo(formats, true)
}

func findWorstVideo(formats []ytdl.Format) (video ytdl.Format, found bool) {
	return _findVideo(formats, false)
}

// medium is 1080p, 720p or 480p
// prefers 1080p > 720p > 480p
func findMediumVideo(formats []ytdl.Format) (video ytdl.Format, found bool) {
	var currBitrate int
	allowedRes := map[string]bool{"1080p": true, "720p": true, "480p": true}

	for _, f := range formats {
		if f.VideoEncoding == "" || f.AudioEncoding != "" {
			continue
		}

		bitrateI, exists := f.Meta["bitrate"]
		if !exists {
			continue
		}
		bitrate, err := strconv.Atoi(bitrateI.(string))
		if err != nil {
			continue
		}

		allowed, inRange := allowedRes[f.Resolution]
		if inRange && allowed && bitrate > currBitrate {
			video = f
			currBitrate = bitrate
			found = true
		}
	}

	return
}

func getFormats(strs ...string) []ytdl.Format {
	var formats []ytdl.Format

	for _, str := range strs {
		formatStrs := strings.Split(str, ",")

		for _, formatStr := range formatStrs {
			query, err := url.ParseQuery(formatStr)
			if err != nil {
				continue
			}

			itag, err := strconv.Atoi(query.Get("itag"))
			if err != nil || itag <= 0 {
				continue
			}

			format, _ := ytdl.NewFormat(itag)
			format.Meta = make(map[string]interface{})
			if strings.HasPrefix(query.Get("conn"), "rtmp") {
				format.Meta["rtmp"] = true
			}
			for k, v := range query {
				if len(v) == 1 {
					format.Meta[k] = v[0]
				} else {
					format.Meta[k] = v
				}
			}
			formats = append(formats, format)
		}
	}

	return formats
}
