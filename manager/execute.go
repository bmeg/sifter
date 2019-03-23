package manager

func (pb *Playbook) Execute(man *Manager, inputs map[string]interface{}) error {
	run := man.NewRuntime()

	for _, step := range pb.Steps {
		if step.MatrixLoad != nil {
			task := run.NewTask(inputs)
			if err := step.MatrixLoad.Run(task); err != nil {
				return err
			}
		}
		if step.ManifestLoad != nil {
			task := run.NewTask(inputs)
			if err := step.ManifestLoad.Run(task); err != nil {
				return err
			}
		}
		if step.Download != nil {
			task := run.NewTask(inputs)
			if err := step.Download.Run(task); err != nil {
				return err
			}
		}
	}
	return nil
}
