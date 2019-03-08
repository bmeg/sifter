package playbook

import (
  "log"
  "os/exec"
  "path"
  "github.com/bmeg/sifter/manager"
)

func (ps *PrepStep) Run(man manager.Manager) error {
  if ps.ArgsCopy != "" {
    log.Printf("Copy %s to %s", ps.ArgsCopy, man.Workdir)

    cpCmd := exec.Command("cp", "-rf", man.Args[0], path.Join(man.Workdir, ps.ArgsCopy ))
    err := cpCmd.Run()
    return err
  }
  return nil
}
