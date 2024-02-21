cd ./cmd/server
go build -o server *.go
cd ../agent
go build -o agent *.go &
