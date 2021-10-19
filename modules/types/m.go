package types

type M map[string]interface{}

func (m M) Get(key string) interface{} {
	return m[key]
}

func (m M) ValInt(key string) int {
	if v := m.Get(key); v == nil {
		return 0
	} else {
		if i, ok := v.(float64); ok {
			return int(i)
		}
		return v.(int)
	}
}
func (m M) ValString(key string) string {
	if v := m.Get(key); v == nil {
		return ""
	} else {
		return v.(string)
	}
}

func (m M) Put(key string, v interface{}) {
	m[key] = v
}
