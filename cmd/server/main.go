package server

import (
	"fmt"
	"log"

	"github.com/bmeg/sifter/manager"
	"github.com/bmeg/sifter/restapi"
	"github.com/bmeg/sifter/restapi/operations"
	"github.com/spf13/cobra"

	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	loads "github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
)

var webDir string = "./static"
var playbookDir string = "./playbooks"
var port int = 8090
var proxy string = ""

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "server",
	Short: "Start web based server",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("Starting server")

		man, err := manager.Init(playbookDir)
		if err != nil {
			log.Fatalln(err)
		}

		api := operations.NewSifterAPI(swaggerSpec)

		//restapi.StaticDir = webDir
		server := restapi.NewServer(api)
		// set the port this service will be run on
		server.Port = port

		api.PostPlaybookHandler = operations.PostPlaybookHandlerFunc(
			func(params operations.PostPlaybookParams) middleware.Responder {
				fmt.Printf("Playbook Posted:\n%s", params.Manifest)
				pbTxt := []byte(params.Manifest)
				pb := manager.Playbook{}
				if err := manager.Parse(pbTxt, &pb); err != nil {
					log.Printf("Parse Error: %s", err)
				}
				return operations.NewPostPlaybookOK()
			})

		api.GetPlaybookHandler = operations.GetPlaybookHandlerFunc(
			func(params operations.GetPlaybookParams) middleware.Responder {
				body := []*operations.GetPlaybookOKBodyItems0{}
				for _, i := range man.GetPlaybooks() {
					body = append(body, &operations.GetPlaybookOKBodyItems0{Name: i.Name})
				}
				out := operations.NewGetPlaybookOK().WithPayload(body)
				return out
			})

		api.GetStatusHandler = operations.GetStatusHandlerFunc(
			func(params operations.GetStatusParams) middleware.Responder {
				log.Printf("Status requested")
				body := operations.GetStatusOKBody{
					Current:     man.GetCurrent(),
					EdgeCount:   man.GetEdgeCount(),
					StepNum:     man.GetStepNum(),
					StepTotal:   man.GetStepTotal(),
					VertexCount: man.GetVertexCount(),
				}
				out := operations.NewGetStatusOK().WithPayload(&body)
				return out
			})
		api.PostPlaybookIDGraphHandler = operations.PostPlaybookIDGraphHandlerFunc(
			func(params operations.PostPlaybookIDGraphParams) middleware.Responder {
				inputs := map[string]interface{}{}
				if err := manager.ParseDataString(params.Params, &inputs); err != nil {
					log.Printf("Error on input %s : %s", params.Params, err)
					//TODO: return error here
					out := operations.NewPostPlaybookIDGraphOK()
					return out
				}
				log.Printf("Starting import playbook: %s %s", params.ID, inputs)
				if pb, ok := man.GetPlaybook(params.ID); ok {
					go pb.Execute(man, params.Graph, inputs)
					out := operations.NewPostPlaybookIDGraphOK()
					return out
				}
				//TODO: return error here
				out := operations.NewPostPlaybookIDGraphOK()
				return out
			})
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
				} else if strings.HasPrefix(r.URL.Path, "/v1") {
					proxyHandler.ServeHTTP(w, r)
				} else {
					http.FileServer(http.Dir(webDir)).ServeHTTP(w, r)
				}
			}),
		)

		// serve API
		defer server.Shutdown()
		defer man.Close()
		if err := server.Serve(); err != nil {
			log.Fatalln(err)
		}
		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVar(&webDir, "web", webDir, "Web Server Content Dir")
	flags.StringVar(&playbookDir, "playbooks", playbookDir, "Playbook Dir")
	flags.StringVar(&proxy, "proxy", proxy, "Proxy")
	flags.IntVar(&port, "port", port, "Server Port")
}
