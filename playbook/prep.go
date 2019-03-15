package playbook

import (
	"log"
	"os/exec"
	"path"

	"github.com/bmeg/sifter/manager"
)

func (ps *PrepStep) Run(man manager.Manager) error {
	if ps.ArgsCopy != "" {
		dstPath := path.Join(man.Workdir, ps.ArgsCopy)
		log.Printf("Copy %s to %s", man.Args[0], dstPath)
		cpCmd := exec.Command("cp", "-rf", man.Args[0], dstPath)
		err := cpCmd.Run()
		return err
	}
	return nil
}
