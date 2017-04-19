package retag

import (
	"bytes"

	"github.com/golang/protobuf/protoc-gen-go/generator"
)

func init() {
	generator.RegisterPlugin(new(retag))
}

type retag struct {
	gen  *generator.Generator
	tags map[string]string
}

// Name returns the name of this plugin, "settag"
func (s *retag) Name() string {
	return "retag"
}

// Init initializes the plugin.
func (s *retag) Init(gen *generator.Generator) {
	s.gen = gen
}

func (s *retag) P(args ...interface{}) { s.gen.P(args...) }

// Generate generates code for the services in the given file.
func (s *retag) Generate(file *generator.FileDescriptor) {
	s.reSetTags(file)
}

// GenerateImports generates the import declaration for this file.
func (s *retag) GenerateImports(file *generator.FileDescriptor) {}

func (s *retag) reSetTags(file *generator.FileDescriptor) {
	//pf := s.gen.Request.ProtoFile[len(s.gen.Request.ProtoFile)-1]

	//log.Infof("pf:%+v", pf)

	befor := s.gen.Buffer.Bytes()
	after := bytes.Replace(befor, []byte("json"), []byte("xml"), -1)
	s.gen.Buffer.Reset()
	s.gen.Buffer.Write(after)
	// log.Info("------------无奈的分割线------------------")
	// log.Info(string(s.gen.Buffer.Bytes()))
	// log.Info("------------无奈的分割线------------------")
}
