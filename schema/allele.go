
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
  id := fmt.Sprintf("%s:%s:%d:%d:%s:%s", al.Genome, al.Chromosome,
                                 al.Start, al.End,
                                 al.ReferenceBases,
                                 al.AlternateBases)
  a := gripql.Vertex{Gid:fmt.Sprintf("%x", sha1.Sum([]byte(id))), Label:"Allele", Data:protoutil.AsStruct(map[string]interface{}{
      "genome" : al.Genome,
      "chromosome" : al.Chromosome,
      "start" : al.Start,
      "end" : al.End,
      "string" : al.Strand,
      "reference_bases" : al.ReferenceBases,
      "alternate_bases" : al.AlternateBases,
      "hugo_symbol" : al.HugoSymbol,
      "ensembl_transcript" : al.EnsemblTranscript,
      "type" : al.Type,
      "effect" : al.Effect,
      "dbSNP_RS" : al.DBSNP_RS,
  })}
  return []*gripql.Vertex{&a}, []*gripql.Edge{}
}
