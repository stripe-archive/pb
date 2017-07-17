package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/stackmachine/pb/diff"
)

// protodiff -I . app.proto
// Can I somehow embed protoc in protodiff? I probably shouldn't
// Junit output maybe?
func main() {
	var err error
	flag.Parse()
	// TODO: If you call protodiff -h this won't work
	if flag.NArg() == 0 {
		err = runecho()
	} else {
		err = rundiff()
	}
	if err != nil {
		log.Fatal(err)
	}
}

func runecho() error {
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

	// TODO: Generate a unique name here
	name := "codegen.req"
	content := string(data)
	// marsh := jsonpb.Marshaler{}
	// content, err := marsh.MarshalToString(&req)
	// if len(req.FileToGenerate) == 0 {
	// 	return fmt.Errorf("failed to serialize req: %s", err)
	// }

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

func rundiff() error {
	var prev, curr plugin.CodeGeneratorRequest
	var cmd *exec.Cmd

	// Create temporary directory
	dir, err := ioutil.TempDir("", "protodiff")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up

	execpath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	protofile := flag.Arg(0)
	echopath := "--plugin=protoc-gen-echo=" + execpath
	prevdir := filepath.Join(dir, "prev")
	prevgen := filepath.Join(prevdir, "codegen.req")
	currdir := filepath.Join(dir, "curr")
	currgen := filepath.Join(currdir, "codegen.req")

	if err := os.Mkdir(currdir, 0755); err != nil {
		log.Fatal(err)
	}

	if err := os.Mkdir(prevdir, 0755); err != nil {
		log.Fatal(err)
	}

	// TODO: Read in this commit from the environment
	startrev := "ad60c723cdd4f40a52f5e64d6d6866471b229b98"
	cmd = exec.Command("git", "checkout", startrev)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout for %s failed: %s %s", startrev, err, out)
	}

	// Run protoc
	cmd = exec.Command("protoc", echopath, "--echo_out="+prevdir, protofile)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("generating protoc output failed: %s %s %s", startrev, err, out)
	}

	// TODO: Read in this commit from the environment
	endrev := "d452c6bd95548cd709ede11a52937b95c172ccaa"
	cmd = exec.Command("git", "checkout", endrev)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout for %s failed: %s %s", endrev, err, out)
	}

	// Run protoc again
	cmd = exec.Command("protoc", echopath, "--echo_out="+currdir, protofile)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("generating protoc output failed: %s %s %s", endrev, err, out)
	}

	preq, err := ioutil.ReadFile(prevgen)
	if err != nil {
		return fmt.Errorf("reading protoc err")
	}
	if err := proto.Unmarshal(preq, &prev); err != nil {
		fmt.Errorf("parsing %s at revision %s failed: %s", err)
	}

	creq, err := ioutil.ReadFile(currgen)
	if err != nil {
		log.Fatal(err)
	}
	if err := proto.Unmarshal(creq, &curr); err != nil {
		log.Fatalf("parsing curr proto: %s", err)
	}

	report, err := diff.Diff(&prev, &curr)
	fmt.Println("Ran report...")
	if len(report.Changes) > 0 {
		for _, change := range report.Changes {
			fmt.Println(change.String())
		}
		os.Exit(1)
	}
}
