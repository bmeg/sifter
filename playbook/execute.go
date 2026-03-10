package playbook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

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

func (pb *Playbook) PrepConfig(inputParams map[string]string, workdir string) (map[string]string, error) {

	playbookParams := map[string]string{}
	for k, v := range pb.Params {
		if _, ok := inputParams[k]; ok {
			if v.IsFile() || v.IsDir() {
				var defaultPath = inputParams[k]
				if !filepath.IsAbs(inputParams[k]) {
					defaultPath = filepath.Join(workdir, inputParams[k])
				}
				playbookParams[k], _ = filepath.Abs(defaultPath)
			} else {
				playbookParams[k] = inputParams[k]
			}
		} else {
			if v.Default != nil {
				if v.IsFile() || v.IsDir() {
					var defaultPath = fmt.Sprintf("%v", v.Default)
					if !filepath.IsAbs(defaultPath) {
						dirPath := filepath.Dir(pb.path)
						defaultPath = filepath.Join(dirPath, defaultPath)
					}
					playbookParams[k], _ = filepath.Abs(defaultPath)
				} else {
					playbookParams[k] = fmt.Sprintf("%v", v.Default)
				}
			} else {
				return nil, fmt.Errorf("parameter %s not defined", k)
			}
		}
	}

	workdir, _ = filepath.Abs(workdir)
	missing := map[string]bool{}
	out := map[string]string{}
	for _, v := range pb.GetRequiredParams() {
		if val, ok := playbookParams[v.Name]; ok {
			out[v.Name] = val
			logger.Debug("input: ", v.Name, out[v.Name])
		} else if p, ok := pb.Params[v.Name]; ok {
			if p.Default != nil {
				val := fmt.Sprintf("%v", p.Default)
				if v.IsFile() || v.IsDir() {
					var defaultPath = val
					if !filepath.IsAbs(val) {
						defaultPath = filepath.Join(filepath.Dir(pb.path), val)
					}
					out[v.Name], _ = filepath.Abs(defaultPath)
				} else {
					out[v.Name] = val
				}
			} else {
				missing[v.Name] = true
			}
		} else {
			return nil, fmt.Errorf("parameter %s not defined", v.Name)
		}
	}
	if len(missing) > 0 {
		o := []string{}
		for k := range missing {
			o = append(o, k)
		}
		return nil, fmt.Errorf("missing inputs: %s", strings.Join(o, ","))
	}
	logger.Debug("prep config inputs", "config", out)
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

// stepCaptureState tracks debug capture state for a single step
type stepCaptureState struct {
	pipelineName string
	stepIndex    int
	stepType     string
	count        uint64
	limit        int
	file         *os.File
	mu           sync.Mutex
}

// captureRecord writes a debug record to the capture file
func (s *stepCaptureState) captureRecord(record map[string]any) {
	if s.limit > 0 {
		currentCount := atomic.LoadUint64(&s.count)
		if currentCount >= uint64(s.limit) {
			return
		}
	}

	recordNum := atomic.AddUint64(&s.count, 1)

	envelope := map[string]any{
		"pipeline":   s.pipelineName,
		"step_index": s.stepIndex,
		"step_type":  s.stepType,
		"record_num": recordNum,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"data":       record,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.file != nil {
		data, err := json.Marshal(envelope)
		if err == nil {
			s.file.Write(data)
			s.file.Write([]byte("\n"))
		} else {
			logger.Error("Failed to marshal debug record", "error", err)
		}
	}
}

func (pb *Playbook) Execute(task task.RuntimeTask, captureDir string, captureLimit int) error {
	logger.Debug("Running playbook")
	logger.Debug("Inputs", "config", task.GetConfig())

	wf := flame.NewWorkflow()

	if pb.Name == "" {
		pb.Name = "sifter"
	}
	task.SetName(pb.Name)

	procs := []transform.Processor{}
	joins := []joinStruct{}
	captureFiles := []*os.File{} // Track all open capture files for cleanup

	// Helper function to sanitize filename components
	sanitizeFilename := func(s string) string {
		s = strings.ReplaceAll(s, "*", "")
		s = strings.ReplaceAll(s, "/", "_")
		s = strings.ReplaceAll(s, "\\", "_")
		s = strings.ReplaceAll(s, "transform.", "")  // Remove package prefix for readability
		s = strings.ReplaceAll(s, "extractors.", "") // Remove package prefix for readability
		return s
	}

	// Helper function to create capture state for a step
	createCaptureState := func(pipelineName string, stepIndex int, stepType string) *stepCaptureState {
		if captureDir == "" {
			return nil
		}

		filename := fmt.Sprintf("%s.%d.%s.ndjson", pipelineName, stepIndex, sanitizeFilename(stepType))
		filepath := filepath.Join(captureDir, filename)

		file, err := os.Create(filepath)
		if err != nil {
			logger.Error("Failed to create debug capture file", "path", filepath, "error", err)
			return nil
		}

		captureFiles = append(captureFiles, file)
		logger.Debug("Created debug capture file", "path", filepath)

		return &stepCaptureState{
			pipelineName: pipelineName,
			stepIndex:    stepIndex,
			stepType:     stepType,
			count:        0,
			limit:        captureLimit,
			file:         file,
		}
	}

	outNodes := map[string]flame.Emitter[map[string]any]{}
	inNodes := map[string]flame.Receiver[map[string]any]{}
	outputs := map[string]OutputProcessor{}

	for n, v := range pb.Inputs {
		logger.Debug("Setting Up", "name", n)
		s, err := v.Start(task)
		if err == nil {
			sourceNode := flame.AddSourceChan(wf, s)

			captureState := createCaptureState(n, 0, v.GetType().String())
			if captureState != nil {
				captureMapper := flame.AddMapper(wf, func(record map[string]any) map[string]any {
					captureState.captureRecord(record)
					return record
				})
				captureMapper.Connect(sourceNode)
				outNodes[n] = captureMapper
			} else {
				outNodes[n] = sourceNode
			}
		} else {
			logger.Error("Source error", "error", err)
			return err
		}
	}

	for k, v := range pb.Pipelines {
		var lastStep flame.Emitter[map[string]any]
		var firstStep flame.Receiver[map[string]any]
		for i, s := range v {

			b, err := s.Init(task)
			if err != nil {
				logger.Error("Pipeline error", "name", k, "error", err)
				return err
			}

			procs = append(procs, b)

			if mProcess, ok := b.(transform.NodeProcessor); ok {
				logger.Debug("PipelineSetup", "name", k, "step", i, "processor", fmt.Sprintf("%T", mProcess))

				// Create capture state for this step
				captureState := createCaptureState(k, i, fmt.Sprintf("%T", mProcess))

				// Wrap the process function if capture is enabled
				var processFunc func(map[string]any) []map[string]any
				if captureState != nil {
					processFunc = func(record map[string]any) []map[string]any {
						out := mProcess.Process(record)
						for _, r := range out {
							captureState.captureRecord(r)
						}
						return out
					}
				} else {
					processFunc = mProcess.Process
				}

				c := flame.AddFlatMapper(wf, processFunc)
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

				// Create capture state for this step
				captureState := createCaptureState(k, i, fmt.Sprintf("%T", mProcess))

				// Wrap the process function if capture is enabled
				var processFunc func(map[string]any) map[string]any
				if captureState != nil {
					processFunc = func(record map[string]any) map[string]any {
						out := mProcess.Process(record)
						captureState.captureRecord(out)
						return out
					}
				} else {
					processFunc = mProcess.Process
				}

				var c flame.Node[map[string]any, map[string]any]
				if mProcess.PoolReady() {
					logger.Debug("Starting pool worker")
					c = flame.AddMapperPool(wf, processFunc, 4) // TODO: config pool count
				} else {
					c = flame.AddMapper(wf, processFunc)
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

				// Create capture state for this step
				captureState := createCaptureState(k, i, fmt.Sprintf("%T", mProcess))

				// Wrap the process function if capture is enabled
				var processFunc func(map[string]any) []map[string]any
				if captureState != nil {
					processFunc = func(record map[string]any) []map[string]any {
						out := mProcess.Process(record)
						for _, r := range out {
							captureState.captureRecord(r)
						}
						return out
					}
				} else {
					processFunc = mProcess.Process
				}

				var c flame.Node[map[string]any, map[string]any]
				if mProcess.PoolReady() {
					//	log.Printf("Starting pool worker")
					c = flame.AddFlatMapperPool(wf, processFunc, 4) // TODO: config pool count
				} else {
					c = flame.AddFlatMapper(wf, processFunc)
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
				// Note: StreamProcessor uses channels, not suitable for simple record capture
				// Would need to wrap the entire channel processing, which is complex
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
				// Note: JoinProcessor uses channels, not suitable for simple record capture
				// Would need to wrap the entire channel processing, which is complex
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

				// Create capture state for this step (pre-reduce)
				captureStateInput := createCaptureState(k, i, fmt.Sprintf("%T-input", rProcess))
				captureStateOutput := createCaptureState(k, i, fmt.Sprintf("%T-output", rProcess))

				wrap := reduceWrapper{rProcess}

				// Wrap addKeyValue if capturing input
				var addKeyValueFunc func(map[string]any) flame.KeyValue[string, map[string]any]
				if captureStateInput != nil {
					addKeyValueFunc = func(x map[string]any) flame.KeyValue[string, map[string]any] {
						captureStateInput.captureRecord(x)
						return wrap.addKeyValue(x)
					}
				} else {
					addKeyValueFunc = wrap.addKeyValue
				}

				// Wrap reduce function if capturing output
				var reduceFunc func(string, map[string]any, map[string]any) map[string]any
				if captureStateOutput != nil {
					reduceFunc = func(key string, acc map[string]any, val map[string]any) map[string]any {
						result := rProcess.Reduce(key, acc, val)
						captureStateOutput.captureRecord(result)
						return result
					}
				} else {
					reduceFunc = rProcess.Reduce
				}

				k := flame.AddMapper(wf, addKeyValueFunc)
				r := flame.AddReduceKey(wf, reduceFunc, rProcess.GetInit())
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

				// Create capture state for this step
				captureStateInput := createCaptureState(k, i, fmt.Sprintf("%T-input", rProcess))
				captureStateOutput := createCaptureState(k, i, fmt.Sprintf("%T-output", rProcess))

				wrap := accumulateWrapper{rProcess}

				// Wrap addKeyValue if capturing input
				var addKeyValueFunc func(map[string]any) flame.KeyValue[string, map[string]any]
				if captureStateInput != nil {
					addKeyValueFunc = func(x map[string]any) flame.KeyValue[string, map[string]any] {
						captureStateInput.captureRecord(x)
						return wrap.addKeyValue(x)
					}
				} else {
					addKeyValueFunc = wrap.addKeyValue
				}

				// Wrap accumulate function if capturing output
				var accumulateFunc func(string, []map[string]any) map[string]any
				if captureStateOutput != nil {
					accumulateFunc = func(key string, vals []map[string]any) map[string]any {
						result := rProcess.Accumulate(key, vals)
						captureStateOutput.captureRecord(result)
						return result
					}
				} else {
					accumulateFunc = rProcess.Accumulate
				}

				k := flame.AddMapper(wf, addKeyValueFunc)
				r := flame.AddAccumulate(wf, accumulateFunc)
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

	for k, v := range pb.Outputs {
		if v.JSON != nil {
			proc, err := v.JSON.Init(task)
			if err == nil {
				if srcNode, ok := outNodes[v.JSON.From]; ok {
					s := flame.AddSink(wf, proc.Process)
					outputs[k] = proc
					s.Connect(srcNode)
				}
			}
		} else if v.Table != nil {
			proc, err := v.Table.Init(task)
			if err == nil {
				if srcNode, ok := outNodes[v.Table.From]; ok {
					s := flame.AddSink(wf, proc.Process)
					outputs[k] = proc
					s.Connect(srcNode)
				}
			}
		} else if v.Graph != nil {
			proc, err := v.Graph.Init(task)
			if err == nil {
				if srcNode, ok := outNodes[v.Graph.From]; ok {
					s := flame.AddSink(wf, proc.Process)
					outputs[k] = proc
					s.Connect(srcNode)
				}
			}
		}
	}

	//log.Printf("WF: %#v", wf)

	wf.Start()
	logger.Debug("Workflow Started")

	wf.Wait()

	for p := range procs {
		procs[p].Close()
	}

	for k := range outputs {
		outputs[k].Close()
	}

	// Close all debug capture files
	for _, f := range captureFiles {
		if f != nil {
			f.Close()
		}
	}

	task.Close()
	return nil
}
