package manager

import (
	"log"
)

func (pb *Playbook) Execute(man *Manager, graph string, inputs map[string]interface{}) error {
	run, err := man.NewRuntime(graph)
	run.Printf("Starting Playbook")
	defer run.Close()
	defer run.Printf("Playbook done")
	if err != nil {
		return err
	}

	run.OutputCallback = func(name, value string) error {
		inputs[name] = value
		return nil
	}

	for _, step := range pb.Steps {
		if step.TransposeFile != nil {
			task := run.NewTask(inputs)
			if err := step.TransposeFile.Run(task); err != nil {
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
		} else if step.VCFLoad != nil {
			task := run.NewTask(inputs)
			if err := step.VCFLoad.Run(task); err != nil {
				run.Printf("VCF Load Error: %s", err)
				return err
			}
		} else if step.TableLoad != nil {
			task := run.NewTask(inputs)
			if err := step.TableLoad.Run(task); err != nil {
				run.Printf("Table Load Error: %s", err)
				return err
			}
		} else {
			log.Printf("Unknown Step")
		}
	}
	return nil
}
