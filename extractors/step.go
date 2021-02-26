package extractors

import (
	"log"

	"github.com/bmeg/sifter/manager"
)

type Extractor struct {
	Description   string             `json:"description"  jsonschema_description:"Human Readable description of step"`
	Download      *DownloadStep      `json:"download" jsonschema_description:"Download a File"`
	Untar         *UntarStep         `json:"untar" jsonschema_description:"Untar a file"`
	XMLLoad       *XMLLoadStep       `json:"xmlLoad"`
	TransposeFile *TransposeFileStep `json:"transposeFile" jsonschema_description:"Take a matrix TSV and transpose it (row become columns)"`
	TableLoad     *TableLoadStep     `json:"tableLoad" jsonschema_description:"Run transform pipeline on a TSV or CSV"`
	JSONLoad      *JSONLoadStep      `json:"jsonLoad" jsonschema_description:"Run a transform pipeline on a multi line json file"`
	SQLDumpLoad   *SQLDumpStep       `json:"sqldumpLoad" jsonschema_description:"Parse the content of a SQL dump to find insert and run a transform pipeline"`
	SQLiteLoad    *SQLiteStep        `json:"sqliteLoad"`
	FileGlob      *FileGlobStep      `json:"fileGlob" jsonschema_description:"Scan a directory and run a ETL pipeline on each of the files"`
	Script        *ScriptStep        `json:"script" jsonschema_description:"Execute a script"`
	GripperLoad   *GripperLoadStep   `json:"gripperLoad" jsonschema_description:"Use a GRIPPER server to get data and run a transform pipeline"`
	AvroLoad      *AvroLoadStep      `json:"avroLoad" jsonschema_description:"Load data from avro file"`
}

func (step *Extractor) Run(run *manager.Runtime, playBookPath string, inputs map[string]interface{}) error {

	if step.TransposeFile != nil {
		task := run.NewTask(playBookPath, inputs)
		log.Printf("Running Transpose")
		if err := step.TransposeFile.Run(task); err != nil {
			run.Printf("Tranpose Step Error: %s", err)
			return err
		}
		task.Close()
		/*
		  } else if step.ManifestLoad != nil {
		    task := run.NewTask(inputs)
		    log.Printf("Running ManifestLoad")
		    if err := step.ManifestLoad.Run(task); err != nil {
		      run.Printf("ManifestLoad Error: %s", err)
		      return err
		    } */
	} else if step.Download != nil {
		task := run.NewTask(playBookPath, inputs)
		log.Printf("Running Download")
		if err := step.Download.Run(task); err != nil {
			run.Printf("Download Error: %s", err)
			return err
		}
		task.Close()
	} else if step.Untar != nil {
		task := run.NewTask(playBookPath, inputs)
		log.Printf("Running untar")
		if err := step.Untar.Run(task); err != nil {
			run.Printf("Untar Error: %s", err)
			return err
		}
		task.Close()
		/*
		  } else if step.VCFLoad != nil {
		    task := run.NewTask(inputs)
		    log.Printf("Running VCFLoad")
		    if err := step.VCFLoad.Run(task); err != nil {
		      run.Printf("VCF Load Error: %s", err)
		      return err
		    } */
	} else if step.TableLoad != nil {
		task := run.NewTask(playBookPath, inputs)
		log.Printf("Running TableLoad")
		if err := step.TableLoad.Run(task); err != nil {
			run.Printf("Table Load Error: %s", err)
			return err
		}
		task.Close()
	} else if step.GripperLoad != nil {
		task := run.NewTask(playBookPath, inputs)
		log.Printf("Running GripperLoad")
		if err := step.GripperLoad.Run(task); err != nil {
			run.Printf("Gripper Load Error: %s", err)
			return err
		}
		task.Close()
	} else if step.JSONLoad != nil {
		task := run.NewTask(playBookPath, inputs)
		log.Printf("Running JSONLoad")
		if err := step.JSONLoad.Run(task); err != nil {
			run.Printf("JSON Load Error: %s", err)
			return err
		}
		task.Close()
	} else if step.XMLLoad != nil {
		task := run.NewTask(playBookPath, inputs)
		log.Printf("Running XMLLoad")
		if err := step.XMLLoad.Run(task); err != nil {
			run.Printf("XML Load Error: %s", err)
			return err
		}
		log.Printf("XMLLoad Done")
		task.Close()
	} else if step.AvroLoad != nil {
		task := run.NewTask(playBookPath, inputs)
		log.Printf("Running AvroLoad")
		if err := step.AvroLoad.Run(task); err != nil {
			run.Printf("Avro Load Error: %s", err)
			return err
		}
		log.Printf("AvroLoad Done")
		task.Close()
	} else if step.SQLDumpLoad != nil {
		task := run.NewTask(playBookPath, inputs)
		log.Printf("Running SQLDumpLoad")
		if err := step.SQLDumpLoad.Run(task); err != nil {
			run.Printf("SQLDumpLoad Error: %s", err)
			return err
		}
		task.Close()
	} else if step.SQLiteLoad != nil {
		task := run.NewTask(playBookPath, inputs)
		log.Printf("Running SQLiteLoad")
		if err := step.SQLiteLoad.Run(task); err != nil {
			run.Printf("SQLiteLoad Error: %s", err)
			return err
		}
		task.Close()
	} else if step.FileGlob != nil {
		task := run.NewTask(playBookPath, inputs)
		log.Printf("Running FileGlob")
		if err := step.FileGlob.Run(task); err != nil {
			run.Printf("FileGlob Error: %s", err)
			return err
		}
		task.Close()
	} else if step.Script != nil {
		task := run.NewTask(playBookPath, inputs)
		log.Printf("Running Script")
		if err := step.Script.Run(task); err != nil {
			run.Printf("Script Error: %s", err)
			return err
		}
		task.Close()
	} else {
		log.Printf("Unknown Step")
	}
	return nil
}
