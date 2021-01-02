package pipeline

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	//"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/loader"
	"github.com/hashicorp/go-getter"

	"github.com/jlaffaye/ftp"

	"github.com/bmeg/sifter/datastore"
)

type Task struct {
	Name            string
	Runtime         *Runtime
	Workdir         string
	Inputs          map[string]interface{}
	DataStore       datastore.DataStore
	AllowLocalFiles bool
}

func (m *Task) Child(name string) *Task {
	cname := fmt.Sprintf("%s.%s", m.Name, name)
	return &Task{Name: cname, Runtime: m.Runtime, Workdir: m.Workdir, Inputs: m.Inputs, AllowLocalFiles: m.AllowLocalFiles, DataStore: m.DataStore}
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

	if strings.HasPrefix(src, "ftp:") {
		u, err := url.Parse(src)
		if err != nil {
			return "", err
		}
		c, err := ftp.Dial(u.Host+":21", ftp.DialWithTimeout(5*time.Second))
		if err != nil {
			return "", err
		}
		err = c.Login("anonymous", "anonymous")
		if err != nil {
			return "", err
		}
		r, err := c.Retr(u.Path)
		if err != nil {
			return "", err
		}
		defer r.Close()

		log.Printf("Saving to %s", dest)
		f, err := os.Create(dest)
		if err != nil {
			return "", err
		}
		defer f.Close()

		downloadSize, _ := io.Copy(f, r)
		log.Printf("Downloaded %d bytes", downloadSize)
		return dest, nil
	}

	if strings.HasPrefix(src, "s3:") {
		s3Key := os.Getenv("AWS_ACCESS_KEY_ID")
		s3Secret := os.Getenv("AWS_SECRET_ACCESS_KEY")
		s3Endpoint := os.Getenv("AWS_ENDPOINT")
		if s3Endpoint != "" {
			u, err := url.Parse(src)
			if err != nil {
				return "", err
			}
			//"s3::http://127.0.0.1:9000/test-bucket/hello.txt?aws_access_key_id=KEYID&aws_access_key_secret=SECRETKEY&region=us-east-2"
			src = fmt.Sprintf("s3::%s/%s%s", s3Endpoint, u.Host, u.Path)
		}
		if s3Key != "" && s3Secret != "" {
			src = src + fmt.Sprintf("?aws_access_key_id=%s&aws_access_key_secret=%s", s3Key, s3Secret)
		}
		src = src + "&archive=false"
	} else {
		src = src + "?archive=false"
	}

	return dest, getter.GetFile(dest, src)
}

func (m *Task) Emit(name string, e map[string]interface{}) error {
	return m.Runtime.Emit(name, e)
}

func (m *Task) EmitObject(prefix string, c string, e map[string]interface{}) error {
	return m.Runtime.EmitObject(prefix, c, e)
}

func (m *Task) EmitTable(prefix string, columns []string, sep rune) loader.TableEmitter {
	return m.Runtime.EmitTable(prefix, columns, sep)
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

func (m *Task) GetDataStore() (datastore.DataStore, error) {
	return m.DataStore, nil //DEBUG: fix this
}
