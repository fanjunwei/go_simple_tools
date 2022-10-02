package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

//UDP协议
func main() {

	port := 789
	tag := ""
	if len(os.Args) >= 2 {
		for index := range os.Args {
			data := os.Args[index]
			if tag == "" {
				 if data == "-p" {
					tag = "p"
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
			}
		}
	}

	var listen, err = net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(0, 0, 0, 0),
		Port: port})
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	defer listen.Close() //关闭监听
	for true {
		var bf [1024]byte
		udp, addr, err := listen.ReadFromUDP(bf[:]) //接受UDP数据
		if err != nil {
			fmt.Println("read udp failed,err", err)
			continue
		}
		fmt.Printf("data:%v,addr:%v", string(bf[:udp]), addr)
		//写入数据
		_, err = listen.WriteToUDP(bf[:udp], addr)
		if err != nil {
			fmt.Println("write udp error", err)
			continue
		}
	}
}
