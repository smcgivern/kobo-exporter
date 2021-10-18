package main

import "testing"

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
