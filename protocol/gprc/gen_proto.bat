protoc --go_out=.. ./grpcRequest.proto
protoc --go_out=.. ./grpcResponse.proto

protoc --go_out=plugins=grpc:.. grpc.proto
