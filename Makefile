build: 
	@go build -o bin/sshwebhook main.go

run: build
	@clear
	@./bin/sshwebhook