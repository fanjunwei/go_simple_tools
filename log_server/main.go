package main

import (
	"net/http"
	"log"
	"encoding/json"
	"fmt"
)

type LogReq struct {
	FileName string `json:"file_name"`
	Text     string `json:"text"`
}

type LogRes struct {
	Success bool `json:"success"`
}

func sayhelloName(w http.ResponseWriter, r *http.Request) {
	var logReq LogReq
	var buffer []byte;
	if r.Method == "POST" {
		r.Body.Read(buffer)
		json.Unmarshal(buffer, &logReq)
		res := LogRes{
			Success: true,
		}
		out, err := json.Marshal(res)
		if err != nil {
			w.Write([]byte(fmt.Sprint(err)))
			w.WriteHeader(500)
		} else {
			w.Write(out)
		}
	} else {
		w.Write([]byte("Method Not Allowed"))
		w.WriteHeader(405)
	}

}
func main() {
	http.HandleFunc("/", sayhelloName)                 //设置访问的路由
	err := http.ListenAndServe("127.0.0.1:35678", nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
