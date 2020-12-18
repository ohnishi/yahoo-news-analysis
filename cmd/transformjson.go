package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// newsArticleJSON transformしたニュース記事データ
type newsArticleJSON struct {
	Date  string `json:"date"`
	URL   string `json:"url"`
	Name  string `json:"name"`
	Title string `json:"title"`
}

// transformJSON fetchしたRSSファイルからターゲット日に更新された記事を抽出する
func transformJSON(src, dest string, date time.Time) error {
	feeds, err := readYahooRSSFeed(filepath.Join(src, "rss.jsonl"))
	if err != nil {
		return errors.WithMessage(err, "failed to read rss.json")
	}

	dateStr := date.Format("20060102")
	articleMap, err := toArticleMap(feeds, src, dateStr, date)
	if err != nil {
		return err
	}

	return writeArticleJSOL(dest, dateStr, "rss.jsonl", articleMap)
}

// RSS設定JSONとfetchしたRSSファイルからターゲット日付のニュース記事を抽出して保存する
func toArticleMap(feeds []YahooRSSFeed, src, dateStr string, date time.Time) (map[string]newsArticleJSON, error) {
	m := make(map[string]newsArticleJSON)
	fileDir := filepath.Join(src, dateStr)
	for _, feed := range feeds {
		filePath := filepath.Join(fileDir, feed.ID)
		stat, err := os.Stat(filePath)
		if err != nil || stat.IsDir() {
			// RSSリストが更新されてfetchファイルが存在しないケース
			continue
		}
		rss, err := os.Open(filePath)
		if err != nil {
			// RSSファイルの読み込み失敗しても処理は止めずに warnnig log を出力する
			fmt.Println("failed to open RSS file.", zap.String("path", filePath), zap.Error(err))
			continue
		}

		gfp := gofeed.NewParser()
		feed, parseErr := gfp.Parse(rss)
		closeErr := rss.Close()
		if closeErr != nil {
			return nil, errors.Wrapf(closeErr, "failed to close a rss reader: %s", filePath)
		}
		if parseErr != nil {
			// RSSの解析に失敗しても処理は止めずに warnnig log を出力する
			fmt.Println("failed to parse RSS.", zap.String("path", filePath), zap.Error(parseErr))
			continue
		}
		for _, item := range feed.Items {
			if _, ok := m[item.Link]; ok {
				continue
			}

			var articleDate time.Time
			if item.PublishedParsed != nil {
				articleDate = item.PublishedParsed.In(time.Local)
			} else if item.UpdatedParsed != nil {
				articleDate = item.UpdatedParsed.In(time.Local)
			} else {
				articleDate = date
			}
			if dateStr != articleDate.Format("20060102") {
				continue
			}

			json := newsArticleJSON{
				Date:  articleDate.Format(time.RFC3339),
				URL:   item.Link,
				Name:  feed.Title,
				Title: item.Title,
			}
			m[item.Link] = json
		}
	}
	return m, nil
}

// ニュース記事データをファイルに保存します
func writeArticleJSOL(out, date, fileName string, m map[string]newsArticleJSON) error {
	if len(m) == 0 {
		return nil
	}
	f, err := createOutFile(filepath.Join(out, date, fileName))
	if err != nil {
		return err
	}
	defer f.Close()

	for _, json := range m {
		err = appendOutFile(f, json)
		if err != nil {
			return err
		}
	}
	if err := f.Sync(); err != nil {
		return errors.Wrap(err, "failed to sync file")
	}
	return nil
}
