package main

import (
	"log"

	"github.com/NoirSneaker/monitor"
)

func initShopify(m *monitor.Monitor, link string) {
	retries := 0
	shopifyProducts, err := m.GetShopifyProducts(link)
	for err != nil {
		retries++
		log.Print("[init] error getting shopify products: ", err, "... Retrying! LINK: ", link)
		if retries > 3 {
			return
		}
		shopifyProducts, err = m.GetShopifyProducts(link)
	}

	for _, product := range shopifyProducts.Products {
		for _, variant := range product.Variants {
			err = addShopifyItem(int(variant.ID), variant.UpdatedAt, variant.Available, variant.Title)
			if err != nil {
				log.Print("[init] error adding shopify item", err)
			}
		}
	}
}

func initSupreme(m *monitor.Monitor) {
	supremeProducts, err := m.GetSupremeProducts()
	for err != nil {
		log.Print("[init] error getting supreme products", err, "Retrying ...")
		supremeProducts, err = m.GetSupremeProducts()
	}

	for _, product := range supremeProducts.ProductsAndCategories.Jackets {
		customProduct, err := m.GetSupremeProduct(product.ID)
		for err != nil {
			log.Print("[init] error getting supreme product.. ", err, "Retrying...")
			customProduct, err = m.GetSupremeProduct(product.ID)
		}

		for _, style := range customProduct.Styles {
			for _, size := range style.Sizes {
				err = addSupremeItem(size.StockLevel, size.Name, size.ID)
				if err != nil {
					log.Print("[init] error adding supreme item", err)
				}
			}
		}
	}

	for _, product := range supremeProducts.ProductsAndCategories.Bags {
		customProduct, err := m.GetSupremeProduct(product.ID)
		for err != nil {
			log.Print("[init] error getting supreme product", err)
			customProduct, err = m.GetSupremeProduct(product.ID)
		}

		for _, style := range customProduct.Styles {
			for _, size := range style.Sizes {
				err = addSupremeItem(size.StockLevel, size.Name, size.ID)
				if err != nil {
					log.Print("[init] error adding supreme item", err)
				}
			}
		}
	}

	for _, product := range supremeProducts.ProductsAndCategories.Pants {
		customProduct, err := m.GetSupremeProduct(product.ID)
		for err != nil {
			log.Print("[init] error getting supreme product", err)
			customProduct, err = m.GetSupremeProduct(product.ID)
		}

		for _, style := range customProduct.Styles {
			for _, size := range style.Sizes {
				err = addSupremeItem(size.StockLevel, size.Name, size.ID)
				if err != nil {
					log.Print("[init] error adding supreme item", err)
				}
			}
		}
	}

	for _, product := range supremeProducts.ProductsAndCategories.Accessories {
		customProduct, err := m.GetSupremeProduct(product.ID)
		for err != nil {
			log.Print("[init] error getting supreme product", err)
			customProduct, err = m.GetSupremeProduct(product.ID)
		}

		for _, style := range customProduct.Styles {
			for _, size := range style.Sizes {
				err = addSupremeItem(size.StockLevel, size.Name, size.ID)
				if err != nil {
					log.Print("[init] error adding supreme item", err)
				}
			}
		}
	}

	for _, product := range supremeProducts.ProductsAndCategories.Skate {
		customProduct, err := m.GetSupremeProduct(product.ID)
		for err != nil {
			log.Print("[init] error getting supreme product", err)
			customProduct, err = m.GetSupremeProduct(product.ID)
		}

		for _, style := range customProduct.Styles {
			for _, size := range style.Sizes {
				err = addSupremeItem(size.StockLevel, size.Name, size.ID)
				if err != nil {
					log.Print("[init] error adding supreme item", err)
				}
			}
		}
	}

	for _, product := range supremeProducts.ProductsAndCategories.Shoes {
		customProduct, err := m.GetSupremeProduct(product.ID)
		for err != nil {
			log.Print("[init] error getting supreme product", err)
			customProduct, err = m.GetSupremeProduct(product.ID)
		}

		for _, style := range customProduct.Styles {
			for _, size := range style.Sizes {
				err = addSupremeItem(size.StockLevel, size.Name, size.ID)
				if err != nil {
					log.Print("[init] error adding supreme item", err)
				}
			}
		}
	}

	for _, product := range supremeProducts.ProductsAndCategories.Hats {
		customProduct, err := m.GetSupremeProduct(product.ID)
		for err != nil {
			log.Print("[init] error getting supreme product", err)
			customProduct, err = m.GetSupremeProduct(product.ID)
		}

		for _, style := range customProduct.Styles {
			for _, size := range style.Sizes {
				err = addSupremeItem(size.StockLevel, size.Name, size.ID)
				if err != nil {
					log.Print("[init] error adding supreme item", err)
				}
			}
		}
	}

	for _, product := range supremeProducts.ProductsAndCategories.TopsSweaters {
		customProduct, err := m.GetSupremeProduct(product.ID)
		for err != nil {
			log.Print("[init] error getting supreme product", err)
			customProduct, err = m.GetSupremeProduct(product.ID)
		}

		for _, style := range customProduct.Styles {
			for _, size := range style.Sizes {
				err = addSupremeItem(size.StockLevel, size.Name, size.ID)
				if err != nil {
					log.Print("[init] error adding supreme item", err)
				}
			}
		}
	}

	for _, product := range supremeProducts.ProductsAndCategories.Jackets {
		customProduct, err := m.GetSupremeProduct(product.ID)
		for err != nil {
			log.Print("[init] error getting supreme product", err)
			customProduct, err = m.GetSupremeProduct(product.ID)
		}

		for _, style := range customProduct.Styles {
			for _, size := range style.Sizes {
				err = addSupremeItem(size.StockLevel, size.Name, size.ID)
				if err != nil {
					log.Print("[init] error adding supreme item", err)
				}
			}
		}
	}

	for _, product := range supremeProducts.ProductsAndCategories.Sweatshirts {
		customProduct, err := m.GetSupremeProduct(product.ID)
		for err != nil {
			log.Print("[init] error getting supreme product", err)
			customProduct, err = m.GetSupremeProduct(product.ID)
		}

		for _, style := range customProduct.Styles {
			for _, size := range style.Sizes {
				err = addSupremeItem(size.StockLevel, size.Name, size.ID)
				if err != nil {
					log.Print("[init] error adding supreme item", err)
				}
			}
		}
	}

	for _, product := range supremeProducts.ProductsAndCategories.Shirts {
		customProduct, err := m.GetSupremeProduct(product.ID)
		for err != nil {
			log.Print("[init] error getting supreme product", err)
			customProduct, err = m.GetSupremeProduct(product.ID)
		}

		for _, style := range customProduct.Styles {
			for _, size := range style.Sizes {
				err = addSupremeItem(size.StockLevel, size.Name, size.ID)
				if err != nil {
					log.Print("[init] error adding supreme item", err)
				}
			}
		}
	}

	for _, product := range supremeProducts.ProductsAndCategories.TShirts {
		customProduct, err := m.GetSupremeProduct(product.ID)
		for err != nil {
			log.Print("[init] error getting supreme product", err)
			customProduct, err = m.GetSupremeProduct(product.ID)
		}

		for _, style := range customProduct.Styles {
			for _, size := range style.Sizes {
				err = addSupremeItem(size.StockLevel, size.Name, size.ID)
				if err != nil {
					log.Print("[init] error adding supreme item", err)
				}
			}
		}
	}

	for _, product := range supremeProducts.ProductsAndCategories.Shorts {
		customProduct, err := m.GetSupremeProduct(product.ID)
		for err != nil {
			log.Print("[init] error getting supreme product", err)
			customProduct, err = m.GetSupremeProduct(product.ID)
		}

		for _, style := range customProduct.Styles {
			for _, size := range style.Sizes {
				err = addSupremeItem(size.StockLevel, size.Name, size.ID)
				if err != nil {
					log.Print("[init] error adding supreme item", err)
				}
			}
		}
	}

	for _, product := range supremeProducts.ProductsAndCategories.New {
		customProduct, err := m.GetSupremeProduct(product.ID)
		for err != nil {
			log.Print("[init] error getting supreme product", err)
			customProduct, err = m.GetSupremeProduct(product.ID)
		}

		for _, style := range customProduct.Styles {
			for _, size := range style.Sizes {
				err = addSupremeItem(size.StockLevel, size.Name, size.ID)
				if err != nil {
					log.Print("[init] error adding supreme item", err)
				}
			}
		}
	}
}
