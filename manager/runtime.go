package manager

type Runtime struct {
	man *Manager
	dir string
}

func (run *Runtime) NewTask(inputs map[string]interface{}) *Task {
	return &Task{run.man, run.dir, inputs}
}
