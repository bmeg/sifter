package server

import (
	"fmt"
	"log"

	"github.com/bmeg/sifter/webserver"

	"github.com/bmeg/sifter/manager"
	"github.com/bmeg/sifter/restapi/operations"
	"github.com/spf13/cobra"

	"github.com/go-openapi/runtime/middleware"
)

var webDir string = "./static"
var playbookDir string = "./playbooks"
var port int = 8090
var proxy string = ""
var workDir string = "./"
var gripServer string = "grip://localhost:8202"

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "server",
	Short: "Start web based server",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		log.Printf("Starting server")

		man, err := manager.Init(manager.Config{GripServer: gripServer, WorkDir: workDir, PlaybookDirs: []string{playbookDir}})
		if err != nil {
			log.Fatalln(err)
		}

		postPlaybookHandler := operations.PostPlaybookHandlerFunc(
			func(params operations.PostPlaybookParams) middleware.Responder {
				fmt.Printf("Playbook Posted:\n%s", params.Manifest)
				pbTxt := []byte(params.Manifest)
				pb := manager.Playbook{}
				if err := manager.Parse(pbTxt, &pb); err != nil {
					log.Printf("Parse Error: %s", err)
				}
				return operations.NewPostPlaybookOK()
			})

		getPlaybookHandler := operations.GetPlaybookHandlerFunc(
			func(params operations.GetPlaybookParams) middleware.Responder {
				body := []*operations.GetPlaybookOKBodyItems0{}
				for _, i := range man.GetPlaybooks() {
					body = append(body, &operations.GetPlaybookOKBodyItems0{Name: i.Name})
				}
				out := operations.NewGetPlaybookOK().WithPayload(body)
				return out
			})

		getStatusHandler := manager.NewManagerStatusHandler(man)

		postPlaybookIDGraphHandler := operations.PostPlaybookIDGraphHandlerFunc(
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

		conf := webserver.WebServerHandler{
			PostPlaybookHandler:        postPlaybookHandler,
			GetPlaybookHandler:         getPlaybookHandler,
			GetStatusHandler:           getStatusHandler,
			PostPlaybookIDGraphHandler: postPlaybookIDGraphHandler,
			FileHandler:                webserver.FileHandler(webDir),
		}

		webserver.RunServer(conf, port, proxy)

		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVar(&webDir, "web", webDir, "Web Server Content Dir")
	flags.StringVar(&playbookDir, "playbooks", playbookDir, "Playbook Dir")
	flags.StringVar(&proxy, "proxy", proxy, "Proxy")
	flags.StringVar(&gripServer, "server", gripServer, "GRIP Server")
	flags.StringVar(&workDir, "workdir", workDir, "Workdir")
	flags.IntVar(&port, "port", port, "Server Port")
}
