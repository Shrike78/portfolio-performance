package secondapensione

import (
	"github.com/shrike78/portfolio-perfomance/pkg/security"
	"github.com/shrike78/portfolio-perfomance/pkg/security/loaders/investingcom"
)

var isinToCurrId = map[string]string{
	"0P0000CWZD": "1078538", //Amundi SecondaPensione Espansione ESG
	"0P0000CWZE": "1078581", //Amundi SecondaPensione Prudente ESG
	"0P0000CX0O": "1078593", //Amundi SecondaPensione Garantita ESG
	"0P0000CWZJ": "1078582", //Amundi SecondaPensione Bilanciata ESG
	"0P0000CWZY": "1078585", //Amundi SecondaPensione Sviluppo ESG
}

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

func (s *SecondaPensione) LoadQuotes() ([]security.Quote, error) {
	var investingComLoader security.QuoteLoader
	investingComLoader = investingcom.New(s.name, s.isin, isinToCurrId[s.isin])
	return investingComLoader.LoadQuotes()
}
