package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cinar/indicator/v2/asset"
	"github.com/cinar/indicator/v2/helper"
	"github.com/cinar/indicator/v2/strategy"
	"github.com/cinar/indicator/v2/strategy/momentum"
	"github.com/cinar/indicator/v2/strategy/trend"
	"github.com/cinar/indicator/v2/strategy/volatility"

	portfolio_performance "github.com/vextasy/strategise/app"
)

const datadir = "/Users/john/Downloads/PPData"
const reportdir = "/Users/john/Downloads/PPReport"

func main() {
	// Read the Portfolio Performance XML file
	r, err := portfolio_performance.NewPortfolioPerformanceRepository(datadir + "/portfolio.xml")
	if err != nil {
		fmt.Println("Error reading XML file:", err)
		return
	}

	assets, _ := r.Assets()

	for ai := range assets {
		strategies := []strategy.Strategy{
			volatility.NewBollingerBandsStrategy(),
			////volatility.NewSuperTrendStrategy(),
			trend.NewMacdStrategy(),
			trend.NewMacdStrategyWith(5, 35, 5),
			//trend.NewApoStrategy(),
			//trend.NewAroonStrategy(),
			//trend.NewBopStrategy(),
			////trend.NewCciStrategy(),
			////trend.NewDemaStrategy(),
			//trend.NewGoldenCrossStrategy(),
			////trend.NewKamaStrategy(),
			//trend.NewKdjStrategy(),
			trend.NewQstickStrategy(),
			////trend.NewTrimaStrategy(),
			////trend.NewTripleMovingAverageCrossoverStrategy(),
			////trend.NewTrixStrategy(),
			////trend.NewTsiStrategy(),
			////trend.NewVwmaStrategy(),
			//momentum.NewRsiStrategy(),
			////momentum.NewAwesomeOscillatorStrategy(),
			////momentum.NewStochasticRsiStrategy(),
			////momentum.NewTripleRsiStrategy(),
			//compound.NewMacdRsiStrategy(),
		}
		for si := range strategies {
			snapshots, err := r.Get(assets[ai])
			if err != nil {
				fmt.Println("Error reading asset snapshots:", err)
				return
			}
			data0, data1, _ := duplicateChan(snapshots)
			runAction(strategies[si], assets[ai], data0)
			runReport(strategies[si], assets[ai], data1)
		}
	}
}

// duplicateChan creates a new channel containing a copy of the data from 'cin'
// and returns it and its length.
func duplicateChan(cin <-chan *asset.Snapshot) (<-chan *asset.Snapshot, <-chan *asset.Snapshot, int) {
	slice := helper.ChanToSlice(cin)
	cout0 := helper.SliceToChan(slice)
	cout1 := helper.SliceToChan(slice)
	return cout0, cout1, len(slice)
}

func runReport(st strategy.Strategy, assetName string, data <-chan *asset.Snapshot) {
	data, _, datalen := duplicateChan(data)
	// Detect certain strategies that require a minimum amount of data.
	if notEnoughData(st, assetName, datalen) {
		return
	}
	rep := st.Report(data)
	filepath := fmt.Sprintf("%s/%s--%s.html", reportdir, cleanFilename(assetName), cleanFilename(st.Name()))
	err := rep.WriteToFile(filepath)
	if err != nil {
		fmt.Println("Error writing report:", err)
		return
	}
}

func runAction(st strategy.Strategy, assetName string, data <-chan *asset.Snapshot) {
	data, _, datalen := duplicateChan(data)
	// Detect certain strategies that require a minimum amount of data.
	if notEnoughData(st, assetName, datalen) {
		return
	}
	actions := strategy.DenormalizeActions(st.Compute(data))
	actionSlice := helper.ChanToSlice(actions)
	action := actionSlice[len(actionSlice)-1]
	actionString := mkActionString(action)
	setActionFile(actionString, cleanFilename(assetName), cleanFilename(st.Name()))
}

func notEnoughData(st strategy.Strategy, assetName string, datalen int) bool {
	msg := fmt.Sprintf("Ignoring strategy %s for %s due to insufficient data: %s", st.Name(), assetName, st.Name())
	if st.Name()[0:4] == "MACD" {
		if datalen < trend.NewMacdStrategy().Macd.Ema2.Period {
			fmt.Println(msg)
			return true
		}
	}
	if st.Name() == "Awesome Oscillator Strategy" {
		if datalen < momentum.NewAwesomeOscillatorStrategy().AwesomeOscillator.LongSma.Period {
			fmt.Println(msg)
			return true
		}
	}
	return false
}

// mkActionString returns one of BUY, SELL or HOLD for a given strategy.Action
func mkActionString(action strategy.Action) string {
	switch action {
	case strategy.Buy:
		return "BUY"
	case strategy.Sell:
		return "SELL"
	case strategy.Hold:
		return "HOLD"
	}
	return "HOLD"
}

// setActionFile replaces any existing action file with a new one.
// It removes any existing BUY, SELL or HOLD files.
// Arguments are assumed to be safe for forming part of a filename.
func setActionFile(actionString string, assetName string, strategyName string) {
	filePath := func(act string) string {
		return fmt.Sprintf("%s/%s--%s--%s.txt", reportdir, assetName, strategyName, act)
	}
	// Start by removing any existing BUY, SELL or HOLD files.
	for _, act := range []string{"BUY", "SELL", "HOLD"} {
		path := filePath(act)
		os.Remove(path)
	}
	newfile := filePath(actionString)
	fd, _ := os.Create(newfile)
	defer fd.Close()
}

func cleanFilename(filename string) string {
	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "&", "And")
	filename = strings.ReplaceAll(filename, "/", "-")
	return filepath.Clean(filename)
}

func myDuplicate[T any](input <-chan T, count int) []<-chan T {
	outputs := make([]chan T, count)
	result := make([]<-chan T, count)

	inputSlice := helper.ChanToSlice[T](input)

	for i := range outputs {
		outputs[i] = make(chan T, cap(input))
		result[i] = outputs[i]
	}

	for _, output := range outputs {
		o := output
		go func() {
			defer close(o)

			for _, n := range inputSlice {
				o <- n
			}
		}()
	}

	return result
}