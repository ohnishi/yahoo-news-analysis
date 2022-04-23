# yahoo-news-analysis
Yahoo!ニュースの記事タイトルに登場するキーワードを形態素解析によって抽出して、ランク付けします。


### 以下を参考に mecab-ipadic-NEologd を設定します
https://github.com/neologd/mecab-ipadic-neologd/blob/master/README.ja.md

### カレントディレクトリを移動
cd $GOPATH/src/github.com/ohnishi/yahoo-news-analysis

### Yahoo!ニュース のRSSリストを取得します
go run github.com/ohnishi/yahoo-news-analysis/cmd yahoo --dest ~/Desktop/fetch

### RSSからYahoo!ニュースの記事を取得します
go run github.com/ohnishi/yahoo-news-analysis/cmd rss --src ~/Desktop/fetch --dest ~/Desktop/fetch

### Yahoo!ニュースの記事情報をJSONに変換します
go run github.com/ohnishi/yahoo-news-analysis/cmd json --src ~/Desktop/fetch --dest ~/Desktop/transform --date 20201218

### JSONのニュース記事から形態素解析によって、キーワード抽出します
go run github.com/ohnishi/yahoo-news-analysis/cmd analysis --src ~/Desktop/transform --dest ~/Desktop/transform --date 20201218

### 抽出したキーワード数を集計してランク付けしたレポートを生成します
go run github.com/ohnishi/yahoo-news-analysis/cmd markdown --src ~/Desktop/transform --dest ~/Desktop/transform --date 20201218
