package main

import (
	"encoding/json"
	"errors"
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
	UK_CURRENCY = "£"
	FR_CURRENCY = "€"
	DE_CURRENCY = "€"
	PL_CURRENCY = "zł"
)
const (
	UK_WEBHOOK = "https://discordapp.com/api/webhooks/1252001796361162843/5bzY987n-6-hibjcAtmJbPfZoICSfGbV6TDT_XUn_Xr7izhYxBatv4tn3RHuFxczRByZ"
	FR_WEBHOOK = "https://discordapp.com/api/webhooks/1252276637320609803/UO43qxvtq-zJvhgjWIcE5sh8rYZNcyB8cEC1n1SHT5o8QEqA2F64gJQShg-bM-Eu3cAF"
	DE_WEBHOOK = "https://discordapp.com/api/webhooks/1252276571931410522/elu93W1O7uqYY3kaQzaGT4xcgZhT9SPgmffjAoH_r8dlXFGNnKGfWRXD0T35CNT8klgN"
	PL_WEBHOOK = "https://discordapp.com/api/webhooks/1252276519913525400/erbkXOsWU0QQoUqME_gbGepgqJMi5ZZywR-TcokbTUJSsEOPN5QuiT4D7JQVy1eRt1w7"
)
const (
	UK_BASE_URL = "https://www.vinted.co.uk"
	FR_BASE_URL = "https://www.vinted.fr"
	DE_BASE_URL = "https://www.vinted.de"
	PL_BASE_URL = "https://www.vinted.pl"
)

// var session_channel chan string
var error_channel chan string
var product_channel chan data.Item

type Monitor struct {
	Client    Client
	FOUND_SKU []int
}
type Client struct {
	TlsClient *tls_client.HttpClient
	url       string
	Region    string
}

type Options struct {
	settings []tls_client.HttpClientOption
}

func NewClient(_url string, region string) (*Client, error) {
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

	return &Client{TlsClient: &client, url: _url, Region: region}, nil
}
func main() {
	// initializing channels
	error_channel = make(chan string)
	product_channel = make(chan data.Item)

	//new client to fetch session header
	var monitors []*Monitor
	urls := []string{
		UK_BASE_URL,
		FR_BASE_URL,
		DE_BASE_URL,
		PL_BASE_URL,
	}
	for _, url := range urls {
		client, err := NewClient(url, url[len(url)-2:])
		if err != nil {
			log.Println("Error creating Client to retrive session Header: " + err.Error())
		}
		monitor := Monitor{
			Client:    *client,
			FOUND_SKU: []int{},
		}
		monitors = append(monitors, &monitor)
	}
	for _, monitor := range monitors {
		go monitor.new_product_monitor(&monitor.Client)
	}

	go read_errors()
	go read_prods()

	select {}

}
func read_prods() {
	for {
		prod := <-product_channel

		if prod.Region == "uk" {
			send_webhook(UK_WEBHOOK, UK_CURRENCY, prod)
		}
		if prod.Region == "fr" {
			send_webhook(FR_WEBHOOK, FR_CURRENCY, prod)

		}
		if prod.Region == "de" {
			send_webhook(DE_WEBHOOK, DE_CURRENCY, prod)

		}
		if prod.Region == "pl" {
			send_webhook(PL_WEBHOOK, PL_CURRENCY, prod)

		}

	}
}
func read_errors() {
	for {
		err := <-error_channel
		log.Println("Error:", err)
	}
}
func send_webhook(u string, currency string, prod data.Item) {
	webhook := &data.Webhook{}

	// Create an embed
	embed := data.Embed{}
	var title = "Vinted Monitor (New Prod) [" + prod.Region + "]"
	embed.SetTitle(title)
	embed.SetColor(0x00ff00) // Green color
	embed.SetThumbnail(prod.Photo.URL)
	embed.SetDescription("New Product Detected ")

	embed.AddField("Title:", prod.Title, false)

	embed.AddField("URL: ", prod.URL, false)
	var s = "Price " + currency

	embed.AddField(s, prod.Price, false)
	var t = "Total price " + currency
	embed.AddField(t, prod.TotalItemPrice, false)
	embed.AddField("User :", prod.User.ProfileURL, false)
	embed.SetFooter(prod.Timestamp.String(), "")
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
func (m *Monitor) new_product_monitor(client *Client) {

	for {
		session, err := get_session(client)
		if err != nil {
			error_channel <- "[" + m.Client.Region + "]" + "Retrying fetching session, Error occured: " + err.Error()
			time.Sleep(2 * time.Second)
			continue
		}
		url := client.url + "/api/v2/catalog/items?page=1&per_page=96&search_text=&catalog_ids=&order=newest_first&size_ids=&brand_ids=&status_ids=&color_ids=&material_ids="
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			error_channel <- "[" + m.Client.Region + "]" + "Retrying, Error occured: " + err.Error()
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
			error_channel <- "[" + m.Client.Region + "]" + "Retrying, Error occured while firing the request: " + err.Error()
			time.Sleep(2 * time.Second)
			continue
		}
		if resp.StatusCode == 200 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				error_channel <- "[" + m.Client.Region + "]" + "Retrying, Error occured: " + err.Error()
				resp.Body.Close()
				time.Sleep(2 * time.Second)
				continue
			}
			var catalog data.CatalogueItems
			err = json.Unmarshal(body, &catalog)

			if err != nil {
				error_channel <- "[" + m.Client.Region + "]" + "Retrying, Error occured: " + err.Error()
				return
			}
			log.Println("[" + m.Client.Region + "]" + "Succesfully made request to products page")

			if len(m.FOUND_SKU) <= 0 {
				for _, item := range catalog.Items {

					m.FOUND_SKU = append(m.FOUND_SKU, int(item.ID))

				}
				log.Println("[" + m.Client.Region + "]" + "First iteration finished")
				continue
			}
			for _, item := range catalog.Items {
				if !slices.Contains(m.FOUND_SKU, int(item.ID)) {

					//append to found skus
					m.FOUND_SKU = append(m.FOUND_SKU, int(item.ID))

					//set few personalized informations
					item.Region = m.Client.Region
					currentTime := time.Now()
					timestamp := currentTime.Round(time.Millisecond)
					item.Timestamp = timestamp

					//broadcast product to channel
					product_channel <- item
					log.Println("["+m.Client.Region+"]"+"Found New Item with SKU: ", item.ID)
				}
			}

		} else {
			error_channel <- fmt.Sprintf("["+m.Client.Region+"]"+"Status Code [%d]", resp.StatusCode)
			time.Sleep(2 * time.Second)

			continue
		}

	}

}

func get_session(client *Client) (string, error) {

	req, err := http.NewRequest(http.MethodHead, client.url, nil)
	if err != nil {
		return "", err
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
		return "", err
	}

	defer resp.Body.Close()

	cookie := resp.Header["Set-Cookie"]
	session := extractSessionCookie(cookie)
	if session == "" {
		return "", errors.New("no valid session cookie found")

	}

	log.Println("Succesfully fetched Session cookie")

	return session, nil

}

func extractSessionCookie(cookies []string) string {
	for _, cookie := range cookies {
		if strings.Contains(cookie, "_vinted_fr_session") {
			return strings.Split(strings.Split(cookie, "_vinted_fr_session=")[1], ";")[0]
		}
	}
	return ""
}
