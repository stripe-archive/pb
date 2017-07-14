package diff

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

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
	if len(report.Problems) == 0 {
		t.Fatal("expected report to have at least one problem")
	}
	if len(report.Problems) > 1 {
		t.Errorf("expected report to have one problem, has %d: %v", len(report.Problems), report)
	}
	if report.Problems[0].String() != problem {
		t.Errorf("expected problem: %s", problem)
		t.Errorf("  actual problem: %s", report.Problems[0].String())
	}
}

func TestChangedFieldType(t *testing.T) {
	difftest(t,
		`
syntax = "proto3";
package helloworld;
message HelloRequest {
  string name = 1;
}
`,
		`
syntax = "proto3";
package helloworld;
message HelloRequest {
  bool name = 1;
}
`,
		"changed types for field name: TYPE_STRING -> TYPE_BOOL",
	)
}

func TestRemovedField(t *testing.T) {
	difftest(t,
		`
syntax = "proto3";
package helloworld;
message HelloRequest {
  string name = 1;
}
`,
		`
syntax = "proto3";
package helloworld;
message HelloRequest {
}
`,
		"removed field name",
	)
}

func TestRemovedMessage(t *testing.T) {
	difftest(t,
		`
syntax = "proto3";
package helloworld;
message HelloRequest {
  string name = 1;
}
`,
		`
syntax = "proto3";
package helloworld;
`,
		"removed message HelloRequest",
	)
}

func TestRemovedEnum(t *testing.T) {
	difftest(t,
		`
syntax = "proto3";
package helloworld;

enum FOO {
  bar = 0;
}
`,
		`
syntax = "proto3";
package helloworld;
`,
		"removed enum FOO",
	)
}

func TestRemovedEnumField(t *testing.T) {
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
}
`,
		"removed enum value bat",
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
