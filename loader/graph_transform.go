package loader

import (
	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/sifter/schema"

	"google.golang.org/protobuf/types/known/structpb"
)

type DataGraphEmitter struct {
	sc *schema.Schemas
	gr GraphEmitter
}

// GraphTransformer creates a DataEmitter object that translates raw objects
// (into graph objects and emits them as vertices and edges
func GraphTransformer(gr GraphEmitter, sc *schema.Schemas) DataEmitter {
	return &DataGraphEmitter{sc, gr}
}

func (dg *DataGraphEmitter) Emit(name string, e map[string]interface{}) error {
	if name == "vertex" || name == "vertices" {
		if ogid, ok := e["_gid"]; ok {
			if gid, ok := ogid.(string); ok {
				if olabel, ok := e["_label"]; ok {
					if label, ok := olabel.(string); ok {
						data := map[string]interface{}{}
						if odata, ok := e["_data"]; ok {
							if idata, ok := odata.(map[string]interface{}); ok {
								data = idata
							}
						}
						d, _ := structpb.NewStruct(data)
						dg.gr.EmitVertex(&gripql.Vertex{Gid: gid, Label: label, Data: d})
					}
				}
			}
		}
	}
	if name == "edge" || name == "edges" {
		var gid string
		if ogid, ok := e["_gid"]; ok {
			if sgid, ok := ogid.(string); ok {
				gid = sgid
			}
		}
		var data *structpb.Struct
		if odata, ok := e["_data"]; ok {
			if gdata, ok := odata.(map[string]interface{}); ok {
				data, _ = structpb.NewStruct(gdata)
			}
		}

		if olabel, ok := e["_label"]; ok {
			if label, ok := olabel.(string); ok {
				if oTo, ok := e["_to"]; ok {
					if sTo, ok := oTo.(string); ok {
						if oFrom, ok := e["_from"]; ok {
							if sFrom, ok := oFrom.(string); ok {
								edge := gripql.Edge{Gid: gid, Label: label, Data: data, To: sTo, From: sFrom}
								dg.gr.EmitEdge(&edge)
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func (dg *DataGraphEmitter) Close() {

}

func (dg *DataGraphEmitter) EmitObject(prefix string, objClass string, e map[string]interface{}) error {
	return nil
}

func (dg *DataGraphEmitter) EmitTable(prefix string, columns []string, sep rune) TableEmitter {
	return nil
}
