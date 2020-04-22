package manager

import (
	"os"
	"io/ioutil"
	"strings"
	"log"
	"path"

	"github.com/bmeg/sifter/schema"
)

func (pb *Playbook) Execute(man *Manager, inputs map[string]interface{}, dir string) error {
	var sc *schema.Schemas
	if pb.Schema != "" {
		log.Printf("Loading Schema: %s", pb.Schema)
		t, err := schema.Load(pb.Schema)
		if err != nil {
			log.Printf("Error: %s", err)
			return err
		}
		log.Printf("Loaded Schema: %s", t.GetClasses())
		sc = &t
	}
	run, err := man.NewRuntime(pb.Name, dir, sc)
	run.Printf("Starting Playbook")
	defer run.Close()
	defer run.Printf("Playbook done")
	if err != nil {
		return err
	}

	//run.LoadSchema(pb.Schema)

	run.OutputCallback = func(name, value string) error {
		inputs[name] = value
		return nil
	}

	log.Printf("Using %s", dir)
	stepFile := path.Join(dir, ".sifter_steps")

	startStep := 0
	content, err := ioutil.ReadFile(stepFile)
	if err == nil {
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			log.Printf("Line: %s", line)
			if line == "OK" {
				startStep = i+1
			}
		}
	}

	f, err := os.OpenFile(stepFile,	os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	log.Printf("StartStep: %d", startStep)

	for i, step := range pb.Steps {
		if i >= startStep {
			log.Printf("Running: %#v", step)
			err := step.Run(run, inputs)
			if err == nil {
				f.WriteString("OK\n")
			} else {
				break
			}
		}
	}
	return nil
}
