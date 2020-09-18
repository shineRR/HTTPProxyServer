package main

import (
	"bufio"
	"fmt"
	"io"
	log "log"
	"net/http"
	"os"
	"time"
)

const PORT = ":1330"

type proxy struct {
}

var blockedSites []string

func getBlockedSites() {
	file, err := os.Open("./block/blacklist")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		blockedSites = append(blockedSites, scanner.Text())
	}
	fmt.Println(blockedSites)
}

func writeLogs(data string) {
	f, err := os.OpenFile("./logs/log", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err = f.WriteString(data + "\n"); err != nil {
		panic(err)
	}
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func isBlockedSite(requestedSite string) bool {
	for _, site := range blockedSites {
		if requestedSite == site {
			return true
		}
	}
	return false
}

func (p *proxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.Hostname())
	req.RequestURI = ""

	if isBlockedSite(req.URL.Hostname()) {
		data := time.Now().Format("2 Jan 2006 15:04:05") + " " + "403 Forbidden" + " " + req.URL.String();
		writeLogs(data)
		io.WriteString(wr, "403 Forbidden")
		return
	}
	log.Println(req.RemoteAddr, " ", req.Method, " ", req.URL)

	if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
		return
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(wr, "Server Error", http.StatusInternalServerError)
		log.Fatal("ServeHTTP:", err)
	}
	defer resp.Body.Close()

	logs := time.Now().Format("2 Jan 2006 15:04:05") + " " + resp.Status + " " + req.URL.String();
	log.Println(logs)
	writeLogs(logs)

	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)
	io.Copy(wr, resp.Body)
}

func main() {
	getBlockedSites()

	handler := &proxy{}

	log.Println("Starting proxy server on ", PORT)
	if err := http.ListenAndServe(PORT, handler); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}