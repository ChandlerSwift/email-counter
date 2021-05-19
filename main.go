package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"database/sql"

	"github.com/emersion/go-imap/client"
	_ "github.com/mattn/go-sqlite3"
	"github.com/wcharczuk/go-chart/v2"
)

func getMessageCount(user string, password string, mailbox string) uint32 {
	log.Println("Connecting...")

	c, err := client.DialTLS("imap.migadu.com:993", nil)
	if err != nil {
		panic(err)
	}

	defer c.Logout()

	if err := c.Login(user, password); err != nil {
		panic(err)
	}
	inbox, err := c.Select(mailbox, true)
	if err != nil {
		panic(err)
	}

	return inbox.Messages
}

func main() {
	db, err := sql.Open("sqlite3", "./data.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Exec(`CREATE TABLE IF NOT EXISTS history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		count INTEGER,
		time DATETIME
	);`)

	messageCount := getMessageCount(os.Getenv("IMAP_USER"), os.Getenv("IMAP_PASS"), "INBOX")
	fmt.Println("Currently ", messageCount, " messages in your inbox.") // TODO: ", a change of +/-$n from yesterday"
	_, err = db.Exec(fmt.Sprintf("INSERT INTO history (count, time) VALUES (%v, '%v');", messageCount, time.Now().Format(time.RFC3339)))
	if err != nil {
		log.Fatal(err)
	}
	generateGraph(db)
}

func generateGraph(db *sql.DB) {
	rows, err := db.Query("SELECT count, time FROM history")
	if err != nil {
		log.Fatal(err)
	}
	xvalues := []time.Time{}
	yvalues := []float64{}
	defer rows.Close()
	rowsProcessed := 0
	most_tabs_open := 0
	peak_tabs_annotations := []chart.Value2{}
	for rows.Next() {
		var count int
		var t time.Time
		err = rows.Scan(&count, &t)
		if err != nil {
			log.Fatal(err)
		}
		xvalues = append(xvalues, t)
		yvalues = append(yvalues, float64(count))

		if count > most_tabs_open {
			peak_tabs_annotations = []chart.Value2{
				{XValue: chart.TimeToFloat64(t), YValue: float64(count), Label: fmt.Sprintf("%v tabs on %v", count, t.Format("Jan 2"))},
			}
			most_tabs_open = count
		} else if count == most_tabs_open {
			peak_tabs_annotations = append(peak_tabs_annotations, chart.Value2{
				XValue: chart.TimeToFloat64(t),
				YValue: float64(count),
				Label:  fmt.Sprintf("%v tabs on %v", count, t.Format("Jan 2")),
			})
		}

		rowsProcessed++
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Processed %v history entries", rowsProcessed)

	graph := chart.Chart{
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: xvalues,
				YValues: yvalues,
			},
			chart.AnnotationSeries{
				Annotations: peak_tabs_annotations,
			},
		},
	}
	f, _ := os.Create("out.png")
	defer f.Close()
	graph.Render(chart.PNG, f)
}
