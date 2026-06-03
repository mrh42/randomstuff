package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"golang.org/x/net/html"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"os"
	"github.com/anyascii/go"
	//readability "codeberg.org/readeck/go-readability/v2"
)

// OpenAI API Response Structures
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type CoronerData struct {
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	CityOfResidence string `json:"city_of_residence"`
	LocationOfDeath string `json:"location_of_death"`
}

const example =  `{"coroner_file_number": "2026-04127",
  "first_name": "Alfredo",
  "last_name": "Ortega",
  "gender": "Male",
  "age": 23",
 "city_of_residence": "Rialto",
  "next_of_kin_notified": true,
  "date_of_death": "05/31/2026",
  "time_of_death": "0655 Hours",
  "date_of_injury": "05/31/2026",
  "time_of_injury": "0557 Hours",
  "location_of_death": "Riverside Community Hospital, 4445 Magnolia Avenue, Riverside, CA 92501",
  "location_of_injury": "34.018287, -117.514078, Riverside, CA 92509",
  "agency_investigating": "California Highway Patrol- Riverside Division"
}`

func extractEntitiesWithVLLM(rawdata string) string {
	apikey := os.Getenv("HUGGING_FACE_HUB_TOKEN")
	//url := "http://100.64.0.9:8000/v1/chat/completions" // Update port if your vLLM differs
	url := "https://router.huggingface.co/v1/chat/completions"

	// Construct the prompt
	systemPrompt := "You are a data extraction bot. Extract the victim's information from the raw data provided. Output ONLY raw JSON like this:" + example

	userPrompt := fmt.Sprintf("Extract data from this : %s", rawdata)

	// Build the payload
	payload := map[string]interface{}{
		//"model": "google/gemma-4-26B-A4B-it",
		//"model": "google/gemma-4-31B-it",
		"model": "openai/gpt-oss-120b",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		// This forces the model to output valid JSON
		"response_format": map[string]string{"type": "json_object"},
		"temperature":     0.1, // Low temp for extraction tasks
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apikey))

	// Execute the HTTP Request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse JSON response: %v\nRaw: %s\n", err, string(body))
		return ""
	}
	if chatResp.Error.Message != "" {
		fmt.Fprintf(os.Stderr, "API Error: %s\n", chatResp.Error.Message)
		os.Exit(1)
	}

	if len(chatResp.Choices) == 0 {
		fmt.Fprintf(os.Stderr, "No choices, API Error: %s\n", chatResp.Error.Message)
		return ""
	}
	return chatResp.Choices[0].Message.Content
}

func sanitizeOutput(text string) string {
	replacer := strings.NewReplacer(
		"\u2018", "'",   // Left single curly quote
		"\u2019", "'",   // Right single curly quote (apostrophe)
		"\u201C", "\"",  // Left double curly quote
		"\u201D", "\"",  // Right double curly quote
		"\u2013", "-",   // En-dash
		"\u2014", "--",  // Em-dash
		"\u2026", "...", // Ellipsis
	)
	
	return replacer.Replace(text)
}

//        Write the reports  using basic HTML tags (<b>, <h2>, <ul>, <li>, <br>). Do not use Markdown.

func generateBrief(coronerJSON string, newsContext string) string {
	apikey := os.Getenv("HUGGING_FACE_HUB_TOKEN")
	url := "https://router.huggingface.co/v1/chat/completions"

	// 1. Formulate the Intelligence Analyst Prompt
	systemPrompt := `You are an Analyst. Your job is to read the official Coroner's data and compare it against public news reports. 
	Write a concise, professional brief summarizing the incident.
        Add details or interesting facts from media to your report that are not present in the official coroner report.
        The new feed might include aticles not related to our coroner victim, be mindful of that possibility.
	Do not hallucinate. If the news reports do not mention the victim, state that there is no public media coverage of the incident.`

	userPrompt := fmt.Sprintf("CORONER DATA:\n%s\n\nPUBLIC NEWS CONTEXT:\n%s", coronerJSON, newsContext)

	// 2. Build the Payload (Notice we DO NOT enforce JSON here)
	payload := map[string]interface{}{
		"model": "google/gemma-4-31B-it",
		//"model": "openai/gpt-oss-120b",
		//"model": "google/gemma-4-26B-A4B-it",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.4, // Slightly higher temp for better narrative flow
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apikey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("Error generating brief: %v", err)
	}
	defer resp.Body.Close()

	// 3. Parse the standard vLLM response
	body, _ := io.ReadAll(resp.Body)
	
	var vllmResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	
	json.Unmarshal(body, &vllmResponse)
	
	if len(vllmResponse.Choices) > 0 {
		return vllmResponse.Choices[0].Message.Content
	}
	return "Failed to generate brief."
}



func fetchPressReleaseLinks() []string {
	var links []string
	
	resp, _ := http.Get("https://www.riversidesheriff.org/m/newsflash?cat=6")
	defer resp.Body.Close()

	// Tokenize the messy HTML
	z := html.NewTokenizer(resp.Body)

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break // End of document
		}

		if tt == html.StartTagToken {
			t := z.Token()
			
			// Look for <a> tags
			if t.Data == "a" {
				// Check attributes for class="article-title-link"
				isTargetLink := false
				href := ""
				
				for _, a := range t.Attr {
					if a.Key == "class" && strings.Contains(a.Val, "article-title-link") {
						isTargetLink = true
					}
					if a.Key == "href" {
						href = a.Val
					}
				}

				// If it matches, reconstruct the full URL and save it
				if isTargetLink && href != "" {
					fullURL := "https://www.riversidesheriff.org" + href
					links = append(links, fullURL)
				}
			}
		}
	}
	return links
}

func getArticleContent(url string) string {
	fmt.Fprintf(os.Stderr, "fetching: %s\n", url)
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	data := string(body)
	start := strings.Index(data, "<div class=\"article-content")
	if start > 0 {
		data = data[start:]
		end := strings.Index(data, "<div class=\"article-footer")
		if end > 0 {
			data = data[:end]
		}
	}
	return data
}

// XML Structures for Bing News RSS
type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Items []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
}

func fetchNews(feed, url string) string {
	fmt.Fprintf(os.Stderr, "Searching News: %s\n", url)
	
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Sprintf("Error fetching news: %v", err)
	}
	defer resp.Body.Close()

	var rss RSS
	if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
		return fmt.Sprintf("Error parsing XML: %v", err)
	}

	if len(rss.Channel.Items) == 0 {
		return feed + ": No news articles found."
	}
	fmt.Fprintf(os.Stderr, "items: %d\n", len(rss.Channel.Items))

	var newsSummary strings.Builder
	newsSummary.WriteString(fmt.Sprintf("a total of %d news items were found\n"))
	for i, item := range rss.Channel.Items {
		if i >= 4 {
			break
		}

		//link := item.Link
		//fmt.Fprintf(os.Stderr, "link: %s\n", link)
		//article, err := readability.FromURL(link, 5*time.Second)
		newsSummary.WriteString(fmt.Sprintf("title: %s\ndescription: %s\n", feed, item.Title, item.Description))
	}

	return strings.TrimSpace(newsSummary.String())
}


func fetchBingNews(firstName, lastName, city string) string {
	city = strings.Replace(city, ",", "", -1)
	query := fmt.Sprintf("%s %s %s", lastName, firstName, city)
	encodedQuery := url.QueryEscape(query)
	url := fmt.Sprintf("https://www.bing.com/news/search?format=RSS&q=%s", encodedQuery)
	return fetchNews("bing", url)
}

func fetchGoogleNews(firstName, lastName, city string) string {

	city = strings.Replace(city, ",", "", -1)
	query := fmt.Sprintf("%s %s %s", lastName, firstName, city)
	encodedQuery := url.QueryEscape(query)
	url := fmt.Sprintf("https://news.google.com/rss/search?q=%s", encodedQuery)
	return fetchNews("google", url)
}


func main() {
	links := fetchPressReleaseLinks()
	//links = []string{"https://www.riversidesheriff.org/m/newsflash/Home/Detail/7181"}

	for _, l := range links {
		data := getArticleContent(l)
		
		//fmt.Printf("(data): %s\n", data)
		jdata := extractEntitiesWithVLLM(data)
		//fmt.Printf("extracted json: %s\n", jdata)

		var coroner CoronerData
		if err := json.Unmarshal([]byte(jdata), &coroner); err != nil {
			fmt.Println("Error parsing extracted JSON:", err)
			continue
		}

		// 3. Search Bing News using the structured data
		if coroner.LastName != "" && coroner.FirstName != "" {
			//newsContext := fetchBingNews(coroner.FirstName, coroner.LastName, coroner.CityOfResidence)

			newsContext := fetchGoogleNews(coroner.FirstName, coroner.LastName, coroner.CityOfResidence)
			//fmt.Fprintf(os.Stderr, "news: %s\n", newsContext)
			
			finalBrief := generateBrief(jdata, newsContext)
			out := anyascii.Transliterate(finalBrief)

			//out := sanitizeOutput(finalBrief)
			fmt.Println(out)
			fmt.Println("\n----------\n")
		} else {
			fmt.Println("Could not extract a valid name to search.")
		}
		
	}
}


