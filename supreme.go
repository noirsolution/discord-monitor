package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/NoirSneaker/monitor"
)

// Sizes is from the monitor package
type Sizes struct {
	Name       string `json:"name"`
	ID         int    `json:"id"`
	StockLevel int    `json:"stock_level"`
}

// Product is from the monitor package
type Product struct {
	Name          string "json:\"name\""
	ID            int    "json:\"id\""
	ImageURL      string "json:\"image_url\""
	ImageURLHi    string "json:\"image_url_hi\""
	Price         int    "json:\"price\""
	SalePrice     int    "json:\"sale_price\""
	NewItem       bool   "json:\"new_item\""
	Position      int    "json:\"position\""
	CategoryName  string "json:\"category_name\""
	PriceEuro     int    "json:\"price_euro\""
	SalePriceEuro int    "json:\"sale_price_euro\""
}

func addSupremeItem(stock int, sizeName string, sizeID int) error {
	_, err := db.Exec("insert into supreme(sizeID, stock, size) values($1, $2, $3)", sizeID, stock, sizeName)
	if err != nil {
		return err
	}

	return nil
}

// containsPrecise check if the whole string is in array
func containsPrecise(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func buildSupremeWebhook(product Product, sizes, styles, currency string, customProduct *monitor.SupremeProduct, typeEvent string) monitor.Webhook {
	realPrice := fmt.Sprintf("$ %v", strconv.Itoa(product.Price/100))
	htmlEscape := regexp.MustCompile("</?[^>]+(>|$)")
	escapedName := htmlEscape.ReplaceAllString(product.Name, "")

	var EuBlocked string
	var RuBlocked string

	if customProduct.NonEuBlocked == nil || *customProduct.NonEuBlocked == true {
		EuBlocked = "TRUE"
	} else if *customProduct.NonEuBlocked == false {
		EuBlocked = "FALSE"
	}

	if customProduct.RussiaBlocked == nil || *customProduct.RussiaBlocked == true {
		RuBlocked = "TRUE"
	} else if *customProduct.RussiaBlocked == false {
		RuBlocked = "FALSE"
	}

	return monitor.Webhook{
		Username:  "NoirMonitor",
		AvatarURL: "",
		Embeds: []monitor.Embeds{
			{
				Title:     fmt.Sprintf("%s: %s", typeEvent, escapedName),
				URL:       fmt.Sprintf("https://supremenewyork.com/shop/%v", product.ID),
				Color:     0,
				Thumbnail: monitor.Thumbnail{URL: "https:" + product.ImageURL},
				Fields: []monitor.Fields{
					{
						Name:   "Price",
						Value:  realPrice,
						Inline: true,
					},
					{
						Name:   "Size(s)",
						Value:  sizes,
						Inline: true,
					},
					{
						Name:   "Purchsable Quantity",
						Value:  strconv.Itoa(customProduct.PurchasableQty),
						Inline: true,
					},
					{
						Name:   "New Item",
						Value:  strings.ToUpper(strconv.FormatBool(customProduct.NewItem)),
						Inline: true,
					},
					{
						Name:   "Styles",
						Value:  styles,
						Inline: true,
					},
					{
						Name:   "Currency",
						Value:  currency,
						Inline: true,
					},
					{
						Name:   "Blocked from EU",
						Value:  EuBlocked,
						Inline: true,
					},
					{
						Name:   "Blocked from RU",
						Value:  RuBlocked,
						Inline: true,
					},
					{
						Name:   "Blocked from CA",
						Value:  strings.ToUpper(strconv.FormatBool(customProduct.CanadaBlocked)),
						Inline: true,
					},
				},
				Timestamp: time.Now().Format("2006-01-02T15:04:05-0700"),
				Footer:    monitor.Footer{Text: "SUPREME Monitor by NoirMonitor"},
			},
		},
	}
}

func fetchSupremeItem(size Sizes) (bool, bool, bool, []string, error) {
	var rowsPresent bool
	var restock bool
	var soldOut bool
	var sizes []string

	rows, err := db.Query("select sizeID, stock, size from supreme WHERE sizeID = $1", size.ID)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		return false, false, false, nil, err
	}
	for rows.Next() {
		rowsPresent = true
		var id int
		var stock int
		var sizeName string
		err = rows.Scan(&id, &stock, &sizeName)
		if err != nil {
			return false, false, false, nil, err
		}
		if size.StockLevel > stock && size.Name == sizeName {
			if !containsPrecise(sizes, size.Name) {
				sizes = append(sizes, fmt.Sprintf(size.Name))
			}
			restock = true
			_, err = db.Exec("UPDATE supreme SET stock = $1 WHERE sizeID = $2", size.StockLevel, size.ID)
			if err != nil {
				log.Fatal(err)
			}
		} else if size.StockLevel == 0 && size.StockLevel != stock && size.Name == sizeName {
			soldOut = true
			_, err = db.Exec("UPDATE supreme SET stock = $1 WHERE sizeID = $2", size.StockLevel, size.ID)
			if err != nil {
				log.Fatal(err)
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

	return rowsPresent, restock, soldOut, sizes, nil
}

func handleProducts(m *monitor.Monitor, product Product) {
	retries := 0
	customProduct, err := m.GetSupremeProduct(product.ID)
	for err != nil {
		retries++
		if retries > 5 {
			return
		}
	}

	var rowsPresent bool
	var soldOut bool
	var restock bool
	var currency string

	var sizes []string
	var styles []string

	for _, style := range customProduct.Styles {
		styles = append(styles, style.Name)
		for _, size := range style.Sizes {
			retries = 0
			rowsPresent, restock, soldOut, sizes, err = fetchSupremeItem(size)
			for err != nil {
				retries++
				if retries > 5 {
					return
				}
			}

			if !rowsPresent {
				log.Printf("Not in Database: %v. Adding it.", product.ID)
				webhookSend := buildSupremeWebhook(product, size.Name, strings.Join(styles, " "), currency, customProduct, "NEW PRODUCT")
				err = m.SendDiscordWebhook(webhookSend)

				if err != nil {
					log.Print("err sending webhook: ", err)
				}

				err = addSupremeItem(size.StockLevel, size.Name, size.ID)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
		currency = style.Currency
	}

	if soldOut {
		webhookSend := buildSupremeWebhook(product, strings.Join(sizes, " "), strings.Join(styles, " "), currency, customProduct, "SOLD OUT")
		err = m.SendDiscordWebhook(webhookSend)

		if err != nil {
			log.Print("err sending webhook: ", err)
		}
	}

	if restock {
		webhookSend := buildSupremeWebhook(product, strings.Join(sizes, " "), strings.Join(styles, " "), currency, customProduct, "RESTOCK")
		err = m.SendDiscordWebhook(webhookSend)

		if err != nil {
			log.Print("err sending webhook: ", err)
		}
	}
}

func monitorSupreme(m *monitor.Monitor) {
	concurrentGoroutines := make(chan struct{}, threads)
	retries := 0
	supremeProducts, err := m.GetSupremeProducts()
	for err != nil {
		retries++
		if retries > 5 {
			return
		}
	}

	var wg sync.WaitGroup
	for _, product := range supremeProducts.ProductsAndCategories.Jackets {
		wg.Add(1)
		go func(m *monitor.Monitor, product Product) {
			defer wg.Done()
			concurrentGoroutines <- struct{}{}
			err := scanner.Err()
			if err != nil {
				log.Print("ok")
			}

			handleProducts(m, product)
			<-concurrentGoroutines
		}(m, product)
	}

	wg.Wait()

	for _, product := range supremeProducts.ProductsAndCategories.Bags {
		wg.Add(1)
		go func(m *monitor.Monitor, product Product) {
			defer wg.Done()
			concurrentGoroutines <- struct{}{}
			err := scanner.Err()
			if err != nil {
				log.Print("ok")
			}

			handleProducts(m, product)
			<-concurrentGoroutines
		}(m, product)
	}

	wg.Wait()

	for _, product := range supremeProducts.ProductsAndCategories.Pants {

		wg.Add(1)
		go func(m *monitor.Monitor, product Product) {
			defer wg.Done()
			concurrentGoroutines <- struct{}{}
			err := scanner.Err()
			if err != nil {
				log.Print("ok")
			}

			handleProducts(m, product)
			<-concurrentGoroutines
		}(m, product)
	}

	wg.Wait()

	for _, product := range supremeProducts.ProductsAndCategories.Accessories {
		wg.Add(1)
		go func(m *monitor.Monitor, product Product) {
			defer wg.Done()
			concurrentGoroutines <- struct{}{}
			err := scanner.Err()
			if err != nil {
				log.Print("ok")
			}

			handleProducts(m, product)
			<-concurrentGoroutines
		}(m, product)
	}

	wg.Wait()

	for _, product := range supremeProducts.ProductsAndCategories.Skate {
		wg.Add(1)
		go func(m *monitor.Monitor, product Product) {
			defer wg.Done()
			concurrentGoroutines <- struct{}{}
			err := scanner.Err()
			if err != nil {
				log.Print("ok")
			}

			handleProducts(m, product)
			<-concurrentGoroutines
		}(m, product)
	}

	wg.Wait()

	for _, product := range supremeProducts.ProductsAndCategories.Shoes {
		wg.Add(1)
		go func(m *monitor.Monitor, product Product) {
			defer wg.Done()
			concurrentGoroutines <- struct{}{}
			err := scanner.Err()
			if err != nil {
				log.Print("ok")
			}

			handleProducts(m, product)
			<-concurrentGoroutines
		}(m, product)
	}

	wg.Wait()

	for _, product := range supremeProducts.ProductsAndCategories.Hats {
		wg.Add(1)
		go func(m *monitor.Monitor, product Product) {
			defer wg.Done()
			concurrentGoroutines <- struct{}{}
			err := scanner.Err()
			if err != nil {
				log.Print("ok")
			}

			handleProducts(m, product)
			<-concurrentGoroutines
		}(m, product)
	}

	wg.Wait()

	for _, product := range supremeProducts.ProductsAndCategories.TopsSweaters {
		wg.Add(1)
		go func(m *monitor.Monitor, product Product) {
			defer wg.Done()
			concurrentGoroutines <- struct{}{}
			err := scanner.Err()
			if err != nil {
				log.Print("ok")
			}

			handleProducts(m, product)
			<-concurrentGoroutines
		}(m, product)
	}

	wg.Wait()

	for _, product := range supremeProducts.ProductsAndCategories.Jackets {
		wg.Add(1)
		go func(m *monitor.Monitor, product Product) {
			defer wg.Done()
			concurrentGoroutines <- struct{}{}
			err := scanner.Err()
			if err != nil {
				log.Print("ok")
			}

			handleProducts(m, product)
			<-concurrentGoroutines
		}(m, product)
	}

	wg.Wait()

	for _, product := range supremeProducts.ProductsAndCategories.Sweatshirts {
		wg.Add(1)
		go func(m *monitor.Monitor, product Product) {
			defer wg.Done()
			concurrentGoroutines <- struct{}{}
			err := scanner.Err()
			if err != nil {
				log.Print("ok")
			}

			handleProducts(m, product)
			<-concurrentGoroutines
		}(m, product)
	}

	wg.Wait()

	for _, product := range supremeProducts.ProductsAndCategories.Shirts {
		wg.Add(1)
		go func(m *monitor.Monitor, product Product) {
			defer wg.Done()
			concurrentGoroutines <- struct{}{}
			err := scanner.Err()
			if err != nil {
				log.Print("ok")
			}

			handleProducts(m, product)
			<-concurrentGoroutines
		}(m, product)
	}

	wg.Wait()

	for _, product := range supremeProducts.ProductsAndCategories.TShirts {
		wg.Add(1)
		go func(m *monitor.Monitor, product Product) {
			defer wg.Done()
			concurrentGoroutines <- struct{}{}
			err := scanner.Err()
			if err != nil {
				log.Print("ok")
			}

			handleProducts(m, product)
			<-concurrentGoroutines
		}(m, product)
	}

	wg.Wait()

	for _, product := range supremeProducts.ProductsAndCategories.Shorts {
		wg.Add(1)
		go func(m *monitor.Monitor, product Product) {
			defer wg.Done()
			concurrentGoroutines <- struct{}{}
			err := scanner.Err()
			if err != nil {
				log.Print("ok")
			}

			handleProducts(m, product)
			<-concurrentGoroutines
		}(m, product)
	}

	wg.Wait()

	for _, product := range supremeProducts.ProductsAndCategories.New {
		wg.Add(1)
		go func(m *monitor.Monitor, product Product) {
			defer wg.Done()
			concurrentGoroutines <- struct{}{}
			err := scanner.Err()
			if err != nil {
				log.Print("ok")
			}

			handleProducts(m, product)
			<-concurrentGoroutines
		}(m, product)
	}

	wg.Wait()

	return
}
