
package steps

import (
  "log"
  "github.com/bmeg/sifter/pipeline"
)


type Step struct {
	Desc         string            `json:"desc"`
	//ManifestLoad *ManifestLoadStep `json:"manifestLoad"`
	Download     *DownloadStep     `json:"download"`
	Untar        *UntarStep        `json:"untar"`
	//VCFLoad      *VCFStep          `json:"vcfLoad"`
	TableLoad    *TableLoadStep    `json:"tableLoad"`
	JSONLoad     *JSONLoadStep     `json:"jsonLoad"`
  SQLDumpLoad   *SQLDumpStep     `json:"sqldumpLoad"`
	TransposeFile *TransposeFileStep `json:"transposeFile"`
	FileGlob      *FileGlobStep      `json:"fileGlob"`
	Script        *ScriptStep        `json:"script"`
}


func (step *Step) Run(run *pipeline.Runtime, inputs map[string]interface{}) error {

  if step.TransposeFile != nil {
    task := run.NewTask(inputs)
    log.Printf("Running Transpose")
    if err := step.TransposeFile.Run(task); err != nil {
      run.Printf("Tranpose Step Error: %s", err)
      return err
    } /*
  } else if step.ManifestLoad != nil {
    task := run.NewTask(inputs)
    log.Printf("Running ManifestLoad")
    if err := step.ManifestLoad.Run(task); err != nil {
      run.Printf("ManifestLoad Error: %s", err)
      return err
    } */
  } else if step.Download != nil {
    task := run.NewTask(inputs)
    log.Printf("Running Download")
    if err := step.Download.Run(task); err != nil {
      run.Printf("Download Error: %s", err)
      return err
    }
  } else if step.Untar != nil {
    task := run.NewTask(inputs)
    log.Printf("Running untar")
    if err := step.Untar.Run(task); err != nil {
      run.Printf("Untar Error: %s", err)
      return err
    } /*
  } else if step.VCFLoad != nil {
    task := run.NewTask(inputs)
    log.Printf("Running VCFLoad")
    if err := step.VCFLoad.Run(task); err != nil {
      run.Printf("VCF Load Error: %s", err)
      return err
    } */
  } else if step.TableLoad != nil {
    task := run.NewTask(inputs)
    log.Printf("Running TableLoad")
    if err := step.TableLoad.Run(task); err != nil {
      run.Printf("Table Load Error: %s", err)
      return err
    }
  } else if step.JSONLoad != nil {
    task := run.NewTask(inputs)
    log.Printf("Running JSONLoad")
    if err := step.JSONLoad.Run(task); err != nil {
      run.Printf("JSON Load Error: %s", err)
      return err
    }
  } else if step.SQLDumpLoad != nil {
    task := run.NewTask(inputs)
    log.Printf("Running SQLDumpLoad")
    if err := step.SQLDumpLoad.Run(task); err != nil {
      run.Printf("SQLDumpLoad Error: %s", err)
      return err
    }
  } else if step.FileGlob != nil {
    task := run.NewTask(inputs)
    log.Printf("Running FileGlob")
    if err := step.FileGlob.Run(task); err != nil {
      run.Printf("FileGlob Error: %s", err)
      return err
    }
  } else if step.Script != nil {
    task := run.NewTask(inputs)
    log.Printf("Running Script")
    if err := step.Script.Run(task); err != nil {
      run.Printf("Script Error: %s", err)
      return err
    }
  } else {
    log.Printf("Unknown Step")
  }
  return nil
}
