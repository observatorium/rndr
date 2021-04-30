package rndrapi

type Groups  map[string][]Resource

type Resource struct {
	Item   string
	Object []byte
}
