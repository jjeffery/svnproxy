package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"
)

var (
	debug      bool
	help       bool
	listenPort int
	rpAddr     = "https://github.com"
	rpURL      *url.URL
)

func main() {
	flag.IntVar(&listenPort, "port", 8080, "listen tcp port number")
	flag.BoolVar(&debug, "debug", false, "log debug messages")
	flag.BoolVar(&help, "help", false, "show this usage text")
	flag.Parse()

	if help || len(flag.Args()) > 1 {
		_, _ = fmt.Fprintln(os.Stderr, "usage svnproxy [options] [ <url> ]")
		_, _ = fmt.Fprintln(os.Stderr, "options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if len(flag.Args()) == 1 {
		rpAddr = flag.Arg(0)
	}

	if err := doMain(); err != nil {
		log.Fatalln("error:", err)
	}
}

func doMain() error {
	rp, err := newRP()
	if err != nil {
		return err
	}

	listenAddr := ":" + strconv.Itoa(listenPort)

	log.Println("listening on", listenAddr)
	return http.ListenAndServe(listenAddr, rp)
}

func newRP() (http.Handler, error) {
	u, err := url.Parse(rpAddr)
	if err != nil {
		return nil, err
	}

	if u.Path != "/" {
		if u.Path != "" {
			log.Println(`warning: ignoring path specification: using "/"`)
		}
		u.Path = "/"
	}
	if u.RawQuery != "" {
		log.Println("warning: ignoring query string")
		u.RawQuery = ""
	}
	if u.Fragment != "" {
		log.Println("warning: ignoring fragment")
		u.Fragment = ""
	}

	// remember URL
	rpURL = u
	rpAddr = u.String()
	log.Println("remote address:", rpAddr)

	rp := &httputil.ReverseProxy{
		Rewrite:        rewrite,
		ModifyResponse: modifyResponse,
		ErrorLog:       log.New(os.Stderr, "proxy: error: ", 0),
		FlushInterval:  time.Second * 10,
	}

	return rp, nil
}

func rewrite(request *httputil.ProxyRequest) {
	request.SetURL(rpURL)
	if debug {
		r := request.In
		log.Printf("%s %s %s", r.Method, r.URL, r.Proto)
		for key, values := range r.Header {
			for _, value := range values {
				log.Printf("%s: %s", key, value)
			}
		}
		log.Println()
	}
}

func modifyResponse(w *http.Response) error {
	if debug {
		log.Printf("%d %s", w.StatusCode, w.Status)
		for key, values := range w.Header {
			for _, value := range values {
				log.Printf("%s: %s", key, value)
			}
		}
		log.Println()
	}
	return nil
}
