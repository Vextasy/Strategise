package internal

import "github.com/cinar/indicator/v2/helper"

func Dup[T any](input <-chan T) (<-chan T, <-chan T) {
	chans := helper.Duplicate(input, 2)
	return chans[0], chans[1]
}
func Dup3[T any](input <-chan T) (<-chan T, <-chan T, <-chan T) {
	chans := helper.Duplicate(input, 3)
	return chans[0], chans[1], chans[2]
}
func DupN[T any](input <-chan T, count int) []<-chan T {
	return helper.Duplicate(input, count)
}
