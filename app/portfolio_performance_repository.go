package app

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/cinar/indicator/v2/asset"
	"github.com/cinar/indicator/v2/helper"
	"github.com/vextasy/strategise/domain"
	"github.com/vextasy/strategise/internal"
)

// portfolioPerformanceRepository stores data for each secuity in the portfolio.
type portfolioPerformanceRepository struct {
	Securities map[string]domain.Security

	dateFormat string // Date format used by the indicator library
}

// NewPortfolioPerformanceRepository initialises the repository from an XML file.
func NewPortfolioPerformanceRepository(path string) (asset.Repository, error) {

	// Read the XML file
	xmlFile, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading XML file:", err)
		return nil, err
	}

	// Create a Client struct to unmarshal the XML data into
	var client domain.Client

	// Unmarshal the XML data into the client struct
	err = xml.Unmarshal(xmlFile, &client)
	if err != nil {
		fmt.Println("Error unmarshaling XML:", err)
		return nil, err
	}

	r := &portfolioPerformanceRepository{
		Securities: make(map[string]domain.Security),
	}

	// Ensure that all security names are suitable for writing as a filename.
	for i := range client.Securities {
		client.Securities[i].Name = internal.CleanFilename(client.Securities[i].Name)
	}

	// r.dateFormat is the date format expected by the indicator library.
	r.dateFormat, err = internal.GetStructTag(asset.Snapshot{}, "Date", "format")
	if err != nil {
		fmt.Println("Error getting struct tag for asset.Snapshot Date format:", err)
		return nil, err
	}
	// Ensure that Price Date string is in the format expected by the indicator library.
	for _, security := range client.Securities {
		for _, price := range security.Prices {
			_, err := time.Parse(r.dateFormat, price.Date)
			if err != nil {
				fmt.Println("Error parsing Price Date", price.Date, "for security", security.Name, ": ", err)
				return nil, err
			}
		}
	}

	for _, security := range client.Securities {
		// Ensure that prices appear in ascending date order.
		sort.SliceStable(security.Prices, func(i int, j int) bool {
			return security.Prices[i].Date < security.Prices[j].Date
		})

		// Portfolio Performance stores the price.Value scaled up by 1e8.
		for pi := range security.Prices {
			security.Prices[pi].Value = security.Prices[pi].Value / 1e8
		}
		r.Securities[security.Name] = security
	}

	return r, nil
}

// Assets returns the names of all non-retired assets in the repository.
func (r *portfolioPerformanceRepository) Assets() ([]string, error) {
	assets := make([]string, 0, len(r.Securities))
	for _, security := range r.Securities {
		if security.IsRetired == "true" {
			continue
		}
		assets = append(assets, security.Name)
	}
	sort.SliceStable(assets, func(i int, j int) bool {
		return assets[i] < assets[j]
	})
	return assets, nil
}

// Get returns the snapshots for the asset with the given name.
// Our data source only contains daily closing prices
// so we manufacture the opening price as the previous close
// and the high and low prices from the opening and closing prices.
func (r *portfolioPerformanceRepository) Get(name string) (<-chan *asset.Snapshot, error) {
	security, ok := r.Securities[name]
	if !ok {
		return nil, asset.ErrRepositoryAssetNotFound
	}
	c := make(chan *asset.Snapshot)

	go func() {
		defer close(c)

		var last_close float64
		for i, price := range security.Prices {
			var open, high, low, close float64

			close = price.Value
			if i == 0 {
				open = close
			} else {
				open = last_close
			}
			last_close = close
			if close > open {
				high = close
				low = open
			} else {
				high = open
				low = close
			}
			date, _ := time.Parse(r.dateFormat, price.Date)
			c <- &asset.Snapshot{
				Date:   date,
				Open:   open,
				High:   high,
				Low:    low,
				Close:  close,
				Volume: 0,
			}
		}
	}()

	return c, nil
}

// GetSince returns a channel of snapshots for the asset with the given name since the given date.
func (r *portfolioPerformanceRepository) GetSince(name string, date time.Time) (<-chan *asset.Snapshot, error) {
	snapshots, err := r.Get(name)
	if err != nil {
		return nil, err
	}

	snapshots = helper.Filter(snapshots, func(s *asset.Snapshot) bool {
		return s.Date.Equal(date) || s.Date.After(date)
	})

	return snapshots, nil
}

// LastDate returns the date of the last snapshot for the asset with the given name.
func (r *portfolioPerformanceRepository) LastDate(name string) (time.Time, error) {
	var last time.Time

	snapshots, err := r.Get(name)
	if err != nil {
		return last, err
	}

	snapshot, ok := <-helper.Last(snapshots, 1)
	if !ok {
		return last, errors.New("empty asset")
	}

	return snapshot.Date, nil
}

// Append adds the given snapshots to the asset with the given name.
func (r *portfolioPerformanceRepository) Append(name string, snapshots <-chan *asset.Snapshot) error {

	for s := range snapshots {
		security, ok := r.Securities[name]
		if !ok {
			return asset.ErrRepositoryAssetNotFound
		}
		security.Prices = append(security.Prices, domain.Price{
			Date:  s.Date.Format(r.dateFormat),
			Value: s.Close,
		})
		r.Securities[name] = security
	}
	return nil
}
