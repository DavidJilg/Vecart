package shapes

import (
	"log"
	"math"

	"github.com/DavidJilg/Vecart/internal/utils"
)

type Shape struct {
	Lines    []Polyline
	Variants []Shape
	centroid Point
}

func CombineShapes(shapeList *[]Shape) *Shape {
	var lines []Polyline

	for _, shape := range *shapeList {
		lines = append(lines, shape.Lines...)
	}

	return NewShape(lines)
}

func NewShape(lines []Polyline) *Shape {
	shape := Shape{lines, nil, Point{0, 0}}
	shape.CalculateCentroid()

	return &shape
}

func (shape *Shape) EqualTo(otherShape *Shape, precision int, ignoreVariants bool) bool {
	if !shape.centroid.EqualTo(&otherShape.centroid, precision) {
		return false
	}
	if !polylinesEqual(&shape.Lines, &otherShape.Lines, precision) {
		return false
	}

	return ignoreVariants || ShapesEqual(&shape.Variants, &otherShape.Variants, precision, ignoreVariants)
}

func ShapesEqual(shapes *[]Shape, otherShapes *[]Shape, precision int, ignoreVariants bool) bool {
	if shapes == nil && otherShapes == nil {
		return true
	}

	if !(shapes != nil && otherShapes != nil) {
		return false
	}

	if len(*shapes) != len(*otherShapes) {
		return false
	}

	var shapePointers, otherShapePointers []*Shape

	for index := range *shapes {
		shapePointers = append(shapePointers, &((*shapes)[index]))
		otherShapePointers = append(otherShapePointers, &((*otherShapes)[index]))
	}

	for i := 0; i < len(shapePointers); i++ {
		currentShape := shapePointers[i]
		foundEqualShape := false
		for j := 0; j < len(otherShapePointers); j++ {
			currentOtherShape := otherShapePointers[j]
			if currentOtherShape.EqualTo(currentShape, precision, ignoreVariants) {
				foundEqualShape = true
				RemoveShape(otherShapePointers, j)
				break
			}
		}

		if !foundEqualShape {
			return false
		}
	}

	return true
}

func RemoveShape(s []*Shape, index int) []*Shape {
	return append(s[:index], s[index+1:]...)
}

func polylinesEqual(polylines *[]Polyline, otherPolylines *[]Polyline, precision int) bool {
	if polylines == nil && otherPolylines == nil {
		return true
	}

	if !(polylines != nil && otherPolylines != nil) {
		return false
	}

	if len(*polylines) != len(*otherPolylines) {
		return false
	}

	var polylinePointers, otherPolylinePointers []*Polyline

	for index := range *polylines {
		polylinePointers = append(polylinePointers, &(*polylines)[index])
		otherPolylinePointers = append(otherPolylinePointers, &(*otherPolylines)[index])
	}

	for i := 0; i < len(polylinePointers); i++ {
		currentPolyline := polylinePointers[i]
		foundEqualShape := false
		for j := 0; j < len(otherPolylinePointers); j++ {
			currentOtherPolyline := otherPolylinePointers[j]
			if currentOtherPolyline.EqualTo(currentPolyline, precision) {
				foundEqualShape = true
				removePolyline(otherPolylinePointers, j)
				break
			}
		}

		if !foundEqualShape {
			return false
		}
	}

	return true
}

func removePolyline(s []*Polyline, index int) []*Polyline {
	return append(s[:index], s[index+1:]...)
}

func NewSingleLineShape(line Polyline) *Shape {
	shape := Shape{[]Polyline{line}, nil, Point{0, 0}}
	shape.CalculateCentroid()

	return &shape
}

func NewLine(p1, p2 *Point) *Shape {
	return NewSingleLineShape(*NewPolyline(&[]Point{*p1, *p2}, nil))
}

func (shape *Shape) CenterOnOrigin() {
	shape.CalculateCentroid()
	shape.Move(shape.centroid.X*-1.0, shape.centroid.Y*-1.0)
}

func (shape *Shape) CenterOnPoint(point Point) {
	shape.CenterOnOrigin()
	shape.Move(point.X, point.Y)
}

func (shape *Shape) CenterLeftOnOrigin() {
	minX, _, minY, maxY := shape.GetMaxAndMinCoordinates()

	xDiff := minX * -1
	yDiff := ((maxY + minY) / 2) * -1

	if xDiff == 0 && yDiff == 0 {
		return
	}
	shape.Move(xDiff, yDiff)
}

func (shape *Shape) TopLeftOnOrigin() {
	minX, _, minY, _ := shape.GetMaxAndMinCoordinates()

	xDiff := minX * -1
	yDiff := minY * -1

	if xDiff == 0 && yDiff == 0 {
		return
	}
	shape.Move(xDiff, yDiff)
}

func (shape *Shape) BottomLeftOnOrigin() {
	minX, _, _, maxY := shape.GetMaxAndMinCoordinates()

	xDiff := minX * -1
	yDiff := maxY * -1

	if xDiff == 0 && yDiff == 0 {
		return
	}
	shape.Move(xDiff, yDiff)
}

func (shape *Shape) CalculateCentroid() {
	minX, maxX, minY, maxY := shape.GetMaxAndMinCoordinates()
	shape.centroid = Point{minX + (maxX-minX)/2, minY + (maxY-minY)/2}
}

func (shape *Shape) EnsureOriginCover() {
	if shape.OriginCovered() {
		return
	}
	if utils.DebugModeEnabled() {
		log.Println("Shape is not covering origin (0,0). Trying to move it ...")
	}

	size := int(math.Ceil(math.Max(shape.GetSize())))

	for i := 0; i <= size; i++ {
		shapeCopy := shape.TransformCopy(float64(i), 0)
		if shapeCopy.OriginCovered() {
			shape.Lines = shapeCopy.Lines
			shape.centroid = shapeCopy.centroid
			if utils.DebugModeEnabled() {
				log.Println("Moving shape to cover origin sucessful.")
			}
			return
		}

		shapeCopy2 := shape.TransformCopy(0, float64(i))
		if shapeCopy2.OriginCovered() {
			shape.Lines = shapeCopy2.Lines
			shape.centroid = shapeCopy2.centroid
			if utils.DebugModeEnabled() {
				log.Println("Moving shape to cover origin sucessful.")
			}
			return
		}

		shapeCopy3 := shape.TransformCopy(float64(0-i), 0)
		if shapeCopy3.OriginCovered() {
			shape.Lines = shapeCopy3.Lines
			shape.centroid = shapeCopy3.centroid
			if utils.DebugModeEnabled() {
				log.Println("Moving shape to cover origin sucessful.")
			}
			return
		}

		shapeCopy4 := shape.TransformCopy(0, float64(0-i))
		if shapeCopy4.OriginCovered() {
			shape.Lines = shapeCopy4.Lines
			shape.centroid = shapeCopy4.centroid
			if utils.DebugModeEnabled() {
				log.Println("Moving shape to cover origin sucessful.")
			}
			return
		}

	}

	line := shape.Lines[0]
	p1 := line.Points[0]
	p2 := line.Points[1]

	x := 0 - ((p1.X + p2.X) / 2)
	y := 0 - ((p1.Y + p2.Y) / 2)

	shape.Move(x, y)

	if shape.OriginCovered() {
		if utils.DebugModeEnabled() {
			log.Println("Moving shape to cover origin sucessful.")
		}
	} else {
		if utils.DebugModeEnabled() {
			log.Println("Could not ensure origin cover for shape.")
		}
	}

	for index := range shape.Variants {
		shape.Variants[index].EnsureOriginCover()
	}
}

func (shape *Shape) OriginCovered() bool {
	originRectangle := NewPolyline(&[]Point{{-0.5, -0.5}, {0.5, -0.5}, {0.5, 0.5}, {-0.5, 0.5}}, nil)

	for _, line := range shape.Lines {
		if IntersectingPolylines(&line, originRectangle) {
			return true
		}
	}

	return false
}

func (shape *Shape) Intersects(line *Polyline) bool {
	for _, shapeLine := range shape.Lines {
		if IntersectingPolylines(&shapeLine, line) {
			return true
		}
	}

	return false
}

func (shape *Shape) GetSize() (float64, float64) {
	minX, maxX, minY, maxY := shape.GetMaxAndMinCoordinates()
	return maxX - minX, maxY - minY
}

func (shape *Shape) GetMaxSize() (float64, float64) {
	maxX, maxY := math.MaxFloat64*-1, math.MaxFloat64*-1
	for i := 0.0; i < 360; i += 5 {
		x, y := shape.RotateCopy(i, shape.centroid).GetSize()
		if x > maxX {
			maxX = y
		}
		if y > maxY {
			maxX = y
		}
	}

	return maxX, maxY
}

func (shape *Shape) GetBoundries() (float64, float64, float64, float64) {
	minX, maxX, minY, maxY := shape.GetMaxAndMinCoordinates()
	return minX, minY, maxX, maxY
}

func (shape *Shape) GetMaxAndMinCoordinates() (float64, float64, float64, float64) {
	minX := math.MaxFloat64
	maxX := math.MaxFloat64 * -1
	minY := math.MaxFloat64
	maxY := math.MaxFloat64 * -1

	for lineIndex := range shape.Lines {
		for pointIndex := range shape.Lines[lineIndex].Points {
			currentPoint := &shape.Lines[lineIndex].Points[pointIndex]
			minX = math.Min(currentPoint.X, minX)
			minY = math.Min(currentPoint.Y, minY)
			maxX = math.Max(currentPoint.X, maxX)
			maxY = math.Max(currentPoint.Y, maxY)
		}
	}

	return minX, maxX, minY, maxY
}

func (shape *Shape) GetVariant(angle float64) {
	currentVariant := shape.RotateCopy(angle, shape.centroid)
	currentVariant.CenterOnOrigin()
	for shapeVariantIndex := range shape.Variants {
		if shape.Variants[shapeVariantIndex].EqualTo(currentVariant, 5, true) {
			return
		}
	}

	shape.Variants = append(shape.Variants, *currentVariant)
}

func (shape *Shape) IsSingleCircle() bool {
	if len(shape.Lines) != 1 {
		return false
	}

	if shape.Lines[0].OriginalShape == nil {
		return false
	}

	switch shape.Lines[0].OriginalShape.(type) {
	default:
		return false
	case Circle:
		return true
	}
}

func (shape *Shape) Rotate(angle float64, origin Point) {
	for index := range shape.Lines {
		shape.Lines[index].Rotate(angle, origin)
	}
}

func (shape *Shape) RotateCopy(angle float64, origin Point) *Shape {
	copiedShape := shape.Copy()
	copiedShape.Rotate(angle, origin)

	return &copiedShape
}

func (shape *Shape) Move(x float64, y float64) {
	for index := range shape.Lines {
		shape.Lines[index].Move(x, y)
	}
	shape.CalculateCentroid()
}

func (shape *Shape) Transform(a, b, c, d, e, f float64) {
	for index := range shape.Lines {
		shape.Lines[index].Transform(a, b, c, d, e, f)
	}
	shape.CalculateCentroid()
}

func (shape *Shape) Scale(factor float64) {
	shape.CalculateCentroid()

	for index := range shape.Lines {
		shape.Lines[index].Scale(factor, shape.centroid)
	}

	shape.CalculateCentroid()
}

func (shape *Shape) ScaleToSize(size float64) {
	shape.CenterOnPoint(*NewPoint(0, 0))
	shape.CalculateCentroid()
	currentSize := max(shape.GetMaxSize())
	shape.Scale(1 / currentSize)
	shape.Scale(size)
}

func (shape *Shape) MMToPixel(dpi float64) {
	for index := range shape.Lines {
		shape.Lines[index].MMToPixel(dpi)
	}
	shape.CalculateCentroid()
}

func (shape *Shape) PixelToMM(dpi float64) {
	for index := range shape.Lines {
		shape.Lines[index].PixelToMM(dpi)
	}
	shape.CalculateCentroid()
}

func (shape *Shape) ScaleCopy(factor float64) *Shape {
	copiedShape := shape.Copy()
	copiedShape.Scale(factor)

	return &copiedShape
}

func (shape *Shape) TransformCopy(x, y float64) *Shape {
	copiedShape := shape.Copy()
	copiedShape.Move(x, y)

	return &copiedShape
}

func (shape *Shape) Copy() Shape {
	var copiedShape Shape
	for _, line := range shape.Lines {
		copiedShape.Lines = append(copiedShape.Lines, line.Copy())
	}

	return copiedShape
}

func (shape *Shape) ToSVG(style string) string {
	svg := ""

	if len(shape.Lines) == 0 {
		panic(66)
	}

	for index := range shape.Lines {
		svg += shape.Lines[index].ToSVG(style)
		if index != len(shape.Lines)-1 {
			svg += "\n"
		}
	}
	return svg
}

func (shape *Shape) ToJSON() any {
	if len(shape.Lines) == 0 {
		return []any{}
	}

	if len(shape.Lines) == 1 {
		return shape.Lines[0].ToJSON()
	}

	var polylines []any
	for _, polyline := range shape.Lines {
		polylines = append(polylines, polyline.ToJSON())
	}

	return polylines
}

func (shape Shape) GetStrokeStart(bezierTolerance float64, stampMode bool) Point {
	return shape.Lines[0].GetStrokeStart(bezierTolerance, stampMode)
}

func (shape Shape) GetMidpoint() Point {
	minX, minY, maxX, maxY := shape.GetBoundries()
	return Point{minX + (maxX-minX)/2, minY + (maxY-minY)/2}
}

func (shape Shape) ToGCode(activateTool []string, deactivateTool []string, stampMode bool, feedRateXY int) []string {
	lines := make([]string, 0)

	if stampMode {
		center := shape.GetMidpoint()
		lines = append(lines, utils.G00XY(center.X, center.Y))
		lines = append(lines, activateTool...)
		lines = append(lines, deactivateTool...)
		return lines
	}

	for _, line := range shape.Lines {
		lines = append(lines, line.ToGCode(activateTool, deactivateTool, false, feedRateXY)...)
	}

	return lines
}
