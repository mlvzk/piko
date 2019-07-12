package script

import (
	"fmt"
	"io"
	"log"

	"github.com/mlvzk/piko/service"
	"go.starlark.net/starlark"
)

type Script struct {
	thread        *starlark.Thread
	fetchItems    func(target string) ([]service.Item, error)
	isValidTarget func(target string) bool
}

var _ service.Service = (*Script)(nil)

func New(name string, script string) Script {
	thread := &starlark.Thread{Name: "my thread",
		Print: func(_ *starlark.Thread, msg string) { log.Println(msg) },
	}

	fetchBuiltin := starlark.NewBuiltin("fetch", fetch)
	predeclared := starlark.StringDict{
		"fetch": fetchBuiltin,
	}
	globals, err := starlark.ExecFile(thread, name+".starlark", script, predeclared)
	if err != nil {
		panic(err)
	}

	isValidTarget, ok := globals["isValidTarget"].(*starlark.Function)
	if !ok {
		panic("missing isValidTarget function in " + name)
	}

	fetchItems, ok := globals["fetchItems"].(*starlark.Function)
	if !ok {
		panic("missing fetchItems function in " + name)
	}

	return Script{
		thread: thread,
		fetchItems: func(target string) ([]service.Item, error) {
			v, err := starlark.Call(thread, fetchItems, starlark.Tuple{starlark.String(target)}, nil)
			if err != nil {
				return nil, err
			}

			list := v.(*starlark.List)
			iterator := list.Iterate()
			var item starlark.Value
			iterator.Next(&item)
			for item := item; item != nil; iterator.Next(&item) {
				dict := item.(*starlark.Dict)
				meta, found, err := dict.Get(starlark.String("Meta"))
				if !found || err != nil {
					continue
				}
				defaultName, found, err := dict.Get(starlark.String("DefaultName"))
				if !found || err != nil {
					continue
				}

				fmt.Println("meta: ", meta, "defaultName", defaultName)
			}

			return nil, nil
		},
		isValidTarget: func(target string) bool {
			v, err := starlark.Call(thread, isValidTarget, starlark.Tuple{starlark.String(target)}, nil)
			if err != nil {
				panic(err)
			}

			return bool(v.Truth())
		},
	}
}

func (s Script) IsValidTarget(target string) bool {
	return s.isValidTarget(target)
}

func (s Script) Download(meta, options map[string]string) (io.Reader, error) {
	return nil, nil
}

func (s Script) FetchItems(target string) (service.ServiceIterator, error) {
	return &iterator{}, nil
}

type iterator struct {
	end bool
}

func (i *iterator) Next() ([]service.Item, error) {
	i.end = true
	return []service.Item{}, nil
}

func (i iterator) HasEnded() bool { return i.end }
