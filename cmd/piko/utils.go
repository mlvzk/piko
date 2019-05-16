package main

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/mlvzk/piko/service"
)

var formatRegexp = regexp.MustCompile(`%\[[[:alnum:]]*\]`)

func format(formatter string, meta map[string]string) string {
	return formatRegexp.ReplaceAllStringFunc(formatter, func(str string) string {
		// remove "%[" and "]"
		key := str[2 : len(str)-1]

		v, ok := meta[key]
		if !ok {
			return key
		}

		return sanitizeFileName(v)
	})
}

// order of args matters
func mergeStringMaps(maps ...map[string]string) map[string]string {
	merged := map[string]string{}

	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}

	return merged
}

func tryClose(reader interface{}) error {
	if closer, ok := reader.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}

var seps = regexp.MustCompile(`[\r\n &_=+:/]`)

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	name = seps.ReplaceAllString(name, "-")

	return name
}

func prettyPrintItem(item service.Item) string {
	builder := strings.Builder{}

	builder.WriteString("Default Name: " + item.DefaultName + "\n")

	builder.WriteString("Meta:\n")
	for k, v := range item.Meta {
		if k[0] == '_' {
			continue
		}

		builder.WriteString(fmt.Sprintf("\t%s=%s\n", k, v))
	}

	builder.WriteString("Available Options:\n")
	for key, values := range item.AvailableOptions {
		builder.WriteString("\t" + key + ":\n")
		for _, v := range values {
			builder.WriteString("\t\t- " + v + "\n")
		}
	}

	builder.WriteString("Default Options:\n")
	for k, v := range item.DefaultOptions {
		builder.WriteString(fmt.Sprintf("\t%s=%s\n", k, v))
	}

	return builder.String()
}

func truncateString(str string, limit int) string {
	if len(str)-3 <= limit {
		return str
	}

	return string([]rune(str)[:limit]) + "..."
}
