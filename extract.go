package main

import (
	"regexp"
	"sort"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

var removeThink = regexp.MustCompile(`(?s)<think>.*?</think>`)

type ExtractFunc func(string) string

var extractors = map[string]ExtractFunc{
	"default": defaultExtract,
	"result":  resultExtract,
	"code":    codeExtract,
}

func listExtract() []string {
	list := make([]string, 0, len(extractors))
	for k := range extractors {
		list = append(list, k)
	}
	sort.Strings(list)
	return list
}

func extractResponse(selection string) ExtractFunc {
	if e, ok := extractors[selection]; ok {
		return e
	}
	return defaultExtract
}

func defaultExtract(s string) string {
	if md, err := glamour.Render(s, "dark"); err == nil {
		return md
	}
	return s
}

func resultExtract(s string) string {
	clean := removeThink.ReplaceAllString(s, "")
	return defaultExtract(clean)
}

func codeExtract(md string) string {
	src := []byte(md)
	r := text.NewReader(src)
	parser := goldmark.DefaultParser()
	root := parser.Parse(r)
	queue := []ast.Node{root}
	var code strings.Builder
	for len(queue) != 0 {
		current := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if current == nil {
			continue
		}
		if current.Kind() == ast.KindFencedCodeBlock {
			block := current.(*ast.FencedCodeBlock)
			code.WriteString("```")
			code.Write(block.Language(src))
			code.WriteString("\n")
			code.Write(block.Lines().Value(src))
			code.WriteString("```\n")
		}
		queue = append(queue, current.NextSibling(), current.FirstChild())
	}
	return code.String()
}
