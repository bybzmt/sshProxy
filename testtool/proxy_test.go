package main

import (
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

func main() {
	var p = flag.String("proxy", "socks5://127.0.0.1:1080", "socks5 proxy")
	var addr = flag.String("addr", ":8080", "listen on addr:port")
	flag.Parse()

	if *p == "" || *addr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	ur, err := url.Parse(*p)
	if err != nil {
		log.Println("Proxy Error", err)
		os.Exit(1)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy:        http.ProxyURL(ur),
			MaxIdleConns: 100,
		},
	}

	test_url := "http://" + *addr + "/word"

	http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		test(client, test_url, w)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	log.Fatal(http.ListenAndServe(*addr, nil))
}

func test(c *http.Client, u string, w io.Writer) {
	if resp, err := c.Get(u); err != nil {
		w.Write([]byte("Fail " + err.Error()))
	} else {
		defer resp.Body.Close()
		io.Copy(w, resp.Body)
	}
}
