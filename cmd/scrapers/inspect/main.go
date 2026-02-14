package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get("https://www.churchofjesuschrist.org/study/scriptures/ot/gen/1?lang=eng")
	if err != nil {
		fmt.Printf("Error fetching: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("Status: %d\n", resp.StatusCode)
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Printf("Error parsing: %v\n", err)
		return
	}

	// Find first 3 footnote li elements
	count := 0
	doc.Find("footer.study-notes li[data-full-marker]").Each(func(i int, li *goquery.Selection) {
		if count >= 3 {
			return
		}

		fullMarker, _ := li.Attr("data-full-marker")
		id, hasID := li.Attr("id")
		class, _ := li.Attr("class")

		fmt.Printf("=== Footnote %d ===\n", count+1)
		fmt.Printf("data-full-marker: %q\n", fullMarker)
		fmt.Printf("id attribute present: %v\n", hasID)
		if hasID {
			fmt.Printf("id: %q\n", id)
		}
		fmt.Printf("class: %q\n", class)

		// Get the full HTML of this li
		html, _ := goquery.OuterHtml(li)
		if len(html) > 800 {
			fmt.Printf("HTML (first 800 chars):\n%s\n\n", html[:800])
		} else {
			fmt.Printf("HTML:\n%s\n\n", html)
		}

		count++
	})

	if count == 0 {
		fmt.Println("No footnotes found with data-full-marker selector.")
		fmt.Println("Searching for 'footer.study-notes'...")
		footer := doc.Find("footer.study-notes")
		fmt.Printf("Found footer.study-notes elements: %d\n", footer.Length())
		if footer.Length() > 0 {
			fmt.Println("Footer exists but no li[data-full-marker] found.")
		}
	}
}
