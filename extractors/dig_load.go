package extractors


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
  Host          string                  `json:"host" jsonschema_description:"DIG URL"`
	Collection    string                  `json:"collection" jsonschema_description:"DIG collection to target"`
  Transform     transform.TransformPipe `json:"transform" jsonschema_description:"The transform pipeline to run"`
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

  ml.Transform.Init(task)
  out, err := ml.Transform.Start( procChan, task, wg )
  if err != nil {
    return err
  }
  go func() {
    //we don't do anything with the transform output. So just read it and
    //toss it
    for range out {}
  }()

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
  ml.Transform.Close()

	return nil
}
