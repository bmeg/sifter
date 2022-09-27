package transform

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type HashStep struct {
	Field  string `json:"field"`
	Value  string `json:"value"`
	Method string `json:"method"`
}

type hashProcessor struct {
	config HashStep
	hasher hash.Hash
	task   task.RuntimeTask
}

func (fs HashStep) Init(task task.RuntimeTask) (Processor, error) {
	if fs.Method == "sha1" {
		return &hashProcessor{fs, sha1.New(), task}, nil
	} else if fs.Method == "sha256" {
		return &hashProcessor{fs, sha256.New(), task}, nil
	} else if fs.Method == "md5" || fs.Method == "" {
		return &hashProcessor{fs, md5.New(), task}, nil
	}
	return nil, fmt.Errorf("Hash method %s not found", fs.Method)
}

func (fs *hashProcessor) Process(i map[string]any) []map[string]any {
	value, err := evaluate.ExpressionString(fs.config.Value, fs.task.GetConfig(), i)
	if err == nil {
		fs.hasher.Reset()
		fs.hasher.Write([]byte(value))
		i[fs.config.Field] = hex.EncodeToString(fs.hasher.Sum(nil))
	}
	return []map[string]any{i}
}

func (fs *hashProcessor) Close() {

}
