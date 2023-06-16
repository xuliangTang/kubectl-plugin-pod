go build -o /Users/zhoujiahong/go/bin/kubectl-pods main.go

CGO_ENABLED=0  GOOS=linux  GOARCH=amd64 go build -o /Users/zhoujiahong/bin/kubectl-pods main.go