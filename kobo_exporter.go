package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	koboPrice = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kobo_price",
		Help: "Current price of the book.",
	}, []string{"title", "author"})
)

type BookInfo struct {
	price  float64
	title  string
	author string
}

func fetchBook(url string) io.ReadCloser {
	response, err := http.Get(url)

	if err != nil {
		log.Fatal(err)
	}

	return response.Body
}

func hasClass(token html.Token, value string) bool {
	for _, attr := range token.Attr {
		if attr.Key == "class" {
			for _, class := range strings.Fields(attr.Val) {
				if class == value {
					return true
				}
			}
		}
	}

	return false
}

func PriceToFloat(data string) (bool, float64) {
	re := regexp.MustCompile(`\D*(\d+)[.,](\d+)\D*`)
	price, err := strconv.ParseFloat(re.ReplaceAllString(data, "$1.$2"), 64)

	if err != nil {
		return false, 0.0
	}

	return true, price
}

func FindInfo(page io.ReadCloser) (ok bool, info BookInfo) {
	z := html.NewTokenizer(page)
	title := ""
	author := ""
	inWrapper := false
	inTitle := false
	inAuthor := false
	inPrice := false
	depth := 0

	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			return
		case html.StartTagToken, html.EndTagToken:
			t := z.Token()
			if inWrapper {

				if tt == html.EndTagToken {
					depth--
					inWrapper = depth > 0
				} else {
					depth++
				}

				inTitle = t.Data == "h2" && hasClass(t, "title")
				inAuthor = t.Data == "a" && hasClass(t, "contributor-name")
				inPrice = t.Data == "span" && hasClass(t, "price")
			} else {
				if t.Data == "div" && (hasClass(t, "item-info") || hasClass(t, "active-price")) {
					inWrapper = true
					depth = 1
				}
			}
		case html.TextToken:
			if inTitle && title == "" {
				title = strings.TrimSpace(z.Token().Data)
			} else if inAuthor && author == "" {
				author = z.Token().Data
			} else if inPrice {
				priceOK, price := PriceToFloat(z.Token().Data)

				if priceOK {
					return priceOK, BookInfo{title: title, author: author, price: price}
				} else {
					return
				}
			}
		}
	}

	return
}

func init() {
	prometheus.MustRegister(koboPrice)
}

func main() {
	port := flag.Int("port", 8080, "Port for metrics server")

	flag.Parse()

	for _, url := range flag.Args() {
		ok, info := FindInfo(fetchBook(url))

		if ok {
			koboPrice.With(prometheus.Labels{"title": info.title, "author": info.author}).Set(info.price)
		}
	}

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
