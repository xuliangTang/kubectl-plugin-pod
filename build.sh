go build -o /Users/zhoujiahong/go/bin/kubectl-pods pod.go
go build -o /Users/zhoujiahong/go/bin/kubectl-deploy deploy.go

CGO_ENABLED=0  GOOS=linux  GOARCH=amd64 go build -o /Users/zhoujiahong/bin/kubectl-pods pod.go
CGO_ENABLED=0  GOOS=linux  GOARCH=amd64 go build -o /Users/zhoujiahong/bin/kubectl-deploy deploy.go