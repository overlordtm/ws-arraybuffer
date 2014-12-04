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
var size = flag.Int("size", 5000, "size of message (in float32s)")
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

	for {

		messageType, p, err := ws.ReadMessage()
		if err != nil {
			log.Println("read error", err)
		} else {
			log.Println("got msg", messageType, string(p))
		}

		data := makeData(*size)
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
<html>
<body>

<canvas id="myCanvas" width="1000" height="500"
style="border:1px solid #d3d3d3;">
Your browser does not support the HTML5 canvas tag.</canvas>
<div id="fps"></div>

<script src="//ajax.googleapis.com/ajax/libs/jquery/2.1.1/jquery.min.js"></script>

<script>
    var c = document.getElementById("myCanvas");
    var ctx = c.getContext("2d");

    var data = null;


    conn = new WebSocket("ws://" + window.location.host + "/ws");
    conn.binaryType = "arraybuffer"


    last = new Date().getTime()

    function drawLine(data, y) {
        ctx.moveTo(0,y);

        for (i = 0; i < data.length; i++) { 
            var dva = y + (data[i] * 50)
            var ena = i;
            ctx.lineTo(ena,dva);
            ctx.moveTo(ena,dva);
            
        }
    }

    function draw() {

        ctx.beginPath();
        ctx.clearRect(0,0,1000,500);
        // ctx.moveTo(0,0);

        // for (i = 0; i < 1024; i++) { 
        //     var dva = data[i] * 500
        //     var ena = i;
        //     ctx.lineTo(ena,dva);
        //     ctx.moveTo(ena,dva);
            
        // }
        ctx.strokeStyle="#000000";
        drawLine(data.subarray(0,999), 0)
        ctx.stroke();

        ctx.strokeStyle="#FF0000";
        drawLine(data.subarray(1000,1999),50)
        ctx.stroke();

        ctx.strokeStyle="#00FF00";
        drawLine(data.subarray(2000,2999),100)
        ctx.stroke();

        ctx.strokeStyle="#0000FF";
        drawLine(data.subarray(3000,3999),150)
        ctx.stroke();

        ctx.strokeStyle="#cccccc";
        drawLine(data.subarray(4000,4999),200)
        ctx.stroke();

        ctx.closePath();
        

        // window.requestAnimationFrame(function() {
        //     window.requestAnimationFrame(draw);
        // })
    }
    // draw();

    conn.onopen = function() {
        conn.send("foo")
    }

    conn.onmessage = function(e) {
        data = new Float32Array(e.data)
        // draw()
        window.requestAnimationFrame(function() {
            draw();
            conn.send("foo")

            delta = (new Date().getTime() - last)/1000
            fps = 1/delta
            last = new Date().getTime()
            $("#fps").text(fps)
        })
        


    }


</script>

</body>
</html>
`
