package extractors

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/task"
	"github.com/google/shlex"
)

type PluginLoadStep struct {
	CommandLine string `json:"commandLine"`
}

func (ml *PluginLoadStep) Start(task task.RuntimeTask) (chan map[string]interface{}, error) {
	logger.Debug("Starting JSON Load")
	cmdText, err := evaluate.ExpressionString(ml.CommandLine, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}
	cmdLine, err := shlex.Split(cmdText)
	if err != nil {
		return nil, err
	}

	procChan := make(chan map[string]interface{}, 100)
	go func() {
		workdir := task.BaseDir()
		cmd := exec.Command(cmdLine[0], cmdLine[1:]...)
		cmd.Dir = workdir
		stdout, _ := cmd.StdoutPipe()
		cmd.Stderr = os.Stderr
		logger.Debug("Starting: %#v", cmd)
		err := cmd.Start()
		if err != nil {
			logger.Error("plugin exec error: %s", err)
		}

		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			reader := bufio.NewReaderSize(stdout, 102400)
			var isPrefix bool = true
			var err error = nil
			var line, ln []byte

			for err == nil {
				line, isPrefix, err = reader.ReadLine()
				if err != nil && err != io.EOF {
					logger.Error("plugin (%s) input error: %s", ml.CommandLine, err)
				}
				ln = append(ln, line...)
				if !isPrefix {
					if len(ln) > 0 {
						row := map[string]any{}
						err := json.Unmarshal(ln, &row)
						if err == nil {
							procChan <- row
						} else {
							logger.Error("plugin (%s) output error: %s", ml.CommandLine, err)
							logger.Error("unmarshalled line: %s", ln)
						}
						ln = []byte{}
					}
				}
			}

			wg.Done()
		}()

		logger.Debug("plugin has exited: %s\n", ml.CommandLine)
		wg.Wait()

		close(procChan)
	}()
	return procChan, nil
}

func (ml *PluginLoadStep) GetConfigFields() []config.Variable {
	out := []config.Variable{}
	for _, s := range evaluate.ExpressionIDs(ml.CommandLine) {
		out = append(out, config.Variable{Type: "File", Name: config.TrimPrefix(s)})
	}
	return out
}
