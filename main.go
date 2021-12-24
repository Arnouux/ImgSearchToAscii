package main

import (
	"bytes"
	"fmt"
	"html/template"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/gocolly/colly"
	"github.com/nfnt/resize"
)

var templates = template.Must(template.ParseFiles("main.html"))
var levels = []string{" ", "░", "▒", "▓", "█"}

type Page struct {
	Title string
	Body  []byte
	Text  string
}

// handler for main.html
func mainHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		templates.ExecuteTemplate(w, "main.html", &Page{})
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Println(err)
			return
		}
		item := r.FormValue("item")

		// Sanitize input
		title := strings.Title(item)
		item = strings.Replace(item, " ", "_", -1)

		image, err := searchItem(item, r.UserAgent())
		if err != nil {
			fmt.Println(err)
			return
		}

		newImage := resize.Resize(50, 0, image, resize.Lanczos3)
		if err != nil {
			fmt.Println(err)
			return
		}
		var sb strings.Builder
		for y := newImage.Bounds().Min.Y; y < newImage.Bounds().Max.Y; y++ {
			for x := newImage.Bounds().Min.X; x < newImage.Bounds().Max.X; x++ {
				pixel := color.GrayModel.Convert(newImage.At(x, y)).(color.Gray)
				level := pixel.Y / 51 // 51 * 5 = 255
				if level == 5 {
					level--
				}
				sb.WriteString(levels[level])
			}
			sb.WriteString("\n")
		}

		templates.ExecuteTemplate(w, "main.html", &Page{Title: title, Text: sb.String()})

	}

	// renderTemplate(w, "main", &Page{Body: body})
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Useragent can't seem to be usable
func searchItem(item string, userAgent string) (image.Image, error) {

	imgFound := false
	counter := 0
	img := image.Image(nil)

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
	)

	c2 := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
	)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnHTML("img", func(e *colly.HTMLElement) {
		if e.Attr("alt") == "" && !imgFound {

			random := rand.Intn(10)
			if random <= 2+counter {
				link := e.Attr("src")
				imgFound = true
				c2.Visit(link)
			}
			counter += 1
		}

	})

	c2.OnResponse(func(r *colly.Response) {
		reader := bytes.NewReader(r.Body)
		img, _, _ = image.Decode(reader)
	})

	c.Visit("https://www.google.com/search?q=" + item + "&hl=fr&tbm=isch")

	return img, nil
}

func main() {

	http.HandleFunc("/main/", mainHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
