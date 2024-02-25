package main

import (
	"log"
	"net/http"
)

func main() {
	setupApi()
	log.Fatal(http.ListenAndServe(":7000", nil))
}

func setupApi() {
	manager := NewManager()
	http.HandleFunc("/ws", manager.serveWS)
}

// To test from console of a web browser
// socket=new WebSocket("ws://localhost:7000/ws")
// socket.onmessage=(event)=>{console.log(evnet.data)}
// socket.send("Hello")

// To connect from dart  use this function
// setConnection() async {
//     const wsUrl = "ws://192.168.1.6:7000/ws";
//     final channel = IOWebSocketChannel.connect(wsUrl);
//     Map<String, dynamic> messageMap = {
//       "type": "new_message",
//       "payload": "Helylo",
//     };
//     channel.sink.add(jsonEncode(messageMap));
//   }
