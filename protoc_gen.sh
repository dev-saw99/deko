export PATH=$PATH:/usr/local/go:/home/sonu/go:/home/sonu/go/bin
protoc --go_out=./proto/ --go_opt=paths=source_relative -I=./proto/  ./proto/deko_interface.proto --go-grpc_out=./proto/ --go-grpc_opt=paths=source_relative


