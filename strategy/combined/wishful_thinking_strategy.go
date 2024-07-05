// WishfulThinking: MACD, Awesome Oscillator, RSI, and Ulcer Index strategy.

// Macd represents the configuration parameters for calculating the
// Moving Average Convergence Divergence (MACD).
//
//	MACD = 12-Period EMA - 26-Period EMA.
//	Signal = 9-Period EMA of MACD.
//
// Rsi represents the configuration parameter for calculating the Relative Strength Index (RSI).  It is a momentum
// indicator that measures the magnitude of recent price changes to evaluate overbought and oversold conditions.
//
//	RS = Average Gain / Average Loss
//	RSI = 100 - (100 / (1 + RS))
//
// AwesomeOscillator represents the configuration parameter for calculating the Awesome Oscillator (AO). It gauges
// market momentum by comparing short-term price action (5-period average) against long-term trends (34-period
// average). Its value around a zero line reflects bullishness above and bearishness below. Crossings of the
// zero line can signal potential trend reversals. Traders use the AO to confirm existing trends, identify
// entry/exit points, and understand momentum shifts.
//
//	Median Price = ((Low + High) / 2).
//	AO = 5-Period SMA - 34-Period SMA.

package combined

import (
	"fmt"

	"github.com/cinar/indicator/v2/asset"
	"github.com/cinar/indicator/v2/helper"
	"github.com/cinar/indicator/v2/strategy"
	"github.com/cinar/indicator/v2/strategy/momentum"
	"github.com/cinar/indicator/v2/volatility"
	alt_trend "github.com/vextasy/strategise/strategy/trend"
)

const (
	// DefaultMacdRsiStrategyBuyAt defines the default RSI level at which a Buy action is generated.
	DefaultMacdRsiStrategyBuyAt = 30

	// DefaultMacdRsiStrategySellAt defines the default RSI level at which a Sell action is generated.
	DefaultMacdRsiStrategySellAt = 70
)

type WishfulThinkingStrategy struct {
	strategy.Strategy

	// MacdStrategy is the MACD strategy instance.
	//MacdStrategy *alt_trend.BoldMacdStrategy
	MacdStrategy *alt_trend.BoldMacdStrategy

	// AwesomeOacillatorStrategy is the Awesome Oscillator strategy instance.
	AwesomeOscillatorStrategy *momentum.AwesomeOscillatorStrategy

	// RsiStrategy is the RSI strategy instance.
	RsiStrategy *momentum.RsiStrategy

	// OrStrategy is the OR strategy instance.
	OrStrategy *strategy.OrStrategy

	// UlcerIndexStrategy is the Ulcer Index strategy instance.
	//UlcerIndexStrategy *momentum.UlcerIndexStrategy
}

func NewWishfulThinkingStrategy() *WishfulThinkingStrategy {
	return NewWishfulThinkingStrategyWith(
		DefaultMacdRsiStrategyBuyAt,
		DefaultMacdRsiStrategySellAt,
	)
}

func NewWishfulThinkingStrategyWith(buyAt, sellAt float64) *WishfulThinkingStrategy {
	s := &WishfulThinkingStrategy{
		MacdStrategy:              alt_trend.NewBoldMacdStrategy(),
		AwesomeOscillatorStrategy: momentum.NewAwesomeOscillatorStrategy(),
		RsiStrategy:               momentum.NewRsiStrategyWith(buyAt, sellAt),
		OrStrategy:                strategy.NewOrStrategy("OR Strategy"),
	}
	// MACD, Awesome Oscillator, and RSI outcomes are ORed together.
	inner := strategy.NewOrStrategy("Inner OR Strategy")
	inner.Strategies = append(inner.Strategies, s.MacdStrategy, s.AwesomeOscillatorStrategy)
	outer := strategy.NewOrStrategy("Outer OR Strategy")
	outer.Strategies = append(outer.Strategies, inner, s.RsiStrategy)
	s.OrStrategy = outer
	return s
}

// Name returns the name of the strategy.
func (m *WishfulThinkingStrategy) Name() string {
	return fmt.Sprintf("Wishful Thinking Strategy (%.0f, %.0f)",
		m.RsiStrategy.BuyAt,
		m.RsiStrategy.SellAt,
	)
}

// Compute processes the provided asset snapshots and generates a stream of actionable recommendations.
func (m *WishfulThinkingStrategy) Compute(snapshots <-chan *asset.Snapshot) <-chan strategy.Action {
	actions := m.OrStrategy.Compute(snapshots)
	actions = strategy.NormalizeActions(actions)

	return actions
}

func (m *WishfulThinkingStrategy) Report(c <-chan *asset.Snapshot) *helper.Report {
	//
	// snapshots[0] -> dates
	// snapshots[1] -> closings[0] -> macds, signals
	//                 closings[1] -> rsi
	//                 closings[2] -> ao
	//                 closings[3] -> closings chart
	//                 closings[4] -> ulcer index
	// snapshots[2] -> highs       -> ao highs
	// snapshots[3] -> lows        -> ao lows
	// snapshots[4] -> actions     -> annotations
	//              -> outcomes
	// snapshots[5] -> macd_actions-> macd_annotations
	// snapshots[6] -> rsi_actions -> rsi_annotations
	// snapshots[7] -> ao_actions  -> ao_annotations
	//
	snapshots := helper.Duplicate(c, 8)

	dates := asset.SnapshotsAsDates(snapshots[0])
	closings := helper.Duplicate(asset.SnapshotsAsClosings(snapshots[1]), 5)

	// MACD
	macds, signals := m.MacdStrategy.Macd.Compute(closings[0])
	macds = helper.Shift(macds, m.MacdStrategy.Macd.IdlePeriod(), 0)
	signals = helper.Shift(signals, m.MacdStrategy.Macd.IdlePeriod(), 0)

	// RSI
	rsi := m.RsiStrategy.Rsi.Compute(closings[1])
	rsi = helper.Shift(rsi, m.RsiStrategy.Rsi.IdlePeriod(), 0)

	// Awesome Oscillator
	highs := asset.SnapshotsAsHighs(snapshots[2])
	lows := asset.SnapshotsAsLows(snapshots[3])
	ao := m.AwesomeOscillatorStrategy.AwesomeOscillator.Compute(highs, lows)
	ao = helper.Shift(ao, m.AwesomeOscillatorStrategy.AwesomeOscillator.IdlePeriod(), 0)

	// Wishful Thinking outcomes & annotations
	actions, outcomes := strategy.ComputeWithOutcome(m, snapshots[4])
	annotations := strategy.ActionsToAnnotations(actions)
	outcomes = helper.MultiplyBy(outcomes, 100)

	// Ulcer index
	ulcer_index_indicator := volatility.NewUlcerIndex[float64]()
	ulcer_index := ulcer_index_indicator.Compute(closings[4])
	ulcer_index = helper.Shift(ulcer_index, ulcer_index_indicator.IdlePeriod(), 0)

	// Other annotations
	macd_actions := m.MacdStrategy.Compute(snapshots[5])
	macd_annotations := strategy.ActionsToAnnotations(macd_actions)
	rsi_actions := m.RsiStrategy.Compute(snapshots[6])
	rsi_annotations := strategy.ActionsToAnnotations(rsi_actions)
	ao_actions := m.AwesomeOscillatorStrategy.Compute(snapshots[7])
	ao_annotations := strategy.ActionsToAnnotations(ao_actions)

	report := helper.NewReport(m.Name(), dates) // Close
	report.AddChart()                           // MACD
	report.AddChart()                           // RSI
	report.AddChart()                           // OA
	report.AddChart()                           // Outcome
	report.AddChart()                           // Ulcer Index

	report.AddColumn(helper.NewNumericReportColumn("Close", closings[3]))

	report.AddColumn(helper.NewNumericReportColumn("MACD", macds), 1)
	report.AddColumn(helper.NewNumericReportColumn("Signal", signals), 1)
	report.AddColumn(helper.NewAnnotationReportColumn(macd_annotations), 1)

	report.AddColumn(helper.NewNumericReportColumn("RSI", rsi), 2)
	report.AddColumn(helper.NewAnnotationReportColumn(rsi_annotations), 2)

	report.AddColumn(helper.NewNumericReportColumn("AO", ao), 3)
	report.AddColumn(helper.NewAnnotationReportColumn(ao_annotations), 3)

	report.AddColumn(helper.NewNumericReportColumn("Ulcer", ulcer_index), 4)

	report.AddColumn(helper.NewAnnotationReportColumn(annotations), 0)

	report.AddColumn(helper.NewNumericReportColumn("Outcome", outcomes), 5)

	return report
}
