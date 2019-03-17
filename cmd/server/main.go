package server

import (
	"log"

	"github.com/bmeg/sifter/restapi"
	"github.com/bmeg/sifter/restapi/operations"
	"github.com/spf13/cobra"

	loads "github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
)

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
		server := restapi.NewServer(api)
		defer server.Shutdown()

		// set the port this service will be run on
		server.Port = 9800

		// TODO: Set Handle

		api.PostManifestHandler = operations.PostManifestHandlerFunc(
			func(params operations.PostManifestParams) middleware.Responder {
				log.Printf("Manifest Posted")
				return operations.NewPostManifestOK()
			})

		// serve API
		if err := server.Serve(); err != nil {
			log.Fatalln(err)
		}

		return nil
	},
}
