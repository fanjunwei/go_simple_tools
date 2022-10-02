package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

func main() {
	port := 789
	dest := "127.0.0.1"
	tag := ""
	if len(os.Args) >= 2 {
		for index := range os.Args {
			data := os.Args[index]
			if tag == "" {
				if data == "-p" {
					tag = "p"
				} else if data == "-d" {
					tag = "d"
				} else if data == "-h" {
					fmt.Printf(`usage: %s [OPTIONS]
    -p PORT: port number (default: 789)
    `, os.Args[0])
					os.Exit(0)
				}
			} else if tag == "p" {
				var err error
				port, err = strconv.Atoi(data)
				if err != nil {
					fmt.Printf("%s\n", "port must be int")
					os.Exit(1)
				}
				tag = ""
			} else if tag == "d" {
				var err error
				port, err = strconv.Atoi(data)
				if err != nil {
					fmt.Printf("%s\n", "port must be int")
					os.Exit(1)
				}
				tag = ""
			}
		}
	}
	socket, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.ParseIP(dest),
		Port: port})
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	defer socket.Close() //关闭连接
	sendData := []byte("hello server")
	_, err = socket.Write(sendData)
	if err != nil {
		fmt.Println("发送数据失败", err)
		return
	}
	data := make([]byte, 4096)
	udp, addr, err := socket.ReadFromUDP(data) //接收数据
	if err != nil {
		fmt.Println("接受数据失败", err)
		return
	}
	fmt.Printf("recv:%v,addr:%v", string(data[:udp]), addr)
}
