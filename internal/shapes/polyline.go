package shapes

import (
	"fmt"

	"github.com/DavidJilg/Vecart/internal/utils"
)

type Polyline struct {
	Points        []Point
	OriginalShape any
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
	return p1.EqualTo(p3, precision) && p2.EqualTo(p4, precision) || p1.EqualTo(p4, precision) && p2.EqualTo(p3, precision)
}

func (line *Polyline) EqualTo(otherPolyline *Polyline, precision int) bool {
	if (line.OriginalShape == nil) != (otherPolyline.OriginalShape == nil) {
		return false
	}

	if line.OriginalShape != nil {
		if fmt.Sprintf("%T", line.OriginalShape) != fmt.Sprintf("%T", otherPolyline.OriginalShape) {
			return false
		}

		switch originalShapeType := line.OriginalShape.(type) {
		default:
			fmt.Println("Invalid Type for original Shape ", originalShapeType)
			panic(10)
		case *Polygon:
			polygon, _ := line.OriginalShape.(*Polygon)
			otherPolygon, _ := otherPolyline.OriginalShape.(*Polygon)
			return polygon.EqualTo(otherPolygon, precision)
		case *Circle:
			circle, _ := line.OriginalShape.(*Circle)
			otherCircle, _ := otherPolyline.OriginalShape.(*Circle)
			return circle.EqualTo(otherCircle, precision)
		}

	}

	if !equalLineSegments(&line.Points, &otherPolyline.Points, precision) {
		return false
	}

	return true
}

func (line *Polyline) Rotate(angle float64, origin Point) {
	for index := range line.Points {
		line.Points[index].Rotate(angle, origin)
	}

	if line.OriginalShape == nil {
		return
	}

	switch originalShapeType := line.OriginalShape.(type) {
	default:
		fmt.Println("Invalid Type for original Shape ", originalShapeType)
		panic(10)
	case *Polygon:
		polygon, _ := line.OriginalShape.(*Polygon)
		polygon.Rotate(angle, origin)
	case *Circle:
		circle, _ := line.OriginalShape.(*Circle)
		circle.Rotate(angle, origin)
	}
}

func (line *Polyline) Move(x float64, y float64) {
	for index := range line.Points {
		line.Points[index].Move(x, y)
	}

	if line.OriginalShape == nil {
		return
	}

	switch originalShapeType := line.OriginalShape.(type) {
	default:
		fmt.Println("Invalid Type for original Shape ", originalShapeType)
		panic(10)
	case *Polygon:
		polygon, _ := line.OriginalShape.(*Polygon)
		polygon.Move(x, y)
	case *Circle:
		circle, _ := line.OriginalShape.(*Circle)
		circle.Move(x, y)
	}
}

func (line *Polyline) Transform(a, b, c, d, e, f float64) {
	for index := range line.Points {
		line.Points[index].Transform(a, b, c, d, e, f)
	}

	if line.OriginalShape == nil {
		return
	}

	switch originalShapeType := line.OriginalShape.(type) {
	default:
		fmt.Println("Invalid Type for original Shape ", originalShapeType)
		panic(10)
	case *Polygon:
		polygon, _ := line.OriginalShape.(*Polygon)
		polygon.Transform(a, b, c, d, e, f)
	case *Circle:
		circle, _ := line.OriginalShape.(*Circle)
		circle.Transform(a, b, c, d, e, f)
	case *Polyline:
		polyline, _ := line.OriginalShape.(*Polyline)
		polyline.Transform(a, b, c, d, e, f)
	}
}

func (line *Polyline) Scale(factor float64, centroid Point) {
	for index := range line.Points {
		line.Points[index].Scale(factor, centroid)
	}

	if line.OriginalShape == nil {
		return
	}

	switch originalShapeType := line.OriginalShape.(type) {
	default:
		fmt.Println("Invalid Type for original Shape ", originalShapeType)
		panic(10)
	case *Polygon:
		polygon, _ := line.OriginalShape.(*Polygon)
		polygon.Scale(factor, centroid)
	case *Circle:
		circle, _ := line.OriginalShape.(*Circle)
		circle.Scale(factor, centroid)
	}
}

func (line *Polyline) MMToPixel(dpi float64) {
	for index := range line.Points {
		line.Points[index].MMToPixel(dpi)
	}

	if line.OriginalShape == nil {
		return
	}

	switch originalShapeType := line.OriginalShape.(type) {
	default:
		fmt.Println("Invalid Type for original Shape ", originalShapeType)
		panic(10)
	case *Polygon:
		polygon, _ := line.OriginalShape.(*Polygon)
		polygon.MMToPixel(dpi)
	case *Circle:
		circle, _ := line.OriginalShape.(*Circle)
		circle.MMToPixel(dpi)
	}
}

func (line *Polyline) PixelToMM(dpi float64) {
	for index := range line.Points {
		line.Points[index].PixelToMM(dpi)
	}

	if line.OriginalShape == nil {
		return
	}

	switch originalShapeType := line.OriginalShape.(type) {
	default:
		fmt.Println("Invalid Type for original Shape ", originalShapeType)
		panic(10)
	case *Polygon:
		polygon, _ := line.OriginalShape.(*Polygon)
		polygon.PixelToMM(dpi)
	case *Circle:
		circle, _ := line.OriginalShape.(*Circle)
		circle.PixelToMM(dpi)
	}
}

func (polyline *Polyline) GetBoundries() (float64, float64, float64, float64) {
	minX := polyline.Points[0].X
	maxX := polyline.Points[0].X
	minY := polyline.Points[0].Y
	maxY := polyline.Points[0].Y
	for _, point := range polyline.Points {
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

func (line *Polyline) Copy() Polyline {
	var copiedLine Polyline
	for _, point := range line.Points {
		copiedLine.Points = append(copiedLine.Points, point.Copy())
	}

	if line.OriginalShape == nil {
		return copiedLine
	}

	switch originalShapeType := line.OriginalShape.(type) {
	default:
		fmt.Println("Invalid Type for original Shape ", originalShapeType)
		panic(10)
	case *Polygon:
		polygon, _ := line.OriginalShape.(*Polygon)
		copiedLine.OriginalShape = polygon.Copy()
	case *Circle:
		circle, _ := line.OriginalShape.(*Circle)
		copiedLine.OriginalShape = circle.Copy()
	}

	return copiedLine
}

func (line *Polyline) GetLineSegments() []Polyline {
	var lineSegments []Polyline
	for pointIndex := 1; pointIndex < len(line.Points); pointIndex++ {
		lineSegments = append(lineSegments, Polyline{[]Point{line.Points[pointIndex-1], line.Points[pointIndex]}, nil})
	}

	return lineSegments
}

func (line *Polyline) ToSVG(style string) string {
	if line.OriginalShape != nil {
		switch originalShapeType := line.OriginalShape.(type) {
		default:
			fmt.Println("Invalid Type for original Shape ", originalShapeType)
			panic(10)
		case *Polygon:
			polygon, _ := line.OriginalShape.(*Polygon)
			return polygon.ToSVG(style)
		case *Circle:
			circle, _ := line.OriginalShape.(*Circle)
			return circle.ToSVG(style)
		}
	}
	svg := "<polyline points=\""

	for index, point := range line.Points {
		svg += point.ToSVGFormat()
		if index != len(line.Points)-1 {
			svg += " "
		}
	}

	svg += "\" style=\""
	svg += style
	svg += "\" />"
	return svg
}

func (line *Polyline) ToJSON() any {
	if line.OriginalShape != nil {
		switch originalShapeType := line.OriginalShape.(type) {
		default:
			fmt.Println("Invalid Type for original Shape ", originalShapeType)
			panic(10)
		case *Polygon:
			polygon, _ := line.OriginalShape.(*Polygon)
			return polygon.ToJSON()
		case *Circle:
			circle, _ := line.OriginalShape.(*Circle)
			return circle.ToJSON()
		case *Polyline:
			polyline, _ := line.OriginalShape.(*Polyline)
			return polyline.ToJSON()
		}
	}

	jsonMap := make(map[string]any)
	points := []any{}
	jsonMap["type"] = "polyline"
	for _, point := range line.Points {
		points = append(points, point.ToJSON())
	}
	jsonMap["points"] = points

	return jsonMap
}

func (line *Polyline) ToShape(bezierTolerance float64) *Shape {
	if line.OriginalShape != nil {
		switch originalShapeType := line.OriginalShape.(type) {
		default:
			fmt.Println("Invalid Type for original Shape ", originalShapeType)
			panic(10)
		case *Polygon:
			polygon, _ := line.OriginalShape.(*Polygon)
			return polygon.ToShape()
		case *Path:
			path, _ := line.OriginalShape.(*Path)
			return path.ToPolyline(bezierTolerance).ToShape(bezierTolerance)
		case *Circle:
			circle, _ := line.OriginalShape.(*Circle)
			return circle.ToShape()
		case *Polyline:
			polyline, _ := line.OriginalShape.(*Polyline)
			return polyline.ToShape(bezierTolerance)
		}
	}

	return NewShape([]Polyline{*line})
}

func (line Polyline) GetStrokeStart(bezierTolerance float64, stampMode bool) Point {
	if line.OriginalShape != nil {
		switch originalShapeType := line.OriginalShape.(type) {
		default:
			fmt.Println("Invalid Type for original Shape ", originalShapeType)
			panic(10)
		case *Polygon:
			polygon, _ := line.OriginalShape.(*Polygon)
			return polygon.GetStrokeStart(bezierTolerance, stampMode)
		case *Circle:
			circle, _ := line.OriginalShape.(*Circle)
			return circle.GetStrokeStart(stampMode)
		case *Polyline:
			polyline, _ := line.OriginalShape.(*Polyline)
			return polyline.GetStrokeStart(bezierTolerance, stampMode)
		}
	}

	if(stampMode) {
		return line.GetMidpoint()
	}

	return line.Points[0].Copy()
}

func (line Polyline) GetMidpoint() Point {
	minX, minY, maxX, maxY := line.GetBoundries()
	return Point{minX + (maxX - minX) / 2, minY + (maxY - minY) / 2}
}

func (line Polyline) HasDrawableSegments() bool {
	if len(line.Points) < 2 {
		return false
	}

	for index := 1; index < len(line.Points); index++ {
		previousPoint := line.Points[index-1]
		currentPoint := line.Points[index]
		if previousPoint.X != currentPoint.X || previousPoint.Y != currentPoint.Y {
			return true
		}
	}

	return false
}

func (line Polyline) ToGCode(activateTool []string, deactivateTool []string, stampMode bool, feedrateXY int) []string {
	if line.OriginalShape != nil {
		switch originalShapeType := line.OriginalShape.(type) {
		default:
			fmt.Println("Invalid Type for original Shape ", originalShapeType)
			panic(10)
		case *Polygon:
			polygon, _ := line.OriginalShape.(*Polygon)
			return polygon.ToGCode(activateTool, deactivateTool, stampMode, feedrateXY)
		case *Circle:
			circle, _ := line.OriginalShape.(*Circle)
			return circle.ToGCode(activateTool, deactivateTool, stampMode, feedrateXY)
		case *Polyline:
			polyline, _ := line.OriginalShape.(*Polyline)
			return polyline.ToGCode(activateTool, deactivateTool, stampMode, feedrateXY)
		}
	}

	if !stampMode && !line.HasDrawableSegments() {
		return nil
	}

	if stampMode {
		lines := make([]string, 0)
		center := line.GetMidpoint()
		lines = append(lines, utils.G00XY(center.X, center.Y))
		lines = append(lines, activateTool...)
		lines = append(lines, deactivateTool...)
		return lines
	}

	lines := make([]string, 0)
	lines = append(lines, utils.G00XY(line.Points[0].X, line.Points[0].Y))
	lines = append(lines, activateTool...)
	for _, point := range line.Points[1:] {
		lines = append(lines, utils.G01XY(point.X, point.Y, feedrateXY))
	}
	lines = append(lines, deactivateTool...)

	return lines
}
