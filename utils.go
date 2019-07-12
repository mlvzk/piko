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

package piko

import (
	"github.com/mlvzk/piko/service"
	"github.com/mlvzk/piko/service/facebook"
	"github.com/mlvzk/piko/service/fourchan"
	"github.com/mlvzk/piko/service/instagram"
	"github.com/mlvzk/piko/service/script"
	"github.com/mlvzk/piko/service/soundcloud"
	"github.com/mlvzk/piko/service/twitter"
	"github.com/mlvzk/piko/service/youtube"
)

func GetAllServices() []service.Service {
	return []service.Service{
		youtube.New(),
		// imgur.New(),
		instagram.New(),
		fourchan.New(),
		soundcloud.New("a3e059563d7fd3372b49b37f00a00bcf"),
		twitter.New("AAAAAAAAAAAAAAAAAAAAAIK1zgAAAAAA2tUWuhGZ2JceoId5GwYWU5GspY4%3DUq7gzFoCZs1QfwGoVdvSac3IniczZEYXIcDyumCauIXpcAPorE"),
		facebook.New(),
		script.New("imgur", `
def isValidTarget(target):
	return target.find("imgur.com") != -1

def download(item):
	url = "https://i.imgur.com/%(id)s.%(ext)s" % item["Meta"]
	return url

def fetchItems(url):
	doc = fetch(url)

	albumTitle = doc.find("div.post-title-container h1").text
	
	images = doc.find("div.post-images div.post-image-container")
	items = []
	for image in images:
		itemType = image.attr("itemtype")
		items.append({
			"Meta": {
				"id": image.attr("id"),
				"ext": "mp4" if itemType.endswith("VideoObject") else "png",
				"itemType": itemType,
				"albumTitle": albumTitle,
			},
			"DefaultName": "%[id].%[ext]",
		})

	return items

		`),
	}
}
