package main

import (
	"net/http"
	"log"
	"fmt"
	"os"
	"path/filepath"
	"io/ioutil"
	"encoding/json"
)

var (
	logChannel  = make(chan LogReq, 100)
	exitChannel = make(chan int)
	fileHandler = make(map[string]*os.File)
)

type LogReq struct {
	FileName string `json:"file_name"`
	Text     string `json:"text"`
}


func httpWriteLog(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		buffer, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("req error:", err)))
			w.WriteHeader(400)
			return
		}
		var logReq LogReq
		err = json.Unmarshal(buffer, &logReq)
		if err != nil {
			fmt.Println("req error:", err)
			return
		}
		logChannel <- logReq

		w.Write([]byte("ok\n"))
	} else {
		w.Write([]byte("Method Not Allowed"))
		w.WriteHeader(405)
	}

}
func home(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Write([]byte("Log Server OK"))
	} else {
		w.Write([]byte("Method Not Allowed"))
		w.WriteHeader(405)
	}
}
func PathExists(path string) (bool) {
	_, err := os.Stat(path)
	if err == nil {
		return true
	} else {
		return false
	}

}

func getLogFile(logFilePath string) (file *os.File, err error) {
	dir := filepath.Dir(logFilePath)
	file, ok := fileHandler[logFilePath]
	if ok {
		return file, nil
	}
	if !PathExists(dir) {
		os.MkdirAll(dir, 0777)
	}
	file, err = os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err == nil {
		fileHandler[logFilePath] = file
	}
	return
}
func writeLog() {
	for {
		select {
		case <-exitChannel:
			return
		case logData := <-logChannel:
			fd, err := getLogFile(logData.FileName)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Fprintln(fd, logData.Text)
		}
	}
}
func main() {
	go writeLog()
	http.HandleFunc("/", home)
	http.HandleFunc("/log", httpWriteLog)
	err := http.ListenAndServe("127.0.0.1:35678", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	exitChannel <- 0
}
