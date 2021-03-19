//+build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/shabbyrobe/golib/httptools/curl"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	rt := curl.Transport{}
	hc := http.Client{
		Transport: &rt,
	}

	rq, err := http.NewRequest("GET", os.Args[1], nil)
	if err != nil {
		return err
	}
	rq.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.90 Safari/537.36")

	rs, err := hc.Do(rq)
	if err != nil {
		return err
	}
	defer rs.Body.Close()

	for k, vs := range rs.Header {
		fmt.Println("header", k, vs)
	}
	fmt.Println("content length", rs.ContentLength)
	fmt.Println("compressed", rs.Uncompressed)
	fmt.Println("transfer encoding", rs.TransferEncoding)

	body, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return err
	}
	if err := rs.Body.Close(); err != nil {
		return err
	}
	rs.Body.Close()

	fmt.Println(string(body))
	return nil
}
