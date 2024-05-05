build: 
	@go build -o bin/sshwebhook main.go

run: build
	@clear
	@./bin/sshwebhook

echo:
	@go build -o bin/echo-server cmd/echo/main.go
	@./bin/echo-server