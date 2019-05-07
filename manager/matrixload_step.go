package manager

import (
	"fmt"
	"log"
	"os"
	"io"
	"strings"
	"compress/gzip"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/grip/protoutil"
	"github.com/bmeg/sifter/evaluate"

	"github.com/go-gota/gota/dataframe"
	//"github.com/go-gota/gota/series"
)

type MatrixLoadStep struct {
	Input         string                 `json:"input"`
	SkipIfMissing bool                   `json:"skipIfMissing"`
	RowLabel      string                 `json:"rowLabel"`
	RowPrefix     string                 `json:"rowPrefix"`
	RowSkip       int                    `json:"rowSkip"`
	Exclude       []string               `json:"exclude"`
	Transpose     bool                   `json:"transpose"`
	IndexCol      int                    `json:"transpose"`
	NoVertex      bool                   `json:"noVertex"`
	Edge          []EdgeCreateStep       `json:"edge"`
	DestVertex    []VertexCreateStep     `json:"destVertex"`
	ColumnReplace []ColumnReplaceStep    `json:"columnReplace"`
	ColumnExclude []string               `json:"columnExclude"`
}

func contains(s []string, q string) bool {
	for _, i := range s {
		if i == q {
			return true
		}
	}
	return false
}

func (ml *MatrixLoadStep) Run(task *Task) error {

	input, err := evaluate.ExpressionString(ml.Input, task.Inputs, nil)
	inputPath, err := task.Path(input)

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		if ml.SkipIfMissing {
			return nil
		}
		return fmt.Errorf("File Not Found: %s", input)
	}
	log.Printf("Loading: %s", inputPath)
	fhd, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer fhd.Close()

	var hd io.Reader
	if strings.HasSuffix(input, ".gz") || strings.HasSuffix(input, ".tgz") {
		hd, err = gzip.NewReader(fhd)
		if err != nil {
			return err
		}
	} else {
    hd = fhd
  }

	df := dataframe.ReadCSV(hd, dataframe.WithDelimiter('\t'), dataframe.HasHeader(true), dataframe.WithComments('#'))
	cols := df.Names()
	idCol := cols[0]

	ids := df.Col(idCol)

	colMap := map[string][]string{}
	for _, n := range cols[1:] {
		if !contains(ml.ColumnExclude, n) {
			colMap[n] = df.Col(n).Records()
		}
	}

	if !ml.NoVertex {
		for i, g := range ids.Records() {
			o := map[string]interface{}{}
			for k, v := range colMap {
				o[k] = v[i]
			}
			v := gripql.Vertex{Gid: g, Label: ml.RowLabel, Data: protoutil.AsStruct(o)}
			task.EmitVertex(&v)
		}
	}
	return nil
}
