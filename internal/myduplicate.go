package internal

import "github.com/cinar/indicator/v2/helper"

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
