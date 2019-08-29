package main

import (
	"os"
	"net/http"
	"log"
)

func main() {
	if len(os.Args) >= 3 {
		addr := os.Args[1]
		filePaht := os.Args[2]
		http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(filePaht))))
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}

}
