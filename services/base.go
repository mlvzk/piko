package shovel

import "io"

type Item struct {
	Meta             map[string]string
	DefaultName      string
	AvailableOptions map[string]([]string)
	DefaultOptions   map[string]string
}

type ServiceIterator interface {
	Next() ([]Item, error)
	HasEnded() bool
}

type Service interface {
	IsValidTarget(target string) bool
	FetchItems(target string) ServiceIterator
	Download(meta, options map[string]string) (io.ReadCloser, error)
}
