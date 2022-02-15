package main

import (
	"bufio"
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	koboPrice = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kobo_price",
		Help: "Current price of the book.",
	}, []string{"title", "author"})
	koboScrapes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kobo_scrapes",
		Help: "Number of scrapes for this book.",
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

func scrape(url string) {
	ok, info := FindInfo(fetchBook(url))

	if ok {
		koboPrice.With(prometheus.Labels{"title": info.title, "author": info.author}).Set(info.price)
		koboScrapes.With(prometheus.Labels{"title": info.title, "author": info.author}).Inc()
	}
}

func readConfig(path string) (urls []string) {
	file, err := os.Open(path)

	if err != nil {
		log.Fatal("Reading ", path, ": ", err)
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		log.Fatal("Reading ", path, ": ", err)
	}

	return urls
}

func tick(frequency time.Duration, urls []string) {
	ticker := time.NewTicker(frequency)
	i := 0
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				scrape(urls[i])
				i = (i + 1) % len(urls)
			}
		}
	}()
}

func init() {
	prometheus.MustRegister(koboPrice)
	prometheus.MustRegister(koboScrapes)
}

func main() {
	port := flag.Int("port", 8080, "Port for metrics server")
	frequency := flag.Int("frequency", 600, "Scrape frequency in seconds")
	configFile := flag.String("config", "", "Config file (line-delimited URLs)")
	urls := []string{}

	flag.Parse()

	if *configFile != "" {
		urls = readConfig(*configFile)
	} else {
		urls = flag.Args()
	}

	tick(time.Duration(*frequency)*time.Second, urls)

	http.HandleFunc("/", index)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

// Last because of the large HTML string literal.
func index(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, `
<!DOCTYPE html>
<html>
  <head>
    <title>Kobo exporter</title>
    <style type="text/css">
      body, input {
	font-family: "Palatino Linotype", "Palatino", "URW Palladio L", "Book Antiqua", "Baskerville", "Bitstream Charter", "Garamond", "Georgia", serif;
	font-size: 105%;
	color: #000;
	background: #fff;
      }

      body {
	line-height: 1.6;
	width: 50em;
	margin: 0em auto;
      }

      h1, h2, h3, h4, h5, h6 {
	font-family: "Gill Sans", "Gill Sans MT", "GillSans", "Calibri", "Trebuchet MS", sans-serif;
	text-align: center;
	font-weight: normal;
      }

      :link,:link:active,:link:hover, .click {
	color: #0000ff;
	background: transparent;
	cursor: pointer;
	text-decoration: underline;
      }

      :link:visited {
	color: #800080;
	background: transparent;
      }

      #byline {
	clear: both;
	text-align: right;
	font-size: 75%;
	font-style: italic;
      }
    </style>
  </head>
  <body>
    <h1>Kobo exporter</h1>
    <p>
      This is a <a href="https://prometheus.io">Prometheus</a> exporter
      for <a href="https://www.kobo.com">Kobo ebook prices</a>. It
      exports a Prometheus
      <a href="https://prometheus.io/docs/concepts/metric_types/#gauge">gauge</a>
      for a set of books with their current price, and scrapes each
      book's price periodically to update the gauge.
    </p>
    <p>
      The metrics for my list of books are available at
      <a href="./metrics"><code>./metrics</code></a>.
      I use it in conjunction with
      <a href="https://prometheus.io/docs/alerting/latest/alertmanager/">Alertmanager</a>
      to provide alerts when Kobo books I'm interested in go on sale.
    </p>
    <p>
      The <a href="https://github.com/smcgivern/kobo-exporter">source
      is on GitHub</a>.
    </p>
    <div id="byline">
      <p>
	By
	<a href="http://sean.mcgivern.me.uk/">Sean McGivern</a>
      </p>
    </div>
  </body>
</html>
`)
}
