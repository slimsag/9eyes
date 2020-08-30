package main

import (
	"context"
	"log"
	"sort"
	"time"

	"github.com/MichaelS11/go-hx711"
	"github.com/pkg/errors"
)

type scaleResult struct {
	time       time.Time
	raw, grams float64
}

type scale struct {
	results       chan scaleResult
	sampleSize    int
	baseline      float64
	divisor       float64
	invert        bool
	movingAverage int
}

func (s *scale) start(ctx context.Context) error {
	err := hx711.HostInit()
	if err != nil {
		return errors.Wrap(err, "HostInit")
	}

	hx711, err := hx711.NewHx711("6", "5")
	if err != nil {
		return errors.Wrap(err, "NewHx711")
	}
	defer hx711.Shutdown()

	err = hx711.Reset()
	if err != nil {
		return errors.Wrap(err, "Reset")
	}

	var previousSamples []float64
	var samples []float64
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		data, err := hx711.ReadDataRaw()
		if err != nil {
			log.Println("Hx711 read error:", err)
			continue
		}
		if data == -1 {
			continue
		}
		if s.invert {
			data = -data
		}
		samples = append(samples, float64(data))
		if len(samples) >= s.sampleSize {
			//hx711.Reset()
			raw := sample(samples)
			samples = nil
			if s.movingAverage != 0 {
				if len(previousSamples) >= s.movingAverage {
					previousSamples = previousSamples[1:]
				}
				previousSamples = append(previousSamples, raw)
				raw = sample(previousSamples)
			}
			s.results <- scaleResult{
				time:  time.Now(),
				raw:   raw,
				grams: (float64(raw) / s.divisor) - s.baseline,
			}
		}

	}
}

func sample(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	mean := avg(samples)
	sort.Slice(samples, func(i, j int) bool {
		return dist(samples[i], mean) < dist(samples[j], mean)
	})
	// Remove 50% of the least average values. e.g.:
	// [-11692 -11730 -11739 -11739 -11746 -11746 -11994 -5854 -5845 -5835]
	// =>
	// [-11692 -11730 -11739 -11739 -11746]
	n := len(samples) / 2
	if n == 0 {
		n = 1
	}
	return avg(samples[:n])
}

func dist(x, y float64) float64 {
	return abs(abs(x) - abs(y))
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func avg(x []float64) float64 {
	n := 0.0
	for _, v := range x {
		n += v
	}
	return n / float64(len(x))
}
