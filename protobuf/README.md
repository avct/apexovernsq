# Updating the protocol

The protocol spec for protobuf must be compile if it is updated.  To do this you must run:

```protoc --go_out=. entry.proto```
