package playbook

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/shlex"
)

func (pb *Playbook) RunScript(name string) error {
	if sc, ok := pb.Scripts[name]; ok {
		path, _ := filepath.Abs(pb.path)
		workdir := filepath.Join(filepath.Dir(path), sc.Workdir)
		cmdLine, err := shlex.Split(sc.CommandLine)
		if err != nil {
			return err
		}
		cmd := exec.Command(cmdLine[0], cmdLine[1:len(cmdLine)]...)
		cmd.Dir = workdir
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		log.Printf("(%s) %s %s", cmd.Dir, cmd.Path, strings.Join(cmd.Args, " "))
		return cmd.Run()
	}
	return fmt.Errorf("Script %s not found", name)
}
