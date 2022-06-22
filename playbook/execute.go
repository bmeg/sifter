package playbook

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bmeg/flame"
	"github.com/bmeg/sifter/task"
	"github.com/bmeg/sifter/transform"
	"github.com/bmeg/sifter/writers"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (pb *Playbook) PrepConfig(inputs map[string]any, workdir string) map[string]any {

	workdir, _ = filepath.Abs(workdir)

	out := map[string]any{}

	//fill in missing values with default values
	for k, v := range pb.Config {
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

type reduceWrapper struct {
	reducer transform.ReduceProcessor
}

func (rw *reduceWrapper) addKeyValue(x map[string]any) flame.KeyValue[string, map[string]any] {
	return flame.KeyValue[string, map[string]any]{Key: rw.reducer.GetKey(x), Value: x}
}

func (rw *reduceWrapper) removeKeyValue(x flame.KeyValue[string, map[string]any]) []map[string]any {
	return []map[string]any{x.Value}
}

type accumulateWrapper struct {
	accumulator transform.AccumulateProcessor
}

func (rw *accumulateWrapper) addKeyValue(x map[string]any) flame.KeyValue[string, map[string]any] {
	return flame.KeyValue[string, map[string]any]{Key: rw.accumulator.GetKey(x), Value: x}
}

func (rw *accumulateWrapper) removeKeyValue(x flame.KeyValue[string, map[string]any]) []map[string]any {
	return []map[string]any{x.Value}
}

func (pb *Playbook) Execute(task task.RuntimeTask) error {
	log.Printf("Running playbook")

	log.Printf("Inputs: %#v", task.GetConfig())

	wf := flame.NewWorkflow()

	if pb.Name == "" {
		pb.Name = "sifter"
	}
	task.SetName(pb.Name)

	outNodes := map[string]flame.Emitter[map[string]any]{}
	inNodes := map[string]flame.Receiver[map[string]any]{}
	writers := map[string]writers.WriteProcess{}

	for n, v := range pb.Inputs {
		log.Printf("Setting up %s", n)
		s, err := v.Start(task)
		if err == nil {
			c := flame.AddSourceChan(wf, s)
			outNodes[n] = c
		} else {
			log.Printf("Source error: %s", err)
			return err
		}
	}

	for k, v := range pb.Pipelines {
		sub := task.SubTask(k)
		var lastStep flame.Emitter[map[string]any]
		var firstStep flame.Receiver[map[string]any]
		for i, s := range v {
			b, err := s.Init(sub)
			if err != nil {
				log.Printf("Pipeline %s error: %s", k, err)
				return err
			} else {
				if mProcess, ok := b.(transform.MapProcessor); ok {
					log.Printf("Pipeline %s step %d: %T", k, i, b)
					c := flame.AddFlatMapper(wf, mProcess.Process)
					if lastStep != nil {
						c.Connect(lastStep)
					}
					if c != nil {
						lastStep = c
						if firstStep == nil {
							firstStep = c
						}
					} else {
						log.Printf("Error setting up step")
						//throw error?
					}
				} else if rProcess, ok := b.(transform.ReduceProcessor); ok {
					log.Printf("Pipeline reduce %s step %d: %T", k, i, b)
					wrap := reduceWrapper{rProcess}
					k := flame.AddMapper(wf, wrap.addKeyValue)
					r := flame.AddReduceKey(wf, rProcess.Reduce, rProcess.GetInit())
					c := flame.AddFlatMapper(wf, wrap.removeKeyValue)
					if lastStep != nil {
						k.Connect(lastStep)
					}
					r.Connect(k)
					c.Connect(r)
					lastStep = c
					if firstStep == nil {
						firstStep = k
					}
				} else if rProcess, ok := b.(transform.AccumulateProcessor); ok {
					log.Printf("Pipeline accumulate %s step %d: %T", k, i, b)

					wrap := accumulateWrapper{rProcess}
					k := flame.AddMapper(wf, wrap.addKeyValue)
					r := flame.AddAccumulate(wf, rProcess.Accumulate)
					c := flame.AddFlatMapper(wf, wrap.removeKeyValue)
					if lastStep != nil {
						k.Connect(lastStep)
					}
					r.Connect(k)
					c.Connect(r)
					lastStep = c
					if firstStep == nil {
						firstStep = k
					}

				} else {
					log.Printf("Unknown processor type")
				}
			}
		}
		outNodes[k] = lastStep
		inNodes[k] = firstStep
	}

	for k, v := range pb.Outputs {
		sub := task.SubTask(k)
		s, err := v.Init(sub)
		if err == nil {
			c := flame.AddSink(wf, s.Write)
			inNodes[k] = c
			writers[k] = s
		} else {
			log.Printf("output error: %s", err)
		}
	}

	for dst, p := range pb.Pipelines {
		if len(p) > 0 {
			if p[0].From != nil {
				src := string(*p[0].From)
				if src == dst {
					//TODO: more loop detection
					log.Printf("Pipeline Loop detected in %s", dst)
					return fmt.Errorf("Pipeline Loop detected")
				}
				if srcNode, ok := outNodes[src]; ok {
					if dstNode, ok := inNodes[dst]; ok {
						log.Printf("Connecting %s to %s ", src, dst)
						dstNode.Connect(srcNode)
					} else {
						log.Printf("Dest %s not found", dst)
					}
				} else {
					log.Printf("%s source %s not found", dst, src)
				}
			} else {
				log.Printf("First step of pipelines %s not 'from'", dst)
				return fmt.Errorf("First step of pipelines %s not 'from'", dst)
			}
		} else {
			log.Printf("Pipeline %s is empty", dst)
		}
	}

	for dst, v := range pb.Outputs {
		src := v.From()
		if src == dst {
			//TODO: more loop detection
			log.Printf("Pipeline Loop detected in %s", dst)
			return fmt.Errorf("Pipeline Loop detected")
		}
		if srcNode, ok := outNodes[src]; ok {
			if dstNode, ok := inNodes[dst]; ok {
				log.Printf("Connecting %s to %s ", src, dst)
				dstNode.Connect(srcNode)
			} else {
				log.Printf("Dest %s not found", dst)
			}
		} else {
			log.Printf("%s source %s not found", dst, src)
		}
	}

	//log.Printf("WF: %#v", wf)

	wf.Start()
	log.Printf("Workflow Started")

	wf.Wait()

	for k := range writers {
		writers[k].Close()
	}

	task.Close()
	return nil
}
