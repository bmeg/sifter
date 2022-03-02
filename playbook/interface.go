package playbook

import "github.com/bmeg/sifter/task"

type Source interface {
	Start(task.RuntimeTask) chan map[string]interface{}
}

type Process interface {
}
