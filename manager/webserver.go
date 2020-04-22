package manager

import (
	"github.com/bmeg/sifter/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	"log"
	"github.com/bmeg/sifter/pipeline"
)

func NewManagerStatusHandler(man *Manager) operations.GetStatusHandlerFunc {
	return operations.GetStatusHandlerFunc(
		func(params operations.GetStatusParams) middleware.Responder {
			log.Printf("Status requested")
			body := []*operations.GetStatusOKBodyItems0{}
			man.Runtimes.Range(func(key, value interface{}) bool {
				v := value.(*pipeline.Runtime)
				item := &operations.GetStatusOKBodyItems0{
					Current:     v.GetCurrent(),
					StepNum:     v.GetStepNum(),
					StepTotal:   v.GetStepTotal(),
				}
				body = append(body, item)
				return true
			})
			out := operations.NewGetStatusOK().WithPayload(body)
			return out
		})
}
