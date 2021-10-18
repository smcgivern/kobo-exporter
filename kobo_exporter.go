package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

func fetchBook(url string) io.ReadCloser {
	response, err := http.Get(url)

	if err != nil {
		log.Fatal(err)
	}

	return response.Body
}

func hasClass(token html.Token, value string) bool {
	for _, attr := range token.Attr {
		if attr.Key == "class" && attr.Val == value {
			return true
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

func findPrice(url string) (ok bool, price float64) {
	z := html.NewTokenizer(fetchBook(url))
	inWrapper := false
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

				inPrice = t.Data == "span" && hasClass(t, "price")
			} else {
				inWrapper = t.Data == "div" && hasClass(t, "active-price")
			}
		case html.TextToken:
			if inPrice {
				return PriceToFloat(z.Token().Data)
			}
		}
	}

	return
}

func main() {
	ok, price := findPrice("https://www.kobo.com/gb/en/ebook/warlock-8")

	if ok {
		fmt.Println(price)
	} else {
		fmt.Println(price)
	}
}
