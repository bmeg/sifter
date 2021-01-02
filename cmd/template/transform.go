package template

import (
	"github.com/bmeg/sifter/transform"
)

var TransformTemplates = map[string]transform.Pipe{
	"pfb": {
		transform.Step{
			Filter: &transform.FilterStep{
				Field: "{{row.id}}",
				Check: "hasValue",
				Steps: transform.Pipe{
					transform.Step{
						Project: &transform.ProjectStep{
							Mapping: map[string]interface{}{
								"_gid":   "{{row.id.string}}",
								"_label": "{{row.name}}",
							},
							Rename: map[string]string{
								"object": "_data",
							},
						},
					},
					transform.Step{
						FieldProcess: &transform.FieldProcessStep{
							Field: "relations",
							Mapping: map[string]string{
								"src_id": "{{row._gid}}",
							},
							Steps: transform.Pipe{
								transform.Step{
									Project: &transform.ProjectStep{
										Mapping: map[string]interface{}{
											"_to":    "{{row.dst_id}}",
											"_from":  "{{row.src_id}}",
											"_label": "{{row.dst_name}}",
										},
									},
								},
								transform.Step{
									Clean: &transform.CleanStep{
										Fields: []string{"_to", "_from", "_label"},
									},
								},
								transform.Step{
									Emit: &transform.EmitStep{Name: "edges"},
								},
							},
						},
					},
					transform.Step{
						Clean: &transform.CleanStep{
							Fields: []string{"_gid", "_label", "_data"},
						},
					},
					transform.Step{
						Emit: &transform.EmitStep{Name: "vertices"},
					},
				},
			},
		},
	},
}
