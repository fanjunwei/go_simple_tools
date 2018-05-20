package main

import (
	"net/http"
	"log"
	"fmt"
	"os"
	"path/filepath"
	"io/ioutil"
	"encoding/json"
	"time"
)

type LogReq struct {
	FileName string `json:"file_name"`
	Text     string `json:"text"`
}
type LogNode struct {
	Fd         *os.File
	RolloverAt int64
}

var (
	logChannel  = make(chan LogReq, 100)
	exitChannel = make(chan int)
	logNodeMap  = make(map[string]*LogNode)
)

const (
	INTERVAL_SECONDS = 3

)

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
func checkRollover(logFilePath string) {
	logNode, ok := logNodeMap[logFilePath]
	if ok {
		if time.Now().Unix() >= logNode.RolloverAt {
			fmt.Println("on roll")
			err:=logNode.Fd.Close()
			if err!=nil{
				fmt.Println(err)
				return
			}
			ts := (logNode.RolloverAt - INTERVAL_SECONDS)
			dir := filepath.Dir(logFilePath)
			name := filepath.Base(logFilePath)
			t := time.Unix(ts, 0).Format("2006_01_02_15_04_05")
			name = fmt.Sprintf("%s_%s", name, t)
			newPath := filepath.Join(dir, name)
			fmt.Println("newpath:",newPath)
			if PathExists(newPath) {
				os.Remove(newPath)
			}
			fmt.Println("to:",logFilePath,newPath)
			err=os.Rename(logFilePath, newPath)
			if err!=nil{
				fmt.Println(err)
			}
			delete(logNodeMap, logFilePath)
		}
	}
}
func getLogFile(logFilePath string) (file *os.File, err error) {
	checkRollover(logFilePath)
	dir := filepath.Dir(logFilePath)
	logNode, ok := logNodeMap[logFilePath]
	if ok {
		return logNode.Fd, nil
	}
	if !PathExists(dir) {
		os.MkdirAll(dir, 0777)
	}
	file, err = os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	stat, err := file.Stat()
	if err != nil {
		return
	}
	mdTime := stat.ModTime().Unix()
	rolloverAt:=mdTime - mdTime%INTERVAL_SECONDS + INTERVAL_SECONDS
	logNodeMap[logFilePath] = &LogNode{
		Fd:         file,
		RolloverAt: rolloverAt,
	}
	fmt.Println("route at:",time.Unix(rolloverAt,0).Format("2006-01-02_15:04:05"))
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
	var listen string
	if len(os.Args) >= 2 {
		listen = os.Args[1]
	} else {
		listen = "127.0.0.1:35673"
	}
	go writeLog()
	http.HandleFunc("/", home)
	http.HandleFunc("/log", httpWriteLog)
	err := http.ListenAndServe(listen, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	exitChannel <- 0
}
