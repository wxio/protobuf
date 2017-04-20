# retag

a plugin for protoc-gen-go to reset struct tags.

## why

golang protobuf doesn't support custom tags to generated structs. this plugin help to set custom tags to generated protobuf file.

## install

```shell
git clone https://github.com/qianlnk/protobuf.git $GOPATH/src/github.com/golang/
go install $GOPATH/src/github.com/golang/protobuf/protoc-gen-go
```

## usage

Add a comment with syntax `//｀custom_tag1:custom_value1 custom_tag2:custom_value2｀` after fields.

Example:

```proto
syntax = "proto3";

package staff;

message Staff {
    string ID = 1;    //`json:"id,omitempty"   xml:"id,omitempty"`
    string Name = 2;  //`json:"name,omitempty" xml:"name,omitempty"`
    int64 Age = 3;    //`json:"age,omitempty"  xml:"age,omitempty"`
}
```

generate `.pb.go` with command `protoc` as:

```shell
protoc --go_out=plugins=grpc+retag:. example.proto
```

the custom tag will set to strcut:

```golang
type Staff struct {
    ID   string `protobuf:"bytes,1,opt,name=ID"     json:"id,omitempty"    xml:"id,omitempty"`
    Name string `protobuf:"bytes,2,opt,name=Name"   json:"name,omitempty"  xml:"name,omitempty"`
    Age  int64  `protobuf:"varint,3,opt,name=Age"   json:"age,omitempty"   xml:"age,omitempty"`
}
```