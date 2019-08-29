package main

import (
	"log"
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("home 2\n"))
}
func main() {
	addr := "10.211.17.30:9801"
	http.HandleFunc("/", home)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
