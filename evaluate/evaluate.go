package evaluate

type Processor interface {
  Evaluate(inputs... map[string]interface{}) (map[string]interface{}, error)
  EvaluateBool(inputs... map[string]interface{}) (bool, error)
  Close()
}

type Engine interface {
  Compile(code string, method string) (Processor, error)
}

func GetEngine(name string, workdir string) Engine {
  if name == "gpython" {
    return GPythonEngine{}
  }
  if name == "python" {
    //return DockerPythonEngine{}
    return PythonEngine{Workdir:workdir}
  }
  return nil
}
