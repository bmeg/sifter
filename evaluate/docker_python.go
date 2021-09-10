package evaluate

import (
	"log"
	"os"
	"time"

	//"bytes"
	"context"
	"os/exec"
	"strings"

	grpc "google.golang.org/grpc"
)

type DockerPythonEngine struct{ Workdir string }


type Runner interface {
	Compile(code *Code) (*CompileResult, error)
	Call(in *Input) (*Result, error)
	//Process(in chan *Input) (chan *Result, error)
	Close()
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



type DockerRunner struct {
	containerID string
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
	return &DockerRunner{containerID: id, conn: conn, client: client}, nil
}

func (run *DockerRunner) Compile(code *Code) (*CompileResult, error) {
	return run.client.Compile(context.Background(), code)
}

func (run *DockerRunner) Call(in *Input) (*Result, error) {
	return run.client.Call(context.Background(), in)
}

func (run *DockerRunner) Close() {
	log.Printf("Closing docker %s", run.containerID)
	run.conn.Close()
	cmd := exec.Command("docker", "kill", run.containerID)
	cmd.Run()
}
