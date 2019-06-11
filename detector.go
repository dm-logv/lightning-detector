package main

import (
	"bytes"
	"fmt"
	"html/template"
	"image"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	_ "image/jpeg"
)

const (
	// HOST is a webserver host
	HOST = "127.0.0.1"
	// PORT is a webserver port
	PORT = 8080

	// PAGE_TPL is an HTML Page template
	PAGE_TPL = `
		<!DOCTYPE html>
		<html>
			<head></head>
			<body>
				<table>
					{{.}}
				</table>
			</body>
		</html>
	`
	// ROW_TPL is an HTML Table row template
	ROW_TPL = `
		<tr>
			<td><img class="image" src="{{.imageSrc}}" width="500" /></td>
			<td><img class="hist"  src="{{.histSrc}}" /></td>
		</tr>
	`
)

// maxUint32 returns a maximum of uint32 values
func maxUint32(numbers ...uint32) uint32 {
	currentMax := numbers[0]
	for i := 1; i < len(numbers); i++ {
		if currentMax < numbers[i] {
			currentMax = numbers[i]
		}
	}

	return currentMax
}

// makeHistogramm collects average brightness of pixels
func makeHistogram(m *image.Image) *[256]uint32 {
	var hist [256]uint32

	b := (*m).Bounds()
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			r, g, b, _ := (*m).At(x, y).RGBA()

			value := (r + g + b) / 257 / 3

			hist[value]++
		}
	}

	return &hist
}

// plotTty prints histogram by values in STDOUT
func plotTty(values *[256]uint32) {
	hist := *values

	for level := uint32(0); level < uint32(len(hist)-3); level += 3 {
		// Average two values
		ave := maxUint32(hist[level], hist[level+1], hist[level+2])

		// Plot
		fmt.Printf("%d %s\n", level, strings.Repeat(".", int(ave)*20/10000))
	}
}

// parseNPlot is an experiments :)
func parseNPlot(url string) {
	resp, err := http.Get(url)
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
	hist := *makeHistogram(&m)

	// Graph
	plotTty(&hist)
}

// server provieds webserver
func serve(host string, port int, folder string) error {
	http.HandleFunc("/hist", imagesPageHandler)
	http.Handle("/", http.FileServer(http.Dir(folder)))

	server := fmt.Sprintf("%s:%d", host, port)
	log.Printf("Started at https://%s/hist\n", server)

	err := http.ListenAndServe(server, nil)

	return err
}

// imagesPageHandler builds images pages
func imagesPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", (*r).Method, (*r).URL.String())

	rows := bytes.NewBufferString("")

	for _, image := range *loadImages(".") {
		if tmpl, err := template.New("Row").Parse(ROW_TPL); err != nil {
			log.Fatal(err)
		} else {
			data := map[string]interface{}{"imageSrc": image}
			if err = tmpl.Execute(rows, data); err != nil {
				log.Fatal(err)
			}
		}
	}

	if tmpl, err := template.New("Page").Parse(PAGE_TPL); err != nil {
		log.Fatal(err)
	} else {
		if err = tmpl.Execute(w, template.HTML(rows.String())); err != nil {
			log.Fatal(err)
		}
	}
}

// loadImages returns file names in the given folder
func loadImages(path string) *[]string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	var images []string
	for _, file := range files {
		images = append(images, file.Name())
	}

	return &images
}

func main() {
	// parseNPlot(URL)
	serve(HOST, PORT, ".")
}
