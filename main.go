package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/simonfrey/saf_tcp_everse_proxy/pkg/dumptransport"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

var cli struct {
	ListenPort         int      `kong:"default='9200',env='LISTEN_PORT',help='port to listen on'"`
	OriginAddress      string   `kong:"default='https://www.google.com/',env='ORIGIN_ADDRESS',help='upstream address to connect to. Can be IP or name, later one will be resolved'"`
	ExtraOriginHeaders []string `kong:"env='EXTRA_ORIGIN_HEADERS',help='Additional headers to add to the request, in the form of key1=value1,key2=value2'"`
	DumpRequests       bool     `kong:"env='DUMP_REQUEST',help='If set to true then all request and responses will be dumped to the console'"`
}

func main() {
	kong.Parse(&cli)

	origin, err := url.Parse(cli.OriginAddress)
	if err != nil {
		log.Fatal("Failed to parse origin address", err)
	}
	if origin.Host == "" {
		log.Fatal("Failed to parse origin address, no host found")
	}
	if origin.Scheme == "" {
		log.Fatal("Failed to parse origin address, no scheme found")
	}

	extraHeaders := map[string]string{}
	for _, header := range cli.ExtraOriginHeaders {
		parts := strings.Split(header, "=")
		if len(parts) != 2 {
			log.Fatal("Failed to parse extra header", header)
		}
		extraHeaders[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	handler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s", r.Method, r.URL)
			r.Host = origin.Host
			for header, value := range extraHeaders {
				r.Header.Add(header, value)
			}
			p.ServeHTTP(w, r)
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(origin)
	if cli.DumpRequests {
		proxy.Transport = &dumptransport.Transport{}
	}

	http.HandleFunc("/", handler(proxy))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cli.ListenPort), nil))
}
