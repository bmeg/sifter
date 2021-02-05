package schema

import (
	"github.com/bmeg/grip/gripql"
	"google.golang.org/protobuf/types/known/structpb"
)

type AlleleAnnotation struct {
	Allele    Allele
	ID        string
	Label     string
	EdgeLabel string
	Data      map[string]interface{}
}

func (aa *AlleleAnnotation) Render() ([]*gripql.Vertex, []*gripql.Edge) {
	data, _ := structpb.NewStruct(aa.Data)
	av := gripql.Vertex{Gid: aa.ID, Label: aa.Label, Data: data}
	ae := gripql.Edge{Label: aa.EdgeLabel, To: aa.Allele.ID(), From: aa.ID}
	return []*gripql.Vertex{&av}, []*gripql.Edge{&ae}
}
