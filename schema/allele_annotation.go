package schema

import (
  "github.com/bmeg/grip/gripql"
  "github.com/bmeg/grip/protoutil"
)

type AlleleAnnotation struct {
  Allele Allele
  ID     string
  Label  string
  EdgeLabel string
  Data   map[string]interface{}
}


func (aa *AlleleAnnotation) Render() ([]*gripql.Vertex, []*gripql.Edge) {
  av := gripql.Vertex{Gid:aa.ID, Label:aa.Label, Data:protoutil.AsStruct(aa.Data)}
  ae := gripql.Edge{Label:aa.EdgeLabel, To:aa.Allele.ID(), From:aa.ID}
  return []*gripql.Vertex{&av}, []*gripql.Edge{&ae}
}
