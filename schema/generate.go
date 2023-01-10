package schema

import (
	"fmt"

	"github.com/bmeg/grip/gripql"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"google.golang.org/protobuf/types/known/structpb"
)

type GraphElement struct {
	Vertex  *gripql.Vertex
	InEdge  *gripql.Edge
	OutEdge *gripql.Edge
	Field   string
}

func getReferenceIDField(data map[string]any, name string) ([]string, error) {
	out := []string{}
	if d, ok := data[name]; ok {
		//fmt.Printf("Dest id field %#v\n", d)
		if idStr, ok := d.(string); ok {
			out = append(out, idStr)
		} else if idArray, ok := d.([]any); ok {
			for _, g := range idArray {
				if gStr, ok := g.(string); ok {
					out = append(out, gStr)
				} else if gMap, ok := g.(map[string]any); ok {
					if id, ok := gMap["reference"]; ok {
						if idStr, ok := id.(string); ok {
							out = append(out, idStr)
						}
					} else {
						fmt.Printf("Not found in %#v\n", gMap)
					}
				}

			}
		} else if idMap, ok := d.(map[string]any); ok {
			if id, ok := idMap["reference"]; ok {
				if idStr, ok := id.(string); ok {
					out = append(out, idStr)
				}
			}
		}
	}

	return out, nil
}

func getObjectID(data map[string]any, schema *jsonschema.Schema) (string, error) {
	if id, ok := data["id"]; ok {
		if idStr, ok := id.(string); ok {
			return idStr, nil
		}
	}
	return "", fmt.Errorf("object id not found")
}

func (s GraphSchema) Generate(classID string, data map[string]any, clean bool) ([]GraphElement, error) {
	if class := s.GetClass(classID); class != nil {
		if clean {
			var err error
			data, err = s.CleanAndValidate(class, data)
			if err != nil {
				return nil, err
			}
		} else {
			err := class.Validate(data)
			if err != nil {
				return nil, err
			}
		}
		out := make([]GraphElement, 0, 1)

		//TODO: need a way to define the primary ID field
		if id, err := getObjectID(data, class); err == nil {
			fmt.Printf("Vertex %s\n", id)
			dataPB, err := structpb.NewStruct(data)
			if err == nil {
				vert := gripql.Vertex{Gid: id, Label: classID, Data: dataPB}
				out = append(out, GraphElement{Vertex: &vert})
			}

			for name, prop := range class.Properties {
				if ext, ok := prop.Extensions[GraphExtensionTag]; ok {
					gext := ext.(GraphExtension)
					for dstName, backrefLabel := range gext.Backrefs {
						fmt.Printf("Dst: %s\n", dstName)
						if backrefLabelStr, ok := backrefLabel.(string); ok {
							dstIDs, err := getReferenceIDField(data, name)
							if err == nil {
								for _, dstID := range dstIDs {
									edgeOut := gripql.Edge{
										To:    dstID,
										From:  id,
										Label: name,
									}
									edgeIn := gripql.Edge{
										To:    id,
										From:  dstID,
										Label: backrefLabelStr,
									}
									fmt.Printf("edge %s %s %s\n", id, name, dstID)
									out = append(out, GraphElement{OutEdge: &edgeOut})
									out = append(out, GraphElement{InEdge: &edgeIn})
								}
							}
						}
					}
				}
			}
		} else {
			return nil, err
		}
		return out, nil
	}
	return nil, fmt.Errorf("class '%s' not found", classID)
}
