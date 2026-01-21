package playbook

import (
	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/task"
)

type OutputProcessor interface {
	Process(i map[string]any)
}

type OutputProcess struct {
	pb     *Playbook
	config Output
	task   task.RuntimeTask
	count  uint64
}

func (op *OutputProcess) Close() {
	logger.Info("Emit Summary", "name", op.config.Path, "count", op.count)
}

func (op *OutputProcess) Process(i map[string]interface{}) {
	op.count++
	op.task.Emit(op.config.Path, i)
}
