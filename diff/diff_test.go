package diff

import (
	"io/ioutil"
	"testing"

	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

func TestDiff(t *testing.T) {
	var prev, curr plugin.CodeGeneratorRequest

	preq, err := ioutil.ReadFile("../testdata/prev/codegen.req")
	if err != nil {
		t.Fatal(err)
	}
	if err := proto.Unmarshal(preq, &prev); err != nil {
		t.Fatalf("parsing prev proto: %s", err)
	}

	creq, err := ioutil.ReadFile("../testdata/curr/codegen.req")
	if err != nil {
		t.Fatal(err)
	}
	if err := proto.Unmarshal(creq, &curr); err != nil {
		t.Fatalf("parsing curr proto: %s", err)
	}

	if err := Diff(&prev, &curr); err != nil {
		t.Error(err)
	}
}
