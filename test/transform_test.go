package test

//"github.com/bmeg/sifter/loader"

type DebugEmitter struct {
}

func (d *DebugEmitter) Emit(name string, e map[string]interface{}) error {
	return nil
}

func (d *DebugEmitter) Close() {}

func (d *DebugEmitter) EmitObject(prefix string, objClass string, e map[string]interface{}) error {
	return nil
}

/*
func TestPipeline(t *testing.T) {

	testPipe := transform.Pipe{
		transform.Step{
			Project: &transform.ProjectStep{
				Mapping: map[string]interface{}{
					"gid": "{{row._id}}",
				},
			},
		},
	}

	inData := []map[string]interface{}{
		{
			"_id": "1",
		},
		{
			"_id": "2",
		},
	}

	outData := []map[string]interface{}{
		{
			"gid": "1",
		},
		{
			"gid": "2",
		},
	}

	dem := &DebugEmitter{}

	//run := manager.NewRuntime(dem, "./", "test", nil)

	inputs := map[string]interface{}{}
	task := run.NewTask("./", inputs)

	testPipe.Init(task)

	wg := &sync.WaitGroup{}

	inStream := make(chan map[string]interface{}, 10)
	outSteam, err := testPipe.Start(inStream, task, wg)
	if err != nil {
		t.Error(err)
	}

	go func() {
		for _, d := range inData {
			inStream <- d
		}
		close(inStream)
	}()

	count := 0
	for range outSteam {
		count++
	}

	if count != len(outData) {
		t.Errorf("Mismatch count: %d != %d", count, len(outData))
	}

}
*/
