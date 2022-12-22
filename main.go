package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

var cli struct {
	ListenPort         int      `kong:"default='8080',env='LISTEN_PORT',help='port to listen on'"`
	OriginAddress      string   `kong:"default='http://localhost:9200',env='ORIGIN_ADDRESS',help='upstream address to connect to. Can be IP or name, later one will be resolved'"`
	ExtraOriginHeaders []string `kong:"env='EXTRA_ORIGIN_HEADERS',help='Additional headers to add to the request, in the form of key1=value1,key2=value2'"`
}

func main() {
	kong.Parse(&cli)

	targetUrl, err := url.Parse(cli.OriginAddress)
	if err != nil {
		log.Fatal("Failed to parse target address", err)
	}
	if targetUrl.Host == "" {
		log.Fatal("Failed to parse target address, no host found")
	}
	if targetUrl.Scheme == "" {
		log.Fatal("Failed to parse target address, no scheme found")
	}
	if targetUrl.Path != "" {
		log.Fatal("Failed to parse target address, path not allowed")
	}

	extraHeaders := map[string]string{}
	for _, header := range cli.ExtraOriginHeaders {
		parts := strings.Split(header, "=")
		if len(parts) != 2 {
			log.Fatal("Failed to parse extra header", header)
		}
		extraHeaders[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	director := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", targetUrl.Host)
		for header, value := range extraHeaders {
			req.Header.Add(header, value)
		}

		req.URL.Scheme = targetUrl.Scheme
		req.URL.Host = targetUrl.Host
	}

	proxy := &httputil.ReverseProxy{Director: director}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cli.ListenPort), nil))
}
