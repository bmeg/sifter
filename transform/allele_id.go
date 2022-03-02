package transform

import (
	"crypto/sha1"
	"fmt"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type AlleleIDStep struct {
	Prefix         string `json:"prefix"`
	Genome         string `json:"genome"`
	Chromosome     string `json:"chromosome"`
	Start          string `json:"start"`
	End            string `json:"end"`
	ReferenceBases string `json:"reference_bases"`
	AlternateBases string `json:"alternate_bases"`
	Dest           string `json:"dst"`
}

func (al AlleleIDStep) Run(i map[string]interface{}, task task.RuntimeTask) map[string]interface{} {

	genome, _ := evaluate.ExpressionString(al.Genome, task.GetInputs(), i)
	chromosome, _ := evaluate.ExpressionString(al.Chromosome, task.GetInputs(), i)
	start, _ := evaluate.ExpressionString(al.Start, task.GetInputs(), i)
	end, _ := evaluate.ExpressionString(al.End, task.GetInputs(), i)
	ref, _ := evaluate.ExpressionString(al.ReferenceBases, task.GetInputs(), i)
	alt, _ := evaluate.ExpressionString(al.AlternateBases, task.GetInputs(), i)

	id := fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		genome, chromosome,
		start, end,
		ref,
		alt)
	//log.Printf("AlleleStr: %s", id)
	idSha1 := fmt.Sprintf("%x", sha1.Sum([]byte(id)))
	if al.Prefix != "" {
		idSha1 = al.Prefix + idSha1
	}
	o := map[string]interface{}{}
	for k, v := range i {
		o[k] = v
	}
	o[al.Dest] = idSha1
	return o
}
