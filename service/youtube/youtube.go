package youtube

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

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
		PlayerResponseStr string `json:"player_response"`
	} `json:"args"`
	Assets struct {
		JS string `json:"js"`
	} `json:"assets"`
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
	Meta          map[string]string
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
	videoInfo := ytdl.VideoInfo{}
	json.Unmarshal([]byte(meta["_videoInfo"]), &videoInfo)

	if _, err := exec.LookPath("ffmpeg"); err != nil || options["useFfmpeg"] == "no" {
		log.Println("here")
		// no ffmpeg, fallbacking to format with both audio and video
		videoFormat := videoInfo.Formats[0]
		downloadURL, err := videoInfo.GetDownloadURL(videoFormat)
		if err != nil {
			return nil, err
		}

		videoResp, err := http.Get(downloadURL.String())
		if err != nil {
			return nil, err
		}
		videoStream := videoResp.Body

		// this is bad, in order for this to work file name needs to be formatted after Download is called
		meta["ext"] = videoFormat.Extension

		return output{
			ReadCloser: videoStream,
			length:     uint64(videoResp.ContentLength),
		}, nil
	}

	videoInfo.Formats.Sort(ytdl.FormatAudioBitrateKey, true)
	audioFormat := videoInfo.Formats[0]
	videoInfo.Formats.Sort(ytdl.FormatResolutionKey, true)
	videoFormat := videoInfo.Formats[0]

	audioURL, err := videoInfo.GetDownloadURL(audioFormat)
	if err != nil {
		return nil, err
	}

	tmpAudioFile, err := ioutil.TempFile("", "audio*."+audioFormat.Extension)
	if err != nil {
		return nil, err
	}

	// download by 10MB chunks to avoid throttling
	audioStream, audioStreamWriter := io.Pipe()
	go io.Copy(tmpAudioFile, audioStream)
	err = service.DownloadByChunks(audioURL.String(), 0xFFFFF, audioStreamWriter)
	if err != nil {
		return nil, err
	}
	audioStreamWriter.Close()
	audioStream.Close()

	stat, err := tmpAudioFile.Stat()
	if err != nil {
		return nil, err
	}
	audioLength := stat.Size()
	tmpAudioFile.Close()

	videoURL, err := videoInfo.GetDownloadURL(videoFormat)
	if err != nil {
		return nil, err
	}

	videoStream, videoStreamWriter := io.Pipe()
	go service.DownloadByChunks(videoURL.String(), 0xFFFFF, videoStreamWriter)
	videoLength, err := strconv.ParseInt(videoFormat.ValueForKey("clen").(string), 10, 64)
	if err != nil {
		videoLength = -1
	}

	cmd := exec.Command("ffmpeg", "-i", tmpAudioFile.Name(), "-i", "-", "-c", "copy", "-f", "matroska", "-")
	cmd.Stdin = videoStream
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	go cmd.Run()

	if audioLength == -1 || videoLength == -1 {
		return stdout, nil
	}

	return output{
		ReadCloser: stdout,
		// length is not exact, but close enough
		length: uint64(audioLength + videoLength),
	}, nil
}

var ytConfigRegexp = regexp.MustCompile(`ytplayer\.config = (.*?);ytplayer\.load = function()`)

func (i *YoutubeIterator) Next() ([]service.Item, error) {
	i.end = true

	videoInfo, err := ytdl.GetVideoInfo(i.url)
	if err != nil {
		return nil, err
	}

	videoInfoJson, err := json.Marshal(videoInfo)
	if err != nil {
		return nil, err
	}

	item := service.Item{
		Meta: map[string]string{
			"title":      videoInfo.Title,
			"author":     videoInfo.Author,
			"ext":        "mkv",
			"_videoInfo": string(videoInfoJson),
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
