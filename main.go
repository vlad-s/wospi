package main

import (
	"log"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/vlad-s/wospi/spider"
)

var (
	minLength = kingpin.Flag("min-length", "minimum word length, applies to content in body before and after stripping").
			Short('l').Default("4").Int()

	maxDepth = kingpin.Flag("max-depth", "maximum depth for crawling, zero means infinite crawling").
			Short('d').Default("0").Int()

	strictDomain = kingpin.Flag("strict-domain", "strict domain matching, don't add www/non-www to allowed hosts").
			Short('s').Default("false").Bool()

	userAgent = kingpin.Flag("user-agent", "user agent used in HTTP requests").
			Short('u').Default(spider.UserAgent).String()

	stripResult = kingpin.Flag("strip", "also store the stripped version of the result").
			Default("false").Bool()

	stripCharset = kingpin.Flag("strip-charset", "charset used in stripping the result").
			Default("!@#$%^&*()-=`~[]\\{}|;':\",./<>?").String()

	url = kingpin.Arg("url", "URL to scrape").Required().String()
)

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	s := spider.New(&spider.Options{
		MinLength: *minLength,
		MaxDepth:  *maxDepth,

		UserAgent:    *userAgent,
		StrictDomain: *strictDomain,

		StripResult:  *stripResult,
		StripCharset: []rune(*stripCharset),
	})

	if err := s.Run(*url); err != nil {
		log.Printf("couldn't run spider: %s", err)
		os.Exit(1)
	}
}
