package utils

type OrderedMap struct {
	keys   []string
	values map[string]string
}

func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		values: make(map[string]string),
	}
}

func (om *OrderedMap) Set(key, value string) {
	if _, exists := om.values[key]; !exists {
		om.keys = append(om.keys, key)
	}
	om.values[key] = value
}

func (om *OrderedMap) Get(key string) (string, bool) {
	val, exists := om.values[key]
	return val, exists
}

func (om *OrderedMap) Keys() []string {
	return om.keys
}
