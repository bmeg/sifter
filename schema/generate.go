package schema

import (
	"fmt"
	"strings"

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

type reference struct {
	dstID   string
	dstType string
}

func getReferenceIDField(data map[string]any, fieldName string) ([]reference, error) {
	out := []reference{}
	if d, ok := data[fieldName]; ok {
		//fmt.Printf("Dest id field %#v\n", d)
		if idStr, ok := d.(string); ok {
			out = append(out, reference{dstID: idStr})
		} else if idArray, ok := d.([]any); ok {
			for _, g := range idArray {
				if gStr, ok := g.(string); ok {
					out = append(out, reference{dstID: gStr})
				} else if gMap, ok := g.(map[string]any); ok {
					if id, ok := gMap["id"]; ok {
						if idStr, ok := id.(string); ok {
							out = append(out, reference{dstID: idStr})
						}
					} else if id, ok := gMap["reference"]; ok {
						//reference is a FHIR style id pointer, { "reference": "Type/id" }
						if idStr, ok := id.(string); ok {
							a := strings.Split(idStr, "/")
							if len(a) > 1 {
								out = append(out, reference{dstID: a[1], dstType: a[0]})
							}
						}
					} else {
						fmt.Printf("Not found in %#v\n", gMap)
					}
				}
			}
		} else if idMap, ok := d.(map[string]any); ok {
			if id, ok := idMap["id"]; ok {
				if idStr, ok := id.(string); ok {
					out = append(out, reference{dstID: idStr})
				}
			} else if id, ok := idMap["reference"]; ok {
				//reference is a FHIR style id pointer, { "reference": "Type/id" }
				if idStr, ok := id.(string); ok {
					a := strings.Split(idStr, "/")
					if len(a) > 1 {
						out = append(out, reference{dstID: a[1], dstType: a[0]})
					}
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
			//fmt.Printf("Vertex %s\n", id)
			dataPB, err := structpb.NewStruct(data)
			if err == nil {
				vert := gripql.Vertex{Gid: id, Label: classID, Data: dataPB}
				out = append(out, GraphElement{Vertex: &vert})
			}

			for name, prop := range class.Properties {
				if ext, ok := prop.Extensions[GraphExtensionTag]; ok {
					//fmt.Printf("Extension: %#v\n", ext)
					gext := ext.(GraphExtension)
					dstIDs, err := getReferenceIDField(data, name)
					if err == nil {
						for _, dstID := range dstIDs {
							for _, target := range gext.Targets {
								if target.Schema.Title == dstID.dstType || dstID.dstType == "" {
									edgeOut := gripql.Edge{
										To:    dstID.dstID,
										From:  id,
										Label: name,
									}
									out = append(out, GraphElement{OutEdge: &edgeOut})
									if target.Backref != "" {
										edgeIn := gripql.Edge{
											To:    id,
											From:  dstID.dstID,
											Label: target.Backref,
										}
										out = append(out, GraphElement{InEdge: &edgeIn})
									}
								}
							}
						}
					} else {
						return nil, err
					}
				}
			}
		}
		return out, nil
	}
	return nil, fmt.Errorf("class '%s' not found", classID)
}
