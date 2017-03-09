package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

func runlint() error {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("reading input: %s", err)
	}

	var req plugin.CodeGeneratorRequest
	var resp plugin.CodeGeneratorResponse

	if err := proto.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("parsing input proto: %s", err)
	}

	if len(req.FileToGenerate) == 0 {
		return fmt.Errorf("no files to generate")
	}

	name := "codegenreq.txt"
	content := proto.MarshalTextString(&req)

	resp.File = []*plugin.CodeGeneratorResponse_File{
		{
			Name:    &name,
			Content: &content,
		},
	}

	// Send back the results.
	data, err = proto.Marshal(&resp)
	if err != nil {
		return fmt.Errorf("failed to marshal output proto: %s", err)
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write output proto: %s", err)
	}
	return nil
}

func main() {
	if err := runlint(); err != nil {
		log.Fatal(err)
	}
}
