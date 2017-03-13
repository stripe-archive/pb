package diff

import (
	"fmt"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

type ProblemChangedFieldType struct {
	Number  int32
	Field   string
	OldType *descriptor.FieldDescriptorProto_Type
	NewType *descriptor.FieldDescriptorProto_Type
}

func (p ProblemChangedFieldType) String() string {
	return fmt.Sprintf("changed types for field %s: %s -> %s", p.Field, p.OldType, p.NewType)
}

type ProblemRemovedField struct {
	Field string
}

func (p ProblemRemovedField) String() string {
	return fmt.Sprintf("removed field %s", p.Field)
}

type ProblemRemovedEnum struct {
	Enum string
}

func (p ProblemRemovedEnum) String() string {
	return fmt.Sprintf("removed enum %s", p.Enum)
}

type ProblemRemovedMessage struct {
	Message string
}

func (p ProblemRemovedMessage) String() string {
	return fmt.Sprintf("removed message %s", p.Message)
}

type ProblemRemovedFile struct {
	File string
}

func (p ProblemRemovedFile) String() string {
	return fmt.Sprintf("removed message %s", p.File)
}
