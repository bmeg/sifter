
package manager

import (
  "log"
  "github.com/bmeg/sifter/restapi/operations"
  "github.com/go-openapi/runtime/middleware"
)

func NewManagerStatusHandler(man *Manager) operations.GetStatusHandlerFunc {
  return operations.GetStatusHandlerFunc(
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
}
