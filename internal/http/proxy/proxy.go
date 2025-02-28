package proxy

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// https://gist.github.com/yowu/f7dc34bd4736a65ff28d
// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

type Proxy struct {
	RemoteAddr string
}

func (p *Proxy) ProxyHandler(w http.ResponseWriter, req *http.Request) {
	log.Println(req.RemoteAddr, " ", req.Method, " ", req.URL, " ", req.URL.Scheme)

	parsedURL, err := url.Parse(p.RemoteAddr)
	if err != nil {
		http.Error(w, "Invalid remote WMS URL", http.StatusInternalServerError)
		return
	}

	delHopHeaders(req.Header)

	// Forward query parameters from the client request
	query := req.URL.RawQuery
	parsedURL.RawQuery = query

	//if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
	//	appendHostToXForwardHeader(req.Header, clientIP)
	//}

	resp, err := http.Get(parsedURL.String())
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		log.Fatal("ServeHTTP:", err)
	}
	defer resp.Body.Close()

	log.Println(req.RemoteAddr, " ", resp.Status)

	delHopHeaders(resp.Header)

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
