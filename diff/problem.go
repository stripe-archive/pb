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

type ProblemRemovedServiceMethod struct {
	Name string
}

func (p ProblemRemovedServiceMethod) String() string {
	return fmt.Sprintf("removed service method %s", p.Name)
}

type ProblemChangedService struct {
	Name    string
	OldType string
	NewType string
}

func (p ProblemChangedService) String() string {
	return fmt.Sprintf("changed types for service %s: %s -> %s", p.Name, p.OldType, p.NewType)
}

type ProblemRemovedEnumValue struct {
	Name string
}

func (p ProblemRemovedEnumValue) String() string {
	return fmt.Sprintf("removed enum value %s", p.Name)
}

type ProblemChangeEnumValue struct {
	Name     string
	OldValue int32
	NewValue int32
}

func (p ProblemChangeEnumValue) String() string {
	return fmt.Sprintf("changed enum value %s from %d to %d", p.Name, p.OldValue, p.NewValue)
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

type ProblemRemovedService struct {
	Name string
}

func (p ProblemRemovedService) String() string {
	return fmt.Sprintf("removed service %s", p.Name)
}

type ProblemChangedServiceStreaming struct {
	Name      string
	OldStream *bool
	NewStream *bool
}

func (p ProblemChangedServiceStreaming) String() string {
	return fmt.Sprintf("changed service streaming %s", p.Name)
}
