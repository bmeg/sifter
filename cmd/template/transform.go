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
								"gid":   "{{row.id.string}}",
								"label": "{{row.name}}",
							},
							Rename: map[string]string{
								"object": "data",
							},
						},
					},
					transform.Step{
						FieldProcess: &transform.FieldProcessStep{
							Field: "relations",
							Mapping: map[string]string{
								"src_id": "{{row.gid}}",
							},
							Steps: transform.Pipe{
								transform.Step{
									Project: &transform.ProjectStep{
										Mapping: map[string]interface{}{
											"to":    "{{row.dst_id}}",
											"from":  "{{row.src_id}}",
											"label": "{{row.dst_name}}",
										},
									},
								},
								transform.Step{
									Clean: &transform.CleanStep{
										Fields: []string{"to", "from", "label"},
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
							Fields: []string{"gid", "label", "data"},
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
