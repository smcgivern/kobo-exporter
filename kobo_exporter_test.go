package main

import (
	"log"
	"os"
	"testing"
)

func TestPriceToFloat(t *testing.T) {
	cases := []struct {
		input    string
		ok       bool
		expected float64
	}{
		{"1.35", true, 1.35},
		{"£1.35", true, 1.35},
		{"$22.76  AUD", true, 22.76},
		{"$13.59  CAD", true, 13.59},
		{"13,69 €", true, 13.69},
		{"", false, 0},
		{"€", false, 0},
		{"some text", false, 0},
	}
	for _, c := range cases {
		ok, price := PriceToFloat(c.input)
		if ok != c.ok {
			t.Errorf("PriceToFloat(%q) == %v, _, want %v", c.input, ok, c.ok)
		}
		if price != c.expected {
			t.Errorf("PriceToFloat(%q) == _, %v, want %v", c.input, price, c.expected)
		}
	}
}

func TestFindInfo(t *testing.T) {
	cases := []struct {
		file     string
		ok       bool
		expected BookInfo
	}{
		// Multiple currencies
		{"warlock-gb", true, BookInfo{price: 9.89, title: "Warlock", author: "Oakley Hall"}},
		{"warlock-es", true, BookInfo{price: 11.95, title: "Warlock", author: "Oakley Hall"}},
		{"warlock-ca", true, BookInfo{price: 13.59, title: "Warlock", author: "Oakley Hall"}},
		{"warlock-au", true, BookInfo{price: 22.76, title: "Warlock", author: "Oakley Hall"}},

		// No author
		{"broken-stars", true, BookInfo{price: 3.99, title: "Broken Stars", author: ""}},

		// Multiple authors
		{"a-memory-of-light", true, BookInfo{price: 6.99, title: "A Memory Of Light", author: "Robert Jordan"}},

		// Cannot be parsed as HTML
		{"empty-file", false, BookInfo{price: 0, title: "", author: ""}},

		// Can't read price
		{"pride-and-prejudice-free", false, BookInfo{price: 0, title: "", author: ""}},
	}
	for _, c := range cases {
		fixture, err := os.Open("testdata/" + c.file)

		if err != nil {
			log.Fatal(err)
		}

		ok, info := FindInfo(fixture)
		if ok != c.ok {
			t.Errorf("FindInfo(file:%q) == %v, _, want %v", c.file, ok, c.ok)
		}
		if info != c.expected {
			t.Errorf("FindInfo(file:%q) == _, %v, want %v", c.file, info, c.expected)
		}
	}
}
