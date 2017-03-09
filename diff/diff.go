package diff

import (
	"fmt"
	"log"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

func Diff(previous, current *plugin.CodeGeneratorRequest) error {
	curr := map[string]*descriptor.FileDescriptorProto{}

	for _, protoFile := range current.ProtoFile {
		curr[*protoFile.Name] = protoFile
	}

	for _, protoFile := range previous.ProtoFile {
		next, exists := curr[*protoFile.Name]
		if !exists {
			continue
		}
		if err := diffFile(protoFile, next); err != nil {
			return err
		}
	}

	return nil
}

func diffFile(previous, current *descriptor.FileDescriptorProto) error {
	curr := map[string]*descriptor.DescriptorProto{}

	for _, msg := range current.MessageType {
		curr[*msg.Name] = msg
	}

	for _, msg := range previous.MessageType {
		next, exists := curr[*msg.Name]
		if !exists {
			continue
		}
		if err := diffMsg(msg, next); err != nil {
			return err
		}
	}

	return nil
}

func diffMsg(previous, current *descriptor.DescriptorProto) error {
	curr := map[string]*descriptor.FieldDescriptorProto{}

	for _, field := range current.Field {
		curr[*field.Name] = field
	}

	for _, field := range previous.Field {
		next, exists := curr[*field.Name]
		log.Printf("%v", field)
		log.Printf("%v", next)
		if !exists {
			continue
		}
		if *field.Type != *next.Type {
			return fmt.Errorf("changed types for field %s: %s -> %s", *field.Name, field.Type, next.Type)
		}
	}

	return nil
}
