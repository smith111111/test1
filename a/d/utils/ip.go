package utils

import (
	"log"
	"net"
	"strings"
	"errors"
)

var PublicIP = IP()

func IP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Println(err)
		return ""
	}
	ipList := make([]string, 0)
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip == nil || ip.IsLoopback() {
			continue
		}
		ip = ip.To4()
		if ip == nil {
			continue
		}
		if isPrivate(ip) {
			ipList = append(ipList, ip.String())
			continue
		}
		return ip.String()
	}
	if len(ipList) > 0 {
		return ipList[0]
	}
	return ""
}

func isPrivate(ip net.IP) bool {
	ipStr := strings.Split(ip.String(), ".")
	if len(ipStr) != 4 {
		return true
	}
	if ipStr[0] == "10" {
		return true
	}
	if ipStr[0] == "172" && (ipStr[1] >= "16" && ipStr[1] <= "31") {
		return true
	}
	if ipStr[0] == "192" && ipStr[1] == "168" {
		return true
	}
	return false
}

func isPrivateIPv4(ip net.IP) bool {
	return ip != nil &&
		(ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) || ip[0] == 192 && ip[1] == 168)
}

func privateIPv4() (net.IP, error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if isPrivateIPv4(ip) {
			return ip, nil
		}
	}
	return nil, errors.New("no private ip address")
}

func Lower16BitPrivateIP() (uint16, error) {
	ip, err := privateIPv4()
	if err != nil {
		return 0, err
	}

	return uint16(ip[2])<<8 + uint16(ip[3]), nil
}