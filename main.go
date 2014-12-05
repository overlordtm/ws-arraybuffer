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
	"strconv"
)

var addr = flag.String("addr", ":8080", "http service address")
var size = flag.Int("size", 1000, "size of message (in float32s)")
var homeTempl = template.Must(template.New("base").Parse(tpl))

func main() {
	flag.Parse()

	fmt.Printf("Serving: %s\n", *addr)
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
	ReadBufferSize:  1024 * 5,
	WriteBufferSize: 1024 * 5,
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

	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			log.Println("read error", err)
			break
		}

		length, err := strconv.ParseInt(string(p), 10, 32)
		if err != nil {
			log.Println("read error", err)
			break
		}
		log.Println("len: ", length)

		data := makeData(int(length * 5))
		buf := new(bytes.Buffer)

		binary.Write(buf, binary.LittleEndian, data)
		ws.WriteMessage(websocket.BinaryMessage, buf.Bytes())
	}
}

func makeData(size int) (data []float32) {

	data = make([]float32, size)

	for i := 0; i < size; i++ {
		data[i] = rand.Float32() * 50
	}
	return
}

var tpl = `
<!DOCTYPE html>
<html>
<body>

<canvas id="myCanvas" width="1000" height="500" style="display: fixed; border:1px solid #d3d3d3;">
Your browser does not support the HTML5 canvas tag.</canvas>
<div id="fps"></div>

<script>

"use strict"

 var div = document.getElementById("fps"),
    ctx = document.getElementById("myCanvas").getContext("2d"),
    ctr = 0,
    sigLen = 1E3,
    dataLen = sigLen * 5,
    avgFps = 30,
    data = new Float32Array(dataLen);

var sig1Start = 0;
var sig1End = sigLen -1;
var sig2Start = sig1End + 1;
var sig2End = sig2Start + sigLen - 1;
var sig3Start = sig2End + 1;
var sig3End = sig3Start + sigLen - 1;
var sig4Start = sig3End + 1;
var sig4End = sig4Start + sigLen - 1;
var sig5Start = sig3End + 1;
var sig5End = sig5Start + sigLen - 1;

var black = "#000000";
var red = "#FF0000";
var green = "#00FF00";
var blue = "#0000FF";
var grey = "#cccccc";


var conn = new WebSocket("ws://" + window.location.host + "/ws");
conn.binaryType = "arraybuffer";
var last = (new Date).getTime();

function d(data, y) {
    ctx.moveTo(0, y);
    var f = 0;
    for (var i = 0; i < sigLen; i++) {
        f = (y + 0.5 + data[i]) | 0
        ctx.lineTo(i, f)
        ctx.moveTo(i, f)
    }
}

function drawAll() {
    ctx.beginPath();

    ctx.clearRect(0, 0, 1E3, 500); // clear canvas

    ctx.strokeStyle = black;
    d(data.subarray(sig1Start, sig1End), 0);
    ctx.stroke();
    ctx.closePath();

    ctx.beginPath();
    ctx.strokeStyle = red;
    d(data.subarray(sig2Start, sig2End), 50);
    ctx.stroke();
    ctx.closePath();

    ctx.beginPath();
    ctx.strokeStyle = green;
    d(data.subarray(sig3Start, sig3End), 100);
    ctx.stroke();
    ctx.closePath();

    ctx.beginPath();
    ctx.strokeStyle = blue;
    d(data.subarray(sig4Start, sig4End), 150);
    ctx.stroke();
    ctx.closePath();

    ctx.beginPath();
    ctx.strokeStyle = grey;
    d(data.subarray(sig5Start, sig5End), 200);
    ctx.stroke();
    ctx.closePath();

    conn.send(sigLen); // request new data via WS
    ctr++;
}

function getDataAndDraw(e) {
    data.set(new Float32Array(e.data));
    window.requestAnimationFrame(drawAll);
};

conn.onopen = function() {
    conn.send(sigLen)
};

conn.onmessage = getDataAndDraw;

setInterval(function() {
    var delta = ((new Date).getTime() - last) / 1E3;
    last = (new Date).getTime();
    var fps = ctr / delta;
    avgFps = 0.25 * avgFps + 0.75 * fps;
    
    div.innerHTML = sigLen + ":" + parseInt(fps) + ":" + parseInt(avgFps);

    ctr = 0;
}, 1000)

setInterval(function() {

    if (avgFps < 10) {
        sigLen = sigLen / 2;
        dataLen = sigLen * 5;

        sig1End = sigLen -1;
        sig2Start = sig1End + 1;
        sig2End = sig2Start + sigLen - 1;
        sig3Start = sig2End + 1;
        sig3End = sig3Start + sigLen - 1;
        sig4Start = sig3End + 1;
        sig4End = sig4Start + sigLen - 1;
        sig5Start = sig3End + 1;
        sig5End = sig5Start + sigLen - 1;
    }

}, 5000)

</script>

</body>
</html>
`
