gen: 
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
	--go_grpc_out=pb  --go_grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb  --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=swagger \
	proto/*.proto  

clean:
	rm pb/*.go

run-server:
	go run cmd/server/main.go -port ${port} -server-type ${type}

run-client:
	go run cmd/client/main.go -address ${address}

run-nginx-docker:
	docker run --name pcbook-lb -p 8080:8080 -v ${PWD}/nginx.conf:/etc/nginx/nginx.conf:ro --network="host"  -d nginx 