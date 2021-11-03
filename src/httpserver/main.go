package main

import (
	"fmt"
	"github.com/golang/glog"
	"io"
	"log"
	"net"
	"net/http"
	"runtime"
	"strings"
)

func main() {
	glog.V(2).Info("Starting http server...")
	myHandler := http.HandlerFunc(rootHandle)
	http.Handle("/", WithLogging(myHandler))
	http.HandleFunc("/healthz", healthz)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func healthz(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "ok\n")
}

func rootHandle(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("version", runtime.Version())
	for k, v := range r.Header {
		io.WriteString(w, fmt.Sprintf("%s=%s\n", k, v))
	}
}

type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

func WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &StatusRecorder{
			ResponseWriter: w,
			Status:         200,
		}
		h.ServeHTTP(recorder, r)
		log.Printf("request status: %d, ipAddress: %s", recorder.Status, ClientPublicIP(r))
	})
}

func ClientPublicIP(r *http.Request) string {
	var ip string
	for _, ip = range strings.Split(r.Header.Get("X-Forwarded-For"), ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			return ip
		}
	}
	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}
