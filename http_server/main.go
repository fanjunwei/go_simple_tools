package main

import (
	"log"
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("home 1\n"))
}
func main() {
	addr := "127.0.0.1:9801"
	http.HandleFunc("/", home)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
