package soundcloud

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

	"github.com/mlvzk/piko/service"
)

type trackData struct {
	CommentCount        int    `json:"comment_count"`
	Downloadable        bool   `json:"downloadable"`
	CreatedAt           string `json:"created_at"`
	Description         string `json:"description"`
	OriginalContentSize int    `json:"original_content_size"`
	Title               string `json:"title"`
	Duration            int    `json:"duration"`
	OriginalFormat      string `json:"original_format"`
	ArtworkURL          string `json:"artwork_url"`
	Streamable          bool   `json:"streamable"`
	TagList             string `json:"tag_list"`
	Genre               string `json:"genre"`
	DownloadURL         string `json:"download_url"`
	ID                  int    `json:"id"`
	State               string `json:"state"`
	RepostsCount        int    `json:"reposts_count"`
	LastModified        string `json:"last_modified"`
	Commentable         bool   `json:"commentable"`
	Policy              string `json:"policy"`
	FavoritingsCount    int    `json:"favoritings_count"`
	Kind                string `json:"kind"`
	Sharing             string `json:"sharing"`
	URI                 string `json:"uri"`
	AttachmentsURI      string `json:"attachments_uri"`
	DownloadCount       int    `json:"download_count"`
	License             string `json:"license"`
	UserID              int    `json:"user_id"`
	EmbeddableBy        string `json:"embeddable_by"`
	MonetizationModel   string `json:"monetization_model"`
	WaveformURL         string `json:"waveform_url"`
	Permalink           string `json:"permalink"`
	PermalinkURL        string `json:"permalink_url"`
	User                struct {
		ID           int    `json:"id"`
		Kind         string `json:"kind"`
		Permalink    string `json:"permalink"`
		Username     string `json:"username"`
		LastModified string `json:"last_modified"`
		URI          string `json:"uri"`
		PermalinkURL string `json:"permalink_url"`
		AvatarURL    string `json:"avatar_url"`
	} `json:"user"`
	StreamURL     string `json:"stream_url"`
	PlaybackCount int    `json:"playback_count"`
}

type Soundcloud struct {
	clientID string
}
type SoundcloudIterator struct {
	clientID   string
	baseApiURL string
	url        string
	end        bool
}

func NewSoundcloud(clientID string) Soundcloud {
	return Soundcloud{
		clientID: clientID,
	}
}

func (s Soundcloud) IsValidTarget(target string) bool {
	return strings.Contains(target, "soundcloud.com/")
}

func (s Soundcloud) FetchItems(target string) service.ServiceIterator {
	return &SoundcloudIterator{
		url:        target,
		baseApiURL: "https://api.soundcloud.com",
		clientID:   s.clientID,
	}
}

func (s Soundcloud) Download(meta, options map[string]string) (io.Reader, error) {
	downloadURL, downloadExists := meta["_downloadURL"]
	if !downloadExists {
		return nil, errors.New("Download URL doesn't exist")
	}

	resp, err := http.Get(downloadURL + "?client_id=" + s.clientID)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GET %v returned a wrong status code - %v", downloadURL, resp.StatusCode)
	}

	contentDisp := resp.Header.Get("Content-Disposition")
	if strings.Contains(contentDisp, "filename=") {
		dotParts := strings.Split(contentDisp, ".")
		lastPart := dotParts[len(dotParts)-1]
		meta["ext"] = lastPart[:len(lastPart)-1]
	} else {
		dotParts := strings.Split(resp.Request.URL.Path, ".")
		meta["ext"] = dotParts[len(dotParts)-1]
	}

	return resp.Body, nil
}

func (i *SoundcloudIterator) Next() ([]service.Item, error) {
	i.end = true

	u, err := makeResolveUrl(i.baseApiURL, i.url, i.clientID)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GET %v returned a wrong status code - %v", i.url, resp.StatusCode)
	}

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	trackResp := trackData{}
	json.Unmarshal(respData, &trackResp)

	downloadURL := ""
	if trackResp.Downloadable {
		downloadURL = trackResp.DownloadURL
	} else if trackResp.Streamable {
		downloadURL = trackResp.StreamURL
	} else {
		return nil, errors.New("Track is neither downloadable or streamable")
	}

	return []service.Item{
		service.Item{
			Meta: map[string]string{
				"id":           strconv.Itoa(trackResp.ID),
				"title":        trackResp.Title,
				"username":     trackResp.User.Username,
				"playCount":    strconv.Itoa(trackResp.PlaybackCount),
				"duration":     strconv.Itoa(trackResp.Duration),
				"createdAt":    trackResp.CreatedAt,
				"ext":          "mp3",
				"_downloadURL": downloadURL,
			},
			DefaultName: "%[title].%[ext]",
		},
	}, nil
}

func (i SoundcloudIterator) HasEnded() bool {
	return i.end
}

func makeResolveUrl(baseApiURL, targetURL, client_id string) (string, error) {
	u, err := url.Parse(baseApiURL)
	if err != nil {
		return "", err
	}

	u.Path = "resolve"

	v := url.Values{}
	v.Set("client_id", client_id)
	v.Set("url", targetURL)
	u.RawQuery = v.Encode()

	return u.String(), nil
}
