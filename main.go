package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var addr = flag.String("addr", ":8080", "http service address")
var interval = flag.Duration("int", 100*time.Millisecond, "interval to send message")
var size = flag.Int("size", 10000, "size of message (in float32s)")
var homeTempl = template.Must(template.New("base").Parse(tpl))

func main() {
	flag.Parse()

	fmt.Printf("Serving: %s\n", *addr)
	fmt.Printf("Interval: %s\n", *interval)
	fmt.Printf("Size (float32): %d\n", *size)

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTempl.Execute(w, r.Host)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	for _ = range time.Tick(*interval) {

		data := makeData(*size)

		rand.Float32()

		buf := new(bytes.Buffer)

		binary.Write(buf, binary.LittleEndian, data)
		ws.WriteMessage(websocket.BinaryMessage, buf.Bytes())
	}

}

func makeData(size int) (data []float32) {

	data = make([]float32, size)

	for i := 0; i < size; i++ {
		data[i] = rand.Float32()
	}
	return
}

var tpl = `
<!DOCTYPE html>
<html lang="en">
<head>
<title>Binary Example</title>
<script src="//ajax.googleapis.com/ajax/libs/jquery/2.0.3/jquery.min.js"></script>
<script type="text/javascript">
    $(function() {

    var ctr = 0
    var last = new Date().getTime()

    setInterval(function(){
        now = new Date().getTime()
        delta = now - last
        last = now

        mps = ctr / delta * 1000
        ctr = 0

        console.log("msg/s: ", mps)
    }, 1000);

    if (window["WebSocket"]) {
        conn = new WebSocket("ws://localhost:8080/ws");
        conn.binaryType = "arraybuffer"

        conn.onmessage = function(e) {
            var x = new Float32Array(e.data)
            ctr++
        }
    } else {
        console.log("Your browser does not support WebSockets")
    }
    });
</script>
<style type="text/css">

</style>
</head>
<body>

</body>
</html>
`
