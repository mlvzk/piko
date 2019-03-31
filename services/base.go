package shovel

import "io"

type Item struct {
	Meta             map[string]string
	DefaultName      string
	AvailableOptions map[string](map[string]string)
	DefaultOptions   map[string]string
}

type Service interface {
	IsValidTarget(target string) bool
	FetchItems(target string) ([]Item, error)
	Download(meta, options map[string]string) (io.ReadCloser, error)
}
