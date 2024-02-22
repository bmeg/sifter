package transform

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/task"
	"github.com/google/shlex"
)

type PluginStep struct {
	CommandLine string `json:"commandLine"`
}

type pluginProcess struct {
	config *PluginStep
	task   task.RuntimeTask
}

func (ps *PluginStep) Init(task task.RuntimeTask) (Processor, error) {
	return &pluginProcess{config: ps, task: task}, nil
}

func (ps *pluginProcess) Process(in chan map[string]any, out chan map[string]any) {

	cmdLine, err := shlex.Split(ps.config.CommandLine)
	if err != nil {
		logger.Error("Plugin Error: %s", err)
	} else {
		workdir := ps.task.BaseDir()
		cmd := exec.Command(cmdLine[0], cmdLine[1:]...)
		cmd.Dir = workdir
		stdin, _ := cmd.StdinPipe()
		stdout, _ := cmd.StdoutPipe()
		cmd.Stderr = os.Stderr
		logger.Debug("Starting: %#v", cmd)
		err := cmd.Start()
		if err != nil {
			logger.Error("plugin exec error: %s", err)
		}

		go func() {
			for i := range in {
				line, _ := json.Marshal(i)
				stdin.Write(line)
				stdin.Write([]byte("\n"))
			}
			stdin.Close()
		}()
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
					logger.Error("plugin (%s) input error: %s", ps.config.CommandLine, err)
				}
				ln = append(ln, line...)
				if !isPrefix {
					if len(ln) > 0 {
						row := map[string]any{}
						err := json.Unmarshal(ln, &row)
						if err == nil {
							out <- row
						} else {
							logger.Error("plugin output error: %s", err)
							logger.Error("unmarshalled line: %s", ln)
						}
						ln = []byte{}
					}
				}
			}
			wg.Done()
		}()
		logger.Debug("plugin has exited: %s\n", ps.config.CommandLine)
		wg.Wait()
	}
}

func (ps *pluginProcess) Close() {

}
