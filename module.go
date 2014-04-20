package jantar

const (
	MODULE_TEMPLATE_MANAGER = iota
	MODULE_ROUTER           = iota
)

var (
	moduleData = make(map[int]interface{})
)

func setModule(key int, value interface{}) {
	moduleData[key] = value
}

func GetModule(key int) interface{} {
	return moduleData[key]
}

func GetModuleOk(key int) (interface{}, bool) {
	data, ok := moduleData[key]
	return data, ok
}
