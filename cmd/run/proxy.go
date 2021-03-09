package run

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type LoadProxyServer struct {
	DestURL    string
	waitScreen bool
	group      *sync.WaitGroup
	proxy      *httputil.ReverseProxy
}

func NewLoadProxyServer(proxyURL string) *LoadProxyServer {
	rpURL, _ := url.Parse(proxyURL)

	return &LoadProxyServer{DestURL: proxyURL, group: &sync.WaitGroup{}, waitScreen: true, proxy: httputil.NewSingleHostReverseProxy(rpURL)}
}

func (lp *LoadProxyServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if lp.waitScreen {
		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := []byte(`<html><meta http-equiv="refresh" content="5" />Sifter Loading Data</html>`)
		res.Write(data)
	} else {
		lp.proxy.ServeHTTP(res, req)
	}
}

func (lp *LoadProxyServer) Start() error {
	// create a new handler
	lp.group.Add(1)
	go func() {
		http.ListenAndServe(":9999", lp)
		lp.group.Done()
	}()
	return nil
}

func (lp *LoadProxyServer) StartProxy() error {
	lp.waitScreen = false
	lp.group.Wait()
	return nil
}
