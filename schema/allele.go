
package schema

import (
  "fmt"
  "crypto/sha1"
  "github.com/bmeg/grip/gripql"
  "github.com/bmeg/grip/protoutil"
)

type Allele struct {
  AlleleId string `json:"allele_id"`
  Genome   string `json:"genome"`
  Chromosome string `json:"chromosome"`
  Start    uint64      `json:"start"`
  End      uint64      `json:"end"`
  Strand   string   `json:"strand"`
  ReferenceBases string `json:"reference_bases"`
  AlternateBases string `json:"alternate_bases"`
  HugoSymbol     string `json:hugo_symbol`
  EnsemblTranscript string `json:"ensembl_transcript"`
  Type           string  `json:"type"`
  Effect string `json:"effect"`
  DBSNP_RS string `json:"dbSNP_RS"`
}


func (al *Allele) Render() ([]*gripql.Vertex, []*gripql.Edge) {
  data := map[string]interface{}{
      "genome" : al.Genome,
      "chromosome" : al.Chromosome,
      "start" : al.Start,
      "end" : al.End,
      "reference_bases" : al.ReferenceBases,
      "alternate_bases" : al.AlternateBases,
  }

  if len(al.HugoSymbol) > 0 {
    data["hugo_symbol"] = al.HugoSymbol
  }
  if len(al.EnsemblTranscript) > 0 {
    data["ensembl_transcript"] = al.EnsemblTranscript
  }
  if len(al.Type) > 0 {
    data["type"] = al.Type
  }
  if len(al.Effect) > 0 {
    data["effect"] = al.Effect
  }
  if len(al.DBSNP_RS) > 0 {
    data["dbSNP_RS"] = al.DBSNP_RS
  }

  a := gripql.Vertex{Gid:al.ID(), Label:"Allele", Data:protoutil.AsStruct(data)}
  return []*gripql.Vertex{&a}, []*gripql.Edge{}
}

func (al *Allele) ID() string {
  id := fmt.Sprintf("%s:%s:%d:%d:%s:%s", al.Genome, al.Chromosome,
                                 al.Start, al.End,
                                 al.ReferenceBases,
                                 al.AlternateBases)
  return fmt.Sprintf("%x", sha1.Sum([]byte(id)))
}
