package server

import (
	"fmt"
	"log"

	"github.com/bmeg/sifter/manager"
	"github.com/bmeg/sifter/restapi"
	"github.com/bmeg/sifter/restapi/operations"
	"github.com/spf13/cobra"

	"net/http"
	"strings"

	"github.com/bmeg/sifter/playbook"

	loads "github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
)

var webDir string = "./static"
var port int = 8090

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

		man, err := manager.Init([]string{})
		if err != nil {
			log.Fatalln(err)
		}

		api := operations.NewSifterAPI(swaggerSpec)

		//restapi.StaticDir = webDir
		server := restapi.NewServer(api)
		// set the port this service will be run on
		server.Port = port

		api.PostManifestHandler = operations.PostManifestHandlerFunc(
			func(params operations.PostManifestParams) middleware.Responder {
				fmt.Printf("Manifest Posted:\n%s", params.Manifest)
				pbTxt := []byte(params.Manifest)
				pb := playbook.Playbook{}
				if err := playbook.Parse(pbTxt, &pb); err != nil {
					log.Printf("Parse Error: %s", err)
				}
				return operations.NewPostManifestOK()
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
		server.ConfigureAPI()

		origHandler := server.GetHandler()
		server.SetHandler(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, "/api") {
					origHandler.ServeHTTP(w, r)
				} else {
					http.FileServer(http.Dir(webDir)).ServeHTTP(w, r)
				}
			}),
		)

		// serve API
		defer server.Shutdown()
		if err := server.Serve(); err != nil {
			log.Fatalln(err)
		}

		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVar(&webDir, "webdir", webDir, "Web Server Content Dir")
	flags.IntVar(&port, "port", port, "Server Port")
}
