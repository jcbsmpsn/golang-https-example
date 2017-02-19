package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{},
		},
	}

	resp, err := client.Get("https://127.0.0.1:8443")
	if err != nil {
		log.Println(err)
		return
	}

	htmlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("%v\n", resp.Status)
	fmt.Printf(string(htmlData))
}
