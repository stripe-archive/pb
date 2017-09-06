package diff

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
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
		"changed_client_streaming": "changed service streaming Invoke",
		"changed_enum_value":       "changed enum value bat from 1 to 2",
		"changed_field_label":      "changed label for field name: LABEL_OPTIONAL -> LABEL_REPEATED",
		"changed_field_type":       "changed types for field name: TYPE_STRING -> TYPE_BOOL",
		"changed_package":          "descriptor_set_out",
		"changed_service_input":    "changed types for service Invoke: .helloworld.FooRequest -> .helloworld.BarRequest",
		"changed_service_output":   "changed types for service Invoke: .helloworld.FooResponse -> .helloworld.BarResponse",
		"removed_enum":             "removed enum FOO",
		"removed_enum_field":       "removed enum value bat",
		"removed_field":            "removed field name",
		"removed_message":          "removed message HelloRequest",
		"removed_service":          "removed service Foo",
		"removed_service_method":   "removed service method Bar",
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
