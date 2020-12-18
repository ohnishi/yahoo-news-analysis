package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// DatesFlagFormat は`--date`フラグで用いる日付のフォーマットを表す。
const DatesFlagFormat = "20060102"

// StringSliceVarSetter はStringSliceフラグをセットするインタフェースを表す
type StringSliceVarSetter interface {
	StringSliceVar(p *[]string, name string, value []string, usage string)
}

func createOutFile(path string) (*os.File, error) {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create output directory: %s", dir)
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file: %s", path)
	}
	return f, nil
}

func appendOutFile(f *os.File, v interface{}) error {
	jsonl, err := toJSON(v)
	if err != nil {
		return err
	}
	if _, err := f.Write([]byte(jsonl)); err != nil {
		return errors.Wrapf(err, "failed to write line: %s", jsonl)
	}
	return nil
}

func toJSON(r interface{}) (string, error) {
	jsonStr, err := json.Marshal(r)
	if err != nil {
		return "", errors.Wrapf(err, "could not marshal: %v", r)
	}
	return fmt.Sprintf("%s\n", jsonStr), nil
}

type Content struct {
	FormatDate string        `json:"format_date"`
	Date       string        `json:"date"`
	Items      []ContentItem `json:"items"`
}

type ContentItem struct {
	Word     string    `json:"word"`
	Count    int       `json:"count"`
	Articles []Article `json:"articles"`
}

type Article struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type YahooRSSFeed struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type NewsArticleJSON struct {
	Date     string `json:"date"`
	URL      string `json:"url"`
	Name     string `json:"name"`
	Title    string `json:"title"`
	Category string `json:"category"`
}

func readYahooRSSFeed(path string) ([]YahooRSSFeed, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file: %s", path)
	}
	defer f.Close()

	var articles []YahooRSSFeed
	d := json.NewDecoder(f)
	for d.More() {
		var article YahooRSSFeed
		if err := d.Decode(&article); err != nil {
			return nil, errors.Wrapf(err, "could not unmarshal: %v", article)
		}
		articles = append(articles, article)
	}
	return articles, nil
}

func withLoggingE(fn func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return withLogging(fn, cmd, args)
	}
}

func withLogging(fn func(cmd *cobra.Command, args []string) error, cmd *cobra.Command, args []string) error {
	err := fn(cmd, args)
	if err == nil {
		return nil
	}

	if err == flag.ErrHelp {
		return cmd.Help()
	}

	cmd.Printf("Error: %+v\n", err)
	if isFlagError(err) {
		cmd.Usage()
	}
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	return err
}

func eachDate(date []string, fn func(time.Time) error) error {
	step, err := getStepFunc("daily")
	if err != nil {
		return err
	}
	return eachByStep(date, step, fn)
}

// getStepFunc は period に応じた対象日付の次の期間の始めの日付を取得する関数を取得する
func getStepFunc(period string) (func(time.Time) time.Time, error) {
	switch period {
	case "daily":
		return func(d time.Time) time.Time {
			return d.AddDate(0, 0, 1)
		}, nil
	case "weekly":
		return func(d time.Time) time.Time {
			return d.AddDate(0, 0, 7)
		}, nil
	case "monthly":
		return func(d time.Time) time.Time {
			return d.AddDate(0, 1, 0)
		}, nil
	case "quarterly":
		return func(d time.Time) time.Time {
			return d.AddDate(0, 3, 0)
		}, nil
	case "yearly":
		return func(d time.Time) time.Time {
			return d.AddDate(1, 0, 0)
		}, nil
	default:
		return nil, errors.Errorf("invalid period: %s", period)
	}
}

func eachByStep(date []string, step func(time.Time) time.Time, fn func(time.Time) error) error {
	switch len(date) {
	case 0:
		return errors.New("one or two date values must be specified")
	case 1:
		d, err := parseLocal(DatesFlagFormat, date[0])
		if err != nil {
			return err
		}
		return fn(d)
	case 2:
		since, err := parseLocal(DatesFlagFormat, date[0])
		if err != nil {
			return err
		}
		until, err := parseLocal(DatesFlagFormat, date[1])
		if err != nil {
			return err
		}
		if since.After(until) {
			since, until = until, since
		}
		var errs error
		for d := since; !d.After(until); d = step(d) {
			err = fn(d)
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}
		return errs
	default:
		return errors.New("more than 2 values cannot be specified for date")
	}

}

func setDatesFlag(f StringSliceVarSetter, p *[]string, purpose string) {
	setRangeFlag(f, p, "date", purpose)
}

func setRangeFlag(f StringSliceVarSetter, p *[]string, name string, purpose string) {
	const (
		format = "%s in 'YYYYmmdd' or period in 'YYYYmmdd,YYYYmmdd' " +
			"(e.g. date: '20180101', period: '20180101,20180131')"
	)
	f.StringSliceVar(p, name, []string{}, fmt.Sprintf(format, purpose))
}

func parseLocal(layout string, value string) (time.Time, error) {
	t, err := time.ParseInLocation(layout, value, time.Local)
	if err != nil {
		return time.Time{}, errors.Wrapf(err, "cannot parse as %q", time.Local)
	}
	return t, nil
}

type flagError struct {
	Message string
	Args    []interface{}
}

func (e flagError) Error() string {
	return fmt.Sprintf(e.Message, e.Args...)
}

func isFlagError(err error) bool {
	_, ok := err.(flagError)
	return ok
}
