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

package twitter

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func getBestM3u8(baseURL, content string) (string, error) {
	lines := strings.Split(content, "\n")
	best, found := "", false

	for i := len(lines) - 1; i >= 0; i-- {
		if strings.Contains(lines[i], ".m3u8") {
			best, found = lines[i], true
			break
		}
	}

	if !found {
		return "", errors.New("Couldn't find the best m3u8")
	}

	res, err := http.Get(baseURL + best)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	contentLines := strings.Split(string(bytes), "\n")
	for i := range contentLines {
		if len(contentLines[i]) > 0 && contentLines[i][0] == '/' {
			contentLines[i] = baseURL + contentLines[i]
		}
	}

	return strings.Join(contentLines, "\n"), nil
}

func m3u8ToMpeg(content string, writer io.WriteCloser) error {
	defer writer.Close()
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		if len(line) < 4 || line[0:4] != "http" {
			continue
		}

		res, err := http.Get(line)
		if err != nil {
			return err
		}
		io.Copy(writer, res.Body)
		res.Body.Close()
	}

	return nil
}
