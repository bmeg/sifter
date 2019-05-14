
package schema

import (
  "log"
  "fmt"
  "github.com/bmeg/grip/gripql"
  "github.com/bmeg/grip/protoutil"

)

type GraphElement struct {
  Vertex *gripql.Vertex
  Edge   *gripql.Edge
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
  return nil, fmt.Errorf("Class %s not found", classID )
}


func (s Schema) Validate(data map[string]interface{}) (map[string]interface{}, error) {
  out := map[string]interface{}{}
  for k, _ := range s.Props {
    if dataV, ok := data[k]; ok {
      //TODO: typecheck here
      out[k] = dataV
    }
  }

  for _, r := range s.Required {
    if _, ok := out[r]; !ok {
      return nil, fmt.Errorf("Required %s not found", r)
    }
  }

  return out, nil
}


func (s Schema) Generate(data map[string]interface{}) ([]GraphElement, error) {
  out := make([]GraphElement, 0, 1+len(s.Links)*2)
  outData := map[string]interface{}{}

  gid := ""
  for k, v := range s.Props {
    if v.Element.SystemAlias == "node_id" {
      if dv, ok := data[k]; ok {
        if ks, ok := dv.(string); ok {
          gid = ks
        }
      }
    } else {
      if x, ok := data[k]; ok {
        outData[k] = x
      }
    }
  }
  ds := protoutil.AsStruct(outData)
  v := gripql.Vertex{Gid: gid, Label: s.Title, Data:ds}

  out = append(out, GraphElement{Vertex:&v})

  for _, l := range s.Links {
    dst := []string{}
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
    }
    for _, d := range dst {
      e := gripql.Edge{From:gid, To: d, Label:l.Label}
      out = append(out, GraphElement{Edge:&e})
      if l.Backref != "" {
        e := gripql.Edge{To:gid, From: d, Label:l.Backref}
        out = append(out, GraphElement{Edge:&e})
      }
    }
  }
  return out, nil
}