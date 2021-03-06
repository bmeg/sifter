package schema

import (
	"fmt"
	"log"

	"github.com/bmeg/grip/gripql"
	"google.golang.org/protobuf/types/known/structpb"

	multierror "github.com/hashicorp/go-multierror"
)

type GraphElement struct {
	Vertex  *gripql.Vertex
	InEdge  *gripql.Edge
	OutEdge *gripql.Edge
}

func (s Schemas) Generate(classID string, data map[string]interface{}) ([]GraphElement, error) {
	if class, ok := s.Classes[classID]; ok {
		d, err := class.Validate(data)
		if err != nil {
			return nil, err
		}
		out, err := class.Generate(d)
		//log.Printf("%s", out)
		return out, err
	}
	return nil, fmt.Errorf("Class '%s' not found", classID)
}

func (s Schemas) Validate(classID string, data map[string]interface{}) (map[string]interface{}, error) {
	if class, ok := s.Classes[classID]; ok {
		return class.Validate(data)
	}
	return nil, fmt.Errorf("Class '%s' not found in %s", classID, s.GetClasses())
}

func (s Schema) Validate(data map[string]interface{}) (map[string]interface{}, error) {
	out := map[string]interface{}{}
	for k := range s.Props {
		if dataV, ok := data[k]; ok {
			//TODO: typecheck here
			out[k] = dataV
		}
	}

	for _, r := range s.Required {
		if _, ok := out[r]; !ok {
			log.Printf("Not Found %s in %s ", r, data)
			return nil, fmt.Errorf("Required field '%s' in '%s' not found", r, s.ID)
		}
	}

	return out, nil
}

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
	dst := []string{}
	if v, ok := data[l.Name]; ok {
		if vString, ok := v.(string); ok {
			dst = []string{vString}
		} else if vStringArray, ok := v.([]string); ok {
			dst = vStringArray
		} else if vObjectArray, ok := v.([]interface{}); ok {
			for _, x := range vObjectArray {
				if y, ok := x.(string); ok {
					dst = append(dst, y)
				}
			}
		} else if vObject, ok := v.(map[string]interface{}); ok {
			if d, ok := vObject["submitter_id"]; ok { //BUG: this is hard coded to expect Gen3 behavior
				if dStr, ok := d.(string); ok {
					dst = append(dst, dStr)
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
	for _, d := range dst {
		e := gripql.Edge{From: gid, To: d, Label: l.Label}
		out = append(out, GraphElement{OutEdge: &e})
		if l.Backref != "" {
			e := gripql.Edge{To: gid, From: d, Label: l.Backref}
			out = append(out, GraphElement{InEdge: &e})
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
