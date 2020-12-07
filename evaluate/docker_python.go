package evaluate

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	//"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	grpc "google.golang.org/grpc"
)

type DockerPythonEngine struct{ Workdir string }

type PythonEngine struct{ Workdir string }

type Runner interface {
	Compile(code *Code) (*CompileResult, error)
	Call(in *Input) (*Result, error)
	//Process(in chan *Input) (chan *Result, error)
	Close()
}

type PythonProcessor struct {
	runner Runner
	fNum   uint32
}

func (d PythonProcessor) Close() {
	d.runner.Close()
}

func (d PythonProcessor) Evaluate(inputs ...map[string]interface{}) (map[string]interface{}, error) {
	i, err := json.Marshal(inputs)
	out, err := d.runner.Call(&Input{Data: string(i), Code: d.fNum})
	if err != nil {
		return nil, err
	}
	if out.Error != "" {
		return nil, fmt.Errorf("%s", out.Error)
	}
	o := map[string]interface{}{}
	err = json.Unmarshal([]byte(out.Data), &o)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (d PythonProcessor) EvaluateBool(inputs ...map[string]interface{}) (bool, error) {
	i, err := json.Marshal(inputs)
	if err != nil {
		log.Printf("Serialization Error: %s", err)
		return false, err
	}
	out, err := d.runner.Call(&Input{Data: string(i), Code: d.fNum})
	if err != nil {
		return false, err
	}
	if out.Error != "" {
		return false, fmt.Errorf("%s", out.Error)
	}
	var o bool
	err = json.Unmarshal([]byte(out.Data), &o)
	if err != nil {
		return false, err
	}
	return o, nil
}

func (d DockerPythonEngine) Compile(code string, method string) (Processor, error) {
	r, err := StartDockerExecutor("bmeg/sifter-exec-python")
	if err != nil {
		return nil, err
	}
	out, err := r.Compile(&Code{Code: code, Function: method})
	if err != nil {
		r.Close()
		return nil, err
	}
	return PythonProcessor{r, out.Id}, nil
}

func (d PythonEngine) Compile(code string, method string) (Processor, error) {
	r, err := StartLocalExecutor(d.Workdir)
	if err != nil {
		return nil, err
	}
	out, err := r.Compile(&Code{Code: code, Function: method})
	if err != nil {
		r.Close()
		return nil, err
	}
	return PythonProcessor{r, out.Id}, nil
}

type LocalRunner struct {
	Port   int
	Cmd    *exec.Cmd
	Conn   *grpc.ClientConn
	Client ExecutorClient
}

func StartLocalExecutor(workdir string) (Runner, error) {

	d, err := ioutil.TempDir(workdir, "sifterexec_") //TODO: use directory from user config
	if err != nil {
		return nil, err
	}

	for _, f := range []string{"sifter-exec.py", "exec_pb2.py", "exec_pb2_grpc.py"} {
		data, err := Asset(f)
		if err != nil {
			return nil, err
		}
		f, err := os.OpenFile(filepath.Join(d, f), os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return nil, err
		}
		f.Write(data)
		f.Close()
	}

	path := filepath.Join(d, "sifter-exec.py")

	cmd := exec.Command(path)
	cmd.Stderr = os.Stderr
	log.Printf("Launching %#v", cmd)

	stdout, _ := cmd.StdoutPipe()
	err = cmd.Start()
	if err != nil {
		log.Printf("Error starting python: %s", err)
		return nil, err
	}

	m := make(chan int)
	err = nil

	go func() error {
		sent := false
		defer close(m)
		var out []byte
		buf := make([]byte, 1024, 1024)
		for {
			n, ierr := stdout.Read(buf)
			if !sent {
				if n > 0 {
					d := buf[:n]
					out = append(out, d...)
				}
				log.Printf("Read %d (%s) Buffer: %s", n, ierr, string(out))
				if strings.Contains(string(out), "\n") {
					t := strings.Split(string(out), "\n")
					log.Printf("Return port: %s", out)
					p, ierr := strconv.Atoi(string(t[0]))
					err = ierr
					m <- p
					sent = true
				}
			}
			if ierr != nil {
				if ierr == io.EOF {
					log.Printf("Executor closed")
					if cmd.ProcessState.ExitCode() != 0 {
						log.Printf("Executor error: %d", cmd.ProcessState.ExitCode())
					}
					ierr = nil
				}
				return ierr
			}
		}
	}()
	var port int
	port = <-m
	serverAddr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := NewExecutorClient(conn)
	return &LocalRunner{Port: port, Cmd: cmd, Conn: conn, Client: client}, err
}

func (run *LocalRunner) Compile(code *Code) (*CompileResult, error) {
	return run.Client.Compile(context.Background(), code)
}

func (run *LocalRunner) Call(in *Input) (*Result, error) {
	return run.Client.Call(context.Background(), in)
}

func (run *LocalRunner) Close() {
	run.Conn.Close()
	run.Cmd.Process.Kill()
	run.Cmd.Wait()
}

type DockerRunner struct {
	containerId string
	conn        *grpc.ClientConn
	client      ExecutorClient
}

func StartDockerExecutor(dockerImage string) (Runner, error) {
	cmd := exec.Command("docker", "run", "-d", "--rm", "-P", dockerImage)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	id := strings.Trim(string(out), " \r\n\t")
	log.Printf("Started Container: %s", id)
	cmd = exec.Command("docker", "port", id)
	out, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	tmp := strings.Split(string(out), " -> ")
	serverAddr := strings.Trim(tmp[1], " \t\r\n")

	var conn *grpc.ClientConn

	for i := 0; i < 5; i++ {
		log.Printf("Contacting: %s", serverAddr)
		conn, err = grpc.Dial(serverAddr, grpc.WithInsecure(), grpc.WithBlock())
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		log.Printf("Returning err: %s", err)
		return nil, err
	}
	client := NewExecutorClient(conn)
	return &DockerRunner{containerId: id, conn: conn, client: client}, nil
}

func (run *DockerRunner) Compile(code *Code) (*CompileResult, error) {
	return run.client.Compile(context.Background(), code)
}

func (run *DockerRunner) Call(in *Input) (*Result, error) {
	return run.client.Call(context.Background(), in)
}

func (run *DockerRunner) Close() {
	log.Printf("Closing docker %s", run.containerId)
	run.conn.Close()
	cmd := exec.Command("docker", "kill", run.containerId)
	cmd.Run()
}
