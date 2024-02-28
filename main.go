package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
)

func main() {
	// Create a context with cancellation capability
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to the system bus
	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to system bus: %v", err)
	}
	defer conn.Close()

	// Get the unit name of the service you want to restart
	unitName := "jaeger-collector.service"
	
	mode := "replace"
	_, err = conn.RestartUnitContext(ctx, unitName, mode, nil)
	if err != nil {
		log.Fatalf("Failed to restart service: %v", err)
	} else {
		fmt.Printf("Service '%s' restarted\n", unitName)
	}

	time.Sleep(3 * time.Second)
	
	// Get bridge interface and its IP and range
	var bridgeInterfaceName string
	var bridgeAddress string
	var bridgeAddress16BitBlock net.IP
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Fatalf("Couldn't get list of the system's netowrk interfaces, %v\n", err.Error())
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Fatalf("Couldn't get interface addresses, %v\n", err.Error())
		}
		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPNet:
				if strings.HasPrefix(i.Name, "br-") && v.IP.To4() != nil {
					bridgeInterfaceName = i.Name
					bridgeAddress = strings.Split(v.IP.To4().String(), "/")[0] // 192.168.65.5
					bridgeAddress = strings.TrimSpace(bridgeAddress)
					bridgeAddress16BitBlock = net.ParseIP(bridgeAddress)
				}
			}
		}
	}

	ipv4mask := net.CIDRMask(16, 32)
	bridgeAddress16BitBlock = bridgeAddress16BitBlock.Mask(ipv4mask)
	rangeIn16BitBlock := bridgeAddress16BitBlock.String() + "/16"

	cmd := exec.Command("ip", "route", "add", rangeIn16BitBlock, "dev", bridgeInterfaceName,
		"proto", "kernel", "scope", "link", "src", bridgeAddress, "table", "table-lan")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to add bridge route: %v", err)
	}
	fmt.Println("Bridge route added successfully.")
}
