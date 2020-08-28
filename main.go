package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
)

func main() {
	macAddr := flag.String("addr", "", "peripheral MAC address")
	flag.Parse()
	hciDevice, err := dev.NewDevice("default")
	if err != nil {
		panic(err)
	}
	ble.SetDefaultDevice(hciDevice)

	filter := func(a ble.Advertisement) bool {
		return true
		//return strings.ToUpper(a.Addr().String()) == strings.ToUpper(*macAddr)
	}

	// Scan for device
	log.Printf("Scanning for %s\n", *macAddr)
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), time.Second*300))
	client, err := ble.Connect(ctx, filter)
	if err != nil {
		panic(err)
	}

	for {
		fmt.Printf("Client side RSSI: %d\n", client.ReadRSSI())
		time.Sleep(time.Second)
	}

}

/*
package main

import (
	"fmt"

	"github.com/MichaelS11/go-hx711"
)

func main() {
	err := hx711.HostInit()
	if err != nil {
		fmt.Println("HostInit error:", err)
		return
	}

	hx711, err := hx711.NewHx711("GPIO6", "GPIO5")
	if err != nil {
		fmt.Println("NewHx711 error:", err)
		return
	}

	defer hx711.Shutdown()

	err = hx711.Reset()
	if err != nil {
		fmt.Println("Reset error:", err)
		return
	}

	var data int
	for {
		data, err = hx711.ReadDataRaw()
		if err != nil {
			fmt.Println("ReadDataRaw error:", err)
			continue
		}

		fmt.Println(data)
	}

}
*/
