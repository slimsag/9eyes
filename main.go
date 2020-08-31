package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/keegancsmith/sqlf"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

var (
	useTracker     = flag.Bool("tracker", true, "scan for tile trackers")
	trackerVerbose = flag.Bool("tracker-verbose", false, "print raw tracker events")
	clientName     = flag.String("client", "unnamed", "name of this client / raspberry pi")

	useScale           = flag.Bool("scale", true, "read Hx711 scale weight values")
	scaleVerbose       = flag.Bool("scale-verbose", false, "print raw scale weight values for determining baseline")
	scaleBaseline      = flag.Float64("scale-baseline", 0.0, "set base weight of scale (with nothing on it)")
	scaleDivisor       = flag.Float64("scale-divisor", 1.0, "set divisor for scale, i.e. what to divide units by to get grams")
	scaleInvert        = flag.Bool("scale-invert", false, "whether to invert numbers from the scale")
	scaleSampleSize    = flag.Int("scale-sample-size", 4, "number of samples to read, denoise, and average to determine scale sensor value")
	scaleMovingAverage = flag.Int("scale-moving-average", 10, "number of past scale sensor values to denoise and average against")
)

func main() {
	flag.Parse()

	// Tracke Tile trackers / BLE beacons.
	t := &tracker{
		results: make(chan trackerResult, 100),
	}
	if *useTracker {
		go func() {
			log.Println("Tracking tile trackers...")
			ctx := context.Background()
			if err := t.start(ctx); err != nil {
				log.Fatal(err)
			}
		}()
	}

	s := &scale{
		results:       make(chan scaleResult, 100),
		baseline:      *scaleBaseline,
		divisor:       *scaleDivisor,
		invert:        *scaleInvert,
		sampleSize:    *scaleSampleSize,
		movingAverage: *scaleMovingAverage,
	}
	if *useScale {
		go func() {
			log.Println("Monitoring Hx711 scale...")
			ctx := context.Background()
			if err := s.start(ctx); err != nil {
				log.Fatal(err)
			}
		}()
	}

	var (
		latestScale                   scaleResult
		latestTrackers                = map[string]trackerResult{}
		updatedScale, updatedTrackers bool

		bufferMu sync.Mutex
		buffer   = make([]result, 0, 100)
	)
	go func() {
		db, err := sql.Open("postgres", "user=postgres dbname=9eyes sslmode=disable password=password")
		if err != nil {
			log.Fatal(err)
		}

		sendBuffered := func() error {
			var trackerInserts, scaleInserts []*sqlf.Query
			maxPostgresParams := 65535
			paramsPerResult := 7
			maxResults := (maxPostgresParams / paramsPerResult) - 1
			for _, r := range buffer[:maxResults] {
				for _, tr := range r.trackers {
					trackerInserts = append(trackerInserts, sqlf.Sprintf("(%v, %v, %v, %v)", r.t, *clientName, tr.addr, tr.rssi))
				}
				scaleInserts = append(scaleInserts, sqlf.Sprintf("(%v, %v, %v)", r.t, *clientName, r.scale.grams))
			}

			tx, err := db.Begin()
			if err != nil {
				return errors.Wrap(err, "begin")
			}

			if len(trackerInserts) > 0 {
				q := sqlf.Sprintf(`INSERT INTO distance(time, location, cat, distance) VALUES %s;`, sqlf.Join(trackerInserts, ", "))
				_, err = db.Exec(q.Query(sqlf.PostgresBindVar), q.Args()...)
				if err != nil {
					tx.Rollback()
					return errors.Wrap(err, "insert distance")
				}
			}

			q := sqlf.Sprintf(`INSERT INTO scale(time, location, weight_g) VALUES %s;`, sqlf.Join(scaleInserts, ", "))
			_, err = db.Exec(q.Query(sqlf.PostgresBindVar), q.Args()...)
			if err != nil {
				tx.Rollback()
				return errors.Wrap(err, "insert scale")
			}
			if err := tx.Commit(); err != nil {
				tx.Rollback()
				return errors.Wrap(err, "commit")
			}
			buffer = buffer[maxResults:]
			return nil
		}
		for {
			time.Sleep(1 * time.Second)
			bufferMu.Lock()
			if len(buffer) > 0 {
				if err := sendBuffered(); err != nil {
					log.Println("sendBuffered:", err)
				}
			}
			bufferMu.Unlock()
		}
	}()
	for {
		select {
		case ev := <-t.results:
			if *trackerVerbose {
				fmt.Println("tracker:", ev.addr, ev.rssi)
			}
			latestTrackers[ev.addr] = ev
		empty:
			for {
				select {
				case ev := <-t.results:
					if *trackerVerbose {
						fmt.Println("tracker:", ev.addr, ev.rssi)
					}
					latestTrackers[ev.addr] = ev
				default:
					break empty
				}
			}
			updatedTrackers = true

			if updatedScale && updatedTrackers {
				bufferMu.Lock()
				buffer = append(buffer, newResult(latestScale, latestTrackers))
				bufferMu.Unlock()
			}
		case latestScale = <-s.results:
			if *scaleVerbose {
				fmt.Println("scale:", "raw:", int(latestScale.raw), "\t\tgrams:", int(latestScale.grams))
			}
			updatedScale = true

			if updatedScale && updatedTrackers {
				bufferMu.Lock()
				buffer = append(buffer, newResult(latestScale, latestTrackers))
				bufferMu.Unlock()
			}
		}
	}
}

type result struct {
	t        time.Time
	scale    scaleResult
	trackers map[string]trackerResult
}

func newResult(scale scaleResult, trackers map[string]trackerResult) result {
	cpy := make(map[string]trackerResult, len(trackers))
	for k, v := range trackers {
		cpy[k] = v
	}
	return result{
		t:        time.Now(),
		scale:    scale,
		trackers: cpy,
	}
}
