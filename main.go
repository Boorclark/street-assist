package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gocolly/colly"
)

type Shelter struct {
	Image      string
	Name       string
	Description string
	SeeMore string
}

var shelters []Shelter
func informationPage(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	city := r.FormValue("city")
	url := fmt.Sprintf("https://www.homelessshelterdirectory.org/city/%s-%s", state, city)
	fmt.Println("State:", state)
	fmt.Println("City:", city)
	// Create a new Collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.homelessshelterdirectory.org"),
	)

	// OnHTML callback for each shelter
	c.OnHTML("div.layout_post_2", func(e *colly.HTMLElement) {
		// Create a new Shelter instance and set its fields
		shelter := Shelter{
			Image:       e.ChildAttr("img", "src"),
			Name:        e.ChildText("h4"),
			Description: e.ChildText("p"),
			SeeMore:     e.ChildAttr("a.btn_red", "href"),
		}

		// Append the Shelter to the list
		shelters = append(shelters, shelter)
	})

	// OnError callback to handle errors
	c.OnError(func(_ *colly.Response, err error) {
		log.Printf("Error scraping: %s", err.Error())
	})

	// OnScraped callback to execute once the scraping is done
	c.OnScraped(func(_ *colly.Response) {
		// Parse the information template
		tmpl, err := template.ParseFiles("templates/information.html")
		if err != nil {
			log.Fatal(err)
		}

		// Generate the HTML and write it to the response
		if err := tmpl.Execute(w, shelters); err != nil {
			log.Fatal(err)
		}
	})

	// Start the scraping process
	if err := c.Visit(url); err != nil {
		log.Printf("Error visiting %s: %s", url, err.Error())
	}
}


func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			http.Redirect(w, r, "/resources.html", http.StatusSeeOther)
			return
		}
		http.ServeFile(w, r, "./templates/home.html")
	})
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/resources.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./templates/resources.html")
	})

	http.HandleFunc("/information.html", informationPage)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

