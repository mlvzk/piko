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
