protoc proto/message.proto --go_out=gen
protoc proto/auth_service.proto --go_out=gen --go-grpc_out=gen --go-grpc_opt=require_unimplemented_servers=false
protoc proto/test_service.proto --go_out=gen --go-grpc_out=gen --go-grpc_opt=require_unimplemented_servers=false
protoc proto/public_service.proto --go_out=gen --go-grpc_out=gen --go-grpc_opt=require_unimplemented_servers=false
