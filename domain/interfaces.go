package domain

import (
	"github.com/cinar/indicator/v2/asset"
)

type PortfolioPerformanceRepository interface {
	// Assets returns the names of all assets in the repository.
	Assets() ([]string, error)

	// Get returns the Snapshots for the given asset.
	Get(name string) ([]*asset.Snapshot, error)
}
