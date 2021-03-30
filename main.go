package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

type bot struct {
	DiscordURL  string `toml:"discordURL"`
	Stock       map[string]map[string]*gpu
	Products    []Product
	writer      *tabwriter.Writer
	lastChecked time.Time
	refresh     time.Duration
}

type Product struct {
	Name string
	URLS []string
}

type gpu struct {
	inStock bool
	url     string
}

func main() {
	welcome := flag.Bool("welcome", false, "send welcome message")
	flag.Parse()

	b := bot{
		Stock:   make(map[string]map[string]*gpu),
		refresh: 10 * time.Minute,
	}

	// Read config file.
	if _, err := toml.DecodeFile("./config.toml", &b); err != nil {
		log.Fatal(err)
	}

	b.writer = new(tabwriter.Writer)
	b.writer.Init(os.Stdout, 8, 8, 3, '\t', 0)

	var count int
	for _, prod := range b.Products {
		for _, _ = range prod.URLS {
			count++
		}
	}
	log.Infof("watching %d gpus", count)

	if *welcome {
		if err := b.startUp(count); err != nil {
			log.Error(err)
		}
	}

	b.checkStock()
	b.printSummary()
	b.lastChecked = time.Now()

	ticker := time.NewTicker(b.refresh)
	for {
		select {
		case <-ticker.C:
			// Update stock list.
			b.checkStock()
			b.printSummary()
			b.lastChecked = time.Now()
		}
	}

}

func (b *bot) printStatus() {
	fmt.Printf("\033[H\033[2J")
	fmt.Printf("\033[A")
	// fmt.Fprintf(b.writer, "%s\t%s\n\n", "GPU", "IN STOCK")

	for _, prod := range b.Products {
		fmt.Fprintf(b.writer, "%s\n\n", prod.Name)
		for name, gpu := range b.Stock[prod.Name] {
			fmt.Fprintf(b.writer, "%s\t%v\n", name, gpu.inStock)
		}
		fmt.Fprintf(b.writer, "\n\n")
	}
	fmt.Fprintf(b.writer, "Last checked: %v", b.lastChecked.Format("Mon Jan 2 15:04:05 MST 2006"))

	b.writer.Flush()
}

func (b *bot) printSummary() {
	for _, prod := range b.Products {
		var count int
		for _, gpu := range b.Stock[prod.Name] {
			if gpu.inStock {
				count++
			}
		}
		log.Infof("%s: %d", prod.Name, count)
	}
}

func (b *bot) checkStock() {
	for _, prod := range b.Products {
		stock, ok := b.Stock[prod.Name]
		if !ok {
			stock = make(map[string]*gpu)
			b.Stock[prod.Name] = stock
		}

		for _, u := range prod.URLS {
			doc, err := getDoc(u)
			if err != nil {
				continue
			}

			name := getName(doc)

			inStock, err := inStock(doc)
			if err != nil {
				log.Error(err)
				continue
			}

			if g, ok := stock[name]; ok {
				if inStock && !g.inStock {
					// Send notification that item has come into stock.
					g.inStock = inStock
					fmt.Printf("%s is now in stock: %s\n", prod.Name, u)
					if err := b.sendDiscord(name, u, prod.Name); err != nil {
						log.Error(err)
					}
				}
				if !inStock && g.inStock {
					// Send notification that item has come into stock.
					g.inStock = inStock
					fmt.Printf("%s is out of stock: %s\n", prod.Name, u)
					if err := b.outOfStock(name, u, prod.Name); err != nil {
						log.Error(err)
					}
				}
			} else {
				stock[name] = &gpu{inStock: inStock, url: u}
				/*if inStock {
					fmt.Printf("%s is in stock: %s\n", prod.Name, u)
					if err := b.sendDiscord(name, u, prod.Name); err != nil {
						log.Error(err)
					}
				}*/
			}
		}
	}
}

/*func (b *bot) sendUpdate(gpu string, url string) {
	attachment := slack.Attachment{
		Color:      "good",
		AuthorName: gpu,
		AuthorLink: url,
		AuthorIcon: "https://avatars2.githubusercontent.com/u/652790",
		Text:       "This GPU is now in stock :smile",
		FooterIcon: "https://platform.slack-edge.com/img/default_application_icon.png",
		Ts:         json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
	}
	msg := slack.WebhookMessage{
		Attachments: []slack.Attachment{attachment},
	}

	err := slack.PostWebhook(webhookURL, &msg)
	if err != nil {
		fmt.Println(err)
	}
}*/

func getDoc(url string) (*goquery.Document, error) {
	// Get the HTML
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	// Convert HTML into goquery document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func getName(doc *goquery.Document) string {
	sel := doc.Find(".product-hero__title")
	return sel.Text()
}

func getModel(doc *goquery.Document) string {
	sel := doc.Find(".product-hero__key-selling-point")
	return sel.Text()
}

func getPic(doc *goquery.Document) string {
	sel := doc.Find(".image-gallery__hero .js-gallery-trigger")
	for _, n := range sel.Nodes {
		if n.FirstChild != nil {
			fmt.Println(*n.FirstChild)
		}
	}
	return ""
}

func inStock(doc *goquery.Document) (bool, error) {
	sel := doc.Find(".purchase-info__price .inc-vat .price")
	if len(sel.Nodes) == 0 {
		return false, nil
	}

	/*sel.Each(func(i int, s *goquery.Selection) {
		fmt.Println(strings.TrimSpace(s.Text()))
	})*/

	return true, nil
}
