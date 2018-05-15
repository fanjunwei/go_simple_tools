package main

import (
	"os"
	"encoding/json"
	"net"
	"fmt"
	"path/filepath"
	"strings"
)

type NetWorkInfo struct {
	Mac     string `json:"mac"`
	Ip      string `json:"ip"`
	Gateway string `json:"gateway"`
}

func main() {
	if len(os.Args) >= 3 {
		netInfos := []NetWorkInfo{}
		json.Unmarshal([]byte(os.Args[1]), &netInfos)
		outdir := os.Args[2]
		macNameMap := make(map[string]string)
		interfaces, err := net.Interfaces()
		if err != nil {
			panic("Error : " + err.Error())
		}
		for _, inter := range interfaces {
			macNameMap[strings.ToLower(inter.HardwareAddr.String())] = inter.Name
		}
		for _, item := range netInfos {
			ifname, ok := macNameMap[strings.ToLower(strings.Replace(item.Mac, "-", ":", -1))]
			if ok {
				var value string
				if item.Ip != "" {
					tmpl := `TYPE=Ethernet
BOOTPROTO=static
DEFROUTE=yes
PEERDNS=yes
PEERROUTES=yes
IPV4_FAILURE_FATAL=no
NAME=%s
DEVICE=%s
ONBOOT=yes
IPADDR=%s
NETMASK=%s
GATEWAY=%s
HWADDR=%s
ZONE=public`
					ip, sub_net, err := net.ParseCIDR(item.Ip)
					if err != nil {
						panic("Error : " + err.Error())
					}
					value = fmt.Sprintf(tmpl, ifname, ifname, ip.String(), net.IP(sub_net.Mask).String(), item.Gateway, strings.ToUpper(item.Mac))
				} else {
					tmpl := `TYPE=Ethernet
BOOTPROTO=dhcp
DEFROUTE=yes
PEERDNS=yes
PEERROUTES=yes
IPV4_FAILURE_FATAL=no
NAME=%s
DEVICE=%s
ONBOOT=yes
HWADDR=%s
ZONE=public`
					value = fmt.Sprintf(tmpl, ifname, ifname, strings.ToUpper(item.Mac))
				}
				configPath := filepath.Join(outdir, fmt.Sprintf("ifcfg-%s", ifname))
				fd, _ := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE, 0644)
				fd.Truncate(0)
				fd.Write([]byte(value))
			}

		}
	}
}
