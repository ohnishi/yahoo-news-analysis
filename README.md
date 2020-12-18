# yahoo-news-analysis
Yahoo!ニュースの記事タイトルから出現するキーワードを形態素解析で抽出するコマンド

### Move current directory

cd $GOPATH/src/github.com/ohnishi/yahoo-news-analysis

### fetch yahoo news rss list
go run github.com/ohnishi/yahoo-news-analysis/cmd yahoo --dest ~/Desktop/fetch

### fetch yahoo news rss
go run github.com/ohnishi/yahoo-news-analysis/cmd rss --src ~/Desktop/fetch --dest ~/Desktop/fetch

### transform yahoo news rss to json
go run github.com/ohnishi/yahoo-news-analysis/cmd json --src ~/Desktop/fetch --dest ~/Desktop/transform --date 20201218

### transform analysis macab
go run github.com/ohnishi/yahoo-news-analysis/cmd analysis --src ~/Desktop/transform --dest ~/Desktop/transform --date 20201218

### transform analysis mecab report markdown
go run github.com/ohnishi/yahoo-news-analysis/cmd markdown --src ~/Desktop/transform --dest ~/Desktop/transform --date 20201218
