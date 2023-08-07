package dumptransport

import (
	"log"
	"net/http"
	"net/http/httputil"
)

type Transport struct {
}

func (p *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqDump, _ := httputil.DumpRequestOut(req, true)
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		log.Printf("REQUEST\n%q\nRESPONSE\nCould not be loaded as of error %q", reqDump, err)
		return nil, err
	}
	respDump, _ := httputil.DumpResponse(resp, true)
	log.Printf("REQUEST\n%q\nRESPONSE\n%q", reqDump, respDump)
	return resp, err
}
