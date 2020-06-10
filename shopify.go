package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/NoirSneaker/monitor"
)

// Variant is all the variants from a Shopify products
type Variant struct {
	ID               int64       `json:"id"`
	Title            string      `json:"title"`
	Option1          string      `json:"option1"`
	Option2          interface{} `json:"option2"`
	Option3          interface{} `json:"option3"`
	Sku              string      `json:"sku"`
	RequiresShipping bool        `json:"requires_shipping"`
	Taxable          bool        `json:"taxable"`
	FeaturedImage    interface{} `json:"featured_image"`
	Available        bool        `json:"available"`
	Price            string      `json:"price"`
	Grams            int         `json:"grams"`
	CompareAtPrice   interface{} `json:"compare_at_price"`
	Position         int         `json:"position"`
	ProductID        int64       `json:"product_id"`
	CreatedAt        string      `json:"created_at"`
	UpdatedAt        string      `json:"updated_at"`
}

func addShopifyItem(sizeID int, updatedAt string, stock bool, sizeName string) error {
	_, err := db.Exec("insert into shopify(sizeID, updated, stock, size) values($1, $2, $3, $4)", sizeID, updatedAt, stock, sizeName)
	if err != nil {
		return err
	}

	return nil
}

func fetchShopifyItem(variant Variant) (bool, bool, bool, []string, error) {
	var rowsPresent bool
	var restock bool
	var soldOut bool
	var variants []string

	rows, err := db.Query("select sizeID, updated, size from shopify WHERE sizeID = $1", variant.ID)
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		return false, false, false, nil, err
	}
	for rows.Next() {
		rowsPresent = true
		var id int
		var updatedAt string
		var variantName string
		err = rows.Scan(&id, &updatedAt, &variantName)
		if err != nil {
			return false, false, false, nil, err
		}

		if variant.UpdatedAt != updatedAt && variant.Title == variantName {
			if variant.Available == true {
				variants = append(variants, fmt.Sprintf(variant.Title))
				restock = true
				_, err = db.Exec("UPDATE shopify SET updated = $1 WHERE sizeID = $2", variant.UpdatedAt, variant.ID)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				_, err = db.Exec("UPDATE shopify SET updated = $1 WHERE sizeID = $2", variant.UpdatedAt, variant.ID)
				if err != nil {
					log.Fatal(err)
				}
				soldOut = true
			}
		}
	}
	err = rows.Err()
	if err != nil {
		return false, false, false, nil, err
	}

	err = rows.Close()
	if err != nil {
		return false, false, false, nil, err
	}

	return rowsPresent, restock, soldOut, variants, nil
}

func buildShopifyWebhook(name, handle, shopURL, image, variants, price, vendor, productType, typeEvent string) monitor.Webhook {
	urlParsed, err := url.Parse(shopURL)
	if err != nil {
		log.Fatal(err)
	}

	embeds := []monitor.Embeds{
		{
			Title: fmt.Sprintf("%s: %s", typeEvent, name),
			URL:   fmt.Sprintf("%s/products/%v", shopURL, handle),
			Color: 0,
			Fields: []monitor.Fields{
				{
					Name:   "Price",
					Value:  "$" + price,
					Inline: true,
				},
				{
					Name:   "Variant(s)",
					Value:  variants,
					Inline: true,
				},
				{
					Name:   "Vendor",
					Value:  vendor,
					Inline: true,
				},
				{
					Name:   "Product type",
					Value:  productType,
					Inline: true,
				},
			},
			Timestamp: time.Now().Format("2006-01-02T15:04:05-0700"),
			Footer:    monitor.Footer{Text: fmt.Sprintf("%s Monitor by NoirMonitor", strings.ToUpper(urlParsed.Host))},
		},
	}

	if image != "" {
		embeds[0].Thumbnail = monitor.Thumbnail{URL: image}
	}

	return monitor.Webhook{
		Username:  "NoirMonitor",
		AvatarURL: "",
		Embeds:    embeds,
	}
}

func monitorShopify(m *monitor.Monitor, link string) {
	retires := 0
	products, err := m.GetShopifyProducts(link)
	for err != nil {
		retires++
		if retires > 5 {
			return
		}
	}

	var rowsPresent bool
	var soldOut bool
	var restock bool
	var price string

	var variants []string

	for _, product := range products.Products {
		for _, variant := range product.Variants {
			if !contains(keywords, product.Title) {
				return
			}

			retires = 0
			price = variant.Price
			rowsPresent, restock, soldOut, variants, err = fetchShopifyItem(variant)
			if err != nil {
				retires++
				if retires > 5 {
					return
				}
			}

			if !rowsPresent {
				log.Printf("Not in Database: %v. Adding it.", variant.ID)
				var imageURL string
				variantJoined := strings.Join(variants, " ")
				vendor := product.Vendor
				productType := product.ProductType
				if len(product.Images) > 0 {
					imageURL = product.Images[0].Src
				}

				if variantJoined == "" {
					variantJoined = "N/A"
				}

				if vendor == "" {
					vendor = "N/A"
				}

				if productType == "" {
					productType = "N/A"
				}

				webhookSend := buildShopifyWebhook(product.Title, product.Handle, link, imageURL, variantJoined, variant.Price, vendor, productType, "NEW PRODUCT")
				err = m.SendDiscordWebhook(webhookSend)

				if err != nil {
					log.Fatal("err sending the webhook:", err)
				}

				err = addShopifyItem(int(variant.ID), variant.UpdatedAt, variant.Available, variant.Title)
				if err != nil {
					log.Fatal("err adding shopify item:", err)
				}

			}
		}

		var imageURL string
		variantJoined := strings.Join(variants, " ")
		vendor := product.Vendor
		productType := product.ProductType
		if len(product.Images) > 0 {
			imageURL = product.Images[0].Src
		}

		if variantJoined == "" {
			variantJoined = "N/A"
		}

		if vendor == "" {
			vendor = "N/A"
		}

		if productType == "" {
			productType = "N/A"
		}

		if restock {
			webhookSend := buildShopifyWebhook(product.Title, product.Handle, link, imageURL, variantJoined, price, vendor, productType, "RESTOCK")
			err = m.SendDiscordWebhook(webhookSend)

			if err != nil {
				log.Fatal("err sending webhook: ", err)
			}
		}

		if soldOut {
			webhookSend := buildShopifyWebhook(product.Title, product.Handle, link, imageURL, variantJoined, price, vendor, productType, "SOLD OUT")
			err = m.SendDiscordWebhook(webhookSend)

			if err != nil {
				log.Fatal("err sending webhook: ", err)
			}
		}
	}

	time.Sleep(500 * time.Millisecond)

	return
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if strings.Contains(a, str) {
			return true
		}
	}
	return false
}
