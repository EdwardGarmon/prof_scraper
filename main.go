package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gocolly/colly"
)

type name struct {
	FirstName string
	LastName  string
}

type pack struct {
	Text string `json:"text"`
}

type rating struct {
	Values []value `json:"values"`
}

type value struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

const searchURL = "https://www.ratemyprofessors.com/search.jsp?queryoption=HEADER&queryBy=teacherName&schoolName=The+University+of+Texas+at+Dallas&schoolID=1273&query="

const googleSearchURL = "https://www.google.com/search?q="
const googleAddons = "+university+of+texas+at+dallas"

func scrapeRMP(searchTerm name) rating {

	c := colly.NewCollector()

	var values []value

	c.OnHTML(".main", func(e *colly.HTMLElement) {

		var listing string
		listing = e.Text
		var parts = strings.Split(listing, ",")

		var firstName = strings.ToLower(strings.TrimSpace(parts[1]))
		var lastName = strings.ToLower(strings.TrimSpace(parts[0]))

		if firstName == searchTerm.FirstName &&
			lastName == searchTerm.LastName {

			schoolName := e.DOM.Parent().ChildrenFiltered(".sub").Text()

			if strings.Contains(schoolName, "Dallas") {
				link, found := e.DOM.Parent().Parent().Attr("href")
				if found {
					e.Request.Visit(link)

				}

			}

		}
	})

	c.OnHTML(".RatingValue__Numerator-qw8sqy-2", func(e *colly.HTMLElement) {

		values = append(values, value{"Quality", e.Text})
	})

	c.OnHTML(".FeedbackItem__FeedbackNumber-uof32n-1", func(e *colly.HTMLElement) {

		description := e.DOM.Parent().
			ChildrenFiltered(".FeedbackItem__FeedbackDescription-uof32n-2").Text()

		switch description {
		case "Level of Difficulty":
			values = append(values, value{"Difficulty", e.Text})
			break
		case "Would take again":
			values = append(values, value{"Would Take Again", e.Text})
			break

		}

	})

	str := []string{searchURL, searchTerm.FirstName, "+", searchTerm.LastName}
	var searchCompl = strings.Join(str, "")

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit(searchCompl)
	c.Wait()

	if len(values) == 0 {
		link, err := scrapeGoogle(searchTerm)
		if err == nil {
			c.Visit(link)
		}

	} else {
		cache[searchTerm] = rating{values}
	}

	c.Wait()

	var ratingData = rating{values}
	return ratingData

}

func scrapeGoogle(searchTerm name) (string, error) {

	lastlink := ""

	targetURL := strings.Join([]string{googleSearchURL, searchTerm.FirstName, "+", searchTerm.LastName, googleAddons}, "")

	c := colly.NewCollector()

	c.OnHTML("a", func(e *colly.HTMLElement) {

		searchLink := strings.ToLower(e.Attr("href"))
		if strings.Contains(searchLink, "tid") {
			link := e.Attr("href")
			lastlink = strings.Join([]string{"https://www.google.com", link}, "")
		}

	})

	c.Visit(targetURL)
	c.Wait()

	if len(lastlink) == 0 {
		return lastlink, errors.New("professor not found")
	}

	return lastlink, nil
}

func greet(w http.ResponseWriter, r *http.Request) {
	var jsonData, _ = json.Marshal(pack{"helloworld"})
	w.Write(jsonData)
}

func rateProf(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var req name
	for key := range r.Form {
		fmt.Println(key)
		json.Unmarshal([]byte(key), &req)
	}

	cachedRating, ok := cache[req]

	if ok {
		fmt.Println("cache hit! -- sending cached data")
		jsonData, _ := json.Marshal(cachedRating)
		w.Write(jsonData)
	} else {
		jsonData, _ := json.Marshal(scrapeRMP(req))
		w.Write(jsonData)
	}

}

var cache = make(map[name]rating)

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("Port not set --defaulting to 8080")
		port = "8080"
	}

	http.HandleFunc("/rateProf", rateProf)
	http.HandleFunc("/", greet)

	fmt.Printf("listening on %v\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
