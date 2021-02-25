package extractors

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"

	"path/filepath"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/manager"
	shellquote "github.com/kballard/go-shellquote"
)

type ScriptStep struct {
	DockerImage string   `json:"dockerImage" jsonschema_description:"Docker image the contains script environment"`
	Command     []string `json:"command" jsonschema_description:"Command line, written as an array, to be run"`
	CommandLine string   `json:"commandLine" jsonschema_description:"Command line to be run"`
	Stdout      string   `json:"stdout" jsonschema_description:"File to capture stdout"`
	WorkDir     string   `json:"workdir"`
}

func (ss *ScriptStep) Run(task *manager.Task) error {

	var baseCommand []string

	if len(ss.Command) > 0 {
		for _, i := range ss.Command {
			o, err := evaluate.ExpressionString(i, task.Inputs, nil)
			if err != nil {
				return err
			}
			baseCommand = append(baseCommand, o)
		}
	} else if len(ss.CommandLine) > 0 {
		c, err := evaluate.ExpressionString(ss.CommandLine, task.Inputs, nil)
		if err != nil {
			return err
		}
		csplit, err := shellquote.Split(c)
		if err != nil {
			return err
		}
		baseCommand = csplit
	} else {
		return fmt.Errorf("Command line not provided")
	}

	if ss.DockerImage != "" {
		u, err := user.Current()
		if err != nil {
			return err
		}
		volumeMapping := fmt.Sprintf("%s:/var/run/sifter", task.Workdir)

		command := []string{
			"run", "-u", u.Uid, "--rm",
			"-v", volumeMapping, "-w", "/var/run/sifter",
			ss.DockerImage,
		}

		command = append(command, baseCommand...)

		log.Printf("Exec docker %s", command)

		cmd := exec.Command("docker", command...)
		cmd.Stderr = os.Stderr
		if ss.Stdout != "" {
			p, _ := task.Path(ss.Stdout)
			outfile, _ := os.Create(p)
			cmd.Stdout = outfile
			defer outfile.Close()
		} else {
			cmd.Stdout = os.Stdout
		}
		err = cmd.Run()
		return err
	}

	prog := baseCommand[0]
	cmd := exec.Command(prog, baseCommand[1:]...)
	if ss.WorkDir != "" {
		workDir, err := evaluate.ExpressionString(ss.WorkDir, task.Inputs, nil)
		if err == nil {
			cmd.Dir = workDir
		}
	} else {
		baseDir := filepath.Dir(task.SourcePath)
		cmd.Dir = baseDir
	}
	cmd.Stderr = os.Stderr
	if ss.Stdout != "" {
		p, _ := task.Path(ss.Stdout)
		outfile, _ := os.Create(p)
		cmd.Stdout = outfile
		defer outfile.Close()
	} else {
		cmd.Stdout = os.Stdout
	}
	err := cmd.Run()
	return err
}
