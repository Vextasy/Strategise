package main

import (
	"fmt"

	"github.com/cinar/indicator/v2/helper"
	"github.com/cinar/indicator/v2/strategy"
)

func main() {

	ch := helper.SliceToChan([]strategy.Action{strategy.Sell, strategy.Hold, strategy.Hold})
	ch = strategy.DenormalizeActions(ch)
	sl := helper.ChanToSlice(helper.Map(ch, annotate))
	fmt.Println(sl)

}

func annotate(a strategy.Action) string {
	switch a {
	case strategy.Sell:
		return "Sell"

	case strategy.Buy:
		return "Buy"

	default:
		return "Hold"
	}
}
