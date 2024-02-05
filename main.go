package main

import (
	"context"
	"fmt"
	"github.com/coreos/go-systemd/v22/dbus"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
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

	// Get the unit properties
	props, err := conn.GetAllPropertiesContext(ctx, unitName)
	if err != nil {
		log.Fatalf("Failed to get unit properties: %v", err)
	}

	activeState, ok := props["ActiveState"].(string)
	if !ok {
		log.Fatalf("Failed to get active state for unit: %s", unitName)
	}

	if activeState == "active" {
		// Restart the service
		cmd := exec.Command("systemctl", "restart", unitName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Fatalf("Failed to restart service: %v", err)
		}
		fmt.Printf("Service '%s' restarted\n", unitName)
	} else {
		fmt.Printf("Service '%s' is not currently active\n", unitName)
	}

	// Get bridge interface name and its IP
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
			log.Fatal("Couldn't get interface addresses, %v\n", err.Error())
		}
		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPNet:
				if strings.HasPrefix(i.Name, "enp") && v.IP.To4() != nil {
					bridgeInterfaceName = i.Name
					bridgeAddress = strings.Split(v.IP.To4().String(), "/")[0] // 192.168.65.5
					bridgeAddress = strings.TrimSpace(bridgeAddress)
					bridgeAddress16BitBlock = net.ParseIP(bridgeAddress)
				}
			}
		}
	}

	// Get bridge ip address as 16-bit Block
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
