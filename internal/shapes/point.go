package shapes

import (
	"math"
	"strconv"

	"github.com/DavidJilg/Vecart/internal/utils"
)

type Point struct {
	X, Y float64
}

func NewPoint(x, y float64) *Point {
	return &Point{x, y}
}

func (point *Point) Rotate(angle float64, origin Point) {
	angle = (angle * math.Pi) / 180

	newX := origin.X + (point.X-origin.X)*math.Cos(angle) - (point.Y-origin.Y)*math.Sin(angle)
	newY := origin.Y + (point.X-origin.X)*math.Sin(angle) + (point.Y-origin.Y)*math.Cos(angle)

	point.X = newX
	point.Y = newY
}

func (point *Point) RotateCopy(angle float64, origin Point) *Point {
	pointCopy := point.Copy()
	pointCopy.Rotate(angle, origin)
	return &pointCopy
}

func (point *Point) Move(x float64, y float64) {
	point.X += x
	point.Y += y
}

func (point *Point) Transform(a, b, c, d, e, f float64) {
	x := point.X
	y := point.Y
	point.X = a*x + c*y + e
	point.Y = b*x + d*y + f
}

func (point *Point) Scale(factor float64, centroid Point) {
	point.X = factor*(point.X-centroid.X) + centroid.X
	point.Y = factor*(point.Y-centroid.Y) + centroid.Y
}

func (point *Point) Copy() Point {
	return Point{point.X, point.Y}
}

func (point *Point) DistanceTo(otherPoint *Point) float64 {
	return math.Sqrt(math.Pow(otherPoint.X-point.X, 2) + math.Pow(otherPoint.Y-point.Y, 2))
}

func (point *Point) ToSVGFormat() string {
	return strconv.FormatFloat(point.X, 'f', 6, 64) + " " + strconv.FormatFloat(point.Y, 'f', 6, 64)
}

func (point *Point) ToSVG(activateTool []string, deactivateTool []string) []string {
	lines := make([]string, 0)
	lines = append(lines, utils.G00XY(point.X, point.Y))
	lines = append(lines, activateTool...)
	lines = append(lines, deactivateTool...)

	return lines
}

func (point *Point) EqualTo(otherPoint *Point, precision int) bool {
	return float64Equal(point.X, otherPoint.X, precision) && float64Equal(point.Y, otherPoint.Y, precision)
}

func float64Equal(a, b float64, precision int) bool {
	precisionString := "0."
	for range precision {
		precisionString += "0"
	}
	return precisionString == strconv.FormatFloat(math.Abs(a-b), 'f', precision, 64)
}

func (point *Point) MMToPixel(dpi float64) {
	point.X = utils.MMToPixel(point.X, dpi)
	point.Y = utils.MMToPixel(point.Y, dpi)
}

func (point *Point) PixelToMM(dpi float64) {
	point.X = utils.PixelToMM(point.X, dpi)
	point.Y = utils.PixelToMM(point.Y, dpi)
}

func (point *Point) ToJSON() []float64 {
	return []float64{point.X, point.Y}
}

func (point Point) ToGCode(activateTool []string, deactivateTool []string) []string {
	lines := make([]string, 0)
	lines = append(lines, utils.G00XY(point.X, point.Y))
	lines = append(lines, activateTool...)
	lines = append(lines, deactivateTool...)

	return lines
}
