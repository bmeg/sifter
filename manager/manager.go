package manager

import (
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/emitter"
	"github.com/hashicorp/go-getter"
)

type Manager struct {
	Workdir string
	Args    []string
	Output  emitter.Emitter
}

func Init(args []string) (Manager, error) {
	dir, err := ioutil.TempDir("./", "sifterwork_")
	if err != nil {
		log.Fatal(err)
	}
	//s := emitter.StdoutEmitter{}
	//s, _ := emitter.NewMongoEmitter("localhost:27017", "test")
	s, err := emitter.NewGripEmitter("localhost:8202", "test")
	return Manager{dir, args, s}, err
}

func (m Manager) Close() {
	os.RemoveAll(m.Workdir)
}

func (m Manager) Path(p string) string {
	return path.Join(m.Workdir, p)
}

func (m Manager) DownloadFile(url string) (string, error) {
	d := m.Path(path.Base(url))
	return d, getter.GetFile(d, url+"?archive=false")
}

func (m Manager) EmitVertex(v *gripql.Vertex) error {
	return m.Output.EmitVertex(v)
}

func (m Manager) EmitEdge(e *gripql.Edge) error {
	return m.Output.EmitEdge(e)
}
