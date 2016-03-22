package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	REQUEST_STORE         = "http://localhost:9091/terminales/500/"
	REQUEST_COUNTRY       = "http://localhost:9091/terminales/PT/"
	REQUEST_BRAND         = "http://localhost:9091/terminales/Z/"
	REQUEST_BRAND_COUNTRY = "http://localhost:9091/terminales/Z/UK"
)

func TestLookupTerminal(t *testing.T) {
	req, _ := http.NewRequest("GET", REQUEST_STORE, nil)

	w := httptest.NewRecorder()
	terminales(w, req)

	if w.Code != http.StatusTemporaryRedirect {
		t.Fatal(
			"For", REQUEST_STORE,
			"expected status", http.StatusTemporaryRedirect,
			"got", w.Code)
	}

	if w.Header().Get("Location") != LOOKUP_BY_STORE+"500" {
		t.Error(
			"For", REQUEST_STORE,
			"expected location", LOOKUP_BY_STORE+"500",
			"got", w.Header().Get("Location"))
	}
}

func TestLookupByCountry(t *testing.T) {
	req, _ := http.NewRequest("GET", REQUEST_COUNTRY, nil)

	w := httptest.NewRecorder()
	terminales(w, req)

	if w.Code != http.StatusTemporaryRedirect {
		t.Fatal(
			"For", REQUEST_COUNTRY,
			"expected status", http.StatusTemporaryRedirect,
			"got", w.Code)
	}

	if w.Header().Get("Location") != LOOKUP_BY_BRAND+UNKNOWN+AND_COUNTRY+"PT" {
		t.Error(
			"For", REQUEST_COUNTRY,
			"expected location", LOOKUP_BY_BRAND+UNKNOWN+AND_COUNTRY+"PT",
			"got", w.Header().Get("Location"))
	}
}

func TestLookupByBrand(t *testing.T) {
	req, _ := http.NewRequest("GET", REQUEST_BRAND, nil)

	w := httptest.NewRecorder()
	terminales(w, req)

	if w.Code != http.StatusTemporaryRedirect {
		t.Fatal(
			"For", REQUEST_BRAND,
			"expected status", http.StatusTemporaryRedirect,
			"got", w.Code)
	}

	if w.Header().Get("Location") != LOOKUP_BY_BRAND+"Z"+AND_COUNTRY+UNKNOWN {
		t.Error(
			"For", REQUEST_BRAND,
			"expected location", LOOKUP_BY_BRAND+"Z"+AND_COUNTRY+UNKNOWN,
			"got", w.Header().Get("Location"))
	}
}

func TestLookupByBrandAndCountry(t *testing.T) {
	req, _ := http.NewRequest("GET", REQUEST_BRAND_COUNTRY, nil)

	w := httptest.NewRecorder()
	terminales(w, req)

	if w.Code != http.StatusTemporaryRedirect {
		t.Fatal(
			"For", REQUEST_BRAND_COUNTRY,
			"expected status", http.StatusTemporaryRedirect,
			"got", w.Code)
	}

	if w.Header().Get("Location") != LOOKUP_BY_BRAND+"Z"+AND_COUNTRY+"UK" {
		t.Error(
			"For", REQUEST_BRAND_COUNTRY,
			"expected location", LOOKUP_BY_BRAND+"Z"+AND_COUNTRY+"UK",
			"got", w.Header().Get("Location"))
	}
}
