package steps


import (
  "io"
  "log"
  "context"
  "sync"
  "github.com/bmeg/grip/dig"
  "github.com/bmeg/grip/util/rpc"
  "github.com/bmeg/grip/protoutil"

  "github.com/bmeg/sifter/transform"
  "github.com/bmeg/sifter/pipeline"
)


type DigLoadStep struct {
  Host          string                  `json:"host"`
	Collection    string                  `json:"collection"`
  Transform     transform.TransformPipe `json:"transform"`
}

func (ml *DigLoadStep) Run(task *pipeline.Task) error {
  log.Printf("Starting Dig Load")

  rpcConf := rpc.ConfigWithDefaults(ml.Host)
 	log.Printf("Connecting to %s", ml.Host)
 	conn, err := rpc.Dial(context.Background(), rpcConf)
 	if err != nil {
 		log.Printf("RPC Connection error: %s", err)
 		return err
 	}
 	client := dig.NewDigSourceClient(conn)

  procChan := make(chan map[string]interface{}, 100)
  wg := &sync.WaitGroup{}

  ml.Transform.Start( procChan, task, wg )

  req := dig.Collection{Name: ml.Collection}
  log.Printf("Loading: '%s'", ml.Collection)
  cl, err := client.GetRows(context.Background(), &req)
  if err == nil {
    for {
			t, err := cl.Recv()
			if err == io.EOF {
				break
			} else {
        o := protoutil.AsMap(t.Data)
        procChan <- o
      }
    }
  } else {
    log.Printf("Error Getting rows: %s", err)
  }

  log.Printf("Done Loading")
  close(procChan)
  wg.Wait()

	return nil
}
