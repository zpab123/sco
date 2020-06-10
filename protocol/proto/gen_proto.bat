protoc --go_out=.. ./handler.proto
protoc --go_out=.. ./remote.proto

protoc --go_out=plugins=grpc:.. grpc.proto
