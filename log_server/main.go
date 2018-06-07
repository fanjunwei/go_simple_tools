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
	"path"
	"strings"
	"compress/gzip"
	"archive/tar"
	"io"
)

type LogReq struct {
	FileName string `json:"file_name"`
	Text     string `json:"text"`
}
type LogNode struct {
	Fd         *os.File
	RolloverAt int64
}

type TarFile struct {
	File     *os.File
	DestName string
}

var (
	logChannel  = make(chan LogReq, 100)
	exitChannel = make(chan int)
	logNodeMap  = make(map[string]*LogNode)
)

const (
	IntervalSeconds = 60 * 60 * 24
	TimeFormat      = "2006_01_02"
)

//压缩 使用gzip压缩成tar.gz
func Compress(files []*TarFile, dest string) error {
	d, _ := os.Create(dest)
	defer d.Close()
	gw := gzip.NewWriter(d)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	for _, file := range files {
		err := compressForFile(file, "", tw)
		if err != nil {
			return err
		}
	}
	return nil
}

func compressForFile(fileNode *TarFile, prefix string, tw *tar.Writer) error {
	file := fileNode.File
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		var name string
		if fileNode.DestName != "" {
			name = fileNode.DestName
		} else {
			name = info.Name()
		}
		prefix = prefix + "/" + name
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = compressForFile(&TarFile{File: f, DestName: ""}, prefix, tw)
			if err != nil {
				return err
			}
		}
	} else {
		header, err := tar.FileInfoHeader(info, "")
		var name string
		if fileNode.DestName != "" {
			name = fileNode.DestName
		} else {
			name = info.Name()
		}
		header.Name = prefix + "/" + name
		if err != nil {
			return err
		}
		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(tw, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func httpWriteLog(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		buffer, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.Write([]byte(fmt.Sprint("req error:", err)))
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
			err := logNode.Fd.Close()
			if err != nil {
				fmt.Println(err)
				return
			}
			ts := logNode.RolloverAt - IntervalSeconds
			dir := filepath.Dir(logFilePath)
			filenameWithSuffix := filepath.Base(logFilePath)
			fileSuffix := path.Ext(filenameWithSuffix)
			filenameOnly := strings.TrimSuffix(filenameWithSuffix, fileSuffix)
			t := time.Unix(ts, 0).Format(TimeFormat)
			newName := fmt.Sprintf("%s_%s%s", filenameOnly, t, fileSuffix)
			newPath := filepath.Join(dir, newName)
			newTarPath := newPath + ".tar.gz"
			if PathExists(newTarPath) {
				os.Remove(newTarPath)
			}
			file, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				return
			}
			Compress([]*TarFile{{File: file, DestName: newName}}, newTarPath)

			//err = os.Rename(logFilePath, newPath)
			if err != nil {
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
	rolloverAt := mdTime - mdTime%IntervalSeconds + IntervalSeconds
	logNodeMap[logFilePath] = &LogNode{
		Fd:         file,
		RolloverAt: rolloverAt,
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
