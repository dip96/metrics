cd cmd/server
rm -rf server
go build -o server *.go
cd ../agent
rm -rf agent
go build -o agent *.go
