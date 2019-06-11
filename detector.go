package main

import (
	"fmt"
	"image"
	"log"
	"net/http"
	"strings"

	_ "image/jpeg"
)

const (
	URL = "https://cdn.iz.ru/sites/default/files/styles/900x506/public/news-2018-08/20180801_gaf_u39_502.jpg"
)

func maxUint32(numbers ...uint32) uint32 {
	currentMax := numbers[0]
	for i := 1; i < len(numbers); i++ {
		if currentMax < numbers[i] {
			currentMax = numbers[i]
		}
	}

	return currentMax
}

func main() {
	resp, err := http.Get(URL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	m, _, err := image.Decode(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	g := m.Bounds()

	// Get height and width
	h := g.Dy()
	w := g.Dx()

	// Print results
	fmt.Printf("Size: %d x %d\n\n", h, w)

	// Hist
	b := m.Bounds()
	hist := make(map[uint32]uint32)
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			r, g, b, _ := m.At(x, y).RGBA()
			value := (r + g + b) / 257 / 3
			hist[value]++
		}
	}

	// Graph
	for level := uint32(0); level < 258; level += 3 {
		// Average two values
		ave := maxUint32(hist[level], hist[level+1], hist[level+2])

		// Plot
		fmt.Printf("%d %s\n", level, strings.Repeat(".", int(ave)*20/10000))
	}
}
