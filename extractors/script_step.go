package extractors

import (
  "fmt"
  "log"
  "os"
  "os/exec"
  "os/user"

	"github.com/bmeg/sifter/evaluate"
  "github.com/bmeg/sifter/pipeline"
)

type ScriptStep struct {
	DockerImage string   `json:"dockerImage" jsonschema_description:"Docker image the contains script environment"`
	Command     []string `json:"command" jsonschema_description:"Command line to be run"`
  Stdout      string   `json:stdout jsonschema_description:"File to capture stdout"`
}

func (ss *ScriptStep) Run(task *pipeline.Task) error {
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
	for _, i := range ss.Command {
		o, err := evaluate.ExpressionString(i, task.Inputs, nil)
		if err != nil {
			return err
		}
		command = append(command, o)
	}
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
