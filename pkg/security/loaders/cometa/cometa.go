package cometa

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/shrike78/portfolio-perfomance/pkg/security"
)

var cometaSubURLMap = map[string]string{
	"FP-Cometa-Crescita":      "crescita",
	"FP-Cometa-MonetarioPlus": "monetario-plus",
	"FP-Cometa-Sicurezza":     "comparto-sicurezza",
}

type Cometa struct {
	name   string
	isin   string
	subURL string
}

func New(name, isin string) *Cometa {
	return &Cometa{
		name:   name,
		isin:   isin,
		subURL: cometaSubURLMap[isin],
	}
}

func (e *Cometa) Name() string {
	return e.name
}

func (e *Cometa) ISIN() string {
	return e.isin
}

func (s *Cometa) LoadQuotes() ([]security.Quote, error) {
	c := colly.NewCollector()

	url := fmt.Sprintf("https://www.cometafondo.it/andamenti/%s/", s.subURL)

	quotes := []security.Quote{}

	c.OnHTML("#table_2 tbody", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(i int, e *colly.HTMLElement) {
			dateString, valueString := parseRowText(e.ChildTexts("td"))

			if dateString == "" || valueString == "" {
				return
			}

			date, err := time.Parse("02/2006", dateString)
			if err != nil {
				panic(err)
			}

			valueString = strings.Replace(valueString, ",", ".", -1)
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
