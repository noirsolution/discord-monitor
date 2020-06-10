package main

import (
	"bufio"
	"database/sql"
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"github.com/NoirSneaker/monitor"
	_ "github.com/lib/pq"
)

var db *sql.DB
var scanner *bufio.Scanner
var threads = 20
var keywords = []string{
	"bred toe",
	"gold toe",
	"pharrell",
	"holi",
	"free throw line",
	"kendrick",
	"tinker",
	"game royal",
	"yeezy",
	"human race",
	"big bang",
	"dont trip",
	"don't trip",
	"kung fu kenny",
	"playstation",
	"ovo air jordan",
	"ovo jordan",
	"wotherspoon",
	"nike x off-white",
	"off-white x nike",
	"air jordan 1",
	"wave runner",
	"katrina",
	"animal pack",
	"acronym",
	"vf sw",
	"the ten",
	"the 10",
}

func main() {
	boolPtr := flag.Bool("init", false, "init the db")
	supremePtr := flag.Bool("supreme", false, "supreme option")
	flag.Parse()
	shopifySites, err := os.Open("shopify.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer shopifySites.Close()
	scanner = bufio.NewScanner(shopifySites)

	euMonitor := monitor.NewMonitor("webhook_url", []string{
		"login:pass@fr.smartproxy.com:40000",
	})
	supremeUSA := monitor.NewMonitor("webhook_url", []string{
		"login:pass@us.smartproxy.com:10000",
	})
	supremeJP := monitor.NewMonitor("webhook_url", []string{
		"login:pass@jp.smartproxy.com:30000",
	})
	setDb()

	if *boolPtr {
		if *supremePtr {
			initDb(euMonitor, true)
			initDb(supremeUSA, true)
			initDb(supremeJP, true)
		}
		log.Print("SYNCING DB...")
		initDb(euMonitor, false)
	}

	log.Print("MONITORING...")
	for true {
		startMonitor(euMonitor, supremeUSA, supremeJP, *supremePtr)
	}

	defer db.Close()
}

func setDb() {
	var err error
	db, err = sql.Open("postgres", "host=localhost user=youss password=password dbname=monitor sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
}

func startMonitor(euMonitor, supremeUSA, supremeJP *monitor.Monitor, supremeMonitor bool) {
	log.Print("MONITORING...")
	var wg sync.WaitGroup

	concurrentGoroutines := make(chan struct{}, threads)
	if !supremeMonitor {
		for scanner.Scan() {
			wg.Add(1)
			go func(site string) {
				defer wg.Done()
				concurrentGoroutines <- struct{}{}
				err := scanner.Err()
				if err != nil {
					log.Print(err)
				}
				monitorShopify(euMonitor, "https://"+site)
				<-concurrentGoroutines
			}(scanner.Text())
		}
	}

	if supremeMonitor {
		log.Print("supreme")
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := scanner.Err()
			if err != nil {
				log.Print("ok")
			}

			monitorSupreme(supremeUSA)
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := scanner.Err()
			if err != nil {
				log.Print("ERR: ", err)
			}

			monitorSupreme(supremeJP)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := scanner.Err()
			if err != nil {
				log.Print("ok")
			}

			monitorSupreme(euMonitor)
		}()
	}

	wg.Wait()

	time.Sleep(500 * time.Millisecond)

	log.Print("Sleeping...")
}

func initDb(m *monitor.Monitor, supreme bool) {
	sqlStmt := `
		create table IF NOT EXISTS supreme (sizeID int not null, stock integer, size text);
		create table IF NOT EXISTS shopify (sizeID bigint not null, updated text, stock bool, size text);
		`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	if supreme {
		initSupreme(m)
	} else {
		for scanner.Scan() {
			initShopify(m, "https://"+scanner.Text())
		}
	}

}
