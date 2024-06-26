package extractors

import (
	"context"
	"io"

	"github.com/bmeg/grip/gripper"
	"github.com/bmeg/grip/util/rpc"

	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/task"
)

type GripperLoadStep struct {
	Host       string `json:"host" jsonschema_description:"GRIPPER URL"`
	Collection string `json:"collection" jsonschema_description:"GRIPPER collection to target"`
}

func (ml *GripperLoadStep) Start(task.RuntimeTask) (chan map[string]interface{}, error) {
	logger.Debug("Starting GRIPPER Load")

	rpcConf := rpc.ConfigWithDefaults(ml.Host)
	logger.Debug("Connecting to %s", ml.Host)
	conn, err := rpc.Dial(context.Background(), rpcConf)
	if err != nil {
		logger.Error("RPC Connection error: %s", err)
		return nil, err
	}
	client := gripper.NewGRIPSourceClient(conn)

	procChan := make(chan map[string]interface{}, 100)
	go func() {
		req := gripper.Collection{Name: ml.Collection}
		logger.Debug("Loading: '%s'", ml.Collection)
		cl, err := client.GetRows(context.Background(), &req)
		if err == nil {
			for {
				t, err := cl.Recv()
				if err == io.EOF {
					break
				} else {
					o := t.Data.AsMap()
					procChan <- o
				}
			}
		} else {
			logger.Error("Error Getting rows: %s", err)
		}
		logger.Debug("Done Loading")
		close(procChan)
	}()

	return procChan, nil
}
