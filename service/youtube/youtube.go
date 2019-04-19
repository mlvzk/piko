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
	ytConfig := youtubeConfig{}
	json.Unmarshal([]byte(meta["_ytConfig"]), &ytConfig)

	formats := getFormats(ytConfig.Args.AdaptiveFmts, ytConfig.Args.URLEncodedFmtStreamMap)

	if _, err := exec.LookPath("ffmpeg"); err != nil || options["useFfmpeg"] == "no" {
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

	audioFormat := findBestAudio(formats)
	videoFormat := findBestVideo(formats)

	// not actual length, but should be close
	audioLengthMeta, hasAudioLength := audioFormat.Meta["clen"]
	audioLength, _ := strconv.ParseInt(audioLengthMeta.(string), 10, 64)

	videoLengthMeta, hasVideoLength := videoFormat.Meta["clen"]
	videoLength, _ := strconv.ParseInt(videoLengthMeta.(string), 10, 64)

	tmpAudioFile, err := ioutil.TempFile("", "audio*."+audioFormat.Extension)
	if err != nil {
		return nil, err
	}

	audioURL, err := ytdl.GetDownloadURL(audioFormat.Meta, ytConfig.Assets.JS)
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
		},
		DefaultOptions: map[string]string{
			"quality":   "best",
			"useFfmpeg": "yes",
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

func findBestAudio(formats []ytdl.Format) ytdl.Format {
	var best ytdl.Format

	for _, f := range formats {
		if f.AudioEncoding == "" || f.VideoEncoding != "" {
			continue
		}

		if f.AudioBitrate > best.AudioBitrate {
			best = f
		}
	}

	return best
}

func findBestVideo(formats []ytdl.Format) ytdl.Format {
	var best ytdl.Format
	var bestBitrate int

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

		if bitrate > bestBitrate {
			best = f
			bestBitrate = bitrate
		}
	}

	return best
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
