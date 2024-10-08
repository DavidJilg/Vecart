package main

type Pixel struct {
	X1, Y1, X2, Y2   int
	midpoint         Point
	Darkness         int
	AdjustedDarkness float64
	Lines            []Polyline
	Border           Polyline
	LineIntersects   int
}

func NewPixel(X1, Y1, X2, Y2 int, Darkness int) *Pixel {
	midpoint := Point{float64(X1) + 0.5, float64(Y1) + 0.5}

	lines := []Polyline{
		{[]Point{{float64(X1), float64(Y1)}, {float64(X1), float64(Y2)}}, nil},
		{[]Point{{float64(X1), float64(Y2)}, {float64(X2), float64(Y2)}}, nil},
		{[]Point{{float64(X2), float64(Y2)}, {float64(X2), float64(Y1)}}, nil},
		{[]Point{{float64(X2), float64(Y1)}, {float64(X1), float64(Y1)}}, nil},
	}

	border := Polyline{[]Point{{float64(X1), float64(Y1)},
		{float64(X2), float64(Y1)},
		{float64(X2), float64(Y2)},
		{float64(X1), float64(Y2)}}, nil}

	return &Pixel{X1, Y1, X2, Y2, midpoint, Darkness, float64(Darkness), lines, border, 0}
}
