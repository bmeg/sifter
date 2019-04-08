package manager

func (pb *Playbook) Execute(man *Manager, graph string, inputs map[string]interface{}) error {
	run, err := man.NewRuntime(graph)
	run.Printf("Starting Playbook")
	defer run.Close()
	defer run.Printf("Playbook done")
	if err != nil {
		return err
	}

	for _, step := range pb.Steps {
		if step.MatrixLoad != nil {
			task := run.NewTask(inputs)
			if err := step.MatrixLoad.Run(task); err != nil {
				run.Printf("Load Error: %s", err)
				return err
			}
		} else if step.ManifestLoad != nil {
			task := run.NewTask(inputs)
			if err := step.ManifestLoad.Run(task); err != nil {
				run.Printf("Load Error: %s", err)
				return err
			}
		} else if step.Download != nil {
			task := run.NewTask(inputs)
			if err := step.Download.Run(task); err != nil {
				run.Printf("Load Error: %s", err)
				return err
			}
		} else if step.Untar != nil {
			task := run.NewTask(inputs)
			if err := step.Untar.Run(task); err != nil {
				run.Printf("Untar Error: %s", err)
				return err
			}
		}
	}
	return nil
}
