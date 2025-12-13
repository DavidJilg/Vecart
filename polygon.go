package main

type Polygon struct {
	points []Point
}

func NewPolygon(points *[]Point) *Polygon {
	return &Polygon{*points}
}

func NewPolygonShapeWithOriginalPoints(points *[]Point, originalPoints *[]Point) *Shape {
	line := NewPolygon(points).toPolyline()
	line.originalShape = NewPolygon(originalPoints)
	shape := Shape{[]Polyline{*line}, nil, Point{0, 0}}
	shape.calculateCentroid()

	return &shape
}

func NewHeart(size float64) *Shape {
	originalPoints := []Point{
		{277.4, 409.2},
		{277.9, 408.7},
		{278.4, 408.4},
		{279.1, 408.1},
		{279.8, 408.0},
		{280.5, 408.1},
		{281.3, 408.5},
		{282.0, 409.1},
		{282.5, 409.7},
		{282.7, 410.5},
		{282.8, 411.0},
		{282.7, 411.7},
		{282.5, 412.4},
		{282.1, 413.1},
		{281.5, 414.0},
		{277.1, 419.3},
		{272.8, 414.0},
		{272.1, 413.1},
		{271.8, 412.4},
		{271.6, 411.7},
		{271.5, 411.0},
		{271.6, 410.5},
		{271.8, 409.7},
		{272.2, 409.1},
		{273.0, 408.5},
		{273.8, 408.1},
		{274.5, 408.0},
		{275.2, 408.1},
		{275.9, 408.4},
		{276.4, 408.7},
		{276.9, 409.2},
		{277.1, 409.7},
	}
	points := []Point{
		{277.1, 419.3},
		{272.5, 413.6},
		{271.4, 411.0},
		{273.7, 408.1},
		{277.1, 409.7},
		{280.5, 408.1},
		{282.8, 411.0},
		{281.7, 413.6},
	}

	heart := NewPolygonShapeWithOriginalPoints(&points, &originalPoints)
	heart.centerOnPoint(*NewPoint(0, 0))
	heart.calculateCentroid()
	currentSize := max(heart.getMaxSize())
	heart.scale(1 / currentSize)
	heart.scale(size)
	return heart
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
