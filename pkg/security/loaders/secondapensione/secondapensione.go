package secondapensione

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/shrike78/portfolio-perfomance/pkg/security"
)

/*
const (
	SecondaPensioneUrlTemplate = "https://www.secondapensione.it/ezjscore/call/ezjscamundibuzz::sfForwardFront::paramsList=service=ProxyProductSheetV3Front&routeId=_en-GB_879_%s_tab_3"
)
*/

type SecondaPensione struct {
	name string
	isin string
}

func New(name, isin string) *SecondaPensione {
	return &SecondaPensione{
		name: name,
		isin: isin,
	}
}

func (e *SecondaPensione) Name() string {
	return e.name
}

func (e *SecondaPensione) ISIN() string {
	return e.isin
}

/*
func (s *SecondaPensione) LoadQuotes() ([]security.Quote, error) {
	c := colly.NewCollector()

	url := fmt.Sprintf(SecondaPensioneUrlTemplate, s.isin)

	quotes := []security.Quote{}

	c.OnHTML("#tableVl", func(e *colly.HTMLElement) {
		e.ForEach("tbody tr", func(i int, e *colly.HTMLElement) {
			dateString, valueString := parseRowText(e.ChildTexts("td"))

			if dateString == "" || valueString == "" {
				return
			}

			date, err := time.Parse("02/01/2006", dateString)
			if err != nil {
				panic(err)
			}

			closeQuote, err := strconv.ParseFloat(valueString, 32)
			if err != nil {
				panic(err)
			}

			quotes = append(quotes, security.Quote{
				Date:  date,
				Close: float32(closeQuote),
			})
		})
	})

	c.Visit(url)

	return quotes, nil
}

func parseRowText(values []string) (string, string) {
	return values[0], values[1]
}

*/

func (s *SecondaPensione) LoadQuotes() ([]security.Quote, error) {

	// If you know the exact response schema, define a typed struct here.
	// Below is a generic envelope + RawMessage for flexible parsing.
	type ResponseEnvelope struct {
		// Adjust fields to match actual response. Using RawMessage allows delayed parsing.
		Data json.RawMessage `json:"data"`
		// Many APIs include status/metadata; keep generic to be resilient.
		Status     string          `json:"status,omitempty"`
		Error      string          `json:"error,omitempty"`
		Additional json.RawMessage `json:"additional,omitempty"`
	}

	// Endpoint from your curl
	url := fmt.Sprintf("https://www.secondapensione.it/product-services/fdr/share/v3/full/%s", s.isin)
	referUrl := fmt.Sprintf("https://www.secondapensione.it/product/view/%s", s.isin)

	// JSON payload from your curl --data-raw
	requestBody := `{"fields":["translatedLabel.key","label","comp.assetClass1.key","comp.assetClass2.id","performances","class.sri","lastNav.value","lastNav.date","currency.key","currency.iso3Code","navDecimals","lastNav.variationPercent","lastNav.subFundAssets","comp.currency.key","comp.currency.iso3Code","comp.sfdrClassification.key","comp.sfdrClassification.id","morningStarNotation","morningStarDate","isin","shareCode","comp.distributionCommentary","navHistory","couponHistory","guaranteedNav","cpnDecimals","class.indexChangeHistory","base100","partners","calendarPerformances","riskMarkers","perfScenario.byDate","comp.minimumRecommendedHoldingPeriod.key","comp.distribution","launchDate","fund.juridicForm.key","fund.accountingCompanies","resultAssignment.key","registrationCountries","class.referenceBench","fund.custodians","comp.mainManager.label","firstNavDate","bloombergCode","reutersCode","class.valorisationFrequency.key","class.valueUnitSubscription1","class.unitSubscription1.key","class.valueUnitSubscriptionA","class.unitSubscriptionA.key","costs.performanceFeesBenchmark","costs.entry","costs.eap.ogc","costs.eap.perfFees","costs.eap.trx","costs.eap.entry","costs.eap.exit","costs.ea.italyFixedExit","costs.eap.exitNotAcquired","costs.ea.exitMaxAcquired","costs.firstEndExitDate","costs.firstEndExitPct","costs.eap.entryNotAcquired","fund.applicableRight.iso3Code","costs.ea.italyFixedEntry","costs.lastStartExitDate","costs.exit1Year","costs.lastStartExitPct","costs.maxPerformanceFees","costs.ea.ogc","comp.minimumRecommendedHoldingPeriodNumeric","costs.ea.trx","costs.priipsKidPerfNarrative","comp.code"]}`

	// Prepare client and context with sensible timeouts
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Build request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(requestBody))
	if err != nil {
		panic(fmt.Errorf("creating request: %w", err))
	}

	// ===== Set headers from your curl =====
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,it;q=0.8")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Origin", "https://www.secondapensione.it")
	req.Header.Set("Referer", referUrl)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36")
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="143", "Chromium";v="143", "Not A(Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?1")
	req.Header.Set("sec-ch-ua-platform", `"Android"`)

	// Cookies from your curl -b '<cookie string>'
	// You can set the Cookie header directly (works fine for quick replication).
	//req.Header.Set("Cookie", `_pcid=%7B%22browserId%22%3A%22mgbwbkrzgyqgj7mi%22%2C%22_t%22%3A%22mw0de85e%7Cmgbwbl1e%22%7D; _pctx=%7Bu%7DN4IgrgzgpgThIC4B2YA2qA05owMoBcBDfSREQpAeyRCwgEt8oBJAE0RXSwH18yBbAO4AGVlAAcAVigAffgHMARoMWoAjFBABfIA; cookie-agreed-version=1.0.0; amundi_allow_tracking=true; cookie-agreed-categories=["necessary"]; pa_privacy=%22optout%22; cookie-agreed=2; banner_context=amundi-ita-it-retail-secondapensione`)

	// ===== Send request =====
	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Errorf("performing request: %w", err))
	}
	defer resp.Body.Close()

	// Check status code and capture body on errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		panic(fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes)))
	}

	// ===== Parse JSON response =====

	// We need to read the body once; to demonstrate both parses, buffer it:
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Errorf("reading body: %w", err))
	}

	/*
		// Optional: pretty-print the whole JSON response
		var pretty any
		if err := json.Unmarshal(buf, &pretty); err == nil {
			out, _ := json.MarshalIndent(pretty, "", "  ")
			fmt.Printf("\nFull JSON response (pretty):\n%s\n", string(out))
		}
	*/

	type NavPoint struct {
		Date  string  `json:"date"`
		Value float32 `json:"value"`
	}

	type RootItem struct {
		NavHistory []NavPoint `json:"navHistory"`
	}

	var items []RootItem
	if err := json.Unmarshal(buf, &items); err != nil {
		panic(fmt.Errorf("Error parsing JSON: %v", err))
	}

	quotes := []security.Quote{}

	// Access fields
	for _, item := range items {
		fmt.Println("NAV History:")
		for _, nav := range item.NavHistory {
			//fmt.Printf("  %s -> %.3f\n", nav.Date, nav.Value)
			date, err := time.Parse("2006-01-02", nav.Date)
			if err != nil {
				panic(err)
			}
			quotes = append(quotes, security.Quote{
				Date:  date,
				Close: nav.Value,
			})

		}

	}

	return quotes, nil
}
