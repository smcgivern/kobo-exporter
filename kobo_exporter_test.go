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

func TestFindPrice(t *testing.T) {
	cases := []struct {
		file     string
		ok       bool
		expected float64
	}{
		{"warlock-gb", true, 9.89},
		{"warlock-es", true, 11.95},
		{"warlock-ca", true, 13.59},
		{"warlock-au", true, 22.76},
		{"empty-file", false, 0},
		{"pride-and-prejudice-free", false, 0},
	}
	for _, c := range cases {
		fixture, err := os.Open("testdata/" + c.file)

		if err != nil {
			log.Fatal(err)
		}

		ok, price := FindPrice(fixture)
		if ok != c.ok {
			t.Errorf("FindPrice(file:%q) == %v, _, want %v", c.file, ok, c.ok)
		}
		if price != c.expected {
			t.Errorf("FindPrice(file:%q) == _, %v, want %v", c.file, price, c.expected)
		}
	}
}
