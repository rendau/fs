# File server

To generate documentation, `go-swagger` need to be installed:
```
dir=$(mktemp -d) 
git clone https://github.com/go-swagger/go-swagger "$dir" 
cd "$dir"
go install ./cmd/swagger
rm -rf "$dir"
```
