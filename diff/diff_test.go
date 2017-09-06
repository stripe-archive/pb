package diff

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

// Given a directory name and a .proto file, generate a FileDescriptorSet.
//
// Requires protoc to be installed.
func generateFileSet(t *testing.T, prefix, name string) descriptor.FileDescriptorSet {
	var fds descriptor.FileDescriptorSet
	protoDir := filepath.Join("testdata", prefix)
	protoFile := filepath.Join(protoDir, name+".proto")
	fdsDir := filepath.Join("testdata", prefix+"_fds")
	fdsFile := filepath.Join(fdsDir, name+".fds")
	if err := os.MkdirAll(fdsDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Run protoc
	cmd := exec.Command("protoc", "-o", fdsFile, protoFile)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("protoc failed: %s %s", err, out)
	}
	blob, err := ioutil.ReadFile(fdsFile)
	if err != nil {
		t.Fatal(err)
	}
	if err := proto.Unmarshal(blob, &fds); err != nil {
		t.Fatalf("parsing prev proto: %s", err)
	}
	return fds
}

func TestDiffing(t *testing.T) {
	files := map[string]string{
		"changed_field_type":  "changed types for field name: TYPE_STRING -> TYPE_BOOL",
		"changed_field_label": "changed label for field name: LABEL_OPTIONAL -> LABEL_REPEATED",
		"removed_field":       "removed field name",
		"removed_message":     "removed message HelloRequest",
		"removed_enum":        "removed enum FOO",
		"removed_enum_field":  "removed enum value bat",
	}
	for name, problem := range files {
		t.Run(name, func(t *testing.T) {
			prev := generateFileSet(t, "previous", name)
			curr := generateFileSet(t, "current", name)
			// Won't work with --include_imports
			report, err := DiffSet(&prev, &curr)
			if err == nil {
				t.Fatal("expected diff to have an error")
			}
			if len(report.Changes) == 0 {
				t.Fatal("expected report to have at least one problem")
			}
			if len(report.Changes) > 1 {
				t.Errorf("expected report to have one problem, has %d: %v", len(report.Changes), report)
			}
			if report.Changes[0].String() != problem {
				t.Errorf("expected problem: %s", problem)
				t.Errorf("  actual problem: %s", report.Changes[0].String())
			}
		})
	}
}

// A bit of a difficult test setup, as we need to have protoc and
// protoc-gen-echo installed
func difftest(t *testing.T, prevproto, currproto, problem string) {
	t.Parallel()
	var prev, curr plugin.CodeGeneratorRequest
	var cmd *exec.Cmd

	// Create temporary directory
	dir, err := ioutil.TempDir("", "difftest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up

	protofn := filepath.Join(dir, "test.proto")
	prevdir := filepath.Join(dir, "prev")
	prevgen := filepath.Join(prevdir, "codegen.req")
	currdir := filepath.Join(dir, "curr")
	currgen := filepath.Join(currdir, "codegen.req")

	if err := os.Mkdir(currdir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(prevdir, 0755); err != nil {
		t.Fatal(err)
	}

	// Save the first proto
	if err := ioutil.WriteFile(protofn, []byte(prevproto), 0666); err != nil {
		t.Fatal(err)
	}

	// Run protoc
	cmd = exec.Command("protoc", "-I", ".", "--echo_out=./prev/", "test.proto")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("protoc prev failed: %s %s", err, out)
	}

	// Update the first protofile
	if err := ioutil.WriteFile(protofn, []byte(currproto), 0666); err != nil {
		t.Fatal(err)
	}

	// Run protoc again
	cmd = exec.Command("protoc", "-I", ".", "--echo_out=./curr/", "test.proto")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("protoc curr failed: %s %s", err, out)
	}

	preq, err := ioutil.ReadFile(prevgen)
	if err != nil {
		t.Fatal(err)
	}
	if err := proto.Unmarshal(preq, &prev); err != nil {
		t.Fatalf("parsing prev proto: %s", err)
	}

	creq, err := ioutil.ReadFile(currgen)
	if err != nil {
		t.Fatal(err)
	}
	if err := proto.Unmarshal(creq, &curr); err != nil {
		t.Fatalf("parsing curr proto: %s", err)
	}

	report, err := Diff(&prev, &curr)
	if err == nil {
		t.Fatal("expected diff to have an error")
	}
	if len(report.Changes) == 0 {
		t.Fatal("expected report to have at least one problem")
	}
	if len(report.Changes) > 1 {
		t.Errorf("expected report to have one problem, has %d: %v", len(report.Changes), report)
	}
	if report.Changes[0].String() != problem {
		t.Errorf("expected problem: %s", problem)
		t.Errorf("  actual problem: %s", report.Changes[0].String())
	}
}

func TestChangeEnumValue(t *testing.T) {
	difftest(t,
		`
syntax = "proto3";
package helloworld;
enum FOO {
  bar = 0;
  bat = 1;
}
`,
		`
syntax = "proto3";
package helloworld;
enum FOO {
  bar = 0;
  bat = 2;
}
`,
		"changed enum value bat from 1 to 2",
	)
}

func TestRemovedService(t *testing.T) {
	difftest(t,
		`
syntax = "proto3";
package helloworld;
service Foo {
}
`,
		`
syntax = "proto3";
package helloworld;
`,
		"removed service Foo",
	)
}

func TestRemovedServiceMethod(t *testing.T) {
	difftest(t,
		`
syntax = "proto3";
package helloworld;
message Empty {};
service Foo {
  rpc Bar(Empty) returns (Empty) {}
}
`,
		`
syntax = "proto3";
package helloworld;
message Empty {};
service Foo {
}
`,
		"removed service method Bar",
	)
}

func TestChangedServiceInput(t *testing.T) {
	difftest(t,
		`
syntax = "proto3";
package helloworld;

message FooRequest {};
message FooResponse {};
message BarRequest {};

service Foo {
  rpc Invoke(FooRequest) returns (FooResponse) {}
}
`,
		`
syntax = "proto3";
package helloworld;

message FooRequest {};
message BarRequest {};
message FooResponse {};

service Foo {
  rpc Invoke(BarRequest) returns (FooResponse) {}
}
`,
		"changed types for service Invoke: .helloworld.FooRequest -> .helloworld.BarRequest",
	)
}

func TestChangedServiceOutput(t *testing.T) {
	difftest(t,
		`
syntax = "proto3";
package helloworld;

message FooRequest {};
message FooResponse {};
message BarResponse {};

service Foo {
  rpc Invoke(FooRequest) returns (FooResponse) {}
}
`,
		`
syntax = "proto3";
package helloworld;

message FooRequest {};
message FooResponse {};
message BarResponse {};

service Foo {
  rpc Invoke(FooRequest) returns (BarResponse) {}
}
`,
		"changed types for service Invoke: .helloworld.FooResponse -> .helloworld.BarResponse",
	)
}

func TestChangedClientStreaming(t *testing.T) {
	difftest(t,
		`
syntax = "proto3";
package helloworld;

message Empty {};

service Foo {
  rpc Invoke(Empty) returns (Empty) {}
}
`,
		`
syntax = "proto3";
package helloworld;

message Empty {};

service Foo {
  rpc Invoke(stream Empty) returns (Empty) {}
}
`,
		"changed service streaming Invoke",
	)
}

func TestChangedPackage(t *testing.T) {
	difftest(t,
		`
syntax = "proto3";
package foo;
`,
		`
syntax = "proto3";
package bar;
`,
		"changed package from foo to bar",
	)
}
