package playbook

import (
	"log"
	"os"

	"github.com/bmeg/flame"
	"github.com/bmeg/sifter/loader"
	"github.com/bmeg/sifter/task"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

/*
func (pb *Playbook) PrepInputs() {
	workDir, _ = filepath.Abs(workDir)
	outDir, _ = filepath.Abs(outDir)

	for k, v := range pb.Inputs {
		if _, ok := inputs[k]; !ok {
			if v.Default != "" {
				if (v.Type == "File" || v.Type == "Directory") && !download.IsURL(v.Default) {
					log.Printf("Setting input: %s %s", filepath.Dir(pb.path), v.Default)
					defaultPath := filepath.Join(filepath.Dir(pb.path), v.Default)
					inputs[k], _ = filepath.Abs(defaultPath)
				} else {
					inputs[k] = v.Default
				}
			} else if v.Type == "CWD" {
				path, err := os.Getwd()
				if err == nil {
					inputs[k] = path
				}
			} else if v.Type == "OUTPUT_DIR" {
				log.Printf("Setting %s to %s", k, outDir)
				inputs[k] = outDir
			}
		}
	}

	run, err := man.NewRuntime(pb.Name, workDir)
	for k, i := range pb.Inputs {
		if v, ok := inputs[k]; ok {
			if i.Type == "File" || i.Type == "Directory" {
				path := v.(string)
				if download.IsURL(path) {
					log.Printf("Found a URL to download: %s", path)
					tmpTask := run.NewTask(pb.path, map[string]interface{}{})
					dstPath, _ := tmpTask.AbsPath(filepath.Base(path))
					newPath, err := download.ToFile(path, dstPath)
					if err != nil {
						log.Printf("Download Error: %s", err)
						return err
					}
					inputs[k] = newPath
				} else {
					p, _ := filepath.Abs(path)
					if fileExists(p) {
						log.Printf("Using file: %s", p)
						inputs[k] = p
					} else {
						if i.Source != "" {
							newPath, err := download.ToFile(i.Source, p)
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

}
*/

/*

func prepStepLog() {
	log.Printf("Playbook executing in %s", workDir)
	log.Printf("Output to %s", outDir)
	stepFile := path.Join(workDir, ".sifter_steps")

	startStep := 0
	content, err := ioutil.ReadFile(stepFile)
	if err == nil {
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			log.Printf("Line: %s", line)
			if line == "OK" {
				startStep = i + 1
			}
		}
	}

	f, err := os.OpenFile(stepFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	log.Printf("StartStep: %d", startStep)
}
*/

func (pb *Playbook) Execute(man *Manager, inputs map[string]interface{}, workDir string, outDir string) error {

	log.Printf("Running playbook")
	log.Printf("Inputs: %#v", inputs)

	wf := flame.NewWorkflow()

	ld := loader.NewDirLoader(outDir)
	em, _ := ld.NewDataEmitter(nil)

	if pb.Name == "" {
		pb.Name = "sifter"
	}

	task := &task.Task{Name: pb.Name, Inputs: inputs, Workdir: workDir, Emitter: em}

	outNodes := map[string]flame.Emitter[map[string]any]{}
	inNodes := map[string]flame.Receiver[map[string]any]{}

	for n, v := range pb.Sources {
		log.Printf("Setting up %s", n)
		s, err := v.Start(task)
		if err == nil {
			c := flame.AddSourceChan(wf, s)
			outNodes[n] = c
		} else {
			log.Printf("Source error: %s", err)
		}
	}

	for k, v := range pb.Pipelines {
		sub := task.SubTask(k)
		var lastStep flame.Emitter[map[string]any]
		var firstStep flame.Receiver[map[string]any]
		for _, s := range v {
			b, err := s.Init(sub)
			if err != nil {
				log.Printf("Pipeline error: %s", err)
			} else {
				log.Printf("Pipeline %s step: %T", k, b)
				c := flame.AddFlatMapper(wf, b.Process)
				if lastStep != nil {
					c.Connect(lastStep)
				}
				lastStep = c
				if firstStep == nil {
					firstStep = c
				}
			}
		}
		outNodes[k] = lastStep
		inNodes[k] = firstStep
	}

	for dst, src := range pb.Links {
		if srcNode, ok := outNodes[src]; ok {
			if dstNode, ok := inNodes[dst]; ok {
				log.Printf("Connecting %s (%#v) to %s (%#v)", src, srcNode, dst, dstNode)
				dstNode.Connect(srcNode)
			} else {
				log.Printf("Dest %s not found", dst)
			}
		} else {
			log.Printf("Source %s not found", src)
		}
	}

	log.Printf("WF: %#v", wf)

	wf.Start()
	log.Printf("Workflow Started")

	wf.Wait()

	em.Close()
	return nil
}
