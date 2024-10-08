package main

type Polygon struct {
	points []Point
}

func NewPolygon(points *[]Point) *Polygon {
	return &Polygon{*points}
}

func (polygon *Polygon) equalTo(otherPolygon *Polygon, precision int) bool {
	pointsA := append(polygon.points, polygon.points[0])
	pointsB := append(otherPolygon.points, otherPolygon.points[0])
	return equalLineSegments(&pointsA, &pointsB, precision)
}

func (polygon *Polygon) toShape() *Shape {
	line := polygon.toPolyline()
	shape := Shape{[]Polyline{*line}, nil, Point{0, 0}}
	shape.calculateCentroid()

	return &shape
}

func (polygon *Polygon) toPolyline() *Polyline {
	var fullPoints []Point
	for index := range polygon.points {
		fullPoints = append(fullPoints, polygon.points[index].copy())
	}
	fullPoints = append(fullPoints, polygon.points[0])

	return NewPolyline(&fullPoints, polygon)
}

func (polygon *Polygon) rotate(angle float64, origin Point) {
	for index := range polygon.points {
		polygon.points[index].rotate(angle, origin)
	}
}

func (polygon *Polygon) transform(x float64, y float64) {
	for index := range polygon.points {
		polygon.points[index].transform(x, y)
	}
}

func (polygon *Polygon) scale(factor float64, centroid Point) {
	for index := range polygon.points {
		polygon.points[index].scale(factor, centroid)
	}
}

func (polygon *Polygon) mmToPixel(dpi float64) {
	for index := range polygon.points {
		polygon.points[index].mmToPixel(dpi)
	}
}

func (polygon *Polygon) pixelToMM(dpi float64) {
	for index := range polygon.points {
		polygon.points[index].pixelToMM(dpi)
	}
}

func (polygon *Polygon) copy() *Polygon {
	var points []Point
	for _, point := range polygon.points {
		points = append(points, point.copy())
	}
	return &Polygon{points}
}

func (polygon *Polygon) toSVG(style string) string {
	svg := "<polygon  points=\""

	for index, point := range polygon.points {
		svg += point.toSVG()
		if index != len(polygon.points)-1 {
			svg += " "
		}
	}

	svg += "\" style=\""
	svg += style
	svg += "\" />"
	return svg
}

func (polygon *Polygon) toJSON() map[string]any {
	jsonMap := make(map[string]any)

	jsonMap["type"] = "polygon"

	points := []any{}
	for _, point := range polygon.points {
		points = append(points, point.toJSON())
	}
	jsonMap["points"] = points

	return jsonMap
}
