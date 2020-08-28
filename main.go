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

	// SetGain default is 128
	// Gain of 128 or 64 is input channel A, gain of 32 is input channel B
	// hx711.SetGain(128)

	var weight1 float64
	var weight2 float64

	weight1 = 100
	weight2 = 200

	hx711.GetAdjustValues(weight1, weight2)
}
