package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/saintfish/chardet"
	"golang.org/x/net/html/charset"
)

const rssListURL = "https://news.yahoo.co.jp/rss"

type link struct {
	text string
	href string
}

func fetchYahooNewsRSSList(dest string, maxRetry uint) (err error) {
	var links []link
	retry := uint(0)
	for {
		links, err = func() ([]link, error) {
			res, err := http.Get(rssListURL)
			if err != nil {
				return nil, errors.Wrapf(err, "failed request url : %s", rssListURL)
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				return nil, errors.Errorf("status code expected 200 but was %d : url=%s", res.StatusCode)
			}

			feeds, err := getYahooRSSFeeds(res.Body)
			if err != nil {
				return nil, err
			}
			return feeds, nil
		}()
		retry++
		if err == nil || retry > maxRetry {
			break
		}
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		return err
	}

	return saveYahooRSS(dest, "rss.jsonl", links)
}

func getYahooRSSFeeds(r io.Reader) ([]link, error) {
	buf, _ := ioutil.ReadAll(r)

	// 文字コード判定
	det := chardet.NewTextDetector()
	detRslt, _ := det.DetectBest(buf)
	// 文字コード変換
	bReader := bytes.NewReader(buf)
	reader, _ := charset.NewReaderLabel(detRslt.Charset, bReader)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read response body")
		}
		return nil, errors.Wrapf(err, "failed parse response body : %s", string(b))
	}

	var links []link
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && strings.HasPrefix(href, "/rss/") {
			l := link{
				text: s.Text(),
				href: href,
			}
			links = append(links, l)
		}
	})
	return links, nil
}

func saveYahooRSS(out, fileName string, links []link) error {
	if len(links) == 0 {
		return nil
	}
	f, err := createOutFile(filepath.Join(out, fileName))
	if err != nil {
		return err
	}
	defer f.Close()

	for _, link := range links {
		feed := YahooRSSFeed{
			ID:   link.href[1 : len(link.href)-4],
			Name: link.text,
			URL:  "https://news.yahoo.co.jp" + link.href,
		}
		err = appendOutFile(f, feed)
		if err != nil {
			return err
		}
	}
	if err := f.Sync(); err != nil {
		return errors.Wrap(err, "failed to sync file")
	}
	return nil
}
