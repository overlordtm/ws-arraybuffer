# ws-arraybuffer

Demonstration how to use WebSockets and ArrayBuffer to transport binary data. Just open browser on http://localhost:8080 and open developemnt console.

## Install

```
mkdir go
cd go
export GOPATH=`pwd`
go get github.com/overlordtm/ws-arraybuffer
./bin/ws-arraybuffer
```

## Usage

```
./bin/ws-arraybuffer --help
Usage of ./bin/ws-arraybuffer:
  -addr=":8080": http service address
  -int=100ms: interval to send message
  -size=10000: size of message (in float32s)
```