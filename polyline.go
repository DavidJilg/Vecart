package main

import "fmt"

type Polyline struct {
	points        []Point
	originalShape any
}

func NewPolyline(points *[]Point, originalShape any) *Polyline {
	switch originalShapeType := originalShape.(type) {
	default:
		fmt.Println("Invalid Type for original Shape ", originalShapeType)
		panic(10)
	case nil:
		return &Polyline{*points, nil}
	case Polygon:
		polygon, _ := originalShape.(Polygon)
		return &Polyline{*points, &polygon}
	case Circle:
		circle, _ := originalShape.(Circle)
		return &Polyline{*points, &circle}
	case *Polygon:
		polygon, _ := originalShape.(*Polygon)
		return &Polyline{*points, polygon}
	case *Circle:
		circle, _ := originalShape.(*Circle)
		return &Polyline{*points, circle}
	}

}

func equalLineSegments(a *[]Point, b *[]Point, precision int) bool {
	if len(*a) != len(*b) {
		return false
	}

	var lineSegmentsA [][]Point
	var lineSegmentsB [][]Point

	for i := 1; i < len(*a); i++ {
		lineSegmentsA = append(lineSegmentsA, []Point{(*a)[i-1], (*a)[i]})
		lineSegmentsB = append(lineSegmentsB, []Point{(*b)[i-1], (*b)[i]})
	}

	for i := 0; i < len(lineSegmentsA); i++ {
		currentSegment := lineSegmentsA[i]
		foundEqualSegment := false
		for j := 0; j < len(lineSegmentsB); j++ {
			currentOtherSegment := lineSegmentsB[j]

			if equalLineSegment(&currentSegment[0], &currentOtherSegment[1], &currentOtherSegment[0], &currentOtherSegment[1], precision) {
				foundEqualSegment = true
				lineSegmentsB = append(lineSegmentsB[:j], lineSegmentsB[j+1:]...)
				break
			}
		}

		if !foundEqualSegment {
			return false
		}
	}

	return true
}

func equalLineSegment(p1, p2, p3, p4 *Point, precision int) bool {
	return p1.equalTo(p3, precision) && p2.equalTo(p4, precision) || p1.equalTo(p4, precision) && p2.equalTo(p3, precision)
}

func (line *Polyline) equalTo(otherPolyline *Polyline, precision int) bool {
	if (line.originalShape == nil) != (otherPolyline.originalShape == nil) {
		return false
	}

	if line.originalShape != nil {
		if fmt.Sprintf("%T", line.originalShape) != fmt.Sprintf("%T", otherPolyline.originalShape) {
			return false
		}

		switch originalShapeType := line.originalShape.(type) {
		default:
			fmt.Println("Invalid Type for original Shape ", originalShapeType)
			panic(10)
		case *Polygon:
			polygon, _ := line.originalShape.(*Polygon)
			otherPolygon, _ := otherPolyline.originalShape.(*Polygon)
			return polygon.equalTo(otherPolygon, precision)
		case *Circle:
			circle, _ := line.originalShape.(*Circle)
			otherCircle, _ := otherPolyline.originalShape.(*Circle)
			return circle.equalTo(otherCircle, precision)
		}

	}

	if !equalLineSegments(&line.points, &otherPolyline.points, precision) {
		return false
	}

	return true
}

func (line *Polyline) rotate(angle float64, origin Point) {
	for index := range line.points {
		line.points[index].rotate(angle, origin)
	}

	if line.originalShape == nil {
		return
	}

	switch originalShapeType := line.originalShape.(type) {
	default:
		fmt.Println("Invalid Type for original Shape ", originalShapeType)
		panic(10)
	case *Polygon:
		polygon, _ := line.originalShape.(*Polygon)
		polygon.rotate(angle, origin)
	case *Circle:
		circle, _ := line.originalShape.(*Circle)
		circle.rotate(angle, origin)
	}
}

// Todo rename to move
func (line *Polyline) transform(x float64, y float64) {
	for index := range line.points {
		line.points[index].transform(x, y)
	}

	if line.originalShape == nil {
		return
	}

	switch originalShapeType := line.originalShape.(type) {
	default:
		fmt.Println("Invalid Type for original Shape ", originalShapeType)
		panic(10)
	case *Polygon:
		polygon, _ := line.originalShape.(*Polygon)
		polygon.transform(x, y)
	case *Circle:
		circle, _ := line.originalShape.(*Circle)
		circle.transform(x, y)
	}
}

func (line *Polyline) scale(factor float64, centroid Point) {
	for index := range line.points {
		line.points[index].scale(factor, centroid)
	}

	if line.originalShape == nil {
		return
	}

	switch originalShapeType := line.originalShape.(type) {
	default:
		fmt.Println("Invalid Type for original Shape ", originalShapeType)
		panic(10)
	case *Polygon:
		polygon, _ := line.originalShape.(*Polygon)
		polygon.scale(factor, centroid)
	case *Circle:
		circle, _ := line.originalShape.(*Circle)
		circle.scale(factor, centroid)
	}
}

func (line *Polyline) mmToPixel(dpi float64) {
	for index := range line.points {
		line.points[index].mmToPixel(dpi)
	}

	if line.originalShape == nil {
		return
	}

	switch originalShapeType := line.originalShape.(type) {
	default:
		fmt.Println("Invalid Type for original Shape ", originalShapeType)
		panic(10)
	case *Polygon:
		polygon, _ := line.originalShape.(*Polygon)
		polygon.mmToPixel(dpi)
	case *Circle:
		circle, _ := line.originalShape.(*Circle)
		circle.mmToPixel(dpi)
	}
}

func (line *Polyline) pixelToMM(dpi float64) {
	for index := range line.points {
		line.points[index].pixelToMM(dpi)
	}

	if line.originalShape == nil {
		return
	}

	switch originalShapeType := line.originalShape.(type) {
	default:
		fmt.Println("Invalid Type for original Shape ", originalShapeType)
		panic(10)
	case *Polygon:
		polygon, _ := line.originalShape.(*Polygon)
		polygon.pixelToMM(dpi)
	case *Circle:
		circle, _ := line.originalShape.(*Circle)
		circle.pixelToMM(dpi)
	}
}

func (line *Polyline) copy() Polyline {
	var copiedLine Polyline
	for _, point := range line.points {
		copiedLine.points = append(copiedLine.points, point.copy())
	}

	if line.originalShape == nil {
		return copiedLine
	}

	switch originalShapeType := line.originalShape.(type) {
	default:
		fmt.Println("Invalid Type for original Shape ", originalShapeType)
		panic(10)
	case *Polygon:
		polygon, _ := line.originalShape.(*Polygon)
		copiedLine.originalShape = polygon.copy()
	case *Circle:
		circle, _ := line.originalShape.(*Circle)
		copiedLine.originalShape = circle.copy()
	}

	return copiedLine
}

func (line *Polyline) getLineSegments() []Polyline {
	var lineSegments []Polyline
	for pointIndex := 1; pointIndex < len(line.points); pointIndex++ {
		lineSegments = append(lineSegments, Polyline{[]Point{line.points[pointIndex-1], line.points[pointIndex]}, nil})
	}

	return lineSegments
}

func (line *Polyline) toSVG(style string) string {
	if line.originalShape != nil {
		switch originalShapeType := line.originalShape.(type) {
		default:
			fmt.Println("Invalid Type for original Shape ", originalShapeType)
			panic(10)
		case *Polygon:
			polygon, _ := line.originalShape.(*Polygon)
			return polygon.toSVG(style)
		case *Circle:
			circle, _ := line.originalShape.(*Circle)
			return circle.toSVG(style)
		}
	}
	svg := "<polyline points=\""

	for index, point := range line.points {
		svg += point.toSVG()
		if index != len(line.points)-1 {
			svg += " "
		}
	}

	svg += "\" style=\""
	svg += style
	svg += "\" />"
	return svg
}

func (line *Polyline) toJSON() any {
	if line.originalShape != nil {
		switch originalShapeType := line.originalShape.(type) {
		default:
			fmt.Println("Invalid Type for original Shape ", originalShapeType)
			panic(10)
		case *Polygon:
			polygon, _ := line.originalShape.(*Polygon)
			return polygon.toJSON()
		case *Circle:
			circle, _ := line.originalShape.(*Circle)
			return circle.toJSON()
		}
	}

	jsonMap := make(map[string]any)
	points := []any{}
	jsonMap["type"] = "polyline"
	for _, point := range line.points {
		points = append(points, point.toJSON())
	}
	jsonMap["points"] = points

	return jsonMap
}
