package run

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type LoadProxyServer struct {
	port       int
	destURL    string
	waitScreen bool
	count      uint64
	group      *sync.WaitGroup
	proxy      *httputil.ReverseProxy
}

func NewLoadProxyServer(port int, proxyURL string) *LoadProxyServer {
	rpURL, _ := url.Parse(proxyURL)

	return &LoadProxyServer{port: port, destURL: proxyURL, group: &sync.WaitGroup{}, waitScreen: true, proxy: httputil.NewSingleHostReverseProxy(rpURL)}
}

func (lp *LoadProxyServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if lp.waitScreen {
		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		page := fmt.Sprintf(`<html><meta http-equiv="refresh" content="5" /><div style="height: 100%%; margin: auto; text-align: center;"><div>Sifter Loading Data</div><div>%d elements loaded</div></div></html>`, lp.count)
		data := []byte(page)
		res.Write(data)
	} else {
		lp.proxy.ServeHTTP(res, req)
	}
}

func (lp *LoadProxyServer) UpdateCount(count uint64) {
	lp.count = count
}

func (lp *LoadProxyServer) Start() error {
	// create a new handler
	lp.group.Add(1)
	go func() {
		s := fmt.Sprintf(":%d", lp.port)
		http.ListenAndServe(s, lp)
		lp.group.Done()
	}()
	return nil
}

func (lp *LoadProxyServer) StartProxy() error {
	lp.waitScreen = false
	lp.group.Wait()
	return nil
}
