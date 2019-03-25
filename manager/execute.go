package manager

func (pb *Playbook) Execute(man *Manager, graph string, inputs map[string]interface{}) error {
	man.Printf("Starting Playbook")
	run, err := man.NewRuntime(graph)
	defer run.Close()
	defer man.Printf("Playbook done")
	if err != nil {
		return err
	}
	for _, step := range pb.Steps {
		if step.MatrixLoad != nil {
			task := run.NewTask(inputs)
			if err := step.MatrixLoad.Run(task); err != nil {
				man.Printf("Load Error: %s", err)
				return err
			}
		}
		if step.ManifestLoad != nil {
			task := run.NewTask(inputs)
			if err := step.ManifestLoad.Run(task); err != nil {
				man.Printf("Load Error: %s", err)
				return err
			}
		}
		if step.Download != nil {
			task := run.NewTask(inputs)
			if err := step.Download.Run(task); err != nil {
				man.Printf("Load Error: %s", err)
				return err
			}
		}
	}
	return nil
}
