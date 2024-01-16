package movingaverage

import (
	"github.com/gammazero/deque"
	"golang.org/x/exp/constraints"
)

type Number interface {
	constraints.Float | constraints.Integer
}

type MovingAverage[T Number]  struct {
	maxSize int
	data *deque.Deque[T]
	Average float64
}

func New[T Number](size int) *MovingAverage[T] {
	if size <= 0 {
		size = 1
	}
	data := deque.New[T]()
	ma := &MovingAverage[T]{
		maxSize: size,
		data: data,
	}
	return ma
}

func (ma *MovingAverage[T]) Update(i T) float64 {
	len := ma.data.Len()
	if len >= ma.maxSize {
		// forget old datapoints
		ma.Average -= float64(ma.data.PopFront()) / float64(len)
		len -= 1
	} else {
		// make room for the new datapoint
		ma.Average -= float64(ma.Average) / float64(len+1)
	}

	ma.data.PushBack(i)
	len += 1
	ma.Average += float64(i) / float64(len)

	/*
	fmt.Printf("%d[", len)
	for i := 0; i < ma.data.Len(); i++ {
		fmt.Printf("%d,", uint64(ma.data.At(i)))
	}
	fmt.Printf("], %d\n", uint64(ma.Average))
	*/

	return ma.Average
}
