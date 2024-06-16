package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"slices"
	"strings"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/vinted/data"
)

const (
	UK_BASE_URL = "https://www.vinted.co.uk"
	FR_BASE_URL = "https://www.vinted.fr"
	DE_BASE_URL = "https://www.vinted.de"
	PL_BASE_URL = "https://www.vinted.pl"
)

// var session_channel chan string
var error_channel chan string
var product_channel chan Item

// together with the channels it would be better to have it in a struct Monitor
var FOUND_SKU_UK []int

type Client struct {
	TlsClient *tls_client.HttpClient
	url       string
}
type Sessions struct {
	UK_session string
	FR_session string
	// 34DE_session string
	// add more
}
type Options struct {
	settings []tls_client.HttpClientOption
}
type Catalog struct {
	Items []Item `json:"items"`
}
type Item struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Photo     Photo  `json:"photo"`
	SizeTitle string `json:"size_title"`
	Price     string `json:"price"`
}

type Photo struct {
	URL string `json:"url"`
}

func NewClient(_url string) (*Client, error) {
	options := Options{
		settings: []tls_client.HttpClientOption{
			tls_client.WithTimeoutSeconds(15),
			tls_client.WithClientProfile(profiles.Chrome_124),
			tls_client.WithNotFollowRedirects(),
		},
	}
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options.settings...)
	if err != nil {
		return nil, err
	}

	return &Client{TlsClient: &client, url: _url}, nil
}
func main() {
	// initializing channels

	//session_channel = make(chan string)
	error_channel = make(chan string)
	product_channel = make(chan Item)
	//new client to fetch session header
	clientUk, err := NewClient(UK_BASE_URL)
	if err != nil {
		log.Println("Error creating Client to retrive session Header: " + err.Error())
	}

	//only have 1 client as for now, so reuse clientUK
	go read_errors()
	go read_prods()
	go new_product_monitor(clientUk)

	select {}

}
func read_prods() {
	for {
		prod := <-product_channel
		send_webhook("https://discordapp.com/api/webhooks/1252001796361162843/5bzY987n-6-hibjcAtmJbPfZoICSfGbV6TDT_XUn_Xr7izhYxBatv4tn3RHuFxczRByZ", prod)
		log.Println("prod	:", prod)
	}
}
func send_webhook(u string, prod Item) {
	webhook := &data.Webhook{}
	webhook.SetContent("This is a test webhook")

	// Create an embed
	embed := data.Embed{}
	embed.SetTitle("Vinted Monitor (New Prod)")
	embed.SetColor(0x00ff00) // Green color
	embed.SetThumbnail(prod.Photo.URL)
	embed.SetDescription("New Product Detected ")

	embed.AddField("Title:", prod.Title, false)

	embed.AddField("URL: ", prod.URL, false)
	embed.AddField("Price:", prod.Price, false)

	// Add the embed to the webhook
	webhook.AddEmbed(embed)

	// Send the webhook to the specified URL
	webhookURL := u
	err := webhook.Send(webhookURL)
	if err != nil {
		fmt.Println("Error sending webhook:", err)
	} else {
		fmt.Println("Webhook sent successfully")
	}
}
func new_product_monitor(client *Client) {

	for {
		session := get_session(client)

		url := "https://www.vinted.co.uk/api/v2/catalog/items?page=1&per_page=96&search_text=&catalog_ids=&order=newest_first&size_ids=&brand_ids=&status_ids=&color_ids=&material_ids="
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			error_channel <- "Retrying, Error occured: " + err.Error()
			time.Sleep(2 * time.Second)
			continue
		}

		req.Header = http.Header{}

		req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		req.Header.Add("accept-language", "en-US,en;q=0.9")
		req.Header.Add("cookie", fmt.Sprintf("_vinted_fr_session=%s", session))
		req.Header.Add("sec-ch-ua", "\"Google Chrome\";v=\"125\", \"Chromium\";v=\"125\", \"Not.A/Brand\";v=\"24\"")
		req.Header.Add("sec-ch-ua-mobile", "?0")
		req.Header.Add("sec-ch-ua-platform", "\"Windows\"")
		req.Header.Add("upgrade-insecure-requests", "1")
		req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36")

		resp, err := (*client.TlsClient).Do(req)
		if err != nil {
			error_channel <- "Retrying, Error occured while firing the request: " + err.Error()
			time.Sleep(2 * time.Second)
			continue
		}
		if resp.StatusCode == 200 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				error_channel <- "Retrying, Error occured: " + err.Error()
				resp.Body.Close()
				time.Sleep(2 * time.Second)
				continue
			}
			var catalog Catalog
			err = json.Unmarshal(body, &catalog)
			if err != nil {
				error_channel <- "Retrying, Error occured: " + err.Error()
				return
			}
			log.Println("Succesfully made request to products page")

			if len(FOUND_SKU_UK) <= 0 {
				for _, item := range catalog.Items {
					FOUND_SKU_UK = append(FOUND_SKU_UK, item.ID)

				}
				log.Println("First iteration finished")
				continue
			}
			for _, item := range catalog.Items {
				if !slices.Contains(FOUND_SKU_UK, item.ID) {
					FOUND_SKU_UK = append(FOUND_SKU_UK, item.ID)
					product_channel <- item
					log.Println("Found New Item with SKU: ", item.ID)
				}
			}

		} else {
			error_channel <- fmt.Sprintf("Status Code [%d]", resp.StatusCode)
			time.Sleep(2 * time.Second)

			continue
		}

	}

}
func read_errors() {
	for {
		err := <-error_channel
		log.Println("Error:", err)
	}
}
func get_session(client *Client) string {

	req, err := http.NewRequest(http.MethodHead, client.url, nil)
	if err != nil {
		error_channel <- "Retrying fetching session, Error occured: " + err.Error()

	}

	req.Header = http.Header{}
	req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Add("accept-language", "en-US,en;q=0.9")
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("pragma", "no-cache")
	req.Header.Add("priority", "u=0, i")
	req.Header.Add("sec-ch-ua", "\"Google Chrome\";v=\"125\", \"Chromium\";v=\"125\", \"Not.A/Brand\";v=\"24\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Add("sec-fetch-dest", "document")
	req.Header.Add("sec-fetch-mode", "navigate")
	req.Header.Add("sec-fetch-site", "none")
	req.Header.Add("sec-fetch-user", "?1")
	req.Header.Add("upgrade-insecure-requests", "1")
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36")

	resp, err := (*client.TlsClient).Do(req)
	if err != nil {
		error_channel <- "Retrying fetching session, Error occured: " + err.Error()

	}

	resp.Body.Close()

	cookie := resp.Header["Set-Cookie"]
	session := extractSessionCookie(cookie)
	if session == "" {
		error_channel <- "No valid session cookie found"

	}

	log.Println("Succesfully fetched Session cookie")

	return session

}

func extractSessionCookie(cookies []string) string {
	for _, cookie := range cookies {
		if strings.Contains(cookie, "_vinted_fr_session") {
			return strings.Split(strings.Split(cookie, "_vinted_fr_session=")[1], ";")[0]
		}
	}
	error_channel <- "Couldn't find Cookie!"
	return ""
}
