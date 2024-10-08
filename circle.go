package main

import (
	"strconv"
)

type Circle struct {
	center Point
	radius float64
}

func NewCircle(origin Point, radius float64) *Circle {
	return &Circle{origin, radius}
}

func (circle *Circle) equalTo(otherCircle *Circle, precision int) bool {
	if !circle.center.equalTo(&otherCircle.center, precision) {
		return false
	}

	return float64Equal(circle.radius, otherCircle.radius, precision)
}

func (circle *Circle) toPolyline(edges int) *Polyline {
	var points []Point

	points = append(points, *NewPoint(circle.center.X, circle.center.Y-circle.radius))
	edgeAngle := 360.0 / float64(edges)

	for i := 0; i < edges; i++ {
		points = append(points, *points[len(points)-1].rotateCopy(edgeAngle, circle.center))
	}

	return NewPolyline(&points, circle)
}

func (circle *Circle) toShape() *Shape {
	line := circle.toPolyline(6)
	shape := Shape{[]Polyline{*line}, nil, Point{0, 0}}
	shape.calculateCentroid()

	return &shape
}

func (circle *Circle) rotate(angle float64, origin Point) {
	circle.center.rotate(angle, origin)
}

func (circle *Circle) transform(x float64, y float64) {
	circle.center.transform(x, y)
}

func (circle *Circle) scale(factor float64, centroid Point) {
	circle.radius *= factor
	circle.center.X = factor*(circle.center.X-centroid.X) + centroid.X
	circle.center.Y = factor*(circle.center.Y-centroid.Y) + centroid.Y
}

func (circle *Circle) mmToPixel(dpi float64) {
	circle.center.mmToPixel(dpi)
	circle.radius = mmToPixel(circle.radius, dpi)
}

func (circle *Circle) pixelToMM(dpi float64) {
	circle.center.pixelToMM(dpi)
	circle.radius = PixelToMM(circle.radius, dpi)
}

func (circle *Circle) copy() *Circle {
	return &Circle{circle.center.copy(), circle.radius}
}

func (circle *Circle) toSVG(style string) string {
	svg := "<circle cx=\""

	svg += strconv.FormatFloat(circle.center.X, 'f', 2, 64)
	svg += "\" cy=\""
	svg += strconv.FormatFloat(circle.center.Y, 'f', 2, 64)
	svg += "\" r=\""
	svg += strconv.FormatFloat(circle.radius, 'f', 2, 64)
	svg += "\" style=\""
	svg += style
	svg += "\" />"
	return svg
}

func (circle *Circle) toJSON() map[string]any {
	jsonMap := make(map[string]any)

	jsonMap["type"] = "circle"
	jsonMap["center"] = circle.center.toJSON()
	jsonMap["radius"] = circle.radius

	return jsonMap
}
