package playbook

import (
	"github.com/bmeg/grip/gripql"
)

type MatrixLoadStep struct {
	RowLabel      string                 `json:"rowLabel"`
	RowPrefix     string                 `json:"rowPrefix"`
	RowSkip       int                    `json:"rowSkip"`
	Exclude       []string               `json:"exclude"`
	Transpose     bool                   `json:"transpose"`
	IndexCol      int                    `json:"transpose"`
	NoVertex      bool                   `json:"noVertex"`
	Edge          []EdgeCreationStep     `json:"edge"`
	DestVertex    []DestVertexCreateStep `json:"destVertex"`
	ColumnReplace []ColumnReplaceStep    `json:"columnReplace"`
	ColumnExclude []string               `json:"columnExclude"`
}

func (ml *MatrixLoadStep) Load() chan gripql.GraphElement {
	out := make(chan gripql.GraphElement, 10)
	close(out)
	return out
}
