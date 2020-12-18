package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

const tmplStr = `
---
title: "{{ .FormatDate }} に話題になったキーワードランキング"
date: {{ .Date }}
---

{{ range $i, $item := .Items -}}
### {{ rank $i }}位 {{ $item.Word }} （{{ $item.Count }}記事）
{{ range $j, $article := $item.Articles -}}
- [{{ $article.Title }}]({{ $article.URL }})
{{ end }}
{{ end }}
`

func transformMarkdown(src, dest string, date time.Time) (err error) {
	srcPath := filepath.Join(src, date.Format("20060102"), "topic.json")
	f, err := ioutil.ReadFile(srcPath)
	if err != nil {
		return err
	}

	var c Content
	if err = json.Unmarshal(f, &c); err != nil {
		return err
	}

	if len(c.Items) == 0 {
		return errors.New("content size is zero")
	}

	if err = writeContent(dest, date, c); err != nil {
		return err
	}

	return nil
}

func writeContent(dest string, date time.Time, content Content) error {
	f, err := createOutFile(filepath.Join(dest, date.Format("20060102"), "report.md"))
	if err != nil {
		return err
	}
	defer f.Close()

	funcMap := template.FuncMap{
		"rank": func(a int) int { return a + 1 },
	}
	t := template.Must(template.New("funcmap").Funcs(funcMap).Parse(tmplStr))

	// Execute(io.Writer(出力先), データ)
	if err := t.Execute(f, content); err != nil {
		log.Fatal(err)
	}

	if err := f.Sync(); err != nil {
		return errors.Wrap(err, "failed to sync file")
	}
	return nil
}
