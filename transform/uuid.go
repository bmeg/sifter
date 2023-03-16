package transform

import (
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
	"github.com/google/uuid"
)

type UUIDStep struct {
	Field     string `json:"field"`
	Value     string `json:"value"`
	Namespace string `json:"namespace"`
}

type uuidProc struct {
	namespace *uuid.UUID
	config    *UUIDStep
	task      task.RuntimeTask
}

func (ss UUIDStep) Init(task task.RuntimeTask) (Processor, error) {
	var ns *uuid.UUID
	if ss.Namespace != "" {
		t := uuid.NewMD5(uuid.NameSpaceDNS, []byte(ss.Namespace))
		ns = &t
	} else {
		ns = &uuid.NameSpaceURL
	}

	return &uuidProc{ns, &ss, task}, nil
}

func (uu *uuidProc) Process(i map[string]any) []map[string]any {
	out := map[string]any{}
	for k, v := range i {
		out[k] = v
	}
	if uu.config.Value == "" {
		out[uu.config.Field] = uuid.New().String()
		return []map[string]any{out}
	}
	value, err := evaluate.ExpressionString(uu.config.Value, uu.task.GetConfig(), i)
	if err == nil {
		o := uuid.NewSHA1(*uu.namespace, []byte(value))
		out[uu.config.Field] = o.String()
	}
	return []map[string]any{out}
}

func (uu *uuidProc) Close() {}

func (uu *uuidProc) PoolReady() bool {
	return true
}
