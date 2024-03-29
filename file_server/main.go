package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	filePath := "."
	port := "7000"
	addr := "0.0.0.0"
	tag := ""
	if len(os.Args) >= 2 {
		for index := range os.Args {
			data := os.Args[index]
			if tag == "" {
				if data == "-a" {
					tag = "a"
				} else if data == "-p" {
					tag = "p"
				} else if data == "-d" {
					tag = "d"
				} else if data == "-h" {
					fmt.Printf(`usage: %s [OPTIONS]
    -a ADDR: address (default: 0.0.0.0)
    -p PORT: port number (default: 7000)
    -d DIR:  root directory (default: .)`, os.Args[0])
					os.Exit(1)
				}
			} else if tag == "a" {
				addr = data
				tag = ""
			} else if tag == "p" {
				port = data
				tag = ""
			} else if tag == "d" {
				filePath = data
				tag = ""
			}
		}
	}
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(filePath))))
	log.Printf("listen: %s:%s\n", addr, port)
	log.Printf("path: %s\n", filePath)
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", addr, port), nil)
	if err != nil {
		log.Fatal("ERROR: ", err)
	}

}
