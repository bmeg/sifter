package extractors

import (
	"context"
	"io"
	"log"
	"sync"

	"github.com/bmeg/grip/gripper"
	"github.com/bmeg/grip/util/rpc"

	"github.com/bmeg/sifter/manager"
	"github.com/bmeg/sifter/transform"
)

type GripperLoadStep struct {
	Host       string         `json:"host" jsonschema_description:"GRIPPER URL"`
	Collection string         `json:"collection" jsonschema_description:"GRIPPER collection to target"`
	Transform  transform.Pipe `json:"transform" jsonschema_description:"The transform pipeline to run"`
}

func (ml *GripperLoadStep) Run(task *manager.Task) error {
	log.Printf("Starting GRIPPER Load")

	rpcConf := rpc.ConfigWithDefaults(ml.Host)
	log.Printf("Connecting to %s", ml.Host)
	conn, err := rpc.Dial(context.Background(), rpcConf)
	if err != nil {
		log.Printf("RPC Connection error: %s", err)
		return err
	}
	client := gripper.NewGRIPSourceClient(conn)

	procChan := make(chan map[string]interface{}, 100)
	wg := &sync.WaitGroup{}

	ml.Transform.Init(task)
	out, err := ml.Transform.Start(procChan, task, wg)
	if err != nil {
		return err
	}
	go func() {
		//we don't do anything with the transform output. So just read it and
		//toss it
		for range out {
		}
	}()

	req := gripper.Collection{Name: ml.Collection}
	log.Printf("Loading: '%s'", ml.Collection)
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
		log.Printf("Error Getting rows: %s", err)
	}

	log.Printf("Done Loading")
	close(procChan)
	wg.Wait()
	ml.Transform.Close()

	return nil
}
