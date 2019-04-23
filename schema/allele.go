
package schema


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
