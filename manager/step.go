
package manager

import (
  "log"
)

func (step *Step) Run(run *Runtime, inputs map[string]interface{}) error {

  if step.TransposeFile != nil {
    task := run.NewTask(inputs)
    if err := step.TransposeFile.Run(task); err != nil {
      run.Printf("Load Error: %s", err)
      return err
    }
  } else if step.ManifestLoad != nil {
    task := run.NewTask(inputs)
    if err := step.ManifestLoad.Run(task); err != nil {
      run.Printf("Load Error: %s", err)
      return err
    }
  } else if step.Download != nil {
    task := run.NewTask(inputs)
    if err := step.Download.Run(task); err != nil {
      run.Printf("Load Error: %s", err)
      return err
    }
  } else if step.Untar != nil {
    task := run.NewTask(inputs)
    if err := step.Untar.Run(task); err != nil {
      run.Printf("Untar Error: %s", err)
      return err
    }
  } else if step.VCFLoad != nil {
    task := run.NewTask(inputs)
    if err := step.VCFLoad.Run(task); err != nil {
      run.Printf("VCF Load Error: %s", err)
      return err
    }
  } else if step.TableLoad != nil {
    task := run.NewTask(inputs)
    if err := step.TableLoad.Run(task); err != nil {
      run.Printf("Table Load Error: %s", err)
      return err
    }
  } else if step.JSONLoad != nil {
    task := run.NewTask(inputs)
    if err := step.JSONLoad.Run(task); err != nil {
      run.Printf("JSON Load Error: %s", err)
      return err
    }
  } else if step.FileGlob != nil {
    task := run.NewTask(inputs)
    if err := step.FileGlob.Run(task); err != nil {
      run.Printf("FileGlob Error: %s", err)
      return err
    }
  } else if step.Script != nil {
    task := run.NewTask(inputs)
    if err := step.Script.Run(task); err != nil {
      run.Printf("Script Error: %s", err)
      return err
    }
  } else {
    log.Printf("Unknown Step")
  }
  return nil
}
