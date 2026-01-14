package allianzinsieme

import (
	"github.com/shrike78/portfolio-perfomance/pkg/security"
	"github.com/shrike78/portfolio-perfomance/pkg/security/loaders/investingcom"
)

var isinToCurrId = map[string]string{
	"0P0000CWYZ": "1078533", //Allianz Insieme - Linea flessibile
	"0P00017EGO": "1078140", //Allianz Insieme - Linea obbligazionaria breve termine
	"0P00017EGP": "1077906", //Allianz Insieme - Linea obbligazionaria lungo termine
	"0P0000CX0R": "1078595", //Allianz Insieme - Linea obbligazionaria
	"0P0000CWZ4": "1078534", //Allianz Insieme - Linea bilanciata
	"0P0000CWZR": "1078584", //Allianz Insieme - Linea azionaria
	"0P00017EGQ": "1077905", //Allianz Insieme - Linea multiasset
}

type AllianzInsieme struct {
	name string
	isin string
}

func New(name, isin string) *AllianzInsieme {
	return &AllianzInsieme{
		name: name,
		isin: isin,
	}
}

func (e *AllianzInsieme) Name() string {
	return e.name
}

func (e *AllianzInsieme) ISIN() string {
	return e.isin
}

func (s *AllianzInsieme) LoadQuotes() ([]security.Quote, error) {
	var investingComLoader security.QuoteLoader
	investingComLoader = investingcom.New(s.name, isinToCurrId[s.isin])
	return investingComLoader.LoadQuotes()
}
