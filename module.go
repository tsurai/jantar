package jantar

// Constants for accessing jantars core modules
const (
	moduleFirst           = iota
	ModuleTemplateManager = iota
	ModuleRouter          = iota
	moduleLast            = iota
)

var (
	moduleData = make(map[int]interface{})
)

func setModule(key int, value interface{}) {
	moduleData[key] = value
}

// GetModule returns the module specified by key. Returns nil for an invalid key
func GetModule(key int) interface{} {
	if key > moduleFirst && key < moduleLast {
		return moduleData[key]
	}

	return nil
}
