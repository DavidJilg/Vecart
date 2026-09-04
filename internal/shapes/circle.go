package shapes

import (
	"math"
	"strconv"

	"github.com/DavidJilg/Vecart/internal/utils"
)

type Circle struct {
	Center Point
	Radius float64
}

func NewCircle(origin Point, radius float64) *Circle {
	return &Circle{origin, radius}
}

func (circle *Circle) EqualTo(otherCircle *Circle, precision int) bool {
	if !circle.Center.EqualTo(&otherCircle.Center, precision) {
		return false
	}

	return float64Equal(circle.Radius, otherCircle.Radius, precision)
}

func (circle *Circle) ToPolyline(edges int) *Polyline {
	var points []Point

	points = append(points, *NewPoint(circle.Center.X, circle.Center.Y-circle.Radius))
	edgeAngle := 360.0 / float64(edges)

	for i := 0; i < edges; i++ {
		points = append(points, *points[len(points)-1].RotateCopy(edgeAngle, circle.Center))
	}

	return NewPolyline(&points, circle)
}

func (circle *Circle) ToShape() *Shape {
	line := circle.ToPolyline(6)
	shape := Shape{[]Polyline{*line}, nil, Point{0, 0}}
	shape.CalculateCentroid()

	return &shape
}

func (circle *Circle) Rotate(angle float64, origin Point) {
	circle.Center.Rotate(angle, origin)
}

func (circle *Circle) Move(x float64, y float64) {
	circle.Center.Move(x, y)
}

func (circle *Circle) Transform(a, b, c, d, e, f float64) {
	circle.Center.Transform(a, b, c, d, e, f)

	scaleX := math.Sqrt(a*a + b*b)
	scaleY := math.Sqrt(c*c + d*d)
	if math.Abs(scaleX-scaleY) <= 1e-9 {
		circle.Radius *= scaleX
	}
}

func (circle *Circle) Scale(factor float64, centroid Point) {
	circle.Radius *= factor
	circle.Center.X = factor*(circle.Center.X-centroid.X) + centroid.X
	circle.Center.Y = factor*(circle.Center.Y-centroid.Y) + centroid.Y
}

func (circle *Circle) MMToPixel(dpi float64) {
	circle.Center.MMToPixel(dpi)
	circle.Radius = utils.MMToPixel(circle.Radius, dpi)
}

func (circle *Circle) PixelToMM(dpi float64) {
	circle.Center.PixelToMM(dpi)
	circle.Radius = utils.PixelToMM(circle.Radius, dpi)
}

func (circle *Circle) Copy() *Circle {
	return &Circle{circle.Center.Copy(), circle.Radius}
}

func (circle *Circle) GetBoundries() (float64, float64, float64, float64) {
	return circle.Center.X - circle.Radius,
		circle.Center.Y - circle.Radius,
		circle.Center.X + circle.Radius,
		circle.Center.Y + circle.Radius
}

func (circle *Circle) ToSVG(style string) string {
	svg := "<circle cx=\""

	svg += strconv.FormatFloat(circle.Center.X, 'f', 2, 64)
	svg += "\" cy=\""
	svg += strconv.FormatFloat(circle.Center.Y, 'f', 2, 64)
	svg += "\" r=\""
	svg += strconv.FormatFloat(circle.Radius, 'f', 2, 64)
	svg += "\" style=\""
	svg += style
	svg += "\" />"
	return svg
}

func (circle *Circle) ToJSON() map[string]any {
	jsonMap := make(map[string]any)

	jsonMap["type"] = "circle"
	jsonMap["center"] = circle.Center.ToJSON()
	jsonMap["radius"] = circle.Radius

	return jsonMap
}

func (circle Circle) GetStrokeStart(stampMode bool) Point {
	if(stampMode) {
		return Point{X: circle.Center.X, Y: circle.Center.Y}
	}
	return Point{X: circle.Center.X + circle.Radius, Y: circle.Center.Y}
}

func (circle Circle) ToGCode(activateTool []string, deactivateTool []string, stampMode bool, feedrateXY int) []string {
	if stampMode {
		lines := make([]string, 0)
		lines = append(lines, utils.G00XY(circle.Center.X, circle.Center.Y))
		lines = append(lines, activateTool...)
		lines = append(lines, deactivateTool...)
		return lines
	}

	startX := circle.Center.X + circle.Radius
	startY := circle.Center.Y

	radius := circle.Radius

	lines := make([]string, 0)
	lines = append(lines, utils.G00XY(startX, startY))
	lines = append(lines, activateTool...)
	lines = append(lines, utils.G02(startX, startY, -radius, 0, feedrateXY))
	lines = append(lines, deactivateTool...)

	return lines
}
