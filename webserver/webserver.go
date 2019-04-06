
package webserver

import (
  "log"
  	loads "github.com/go-openapi/loads"
  	//"github.com/go-openapi/runtime/middleware"
    "github.com/bmeg/sifter/restapi/operations"
    "github.com/bmeg/sifter/restapi"

    "net/http"
  	"net/http/httputil"
  	"net/url"
  	"strings"

)


type WebServerHandler struct {
  PostPlaybookHandler operations.PostPlaybookHandlerFunc
  GetPlaybookHandler operations.GetPlaybookHandlerFunc
  GetStatusHandler operations.GetStatusHandlerFunc
  PostPlaybookIDGraphHandler operations.PostPlaybookIDGraphHandlerFunc
  FileHandler http.Handler
}


func RunServer(handler WebServerHandler, port int, proxy string) (error){
  swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
  if err != nil {
    log.Fatalln(err)
    return err
  }
  api := operations.NewSifterAPI(swaggerSpec)

  //restapi.StaticDir = webDir
  server := restapi.NewServer(api)
  // set the port this service will be run on
  server.Port = port

  api.PostPlaybookHandler = handler.PostPlaybookHandler
  api.GetPlaybookHandler = handler.GetPlaybookHandler
  api.GetStatusHandler = handler.GetStatusHandler
  api.PostPlaybookIDGraphHandler = handler.PostPlaybookIDGraphHandler

  server.ConfigureAPI()

	var proxyHandler http.Handler = nil
	if proxy != "" {
		u, err := url.Parse(proxy)
		if err != nil {
			log.Printf("Base Proxy Address")
			return err
		}
		proxyHandler = httputil.NewSingleHostReverseProxy(u)
	}

	origHandler := server.GetHandler()
	server.SetHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				origHandler.ServeHTTP(w, r)
			} else if (proxyHandler != nil && strings.HasPrefix(r.URL.Path, "/v1")) {
				proxyHandler.ServeHTTP(w, r)
			} else if (handler.FileHandler != nil) {
				handler.FileHandler.ServeHTTP(w, r)
			}
		}),
	)
  // serve API
  defer server.Shutdown()
  if err := server.Serve(); err != nil {
    log.Fatalln(err)
  }
  return nil
}
