# Protocol Buffer utilities

## pblint

### Installation

    go get -u github.com/stackmachine/pb/cmd/protoc-gen-lint

### Usage

    protoc --lint_out=. helloworld.proto 

## pbdiff

Verify protocol buffer changes are backwards compatible.

### Installation

    go get -u github.com/stackmachine/pb/cmd/protoc-gen-echo
    go get -u github.com/stackmachine/pb/cmd/pbdiff

### Usage

    mkdir -p head prev
    protoc --echo_out=head/ head.proto 
    protoc --echo_out=prev/ prev.proto 
    pbdiff head/codegen.req pre/codegen.req
