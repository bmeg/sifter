package transform

type TableReplace struct {
	Field  string `json:"field"`
	Target string `json:"target"`
}

type tableReplaceInst struct {
	config *TableReplace
	table  map[string]string
}

func (tb *TableReplace) Init(x map[string]string) *tableReplaceInst {
	return &tableReplaceInst{tb, x}
}

func (tr *tableReplaceInst) Close() {}

func (tr *tableReplaceInst) Process(i map[string]interface{}) []map[string]interface{} {

	if _, ok := i[tr.config.Field]; ok {
		out := map[string]interface{}{}
		for k, v := range i {
			if k == tr.config.Field {
				d := k
				if tr.config.Target != "" {
					d = tr.config.Target
				}
				if x, ok := v.(string); ok {
					if n, ok := tr.table[x]; ok {
						out[d] = n
					} else {
						out[d] = x
					}
				} else if x, ok := v.([]interface{}); ok {
					o := []interface{}{}
					for _, y := range x {
						if z, ok := y.(string); ok {
							if n, ok := tr.table[z]; ok {
								o = append(o, n)
							} else {
								o = append(o, z)
							}
						}
					}
					out[d] = o
				} else {
					out[d] = v
				}
			} else {
				out[k] = v
			}
		}
		return []map[string]any{out}
	}
	return []map[string]any{i}
}
