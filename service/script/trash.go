package script

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"go.starlark.net/starlark"
)

type document struct {
	body string
	find *starlark.Builtin
}

func newDocument(body string) document {
	queryDoc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		panic(err)
	}

	return document{
		body: body,
		find: starlark.NewBuiltin("find", find(queryDoc)),
	}
}

func (d document) String() string {
	return `"` + d.body + `"`
}

func (d document) Type() string {
	return "document"
}

func (d document) Freeze() {}

func (d document) Truth() starlark.Bool {
	return d.body != ""
}

func (d document) Hash() (uint32, error) {
	return 0, nil // TODO: make an actual hash
}

func (d document) Attr(name string) (starlark.Value, error) {
	switch name {
	case "body":
		return starlark.String(d.body), nil
	case "find":
		return d.find, nil
	default:
		return nil, starlark.NoSuchAttrError(name)
	}
}

func (d document) AttrNames() []string {
	return []string{"body"}
}

type selection struct{ inner *goquery.Selection }

func (s selection) String() string        { return `"` + s.inner.Text() + `"` }
func (s selection) Type() string          { return "selection" }
func (s selection) Freeze()               {}
func (s selection) Truth() starlark.Bool  { return s.inner.Length() > 0 }
func (s selection) Hash() (uint32, error) { return 0, nil }
func (s selection) AttrNames() []string   { return []string{"text"} }
func (s selection) Attr(name string) (starlark.Value, error) {
	switch name {
	case "text":
		return starlark.String(s.inner.Text()), nil
	case "attr":
		return starlark.NewBuiltin("attr", attr(s.inner)), nil
	default:
		return nil, starlark.NoSuchAttrError(name)
	}
}
func (s selection) Index(i int) starlark.Value {
	return selection{s.inner.Eq(i)}
}
func (s selection) Len() int {
	return s.inner.Length()
}
func (s selection) Iterate() starlark.Iterator {
	return &selectionIterator{s.inner, 0}
}

var (
	_ starlark.Indexable = (*selection)(nil)
	_ starlark.Iterable  = (*selection)(nil)
)

type selectionIterator struct {
	inner *goquery.Selection
	index int
}

func (i *selectionIterator) Next(p *starlark.Value) bool {
	if i.index >= i.inner.Length() {
		return false
	}

	*p = selection{i.inner.Eq(i.index)}
	i.index++
	return true
}

func (i *selectionIterator) Done() { i.inner = nil }

func find(doc *goquery.Document) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var selector string
		starlark.UnpackArgs(b.Name(), args, kwargs, "selector", &selector)
		return selection{doc.Find(selector)}, nil
	}
}

func attr(sel *goquery.Selection) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var name string
		starlark.UnpackArgs(b.Name(), args, kwargs, "name", &name)
		a, _ := sel.Attr(name)
		return starlark.String(a), nil
	}
}

func fetch(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var url string

	starlark.UnpackArgs(b.Name(), args, kwargs, "url", &url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	val := newDocument(string(body))

	return val, nil
}
