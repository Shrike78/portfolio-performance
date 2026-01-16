package cometa

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/shrike78/portfolio-perfomance/pkg/security"
)

/*
Investing.com - ISIN: Curr_ID
Cometa Crescita 		- 0P0000ADYJ: 1055574
Cometa Monetario Plus 	- 0P0000ADRN: 1055575
Cometa Reddito			- 0P0000ADYI: 1055576
*/

var cometaSubURLMap = map[string]string{
	"0P0000ADYJ": "crescita",
	"0P0000ADRN": "monetario-plus",
	"0P0000ADYI": "comparto-sicurezza",
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

			dateString = fmt.Sprintf("15/%s", dateString)
			date, err := time.Parse("02/01/2006", dateString)
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
