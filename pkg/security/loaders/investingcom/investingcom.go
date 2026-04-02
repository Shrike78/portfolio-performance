package investingcom

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
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
	isin    string
	curr_id string
}

const (
	investingComEndpoint = "https://m.it.investing.com/instrument/services/getHistoricalData"
	maxRetries           = 4
	requestTimeout       = 20 * time.Second
)

func New(name, isin string) *InvestingCom {

	isin_curr_id := strings.Split(isin, ".")

	return &InvestingCom{
		name:    name,
		isin:    isin_curr_id[0],
		curr_id: isin_curr_id[1],
	}
}

func (e *InvestingCom) Name() string {
	return e.name
}

func (e *InvestingCom) ISIN() string {
	return e.isin
}

func (e *InvestingCom) Curr_Id() string {
	return e.curr_id
}

func (s *InvestingCom) LoadQuotes() ([]security.Quote, error) {
	client := &http.Client{
		Timeout: requestTimeout,
	}

	var (
		body []byte
		err  error
	)
	for attempt := 1; attempt <= maxRetries; attempt++ {
		body, err = s.fetchHistoricalData(client)
		if err == nil {
			break
		}

		if !isRetryable(err) || attempt == maxRetries {
			return nil, err
		}

		time.Sleep(time.Duration(attempt) * 2 * time.Second)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parse investing.com response for %s: %w", s.isin, err)
	}

	quotes := []security.Quote{}
	var parseErr error

	// Usually it's a <tbody> with <tr>. We'll parse every tr with tds.
	doc.Find("tbody tr").Each(func(_ int, tr *goquery.Selection) {
		if parseErr != nil {
			return
		}

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

		date, err := time.Parse("02.01.2006", rawDate)
		if err != nil {
			parseErr = fmt.Errorf("parse quote date %q for %s: %w", rawDate, s.isin, err)
			return
		}

		price, err := parseNumber(priceStr)
		if err != nil {
			parseErr = fmt.Errorf("parse quote price %q for %s: %w", priceStr, s.isin, err)
			return
		}

		quotes = append(quotes, security.Quote{
			Date:  date,
			Close: float32(price),
		})
	})

	if parseErr != nil {
		return nil, parseErr
	}
	if len(quotes) == 0 {
		return nil, fmt.Errorf("no quotes found in investing.com response for %s", s.isin)
	}

	return quotes, nil
}

func (s *InvestingCom) fetchHistoricalData(client *http.Client) ([]byte, error) {
	form := url.Values{}
	form.Set("curr_id", s.curr_id)
	form.Set("st_date", "01/01/2001")
	form.Set("end_date", time.Now().Format("02/01/2006"))
	form.Set("interval_sec", "Daily")

	req, err := http.NewRequest("POST", investingComEndpoint, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build investing.com request for %s: %w", s.isin, err)
	}

	req.Header.Set("Accept", "text/html, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,it;q=0.8")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Origin", "https://m.it.investing.com")
	req.Header.Set("Priority", "u=1, i")
	req.Header.Set("Referer", "https://m.it.investing.com/funds/")
	req.Header.Set("Sec-CH-UA", `"Google Chrome";v="143", "Chromium";v="143", "Not A(Brand";v="24"`)
	req.Header.Set("Sec-CH-UA-Mobile", "?1")
	req.Header.Set("Sec-CH-UA-Platform", `"Android"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to investing.com failed for %s: %w", s.isin, err)
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("read investing.com response for %s: %w", s.isin, readErr)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusForbidden && looksLikeCloudflareChallenge(body) {
			return nil, retryableError{
				err: fmt.Errorf("investing.com blocked the request with a Cloudflare challenge for %s", s.isin),
			}
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			return nil, retryableError{
				err: fmt.Errorf("investing.com returned HTTP %d for %s", resp.StatusCode, s.isin),
			}
		}

		return nil, fmt.Errorf("investing.com returned HTTP %d for %s: %s", resp.StatusCode, s.isin, truncateBody(body))
	}

	if looksLikeCloudflareChallenge(body) {
		return nil, retryableError{
			err: fmt.Errorf("investing.com returned a Cloudflare challenge page for %s", s.isin),
		}
	}

	return body, nil
}

type retryableError struct {
	err error
}

func (e retryableError) Error() string {
	return e.err.Error()
}

func (e retryableError) Unwrap() error {
	return e.err
}

func isRetryable(err error) bool {
	var retryErr retryableError
	if errors.As(err, &retryErr) {
		return true
	}

	var netErr net.Error
	return errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary())
}

func looksLikeCloudflareChallenge(body []byte) bool {
	content := strings.ToLower(string(body))
	return strings.Contains(content, "just a moment") ||
		strings.Contains(content, "enable javascript and cookies to continue") ||
		strings.Contains(content, "cf_chl_opt")
}

func truncateBody(body []byte) string {
	const maxLen = 300
	text := strings.TrimSpace(string(body))
	if len(text) <= maxLen {
		return text
	}

	return text[:maxLen] + "..."
}
