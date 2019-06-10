package manager

import (
	"github.com/bmeg/sifter/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	"log"
)

func NewManagerStatusHandler(man *Manager) operations.GetStatusHandlerFunc {
	return operations.GetStatusHandlerFunc(
		func(params operations.GetStatusParams) middleware.Responder {
			log.Printf("Status requested")
			body := []*operations.GetStatusOKBodyItems0{}
			man.Runtimes.Range(func(key, value interface{}) bool {
				v := value.(*Runtime)
				item := &operations.GetStatusOKBodyItems0{
					Current:     v.GetCurrent(),
					EdgeCount:   v.GetEdgeCount(),
					StepNum:     v.GetStepNum(),
					StepTotal:   v.GetStepTotal(),
					VertexCount: v.GetVertexCount(),
				}
				body = append(body, item)
				return true
			})
			out := operations.NewGetStatusOK().WithPayload(body)
			return out
		})
}
