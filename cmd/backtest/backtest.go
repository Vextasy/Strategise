package main

import (
	"fmt"

	"github.com/cinar/indicator/v2/strategy"
	"github.com/cinar/indicator/v2/strategy/compound"
	"github.com/cinar/indicator/v2/strategy/momentum"
	"github.com/cinar/indicator/v2/strategy/trend"
	"github.com/cinar/indicator/v2/strategy/volatility"
	"github.com/vextasy/strategise/app"
	"github.com/vextasy/strategise/strategy/combined"
	alt_trend "github.com/vextasy/strategise/strategy/trend"
)

const datadir = "/Users/john/Downloads/PPData"
const backtestdir = "/Users/john/Downloads/PPBacktest"

func main() {

	// Read the Portfolio Performance XML file
	r, err := app.NewPortfolioPerformanceRepository(datadir + "/portfolio.xml")
	if err != nil {
		fmt.Println("Error reading XML file:", err)
		return
	}

	b := strategy.NewBacktest(r, backtestdir)
	b.LastDays = 365
	b.Strategies = []strategy.Strategy{
		combined.NewWishfulThinkingStrategyWith(30, 70),
		combined.NewAwesomeMbuStrategyWith(40, 60),
		strategy.NewBuyAndHoldStrategy(),
		volatility.NewBollingerBandsStrategy(),
		////volatility.NewSuperTrendStrategy(),
		trend.NewMacdStrategy(),
		trend.NewMacdStrategyWith(5, 35, 5),
		alt_trend.NewBoldMacdStrategy(),
		trend.NewApoStrategy(),
		trend.NewAroonStrategy(),
		//trend.NewBopStrategy(),
		////trend.NewCciStrategy(),
		////trend.NewDemaStrategy(),
		//trend.NewGoldenCrossStrategy(),
		////trend.NewKamaStrategy(),
		trend.NewKdjStrategy(),
		trend.NewQstickStrategy(),
		////trend.NewTrimaStrategy(),
		////trend.NewTripleMovingAverageCrossoverStrategy(),
		////trend.NewTrixStrategy(),
		////trend.NewTsiStrategy(),
		////trend.NewVwmaStrategy(),
		momentum.NewRsiStrategy(),
		momentum.NewRsiStrategyWith(40, 60),
		momentum.NewAwesomeOscillatorStrategy(),
		////momentum.NewStochasticRsiStrategy(),
		momentum.NewTripleRsiStrategy(),
		compound.NewMacdRsiStrategy(),
	}
	err = b.Run()

	if err != nil {
		fmt.Println("Error running backtest:", err)
		return
	}
}
