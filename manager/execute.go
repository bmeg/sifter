
package manager

import (
  "log"
)

func (pb *Playbook) Execute(man *Manager) error {

		for _, step := range pb.Steps {
			if step.MatrixLoad != nil {
				log.Printf("%s\n", step.Desc)
				elemStream := step.MatrixLoad.Load()
				for elem := range elemStream {
					log.Printf("%s", elem)
				}
			}
			if step.ManifestLoad != nil {
				log.Printf("Manifest %s\n", step.Desc)
				elemStream := step.ManifestLoad.Load(man.NewTask(map[string]interface{}{}))
				for elem := range elemStream {
					log.Printf("%s", elem)
				}
			}
		}

    return nil
}
