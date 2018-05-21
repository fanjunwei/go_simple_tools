package main

import (
	"os"
	"encoding/json"
	"net"
	"fmt"
	"path/filepath"
	"strings"
	"html/template"
)

type NetWorkInfo struct {
	Mac     string `json:"mac"`
	Ip      string `json:"ip"` //cidr格式
	Gateway string `json:"gateway"`
}

func tpl_write_file(fd *os.File, tpl string, data interface{}) {
	t := template.New("")
	t, err := t.Parse(tpl)
	if err != nil {
		panic("Error : " + err.Error())
	}
	err = t.Execute(fd, data)
	if err != nil {
		panic("Error : " + err.Error())
	}
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

				tpl := `
TYPE=Ethernet
DEFROUTE=yes
PEERDNS=yes
PEERROUTES=yes
IPV4_FAILURE_FATAL=no
NAME={{.ifname}}
DEVICE={{.ifname}}
ONBOOT=yes
HWADDR={{.mac}}
ZONE=public
{{if .static}}
BOOTPROTO=static
IPADDR={{.ip}}
NETMASK={{.netmask}}
{{if ne .gateway "" }}GATEWAY={{.gateway}}{{end}}
{{else}}
BOOTPROTO=dhcp
{{end}}
`
				configPath := filepath.Join(outdir, fmt.Sprintf("ifcfg-%s", ifname))
				fd, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE, 0644)
				if err != nil {
					panic("Error : " + err.Error())
				}
				fd.Truncate(0)
				data := make(map[string]interface{})
				data["ifname"] = ifname
				data["mac"] = strings.ToUpper(item.Mac)
				if item.Ip != "" {

					ip, sub_net, err := net.ParseCIDR(item.Ip)
					if err != nil {
						panic("Error : " + err.Error())
					}
					data["ip"] = ip.String()
					data["netmask"] = net.IP(sub_net.Mask).String()
					data["gateway"] = item.Gateway
					data["static"] = true
				} else {
					data["static"] = false
				}
				tpl_write_file(fd, tpl, data)

			}

		}
	}
}
