package investingcom

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/shrike78/portfolio-perfomance/pkg/security"
)

type InvestingCom struct {
	name    string
	curr_id string
}

func New(name, curr_id string) *InvestingCom {
	return &InvestingCom{
		name:    name,
		curr_id: curr_id,
	}
}

func (e *InvestingCom) Name() string {
	return e.name
}

func (e *InvestingCom) ISIN() string {
	return e.curr_id
}

func (s *InvestingCom) LoadQuotes() ([]security.Quote, error) {

	endpoint := "https://m.it.investing.com/instrument/services/getHistoricalData"

	// Form payload (use & not &amp;)
	// Dates as they appear in your curl (DD/MM/YYYY)
	form := url.Values{}
	// curr_id is the real identifier of the fund
	form.Set("curr_id", s.curr_id)
	//force a far away date as start
	form.Set("st_date", "01/01/2001")
	//take 'now' as end date
	now := time.Now()
	end_date := now.Format("02/01/2006")
	form.Set("end_date", end_date)
	form.Set("interval_sec", "Daily")

	// Build request
	req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(form.Encode()))
	if err != nil {
		panic(fmt.Errorf("new request: %v", err))
	}

	//refererUrl := fmt.Sprintf("https://m.it.investing.com/funds/allianz-insieme-linea-%s-historical-data", s.subURL)

	// Required headers (mirroring your curl)
	req.Header.Set("Accept", "text/html, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,it;q=0.8")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Origin", "https://m.it.investing.com")
	req.Header.Set("Priority", "u=1, i")
	//req.Header.Set("Referer", refererUrl)
	req.Header.Set("Sec-CH-UA", `"Google Chrome";v="143", "Chromium";v="143", "Not A(Brand";v="24"`)
	req.Header.Set("Sec-CH-UA-Mobile", "?1")
	req.Header.Set("Sec-CH-UA-Platform", `"Android"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	// HTTP client with timeout
	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	// Do request
	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Errorf("request failed: %v", err))
	}
	defer resp.Body.Close()

	// Check status code and capture body on errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		panic(fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes)))
	}

	// Parse HTML with goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("parse html: %v", err)
	}

	quotes := []security.Quote{}

	// Usually it's a <tbody> with <tr>. We'll parse every tr with tds.
	doc.Find("tbody tr").Each(func(_ int, tr *goquery.Selection) {
		tds := tr.Find("td")
		if tds.Length() == 0 {
			return
		}

		// Adjust indices based on actual table structure. Often:
		// 0: Date, 1: Price
		get := func(i int) string {
			return strings.TrimSpace(tds.Eq(i).Text())
		}

		rawDate := get(0)
		priceStr := get(1)

		// Normalize European formatting: "1.234,56" -> "1234.56"
		parseNumber := func(s string) (float64, error) {
			s = strings.TrimSpace(s)
			// remove thousands dots
			s = strings.ReplaceAll(s, ".", "")
			// decimal comma -> dot
			s = strings.ReplaceAll(s, ",", ".")
			// remove percent if present
			s = strings.TrimSuffix(s, "%")
			return strconv.ParseFloat(s, 32)
		}

		//no more needed if retrieval is done per day
		//rawDate = fmt.Sprintf("15.%s", rawDate)

		date, err := time.Parse("02.01.2006", rawDate)
		if err != nil {
			panic(err)
		}

		price, _ := parseNumber(priceStr)

		quotes = append(quotes, security.Quote{
			Date:  date,
			Close: float32(price),
		})
	})

	return quotes, nil
}
