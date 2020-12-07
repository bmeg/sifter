package manager

import (
	"os"
	"io/ioutil"
	"strings"
	"log"
	"path"

	"path/filepath"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/schema"
)

func isURL(s string) bool {
	if strings.HasPrefix(s, "http://") { return true}
	if strings.HasPrefix(s, "https://") { return true}
	if strings.HasPrefix(s, "s3://") { return true}
	if strings.HasPrefix(s, "ftp://") { return true}
	return false
}

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

func (pb *Playbook) Execute(man *Manager, inputs map[string]interface{}, dir string) error {

	for k, v := range pb.Inputs {
		if _, ok := inputs[k]; !ok {
			if v.Default != "" {
				if (v.Type == "File" || v.Type == "Directory") && !isURL(v.Default) {
					log.Printf("Setting input: %s %s", filepath.Dir(pb.path), v.Default)
					defaultPath := filepath.Join(filepath.Dir(pb.path), v.Default)
					inputs[k], _ = filepath.Abs(defaultPath)
				} else {
					inputs[k] = v.Default
				}
			}
		}
	}

	if pb.Schema != "" {
		log.Printf("Schema eval inputs: %s %s", pb.Schema, inputs)
		schema, _ := evaluate.ExpressionString(pb.Schema, inputs, nil)
		if !filepath.IsAbs(schema) {
			schema = filepath.Join(filepath.Dir(pb.path), schema)
		}
		pb.Schema = schema
		log.Printf("Schema eval Path: %s", schema)
	}

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

	for k, i := range pb.Inputs {
		if v, ok := inputs[k]; ok {
			if i.Type == "File" || i.Type == "Directory" {
				path := v.(string)
				if isURL(path) {
					log.Printf("Found a URL to download: %s", path)
					tmpTask := run.NewTask(map[string]interface{}{})
					newPath, err := tmpTask.DownloadFile(path, filepath.Base(path))
					if err != nil {
						log.Printf("Download Error: %s", err)
						return err
					}
					inputs[k] = newPath
				} else {
					p, _ := filepath.Abs(path)
					if fileExists(p) {
						inputs[k] = p
					} else {
						if i.Source != "" {
							tmpTask := run.NewTask(map[string]interface{}{})
							newPath, err := tmpTask.DownloadFile(i.Source, p)
							if err != nil {
								log.Printf("Download Error: %s", err)
								return err
							}
							inputs[k] = newPath
						}
					}
				}
			}
		}
	}

	//run.Printf("Starting Playbook")
	//defer run.Printf("Playbook done")

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
			log.Printf("Running Playbook Step: %#v", step)
			err := step.Run(run, inputs)
			if err == nil {
				f.WriteString("OK\n")
				log.Printf("Playbook Step Done")
			} else {
				log.Printf("Playbook Step Error: %s", err)
				break
			}
		}
	}
	log.Printf("Done with steps")
	run.Close()

	return nil
}
