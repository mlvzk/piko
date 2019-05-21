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

package service

import (
	"fmt"
	"io"
	"net/http"
)

func DownloadByChunks(url string, chunkSize uint64, writer io.WriteCloser) error {
	var pos uint64

	for {
		start := pos
		end := pos + chunkSize - 1
		pos += chunkSize

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		n, err := io.Copy(writer, res.Body)
		if err != nil {
			return err
		}
		res.Body.Close()

		if uint64(n) < chunkSize {
			break
		}
	}

	writer.Close()

	return nil
}
