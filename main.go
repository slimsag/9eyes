package main

import (
	"context"
	"flag"
	"log"
)

func main() {
	flag.Parse()

	// Tracke Tile trackers / BLE beacons.
	t := &tracker{
		results: make(chan trackerResult),
	}
	go func() {
		log.Println("Tracking tile trackers...")
		ctx := context.Background()
		if err := t.start(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	for {
		ev := <-t.results
		_ = ev
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
