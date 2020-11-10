package main

import (
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

func proxyHttps(w http.ResponseWriter, r *http.Request) {
	upstream, err := net.DialTimeout("tcp", r.Host, 10*time.Second)

	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)

	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	downstream, _, err := hijacker.Hijack()

	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	go proxy(upstream, downstream)
	go proxy(downstream, upstream)
}

func proxyHttp(w http.ResponseWriter, req *http.Request) {
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func proxy(dst io.WriteCloser, src io.ReadCloser) {
	defer dst.Close()
	defer src.Close()
	io.Copy(dst, src)
}

func main() {
	var (
		listenAddr = flag.String("l", "localhost:8080", "Listen on this address")
	)
	flag.Parse()
	log.Printf("Started proxy on %s", *listenAddr)
	server := &http.Server{
		Addr: *listenAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				proxyHttps(w, r)
			} else {
				proxyHttp(w, r)
			}
		}),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	log.Fatal(server.ListenAndServe())

}
