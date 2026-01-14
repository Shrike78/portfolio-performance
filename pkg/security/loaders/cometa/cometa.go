package cometa

import (
	"github.com/shrike78/portfolio-perfomance/pkg/security"
	"github.com/shrike78/portfolio-perfomance/pkg/security/loaders/investingcom"
)

var isinToCurrId = map[string]string{
	"0P0000ADYJ": "1055574", //Cometa Crescita
	"0P0000ADRN": "1055575", //Cometa Monetario Plus
	"0P0000ADYI": "1055576", //Cometa Reddito
}

type Cometa struct {
	name string
	isin string
}

func New(name, isin string) *Cometa {
	return &Cometa{
		name: name,
		isin: isin,
	}
}

func (e *Cometa) Name() string {
	return e.name
}

func (e *Cometa) ISIN() string {
	return e.isin
}

func (s *Cometa) LoadQuotes() ([]security.Quote, error) {
	var investingComLoader security.QuoteLoader
	investingComLoader = investingcom.New(s.name, s.isin, isinToCurrId[s.isin])
	return investingComLoader.LoadQuotes()
}
