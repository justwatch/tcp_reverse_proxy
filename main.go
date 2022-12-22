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
	ListenPort         int      `kong:"default='9200',env='LISTEN_PORT',help='port to listen on'"`
	OriginAddress      string   `kong:"default='https://www.google.com/',env='ORIGIN_ADDRESS',help='upstream address to connect to. Can be IP or name, later one will be resolved'"`
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

	extraHeaders := map[string]string{}
	for _, header := range cli.ExtraOriginHeaders {
		parts := strings.Split(header, "=")
		if len(parts) != 2 {
			log.Fatal("Failed to parse extra header", header)
		}
		extraHeaders[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	targetQuery := targetUrl.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = targetUrl.Scheme
		req.URL.Host = targetUrl.Host
		req.URL.Path, req.URL.RawPath = joinURLPath(targetUrl, req.URL)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		for header, value := range extraHeaders {
			req.Header.Add(header, value)
		}
	}
	proxy := &httputil.ReverseProxy{Director: director}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cli.ListenPort), nil))
}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
