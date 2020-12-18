package template

import (
	"github.com/bmeg/sifter/extractors"
	"github.com/bmeg/sifter/transform"
)

type addTransform func(extractors.Extractor, transform.Pipe) extractors.Extractor

var ExtractTemplates = map[string]extractors.Extractor{
	"avro": {AvroLoad: &extractors.AvroLoadStep{
		Input: "{{inputs.input}}",
	}},
}

var ExtractorDecorate = map[string]addTransform{
	"avro": func(e extractors.Extractor, p transform.Pipe) extractors.Extractor {
		e.AvroLoad.Transform = p
		return e
	},
}
