package shovel

import (
	"github.com/mlvzk/shovel-go/service"
	"github.com/mlvzk/shovel-go/service/fourchan"
	"github.com/mlvzk/shovel-go/service/imgur"
	"github.com/mlvzk/shovel-go/service/instagram"
	"github.com/mlvzk/shovel-go/service/soundcloud"
	"github.com/mlvzk/shovel-go/service/youtube"
)

func GetAllServices() []service.Service {
	return []service.Service{
		youtube.Youtube{},
		imgur.Imgur{},
		instagram.Instagram{},
		fourchan.Fourchan{},
		soundcloud.NewSoundcloud("a3e059563d7fd3372b49b37f00a00bcf"),
	}
}
