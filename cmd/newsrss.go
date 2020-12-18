package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func fetchYahooNewsRSS(src, dest string, maxRetry uint) error {
	feeds, err := readYahooRSSFeed(filepath.Join(src, "rss.jsonl"))
	if err != nil {
		return errors.WithMessage(err, "failed to read rss.json")
	}

	destDir := filepath.Join(dest, time.Now().Format("20060102"))
	for _, feed := range feeds {
		err = request(destDir, feed, maxRetry)
		if err != nil {
			fmt.Println("failed to fetch RSS", zap.String("url", feed.URL), zap.Error(err))
			continue
		}
	}
	return nil
}

func request(out string, feed YahooRSSFeed, maxRetry uint) (err error) {
	var res *http.Response
	retry := uint(0)
	for {
		res, err = http.Get(feed.URL)
		retry++
		if err == nil || retry > maxRetry {
			break
		}
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		return errors.Wrapf(err, "failed request url : %s", feed.URL)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.Errorf("status code expected 200 but was %d : url=%s", res.StatusCode, feed.URL)
	}

	filePath := filepath.Join(out, feed.ID)
	if err = save(res, filePath); err != nil {
		return err
	}

	return nil
}

func save(res *http.Response, path string) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "failed to create directory: %s", dir)
	}

	out, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "failed to create file: %s", path)
	}
	defer out.Close()

	_, err = io.Copy(out, res.Body)
	return err
}
