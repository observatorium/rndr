package golang

type TemplateAPI struct {
	// Default is a <full package path>.<public function> to be invoked to get valid struct filled in Entry
	Default string
	// Struct is a <full package path>.<public struct> name that should be used as the entry point for API struct.
	Struct string
}
