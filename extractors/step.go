package extractors

import (
	"fmt"
	"reflect"

	"github.com/bmeg/sifter/task"
)

type Source interface {
	Start(task.RuntimeTask) (chan map[string]interface{}, error)
}

type Extractor struct {
	Description string `json:"description"  jsonschema_description:"Human Readable description of step"`
	//Untar         *UntarStep         `json:"untar" jsonschema_description:"Untar a file"`
	XMLLoad *XMLLoadStep `json:"xmlLoad"`
	//TransposeFile *TransposeFileStep `json:"transposeFile" jsonschema_description:"Take a matrix TSV and transpose it (row become columns)"`
	TableLoad *TableLoadStep `json:"tableLoad" jsonschema_description:"Run transform pipeline on a TSV or CSV"`
	JSONLoad  *JSONLoadStep  `json:"jsonLoad" jsonschema_description:"Run a transform pipeline on a multi line json file"`
	//SQLDumpLoad *SQLDumpStep   `json:"sqldumpLoad" jsonschema_description:"Parse the content of a SQL dump to find insert and run a transform pipeline"`
	//SQLiteLoad  *SQLiteStep    `json:"sqliteLoad"`
	//FileGlob      *FileGlobStep      `json:"fileGlob" jsonschema_description:"Scan a directory and run a ETL pipeline on each of the files"`
	//Script        *ScriptStep        `json:"script" jsonschema_description:"Execute a script"`
	GripperLoad *GripperLoadStep `json:"gripperLoad" jsonschema_description:"Use a GRIPPER server to get data and run a transform pipeline"`
	AvroLoad    *AvroLoadStep    `json:"avroLoad" jsonschema_description:"Load data from avro file"`
}

func (ex *Extractor) Start(t task.RuntimeTask) (chan map[string]interface{}, error) {
	v := reflect.ValueOf(ex).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		x := f.Interface()
		if z, ok := x.(Source); ok {
			if !f.IsNil() {
				return z.Start(t)
			}
		}
	}
	return nil, fmt.Errorf(("Extractor not defined"))
}
