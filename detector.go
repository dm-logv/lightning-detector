package main

import (
	"fmt"
	"image"
	"log"
	"net/http"

	_ "image/jpeg"
)

const (
	URL = "https://cdn.iz.ru/sites/default/files/styles/900x506/public/news-2018-08/20180801_gaf_u39_502.jpg"
)

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
	fmt.Printf("Size: %d x %d", h, w)

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

	fmt.Println(hist)
}
