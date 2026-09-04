package shapes

type Polygon struct {
	Points []Point
}

func NewPolygon(points *[]Point) *Polygon {
	return &Polygon{*points}
}

func NewPolygonShapeWithOriginalPoints(points *[]Point, originalPoints *[]Point) *Shape {
	line := NewPolygon(points).ToPolyline()
	line.OriginalShape = NewPolygon(originalPoints)
	shape := Shape{[]Polyline{*line}, nil, Point{0, 0}}
	shape.CalculateCentroid()

	return &shape
}

func NewRect(x float64, y float64, width float64, height float64) *Polygon {
	points := make([]Point, 4)
	points[0] = Point{X: x, Y: y}
	points[1] = Point{X: x + width, Y: y}
	points[2] = Point{X: x + width, Y: y + height}
	points[3] = Point{X: x, Y: y + height}
	return &Polygon{points}
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
	heart.CenterOnPoint(*NewPoint(0, 0))
	heart.CalculateCentroid()
	currentSize := max(heart.GetMaxSize())
	heart.Scale(1 / currentSize)
	heart.Scale(size)
	return heart
}

func (polygon *Polygon) EqualTo(otherPolygon *Polygon, precision int) bool {
	pointsA := append(polygon.Points, polygon.Points[0])
	pointsB := append(otherPolygon.Points, otherPolygon.Points[0])
	return equalLineSegments(&pointsA, &pointsB, precision)
}

func (polygon *Polygon) ToShape() *Shape {
	line := polygon.ToPolyline()
	shape := Shape{[]Polyline{*line}, nil, Point{0, 0}}
	shape.CalculateCentroid()

	return &shape
}

func (polygon *Polygon) ToPolyline() *Polyline {
	var fullPoints []Point
	for index := range polygon.Points {
		fullPoints = append(fullPoints, polygon.Points[index].Copy())
	}
	fullPoints = append(fullPoints, polygon.Points[0])

	return NewPolyline(&fullPoints, polygon)
}

func (polygon *Polygon) Rotate(angle float64, origin Point) {
	for index := range polygon.Points {
		polygon.Points[index].Rotate(angle, origin)
	}
}

func (polygon *Polygon) Move(x float64, y float64) {
	for index := range polygon.Points {
		polygon.Points[index].Move(x, y)
	}
}

func (polygon *Polygon) Transform(a, b, c, d, e, f float64) {
	for index := range polygon.Points {
		polygon.Points[index].Transform(a, b, c, d, e, f)
	}
}

func (polygon *Polygon) Scale(factor float64, centroid Point) {
	for index := range polygon.Points {
		polygon.Points[index].Scale(factor, centroid)
	}
}

func (polygon *Polygon) MMToPixel(dpi float64) {
	for index := range polygon.Points {
		polygon.Points[index].MMToPixel(dpi)
	}
}

func (polygon *Polygon) PixelToMM(dpi float64) {
	for index := range polygon.Points {
		polygon.Points[index].PixelToMM(dpi)
	}
}

func (polygon *Polygon) GetBoundries() (float64, float64, float64, float64) {
	minX := polygon.Points[0].X
	maxX := polygon.Points[0].X
	minY := polygon.Points[0].Y
	maxY := polygon.Points[0].Y
	for _, point := range polygon.Points {
		if point.X < minX {
			minX = point.X
		}
		if point.X > maxX {
			maxX = point.X
		}
		if point.Y < minY {
			minY = point.Y
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}
	return minX, minY, maxX, maxY
}

func (polygon *Polygon) Copy() *Polygon {
	var points []Point
	for _, point := range polygon.Points {
		points = append(points, point.Copy())
	}
	return &Polygon{points}
}

func (polygon *Polygon) ToSVG(style string) string {
	svg := "<polygon  points=\""

	for index, point := range polygon.Points {
		svg += point.ToSVGFormat()
		if index != len(polygon.Points)-1 {
			svg += " "
		}
	}

	svg += "\" style=\""
	svg += style
	svg += "\" />"
	return svg
}

func (polygon *Polygon) ToJSON() map[string]any {
	jsonMap := make(map[string]any)

	jsonMap["type"] = "polygon"

	points := []any{}
	for _, point := range polygon.Points {
		points = append(points, point.ToJSON())
	}
	jsonMap["points"] = points

	return jsonMap
}

func (polygon Polygon) GetStrokeStart(bezierTolerance float64, stampMode bool) Point {
	polyline := polygon.ToPolyline()
	polyline.OriginalShape = nil
	return polyline.GetStrokeStart(bezierTolerance, stampMode)
}

func (polygon Polygon) ToGCode(activateTool []string, deactivateTool []string, stampMode bool, feedRateXY int) []string {

	polyline := polygon.ToPolyline()
	polyline.OriginalShape = nil
	return polyline.ToGCode(activateTool, deactivateTool, stampMode, feedRateXY)
}
