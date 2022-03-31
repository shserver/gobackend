protoc proto/auth_service.proto --grpc-gateway_out=gen --grpc-gateway_opt logtostderr=true \
    --grpc-gateway_opt generate_unbound_methods=true

protoc proto/test_service.proto --grpc-gateway_out=gen --grpc-gateway_opt logtostderr=true \
    --grpc-gateway_opt generate_unbound_methods=true

protoc proto/public_service.proto --grpc-gateway_out=gen --grpc-gateway_opt logtostderr=true \
    --grpc-gateway_opt generate_unbound_methods=true