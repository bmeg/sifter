package transform

type LookupTable map[string]string

func (jp *LookupTable) LookupValue(k string) (string, bool) {
	s, ok := (*jp)[k]
	return s, ok
}

func (jp *LookupTable) LookupRecord(k string) (map[string]any, bool) {
	if x, ok := (*jp)[k]; ok {
		return map[string]any{"value": x}, true
	}
	return nil, false
}
