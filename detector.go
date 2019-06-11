package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"text/template"

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

	// BASE_TPL is an Base64 PNG image template
	BASE_TPL = `data:image/png;base64,{{.}}`
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

// openImage opens and decode image from file
func openImage(path string) *image.Image {
	log.Printf("Open \"%s\"", path)

	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}

	m, _, err := image.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	return &m
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

// plotImage builds graphical representation of the histogram
func plotImage(values *[256]uint32, rect *image.Rectangle) *image.Image {
	m := image.NewRGBA(*rect)

	for x, value := range *values {
		for y := uint32(0); y < value; y++ {
			m.Set(int(x), (*rect).Max.Y-int(y/1000), color.Black)
		}
	}

	var img image.Image = m
	return &img
}

// encodeToBase64 encodes image.Image to the HTML base64 image string
func encodeToBase64(m *image.Image) string {
	buff := new(bytes.Buffer)

	if err := png.Encode(buff, *m); err != nil {
		log.Fatal(err)
	}

	s := base64.StdEncoding.EncodeToString(buff.Bytes())
	base := bytes.NewBufferString("")

	if tmpl, err := template.New("Base").Parse(BASE_TPL); err != nil {
		log.Fatal(err)
	} else {
		if err = tmpl.Execute(base, s); err != nil {
			log.Fatal(err)
		}
	}

	return base.String()
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
	rowsChan := make(chan string)
	var wg sync.WaitGroup

	rect := image.Rect(0, 0, 256, 256)

	for _, path := range *loadImages(".") {
		if tmpl, err := template.New("Row").Parse(ROW_TPL); err != nil {
			log.Fatal(err)
		} else {
			wg.Add(1)

			go func(path string) {
				defer wg.Done()

				row := bytes.NewBufferString("")

				m := openImage(path)
				histData := makeHistogram(m)
				hist := plotImage(histData, &rect)
				base := encodeToBase64(hist)

				data := map[string]interface{}{"imageSrc": path, "histSrc": base}

				if err = tmpl.Execute(row, data); err != nil {
					log.Fatal(err)
				} else {
					rowsChan <- row.String()
				}
			}(path)
		}
	}

	// Collect row templates
	go func() {
		for row := range rowsChan {
			rows.WriteString(row)
		}
	}()

	// Wait for completion
	wg.Wait()

	if tmpl, err := template.New("Page").Parse(PAGE_TPL); err != nil {
		log.Fatal(err)
	} else {
		if err = tmpl.Execute(w, rows.String()); err != nil {
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
