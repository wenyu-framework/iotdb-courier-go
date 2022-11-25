# 编译
## linux
```shell
env CGO_ENABLED=1 CC=x86_64-linux-musl-gcc GOOS=linux GOARCH=amd64 TRAVIS_TAG=temp-musl-v`date +%y%m%d` go build -o collector_linux_x64 --ldflags="-linkmode external -extldflags '-static' -w"
```