package playbook

import (
	"fmt"
	"path/filepath"

	"github.com/bmeg/flame"
	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/task"
	"github.com/bmeg/sifter/transform"
)

/*
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
*/

func (pb *Playbook) PrepConfig(inputs map[string]string, workdir string) (map[string]string, error) {
	workdir, _ = filepath.Abs(workdir)
	out := map[string]string{}
	for _, v := range pb.GetConfigFields() {
		if val, ok := inputs[v.Name]; ok {
			if v.IsFile() || v.IsDir() {
				defaultPath := filepath.Join(workdir, val)
				out[v.Name], _ = filepath.Abs(defaultPath)
			} else {
				out[v.Name] = val
			}
		} else if val, ok := pb.Config[v.Name]; ok {
			if val != nil {
				if v.IsFile() || v.IsDir() {
					defaultPath := filepath.Join(filepath.Dir(pb.path), *val)
					out[v.Name], _ = filepath.Abs(defaultPath)
				} else {
					out[v.Name] = *val
				}
			} else {
				return nil, fmt.Errorf("undefine parameter %s", v.Name)
			}
		} else {
			return nil, fmt.Errorf("config %s not defined", v.Name)
		}
	}
	return out, nil
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

type joinStruct struct {
	node *flame.JoinNode[map[string]any, map[string]any, map[string]any]
	proc transform.JoinProcessor
}

func (pb *Playbook) Execute(task task.RuntimeTask) error {
	logger.Debug("Running playbook")
	logger.Debug("Inputs", "config", task.GetConfig())

	wf := flame.NewWorkflow()

	if pb.Name == "" {
		pb.Name = "sifter"
	}
	task.SetName(pb.Name)

	outNodes := map[string]flame.Emitter[map[string]any]{}
	inNodes := map[string]flame.Receiver[map[string]any]{}

	for n, v := range pb.Inputs {
		logger.Debug("Setting Up", "name", n)
		s, err := v.Start(task)
		if err == nil {
			c := flame.AddSourceChan(wf, s)
			outNodes[n] = c
		} else {
			logger.Error("Source error", "error", err)
			return err
		}
	}

	procs := []transform.Processor{}
	joins := []joinStruct{}

	for k, v := range pb.Pipelines {
		sub := task.SubTask(k)
		var lastStep flame.Emitter[map[string]any]
		var firstStep flame.Receiver[map[string]any]
		for i, s := range v {
			b, err := s.Init(sub)
			if err != nil {
				logger.Error("Pipeline error", "name", k, "error", err)
				return err
			}

			procs = append(procs, b)

			if mProcess, ok := b.(transform.NodeProcessor); ok {
				logger.Debug("PipelineSetup", "name", k, "step", i, "processor", fmt.Sprintf("%T", mProcess))
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
					logger.Error("Error setting up step")
					//throw error?
				}
			} else if mProcess, ok := b.(transform.MapProcessor); ok {
				logger.Debug("Pipeline Pool", "name", k, "step", i, "processor", b)
				var c flame.Node[map[string]any, map[string]any]
				if mProcess.PoolReady() {
					logger.Debug("Starting pool worker")
					c = flame.AddMapperPool(wf, mProcess.Process, 4) // TODO: config pool count
				} else {
					c = flame.AddMapper(wf, mProcess.Process)
				}
				if lastStep != nil {
					c.Connect(lastStep)
				}
				if c != nil {
					lastStep = c
					if firstStep == nil {
						firstStep = c
					}
				} else {
					logger.Error("Error setting up step")
					//throw error?
				}
			} else if mProcess, ok := b.(transform.FlatMapProcessor); ok {
				logger.Debug("Pipeline flatmap", "name", k, "step", i, "processor", b)
				var c flame.Node[map[string]any, map[string]any]
				if mProcess.PoolReady() {
					//	log.Printf("Starting pool worker")
					c = flame.AddFlatMapperPool(wf, mProcess.Process, 4) // TODO: config pool count
				} else {
					c = flame.AddFlatMapper(wf, mProcess.Process)
				}
				if lastStep != nil {
					c.Connect(lastStep)
				}
				if c != nil {
					lastStep = c
					if firstStep == nil {
						firstStep = c
					}
				} else {
					logger.Error("Error setting up step")
					//throw error?
				}
			} else if mProcess, ok := b.(transform.StreamProcessor); ok {
				logger.Info("Pipeline stream %s step %d: %T", k, i, b)
				c := flame.AddStreamer(wf, mProcess.Process)
				if c != nil {
					if lastStep != nil {
						c.Connect(lastStep)
					}
					lastStep = c
					if firstStep == nil {
						firstStep = c
					}
				} else {
					logger.Error("Error setting up step")
					//throw error?
				}
			} else if jProcess, ok := b.(transform.JoinProcessor); ok {
				logger.Debug("Pipeline Join Step")
				c := flame.AddJoin(wf, jProcess.Process)
				if c != nil {
					if lastStep != nil {
						c.ConnectLeft(lastStep)
					} else {
						logger.Debug("Join missing input")
					}
					joins = append(joins, joinStruct{
						node: c,
						proc: jProcess,
					})
					lastStep = c
					if firstStep == nil {
						logger.Error("ERROR: Join can't be the first step")
					}
				}
			} else if rProcess, ok := b.(transform.ReduceProcessor); ok {
				logger.Debug("Pipeline reduce %s step %d: %T", k, i, b)
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
				logger.Debug("Pipeline accumulate %s step %d: %T", k, i, b)

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
				logger.Info("Unknown processor type")
			}
		}
		outNodes[k] = lastStep
		inNodes[k] = firstStep
	}

	for dst, p := range pb.Pipelines {
		if len(p) > 0 {
			if p[0].From != nil {
				src := string(*p[0].From)
				if src == dst {
					//TODO: more loop detection
					logger.Error("Pipeline Loop detected in %s", dst)
					return fmt.Errorf("pipeline loop detected")
				}
				if srcNode, ok := outNodes[src]; ok {
					if dstNode, ok := inNodes[dst]; ok {
						logger.Debug("Connecting", "source", src, "dest", dst)
						dstNode.Connect(srcNode)
					} else {
						logger.Error("Dest not found", "name", dst)
					}
				} else {
					logger.Error("source not found", "dest", dst, "source", src)
				}
			} else {
				logger.Error("First step of pipelines not 'from'", "name", dst)
				return fmt.Errorf("first step of pipelines %s not 'from'", dst)
			}
		} else {
			logger.Error("Pipeline %s is empty", dst)
		}
	}

	//for joins, connect the other end
	for _, i := range joins {
		r := i.proc.GetRightPipeline()
		if srcNode, ok := outNodes[r]; ok {
			logger.Debug("Join Connect", "source", r)
			i.node.ConnectRight(srcNode)
		} else {
			logger.Error("Join source not found", "name", r)
		}
	}

	//log.Printf("WF: %#v", wf)

	wf.Start()
	logger.Debug("Workflow Started")

	wf.Wait()

	for p := range procs {
		procs[p].Close()
	}

	task.Close()
	return nil
}
