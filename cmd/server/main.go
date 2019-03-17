package server

import (
	"log"

	"github.com/bmeg/sifter/restapi"
	"github.com/bmeg/sifter/restapi/operations"
	"github.com/spf13/cobra"

	"net/http"
	"strings"

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

		api := operations.NewSifterAPI(swaggerSpec)

		//restapi.StaticDir = webDir
		server := restapi.NewServer(api)
		// set the port this service will be run on
		server.Port = port

		api.PostManifestHandler = operations.PostManifestHandlerFunc(
			func(params operations.PostManifestParams) middleware.Responder {
				log.Printf("Manifest Posted")
				return operations.NewPostManifestOK()
			})

		api.GetStatusHandler = operations.GetStatusHandlerFunc(
			func(params operations.GetStatusParams) middleware.Responder {
				log.Printf("Status requested")
				return operations.NewGetStatusOK()
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
