package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchCrumbHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Jenkins-Crumb:3c70906d413e89cb6003b86aa4ee7b7e31b4f5d1c1420232aa82c78d35d7f8ec")
	}))
	defer ts.Close()
	basicAuth := BasicAuth{"username", "p@ssword"}
	expected := CrumbHeader{"Jenkins-Crumb", "3c70906d413e89cb6003b86aa4ee7b7e31b4f5d1c1420232aa82c78d35d7f8ec"}
	got, err := fetchCrumbHeader(basicAuth, ts.URL)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if got != expected {
		t.Errorf("Expected: %q, got: %q", expected, got)
	}
}

func TestValidate(t *testing.T) {
	expected := "Jenkinsfile successfully validated.\n"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expected)
	}))
	defer ts.Close()
	basicAuth := BasicAuth{"username", "p@ssword"}
	crumbHeader := CrumbHeader{"Jenkins-Crumb", "3c70906d413e89cb6003b86aa4ee7b7e31b4f5d1c1420232aa82c78d35d7f8ec"}
	got := validate(basicAuth, ts.URL, "", crumbHeader)
	if got != expected {
		t.Errorf("Expected: %q, got: %q", expected, got)
	}
}
