package piko

import (
	"github.com/mlvzk/piko/service"
	"github.com/mlvzk/piko/service/facebook"
	"github.com/mlvzk/piko/service/fourchan"
	"github.com/mlvzk/piko/service/imgur"
	"github.com/mlvzk/piko/service/instagram"
	"github.com/mlvzk/piko/service/soundcloud"
	"github.com/mlvzk/piko/service/twitter"
	"github.com/mlvzk/piko/service/youtube"
)

func GetAllServices() []service.Service {
	return []service.Service{
		youtube.Youtube{},
		imgur.Imgur{},
		instagram.Instagram{},
		fourchan.Fourchan{},
		soundcloud.NewSoundcloud("a3e059563d7fd3372b49b37f00a00bcf"),
		twitter.NewTwitter("AAAAAAAAAAAAAAAAAAAAAIK1zgAAAAAA2tUWuhGZ2JceoId5GwYWU5GspY4%3DUq7gzFoCZs1QfwGoVdvSac3IniczZEYXIcDyumCauIXpcAPorE"),
		facebook.Facebook{},
	}
}
