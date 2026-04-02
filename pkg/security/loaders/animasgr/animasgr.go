package animasgr

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/shrike78/portfolio-perfomance/pkg/security"
)

const (
	pageSize        = 100
	maxPages        = 200
	quotePageURL    = "https://www.animasgr.it/IT/Prodotti/Scheda/?codFondo=%s"
	detailNavURL    = "https://www.animasgr.it/it/Funds/_DetailNav"
	requestTimeout  = 20 * time.Second
	quoteDateLayout = "02.01.06"
)

type AnimaSGR struct {
	name     string
	isin     string
	fundCode string
}

func New(name, isin string) *AnimaSGR {
	parts := strings.SplitN(isin, ".", 2)
	resolvedISIN := strings.TrimSpace(parts[0])
	resolvedFundCode := resolvedISIN
	if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
		resolvedFundCode = strings.TrimSpace(parts[1])
	}

	return &AnimaSGR{
		name:     name,
		isin:     resolvedISIN,
		fundCode: resolvedFundCode,
	}
}

func (e *AnimaSGR) Name() string {
	return e.name
}

func (e *AnimaSGR) ISIN() string {
	return e.isin
}

func (s *AnimaSGR) LoadQuotes() ([]security.Quote, error) {
	startDate, endDate, err := s.fetchAvailableDateRange()
	if err != nil {
		return nil, err
	}

	quotes := make([]security.Quote, 0, pageSize)
	seen := make(map[time.Time]struct{})

	for page := 0; page < maxPages; page++ {
		pageQuotes, err := s.fetchQuotePage(startDate, endDate, page)
		if err != nil {
			return nil, err
		}
		if len(pageQuotes) == 0 {
			break
		}

		for _, quote := range pageQuotes {
			date := quote.Date.UTC()
			if _, ok := seen[date]; ok {
				continue
			}
			seen[date] = struct{}{}
			quotes = append(quotes, quote)
		}

		if len(pageQuotes) < pageSize {
			break
		}
	}

	if len(quotes) == 0 {
		return nil, fmt.Errorf("no ANIMA quotes found for %s", s.isin)
	}

	return quotes, nil
}

func (s *AnimaSGR) fetchAvailableDateRange() (string, string, error) {
	body, err := fetchPage(fmt.Sprintf(quotePageURL, s.fundCode))
	if err != nil {
		return "", "", fmt.Errorf("request ANIMA quote page for %s: %w", s.isin, err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return "", "", fmt.Errorf("parse ANIMA quote page for %s: %w", s.isin, err)
	}

	startDate, ok := doc.Find("#hdDataAvvioQuote").Attr("value")
	if !ok || strings.TrimSpace(startDate) == "" {
		return "", "", fmt.Errorf("ANIMA start date not found for %s", s.isin)
	}

	endDate, ok := doc.Find("#hdDataFineQuote").Attr("value")
	if !ok || strings.TrimSpace(endDate) == "" {
		return "", "", fmt.Errorf("ANIMA end date not found for %s", s.isin)
	}

	return startDate, endDate, nil
}

func (s *AnimaSGR) fetchQuotePage(startDate, endDate string, page int) ([]security.Quote, error) {
	params := url.Values{}
	params.Set("idLanguage", "it")
	params.Set("id", s.fundCode)
	params.Set("dataInizio", startDate)
	params.Set("dataFine", endDate)
	params.Set("numeroPagina", strconv.Itoa(page))
	params.Set("lunghezzaPagina", strconv.Itoa(pageSize))
	params.Set("order", "false")

	body, err := fetchPage(detailNavURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("request ANIMA quote history page %d for %s: %w", page, s.isin, err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse ANIMA quote history page %d for %s: %w", page, s.isin, err)
	}

	quotes := make([]security.Quote, 0, pageSize)
	doc.Find("#tbQuotazioni tr").Each(func(i int, tr *goquery.Selection) {
		if i == 0 {
			return
		}

		cells := tr.Find("td")
		if cells.Length() < 4 {
			return
		}

		rawDate := strings.TrimSpace(cells.Eq(2).Text())
		rawValue := strings.TrimSpace(cells.Eq(3).Text())

		date, err := time.Parse(quoteDateLayout, rawDate)
		if err != nil {
			return
		}

		value, err := parseLocalizedFloat(rawValue)
		if err != nil {
			return
		}

		quotes = append(quotes, security.Quote{
			Date:  date,
			Close: float32(value),
		})
	})

	return quotes, nil
}

func fetchPage(pageURL string) ([]byte, error) {
	client := &http.Client{Timeout: requestTimeout}

	req, err := http.NewRequest(http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "it-IT,it;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, pageURL)
	}

	return body, nil
}

func parseLocalizedFloat(value string) (float64, error) {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, ".", "")
	value = strings.ReplaceAll(value, ",", ".")
	return strconv.ParseFloat(value, 64)
}
