package dumptransport

import (
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

type Transport struct {
}

func (p *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqDump, _ := httputil.DumpRequestOut(req, true)
	requestStart := time.Now()
	resp, err := http.DefaultTransport.RoundTrip(req)
	requestDuration := time.Since(requestStart)
	if err != nil {
		log.Printf("%s %s\n\nDURATION %s\n\nFULL REQUEST\n%s\n\nFULL RESPONSE\nCould not be loaded as of error %q", req.Method, req.URL, requestDuration, reqDump, err)
		return nil, err
	}
	respDump, _ := httputil.DumpResponse(resp, true)
	log.Printf("%s %s\n\nDURATION %s\n\nFULL REQUEST\n%s\n\nFULL RESPONSE%s", req.Method, req.URL, requestDuration, reqDump, respDump)
	return resp, err
}
