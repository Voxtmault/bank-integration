package bca_test

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"testing"
)

func TestForwardProxy(t *testing.T) {
	// proxyURL, err := url.Parse("http://116.68.252.213:3128")
	proxyURL, err := url.Parse("http://wallet.crossnet.co.id:3128")
	if err != nil {
		t.Errorf("Error parsing proxy URL: %v", err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	client := &http.Client{
		Transport: transport,
	}

	// Use the client for your requests
	resp, err := client.Get("https://httpbin.org/ip")
	if err != nil {
		t.Errorf("Error making GET request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response body: %v", err)
	}

	log.Printf("Response body: %s", body)

	// Handle the response
}
