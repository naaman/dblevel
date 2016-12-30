package main

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"math"
	"os"
	"os/signal"
	"strings"
)

const (
	UNHANDLED_ERROR = 1
)

func main() {
	portaudio.Initialize()
	defer portaudio.Terminate()

	m := newDecibelMeter()
	defer m.Close()

	err := m.Start()
	exitOnError(err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig

	err = m.Stop()
	exitOnError(err)
}

type decibelMeter struct {
	*portaudio.Stream
}

func newDecibelMeter() *decibelMeter {
	h, err := portaudio.DefaultHostApi()
	exitOnError(err)

	device := h.DefaultInputDevice

	// TODO: fix this dumb raspberry pi hack
	allDevices, _ := portaudio.Devices()
	for _, d := range allDevices {
		if strings.HasPrefix(d.Name, "USB Device") {
			device = d
			break
		}
	}

	p := portaudio.LowLatencyParameters(device, nil)

	m := new(decibelMeter)
	m.Stream, err = portaudio.OpenStream(p, processAudio)
	exitOnError(err)

	return m
}

// TODO: use a-weighting and linear filtering
func decibel(d []int16) float64 {
	return math.Log10(rootMeanSquare(d)) * 20.0
}

func rootMeanSquare(d []int16) float64 {
	var total float64

	for _, v := range d {
		total += math.Pow(math.Abs(float64(v)), 2)
	}

	return math.Sqrt(total / float64(len(d)))
}

func processAudio(in, _ []int16) {
	fmt.Println(decibel(in))
}

func exitOnError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(UNHANDLED_ERROR)
	}
}
