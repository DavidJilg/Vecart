package general

import "github.com/DavidJilg/Vecart/internal/shapes"

type Pixel struct {
	X1, Y1, X2, Y2   int
	midpoint         shapes.Point
	Darkness         int
	AdjustedDarkness float64
	Lines            []shapes.Polyline
	Border           shapes.Polyline
	LineIntersects   int
}

func NewPixel(X1, Y1, X2, Y2 int, Darkness int) *Pixel {
	midpoint := shapes.Point{X: float64(X1) + 0.5, Y: float64(Y1) + 0.5}

	lines := []shapes.Polyline{
		*shapes.NewPolyline(&[]shapes.Point{{X: float64(X1), Y: float64(Y1)}, {X: float64(X1), Y: float64(Y2)}}, nil),
		*shapes.NewPolyline(&[]shapes.Point{{X: float64(X1), Y: float64(Y2)}, {X: float64(X2), Y: float64(Y2)}}, nil),
		*shapes.NewPolyline(&[]shapes.Point{{X: float64(X2), Y: float64(Y2)}, {X: float64(X2), Y: float64(Y1)}}, nil),
		*shapes.NewPolyline(&[]shapes.Point{{X: float64(X2), Y: float64(Y1)}, {X: float64(X1), Y: float64(Y1)}}, nil),
	}

	border := *shapes.NewPolyline(
		&[]shapes.Point{
			{X: float64(X1), Y: float64(Y1)},
			{X: float64(X2), Y: float64(Y1)},
			{X: float64(X2), Y: float64(Y2)},
			{X: float64(X1), Y: float64(Y2)}},
		nil)

	return &Pixel{X1, Y1, X2, Y2, midpoint, Darkness, float64(Darkness), lines, border, 0}
}
