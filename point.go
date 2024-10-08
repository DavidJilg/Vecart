package main

import (
	"math"
	"strconv"
)

type Point struct {
	X, Y float64
}

func NewPoint(x, y float64) *Point {
	return &Point{x, y}
}

func (point *Point) rotate(angle float64, origin Point) {
	angle = (angle * math.Pi) / 180

	newX := origin.X + (point.X-origin.X)*math.Cos(angle) - (point.Y-origin.Y)*math.Sin(angle)
	newY := origin.Y + (point.X-origin.X)*math.Sin(angle) + (point.Y-origin.Y)*math.Cos(angle)

	point.X = newX
	point.Y = newY
}

func (point *Point) rotateCopy(angle float64, origin Point) *Point {
	pointCopy := point.copy()
	pointCopy.rotate(angle, origin)
	return &pointCopy
}

func (point *Point) transform(x float64, y float64) {
	point.X += x
	point.Y += y
}

func (point *Point) scale(factor float64, centroid Point) {
	point.X = factor*(point.X-centroid.X) + centroid.X
	point.Y = factor*(point.Y-centroid.Y) + centroid.Y
}

func (point *Point) copy() Point {
	return Point{point.X, point.Y}
}

func (point *Point) distanceTo(otherPoint *Point) float64 {
	return math.Sqrt(math.Pow(otherPoint.X-point.X, 2) + math.Pow(otherPoint.Y-point.Y, 2))
}

func (point *Point) toSVG() string {
	return strconv.FormatFloat(point.X, 'f', 6, 64) + " " + strconv.FormatFloat(point.Y, 'f', 6, 64)
}

func (point *Point) equalTo(otherPoint *Point, precision int) bool {
	return float64Equal(point.X, otherPoint.X, precision) && float64Equal(point.Y, otherPoint.Y, precision)
}

func float64Equal(a, b float64, precision int) bool {
	precisionString := "0."
	for range precision {
		precisionString += "0"
	}
	return precisionString == strconv.FormatFloat(math.Abs(a-b), 'f', precision, 64)
}

func (point *Point) mmToPixel(dpi float64) {
	point.X = mmToPixel(point.X, dpi)
	point.Y = mmToPixel(point.Y, dpi)
}

func (point *Point) pixelToMM(dpi float64) {
	point.X = PixelToMM(point.X, dpi)
	point.Y = PixelToMM(point.Y, dpi)
}

func (point *Point) toJSON() []float64 {
	return []float64{point.X, point.Y}
}
