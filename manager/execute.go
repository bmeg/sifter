package manager

func (pb *Playbook) Execute(man *Manager, graph string, inputs map[string]interface{}) error {
	run, err := man.NewRuntime(graph)
	run.Printf("Starting Playbook")
	defer run.Close()
	defer run.Printf("Playbook done")
	if err != nil {
		return err
	}

	run.LoadSchema(pb.Schema)

	run.OutputCallback = func(name, value string) error {
		inputs[name] = value
		return nil
	}

	for _, step := range pb.Steps {
		step.Run(run, inputs)
	}
	return nil
}
