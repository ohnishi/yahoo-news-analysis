package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	mecab "github.com/shogo82148/go-mecab"
)

const ipadic = "/usr/local/lib/mecab/dic/mecab-ipadic-neologd"

var newsArticleNames = []string{"rss.jsonl"}

func transformAnalysis(src, dest string, date time.Time) error {
	dateStr := date.Format("20060102")
	var articles []NewsArticleJSON
	for _, fileName := range newsArticleNames {
		path := filepath.Join(src, dateStr, fileName)
		a, err := readArticles(path)
		if err != nil {
			fmt.Println("failed to open JSONL file.", zap.String("path", path), zap.Error(err))
			continue
		}
		articles = append(articles, a...)
	}

	contentItems := toContents(articles)
	if len(contentItems) >= 30 {
		contentItems = contentItems[:30]
	}

	content := Content{
		FormatDate: date.Format("2006/01/02"),
		Date:       date.Format(time.RFC3339),
		Items:      contentItems,
	}

	if err := writeContentMecab(dest, dateStr, "topic.json", content); err != nil {
		return err
	}
	return nil
}

func writeContentMecab(dest, dateStr, fileName string, c Content) error {
	f, err := createOutFile(filepath.Join(dest, dateStr, fileName))
	if err != nil {
		return err
	}
	defer f.Close()

	err = appendOutFile(f, c)
	if err != nil {
		return err
	}

	if err := f.Sync(); err != nil {
		return errors.Wrap(err, "failed to sync file")
	}
	return nil
}

// ニュース記事情報となるJSONLファイルをreadして返す
func readArticles(path string) ([]NewsArticleJSON, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file: %s", path)
	}
	defer f.Close()

	var articles []NewsArticleJSON
	d := json.NewDecoder(f)
	for d.More() {
		var article NewsArticleJSON
		if err := d.Decode(&article); err != nil {
			return nil, errors.Wrapf(err, "could not unmarshal: %v", article)
		}
		articles = append(articles, article)
	}
	return articles, nil
}

func toContents(articles []NewsArticleJSON) []ContentItem {
	mecab, err := mecab.New(map[string]string{"dicdir": ipadic})
	if err != nil {
		panic(err)
	}
	defer mecab.Destroy()

	m := make(map[string]ContentItem)
	for _, article := range articles {
		title := strings.TrimSpace(strings.ToLower(article.Title))
		i := strings.LastIndex(title, "(")
		if i >= 0 {
			title = title[:i]
		}
		i = strings.LastIndex(title, "（")
		if i >= 0 {
			title = title[:i]
		}
		i = strings.LastIndex(title, "[")
		if i >= 0 {
			title = title[:i]
		}
		i = strings.LastIndex(title, "〈")
		if i >= 0 {
			title = title[:i]
		}
		i = strings.LastIndex(title, "【")
		if i >= 0 {
			title = title[:i]
		}
		i = strings.Index(title, "]")
		if i >= 0 {
			title = title[i:]
		}
		title = strings.ReplaceAll(title, ":", "")
		title = strings.ReplaceAll(title, "にも", "")
		node, err := mecab.ParseToNode(title)
		if err != nil {
			panic(err)
		}

		for ; !node.IsZero(); node = node.Next() {
			features := strings.Split(node.Feature(), ",")
			if features[0] == "名詞" && features[1] == "固有名詞" && features[2] == "人名" && features[3] == "一般" {
				// fmt.Println(node.String())
				word := node.Surface()
				contentItem, ok := m[word]
				if !ok {
					contentItem = ContentItem{
						Word:  word,
						Count: 0,
					}
					m[word] = contentItem
				}
				a := Article{
					Title: article.Title,
					URL:   article.URL,
				}
				contentItem.Articles = append(contentItem.Articles, a)
				contentItem.Count = len(contentItem.Articles)
				m[word] = contentItem
			}
		}
	}
	var ret []ContentItem
	for _, val := range m {
		// fmt.Println(fmt.Sprintf("\"%s\":            {},", key))
		ret = append(ret, val)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].Count > ret[j].Count })
	return ret[:100]
}
