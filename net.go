package netutil

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

	fastping "github.com/tatsushid/go-fastping"
)

func GetHostAddr() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return net.IP{}, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return net.IP{}, err
		}
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
				continue // not an ipv4 address
			}
			return ip, nil
		}
	}
	return net.IP{}, errors.New("are you connected to the network?")
}

func ConfigIPAddress(device, ip string) error {
	cmd := []string{device, ip, "netmask", "255.255.255.0"}
	if err := exec.Command("ifconfig", cmd...).Run(); err != nil {
		fmt.Println(err)
		return err
	}
	ip, err := GetIPAddress(device)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func GetIPAddress(device string) (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Name != device || iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
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
			return ip.String(), nil
		}
	}
	return "", errors.New("Not connected")
}

func GetMacAddress(device string) (addr string) {
	ifaces, err := net.Interfaces()
	if err == nil {

		for _, iface := range ifaces {
			if iface.Name != device || iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
				continue
			}
			addr = iface.HardwareAddr.String()

		}
	}
	return addr
}

func Ping(device string) error {
	ip, err := GetIPAddress(device)
	if err != nil {
		log.Println(err)
		return err
	}
	sub := strings.SplitAfterN(ip, ".", -1)
	host := ""
	for i := 0; i < 3; i++ {
		host = host + sub[i]
	}
	host = host + "1"

	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", host)
	if err != nil {
		log.Println(err)
		return err
	}
	p.AddIPAddr(ra)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		log.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
	}
	p.OnIdle = func() {
		log.Println("ping ok")
	}
	err = p.Run()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func PingIP(dest string) error {
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", dest)
	if err != nil {
		log.Println(err)
		return err
	}
	p.AddIPAddr(ra)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		log.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
	}
	p.OnIdle = func() {
		log.Println("ping ok")
	}
	err = p.Run()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
