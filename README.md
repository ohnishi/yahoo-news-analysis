# yahoo-news-analysis
The keywords appearing in the article titles of Yahoo! News Japan are extracted by morphological analysis and ranked.


### Set up mecab-ipadic-NEologd.
https://github.com/neologd/mecab-ipadic-neologd/blob/master/README.ja.md

### Move current directory
cd $GOPATH/src/github.com/ohnishi/yahoo-news-analysis

### Fetch yahoo news rss list
go run github.com/ohnishi/yahoo-news-analysis/cmd yahoo --dest ~/Desktop/fetch

### Fetch yahoo news rss
go run github.com/ohnishi/yahoo-news-analysis/cmd rss --src ~/Desktop/fetch --dest ~/Desktop/fetch

### Transform yahoo news rss to json
go run github.com/ohnishi/yahoo-news-analysis/cmd json --src ~/Desktop/fetch --dest ~/Desktop/transform --date 20201218

### Transform analysis macab
go run github.com/ohnishi/yahoo-news-analysis/cmd analysis --src ~/Desktop/transform --dest ~/Desktop/transform --date 20201218

### Transform analysis mecab report markdown
go run github.com/ohnishi/yahoo-news-analysis/cmd markdown --src ~/Desktop/transform --dest ~/Desktop/transform --date 20201218
