module github.com/slimsag/9eyes

go 1.15

require (
	github.com/MichaelS11/go-hx711 v1.0.0
	github.com/go-ble/ble v0.0.0-20200407180624-067514cd6e24
	github.com/keegancsmith/sqlf v1.1.0
	github.com/lib/pq v1.8.0
	github.com/paypal/gatt v0.0.0-20151011220935-4ae819d591cf
	github.com/pkg/errors v0.8.1
	periph.io/x/periph v3.6.4+incompatible // indirect
)

replace github.com/MichaelS11/go-hx711 => ../go-hx711
