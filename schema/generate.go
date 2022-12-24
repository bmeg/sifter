package schema

import (
	"fmt"

	"github.com/bmeg/grip/gripql"
)

type GraphElement struct {
	Vertex  *gripql.Vertex
	InEdge  *gripql.Edge
	OutEdge *gripql.Edge
	Field   string
}

func (s GraphSchema) Generate(classID string, data map[string]interface{}) ([]GraphElement, error) {
	if class, ok := s.Classes[classID]; ok {
		err := class.Validate(data)
		if err != nil {
			return nil, err
		}

		out := make([]GraphElement, 0, 1)

		return out, err
	}
	return nil, fmt.Errorf("class '%s' not found", classID)
}

/*
func (l Link) Generate(gid string, data map[string]interface{}) ([]GraphElement, error) {
	if l.Name == "" && len(l.Subgroup) > 0 {
		out := make([]GraphElement, 0, 1)
		for _, s := range l.Subgroup {
			e, err := s.Generate(gid, data)
			if err == nil {
				out = append(out, e...)
			}
		}
		return out, nil
	}
	out := make([]GraphElement, 0, 1)

	type dstData struct {
		id   string
		data map[string]interface{}
	}

	dst := []dstData{}
	if v, ok := data[l.Name]; ok {
		if vString, ok := v.(string); ok {
			dst = append(dst, dstData{id: vString})
		} else if vStringArray, ok := v.([]string); ok {
			for i := range vStringArray {
				dst = append(dst, dstData{id: vStringArray[i]})
			}
		} else if vObjectArray, ok := v.([]interface{}); ok {
			for _, x := range vObjectArray {
				if y, ok := x.(string); ok {
					dst = append(dst, dstData{id: y})
				} else if d, ok := x.(map[string]interface{}); ok {
					//log.Printf("Found structure: %#v", d)
					if l.TargetField != "" {
						if t, ok := d[l.TargetField]; ok {
							if tStr, ok := t.(string); ok {
								td := map[string]interface{}{}
								if l.Properties != nil {
									for k := range l.Properties {
										if v, ok := d[k]; ok {
											td[k] = v
										}
									}
								}
								//log.Printf("Copy: %#v", td)
								dst = append(dst, dstData{id: tStr, data: td})
							}
						}
					}
				} else {
					log.Printf("Unknown list element")
				}
			}
		} else if vObject, ok := v.(map[string]interface{}); ok {
			if d, ok := vObject["submitter_id"]; ok { //BUG: this is hard coded to expect Gen3 behavior
				if dStr, ok := d.(string); ok {
					dst = append(dst, dstData{id: dStr})
				}
			}
		} else {
			log.Printf("Class link field %s Unknown property type: %#v", l.Name, v)
		}
	}
	/*
	  //TODO: This code tries to get the link values using the types found in the schema definition
	  //which is technically correct, but much harder. This code currently breaks on cases where schema uses
	  //`anyOf` definitions. So, for now using the previously mentioned version that doesnt check the schema,
	  //and assumes a string or list of strings
	  if s.Props[l.Name].Element.Type.Type == "string" {
	    if x, ok := data[l.Name].(string); ok {
	      dst = []string{ x }
	    } else {
	      log.Printf("Wrong: %s", data[l.Name])
	    }
	  } else if s.Props[l.Name].Element.Type.Type == "array" {
	    if x, ok := data[l.Name].([]string); ok {
	      dst = x
	    } else if x, ok := data[l.Name].([]interface{}); ok {
	      for _, a := range x {
	        if aStr, ok := a.(string); ok {
	          dst = append(dst, aStr)
	        }
	      }
	    } else {
	      log.Printf("Unknown type: %s %s %s", data[l.Name], l.Name, data)
	    }
	  } else {
	    log.Printf("Class %s field %s Unknown property type: %s", s.Id, l.Name, s.Props[l.Name].Element.Type.Type)
	  }
*/
/*
	for _, d := range dst {
		e := gripql.Edge{From: gid, To: d.id, Label: l.Label}
		if d.data != nil {
			ds, _ := structpb.NewStruct(d.data)
			e.Data = ds
		}
		out = append(out, GraphElement{OutEdge: &e, Field: l.Name})
		if l.Backref != "" {
			e := gripql.Edge{To: gid, From: d.id, Label: l.Backref}
			out = append(out, GraphElement{InEdge: &e, Field: l.Name})
		}
	}
	return out, nil
}

func (s Schema) Generate(data map[string]interface{}) ([]GraphElement, error) {
	out := make([]GraphElement, 0, 1+len(s.Links)*2)
	outData := map[string]interface{}{}
	var result error

	gid := ""
	for k, v := range s.Props {
		if v.Element.SystemAlias == "node_id" {
			if dv, ok := data[k]; ok {
				if ks, ok := dv.(string); ok {
					gid = ks
				}
			} else {
				err := fmt.Errorf("node_id field '%s' not in %s data", k, s.ID)
				log.Printf("node_id field '%s' not in %s data", k, s.ID)
				result = multierror.Append(result, err)
			}
		} else {
			if x, ok := data[k]; ok {
				outData[k] = x
			}
		}
	}
	if s.Edge != nil {
		if tID, ok := data[s.Edge.To]; ok {
			if tIDStr, ok := tID.(string); ok {
				if fID, ok := data[s.Edge.From]; ok {
					if fIDStr, ok := fID.(string); ok {
						ds, _ := structpb.NewStruct(outData)
						e := gripql.Edge{Gid: gid, To: tIDStr, From: fIDStr, Label: s.Edge.Label, Data: ds}
						out = append(out, GraphElement{OutEdge: &e})
					}
				} else {
					log.Printf("Edge from field '%s' missing", s.Edge.From)
					err := fmt.Errorf("Edge from field '%s' missing", s.Edge.From)
					result = multierror.Append(result, err)
				}
			}
		} else {
			log.Printf("Edge to field '%s' missing", s.Edge.To)
		}
	} else {
		if gid == "" {
			log.Printf("GID not found for %s - %s", s.ID, s.Title)
		}
		ds, _ := structpb.NewStruct(outData)
		v := gripql.Vertex{Gid: gid, Label: s.Title, Data: ds}

		out = append(out, GraphElement{Vertex: &v})

		for _, l := range s.Links {
			lo, err := l.Generate(gid, outData)
			if err == nil {
				out = append(out, lo...)
			} else {
				log.Printf("Link error: %s", err)
			}
		}
	}
	return out, result
}

*/
