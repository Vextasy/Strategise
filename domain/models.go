package domain

import (
	"encoding/xml"
	"time"
)

// Portfolio Performance XML structure
type Client struct {
	XMLName    xml.Name   `xml:"client"`
	Securities []Security `xml:"securities>security"`
}

// A Security contains information about a given security within the XML file.
type Security struct {
	Name         string    `xml:"name"`
	CurrencyCode string    `xml:"currencyCode"`
	ISIN         string    `xml:"isin"`
	TickerSymbol string    `xml:"tickerSymbol"`
	Prices       []Price   `xml:"prices>price"`
	IsRetired    string    `xml:"isRetired"` // "false" or "true"
	UpdatedAt    time.Time `xml:"updatedAt"`
}
type Price struct {
	Date  string  `xml:"t,attr"`
	Value float64 `xml:"v,attr"`
}
