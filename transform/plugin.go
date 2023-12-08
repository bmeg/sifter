package transform

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"

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
		log.Printf("Plugin Error: %s", err)
	} else {
		workdir := ps.task.BaseDir()
		cmd := exec.Command(cmdLine[0], cmdLine[1:]...)
		cmd.Dir = workdir
		stdin, _ := cmd.StdinPipe()
		stdout, _ := cmd.StdoutPipe()
		cmd.Stderr = os.Stderr
		log.Printf("Starting: %#v", cmd)
		err := cmd.Start()
		if err != nil {
			log.Printf("plugin exec error: %s", err)
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
					log.Printf("plugin (%s) input error: %s", ps.config.CommandLine, err)
				}
				ln = append(ln, line...)
				if !isPrefix {
					if len(ln) > 0 {
						row := map[string]any{}
						err := json.Unmarshal(ln, &row)
						if err == nil {
							out <- row
						} else {
							log.Printf("plugin output error: %s", err)
							log.Printf("unmarshalled line: %s", ln)
						}
						ln = []byte{}
					}
				}
			}
			wg.Done()
		}()
		log.Printf("plugin has exited: %s\n", ps.config.CommandLine)
		wg.Wait()
	}
}

func (ps *pluginProcess) Close() {

}
