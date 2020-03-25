package pipeline

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bmeg/grip/gripql"
	"github.com/hashicorp/go-getter"

)

type Task struct {
	Runtime *Runtime
	Workdir string
	Inputs  map[string]interface{}
	AllowLocalFiles bool
}

func (m *Task) Path(p string) (string, error) {
	if !strings.HasPrefix(p, "/") {
		p = filepath.Join(m.Workdir, p)
	}
	a, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	if !m.AllowLocalFiles {
		if !strings.HasPrefix(a, m.Workdir) {
			return "", fmt.Errorf("Input file not inside working directory")
		}
	}
	return a, nil
}

func (m *Task) TempDir() string {
	name, _ := ioutil.TempDir(m.Workdir, "tmp")
	return name
}

func (m *Task) DownloadFile(src string, dest string) (string, error) {
	if dest == "" {
		var err error
		dest, err = m.Path(path.Base(src))
		if err != nil {
			return "", err
		}
	} else {
		var err error
		dest, err = m.Path(dest)
		if err != nil {
			return "", err
		}
	}

	if strings.HasPrefix(src, "s3:") {
		s3_key := os.Getenv("AWS_ACCESS_KEY_ID")
		s3_secret := os.Getenv("AWS_SECRET_ACCESS_KEY")
		s3_endpoint := os.Getenv("AWS_ENDPOINT")
		if s3_endpoint != "" {
			u, err := url.Parse(src)
			if err != nil {
				return "", err
			}
			//"s3::http://127.0.0.1:9000/test-bucket/hello.txt?aws_access_key_id=KEYID&aws_access_key_secret=SECRETKEY&region=us-east-2"
			src = fmt.Sprintf("s3::%s/%s%s", s3_endpoint, u.Host, u.Path)
		}
		if s3_key != "" && s3_secret != "" {
			src = src + fmt.Sprintf("?aws_access_key_id=%s&aws_access_key_secret=%s", s3_key, s3_secret)
		}
		src = src + "&archive=false"
	} else {
		src = src + "?archive=false"
	}

	return dest, getter.GetFile(dest, src)
}

func (m *Task) EmitVertex(v *gripql.Vertex) error {
	return m.Runtime.EmitVertex(v)
}

func (m *Task) EmitEdge(e *gripql.Edge) error {
	return m.Runtime.EmitEdge(e)
}

func (m *Task) EmitObject(c string, e map[string]interface{}) error {
	return m.Runtime.EmitObject(c, e)
}

func (m *Task) Output(name string, value string) error {
	if m.Runtime.OutputCallback != nil {
		return m.Runtime.OutputCallback(name, value)
	}
	return fmt.Errorf("Output Callback not set")
}

func (m *Task) Printf(s string, x ...interface{}) {
	m.Runtime.Printf(s, x...)
}
