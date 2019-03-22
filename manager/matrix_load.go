package manager

import (
	"fmt"
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

func (ml *MatrixLoadStep) Run(man *Task) error {
	return fmt.Errorf("Not Implemented")
}
