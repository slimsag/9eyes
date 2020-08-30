package main

import (
	"context"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
	"github.com/go-ble/ble/linux/hci/cmd"
)

type trackerResult struct {
	time time.Time

	// addr of the BLE device
	addr string

	// RSSI signal strength of the device
	rssi int
}

type tracker struct {
	results chan trackerResult
}

func (t *tracker) advHandler(a ble.Advertisement) {
	serviceUUIDs := a.Services()
	isTile := false
	for _, uuid := range serviceUUIDs {
		// https://www.bluetooth.com/specifications/assigned-numbers/company-identifiers/
		// https://sourcegraph.com/github.com/reelyactive/advlib/-/blob/lib/ble/common/assignednumbers/memberservices.js#L20-21
		if uuid.String() == "feed" || uuid.String() == "feec" {
			// Tile, Inc. BLE device
			isTile = true
		}
	}
	if !isTile {
		// do not emit events for non-Tile BLE devices.
		return
	}
	t.results <- trackerResult{
		time: time.Now(),
		addr: a.Addr().String(),
		rssi: a.RSSI(),
	}
}

func (t *tracker) start(ctx context.Context) error {
	device, err := dev.NewDevice("default", ble.OptScanParams(cmd.LESetScanParameters{
		LEScanType:           0x01,   // 0x00: passive, 0x01: active
		LEScanInterval:       0,      // 0x0004 - 0x4000; N * 0.625msec
		LEScanWindow:         0x0004, // 0x0004 - 0x4000; N * 0.625msec
		OwnAddressType:       0x00,   // 0x00: public, 0x01: random
		ScanningFilterPolicy: 0x00,   // 0x00: accept all, 0x01: ignore non-white-listed.
	}))
	if err != nil {
		return err
	}
	ble.SetDefaultDevice(device)
	return ble.Scan(ctx, true, t.advHandler, nil)
}
