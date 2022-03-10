package playbook

import (
	"log"
	"os"
	"path/filepath"

	"github.com/bmeg/flame"
	"github.com/bmeg/sifter/loader"
	"github.com/bmeg/sifter/task"
	"github.com/bmeg/sifter/writers"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (pb *Playbook) PrepInputs(inputs map[string]any, workdir string) map[string]any {

	workdir, _ = filepath.Abs(workdir)

	out := map[string]any{}

	//fill in missing values with default values
	for k, v := range pb.Inputs {
		if _, ok := inputs[k]; !ok {
			if v.Default != "" {
				if v.IsFile() || v.IsDir() {
					defaultPath := filepath.Join(filepath.Dir(pb.path), v.Default)
					out[k], _ = filepath.Abs(defaultPath)
				} else {
					out[k] = v.Default
				}
			}
		} else {
			if v.IsFile() || v.IsDir() {
				if i, ok := inputs[k]; ok {
					if iStr, ok := i.(string); ok {
						newPath := filepath.Join(workdir, iStr)
						out[k], _ = filepath.Abs(newPath)
					}
				}
			} else {
				if i, ok := inputs[k]; ok {
					out[k] = i
				}
			}
		}
	}
	return out
}

func (pb *Playbook) Execute(man *Manager, inputs map[string]interface{}, workDir string, outDir string) error {

	log.Printf("Running playbook")
	log.Printf("Inputs: %#v", inputs)

	wf := flame.NewWorkflow()

	ld := loader.NewDirLoader(outDir)
	em, _ := ld.NewDataEmitter()

	if pb.Name == "" {
		pb.Name = "sifter"
	}

	task := &task.Task{Name: pb.Name, Inputs: inputs, Workdir: workDir, Emitter: em}

	outNodes := map[string]flame.Emitter[map[string]any]{}
	inNodes := map[string]flame.Receiver[map[string]any]{}
	writers := map[string]writers.WriteProcess{}

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

	for k, v := range pb.Sinks {
		sub := task.SubTask(k)
		s, err := v.Init(sub)
		if err == nil {
			c := flame.AddSink(wf, s.Write)
			inNodes[k] = c
			writers[k] = s
		}
	}

	for dst, src := range pb.Links {
		if srcNode, ok := outNodes[src]; ok {
			if dstNode, ok := inNodes[dst]; ok {
				log.Printf("Connecting %s (%T) to %s (%T)", src, srcNode, dst, dstNode)
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

	for k := range writers {
		writers[k].Close()
	}

	em.Close()
	return nil
}
