package main

import (
	"fmt"
	"math"
)

type Shape struct {
	Lines    []Polyline
	Variants []Shape
	centroid Point
}

func NewShape(lines []Polyline) *Shape {
	shape := Shape{lines, nil, Point{0, 0}}
	shape.calculateCentroid()

	return &shape
}

func (shape *Shape) equalTo(otherShape *Shape, precision int, ignoreVariants bool) bool {
	if !shape.centroid.equalTo(&otherShape.centroid, precision) {
		return false
	}
	if !polylinesEqual(&shape.Lines, &otherShape.Lines, precision) {
		return false
	}

	return ignoreVariants || shapesEqual(&shape.Variants, &otherShape.Variants, precision, ignoreVariants)
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
			if currentOtherPolyline.equalTo(currentPolyline, precision) {
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
	shape.calculateCentroid()

	return &shape
}

func NewLine(p1, p2 *Point) *Shape {
	return NewSingleLineShape(*NewPolyline(&[]Point{*p1, *p2}, nil))
}

func (shape *Shape) centerOnOrigin() {
	shape.calculateCentroid()
	shape.transform(shape.centroid.X * -1.0, shape.centroid.Y * -1.0)
}

func (shape *Shape) centerOnPoint(point Point) {
	shape.centerOnOrigin()
	shape.transform(point.X, point.Y)
}

func (shape *Shape) centerLeftOnOrigin() {
	minX, _, minY, maxY := shape.getMaxAndMinCoordinates()

	xDiff := minX * -1
	yDiff := ((maxY + minY) / 2) * -1

	if xDiff == 0 && yDiff == 0 {
		return
	}
	shape.transform(xDiff, yDiff)
}

func (shape *Shape) topLeftOnOrigin() {
	minX, _, minY, _ := shape.getMaxAndMinCoordinates()

	xDiff := minX * -1
	yDiff := minY * -1

	if xDiff == 0 && yDiff == 0 {
		return
	}
	shape.transform(xDiff, yDiff)
}

func (shape *Shape) bottomLeftOnOrigin() {
	minX, _, _, maxY := shape.getMaxAndMinCoordinates()

	xDiff := minX * -1
	yDiff := maxY * -1

	if xDiff == 0 && yDiff == 0 {
		return
	}
	shape.transform(xDiff, yDiff)
}

func (shape *Shape) calculateCentroid() {
	minX, maxX, minY, maxY := shape.getMaxAndMinCoordinates()
	shape.centroid = Point{(minX + maxX)/2, (minY + maxY)/2}
}

func (shape *Shape) ensureOriginCover() {
	if shape.originCovered() {
		return
	}
	if Config.debug {
		fmt.Println("Shape is not covering origin (0,0). Trying to move it ...")
	}

	size := int(math.Ceil(math.Max(shape.getSize())))

	for i := 0; i <= size; i++ {
		shapeCopy := shape.transformCopy(float64(i), 0)
		if shapeCopy.originCovered() {
			shape.Lines = shapeCopy.Lines
			shape.centroid = shapeCopy.centroid
			if Config.debug {
				fmt.Println("Moving shape to cover origin sucessful.")
			}
			return
		}

		shapeCopy2 := shape.transformCopy(0, float64(i))
		if shapeCopy2.originCovered() {
			shape.Lines = shapeCopy2.Lines
			shape.centroid = shapeCopy2.centroid
			if Config.debug {
				fmt.Println("Moving shape to cover origin sucessful.")
			}
			return
		}

		shapeCopy3 := shape.transformCopy(float64(0-i), 0)
		if shapeCopy3.originCovered() {
			shape.Lines = shapeCopy3.Lines
			shape.centroid = shapeCopy3.centroid
			if Config.debug {
				fmt.Println("Moving shape to cover origin sucessful.")
			}
			return
		}

		shapeCopy4 := shape.transformCopy(0, float64(0-i))
		if shapeCopy4.originCovered() {
			shape.Lines = shapeCopy4.Lines
			shape.centroid = shapeCopy4.centroid
			if Config.debug {
				fmt.Println("Moving shape to cover origin sucessful.")
			}
			return
		}

	}

	line := shape.Lines[0]
	p1 := line.points[0]
	p2 := line.points[1]

	x := 0 - ((p1.X + p2.X) / 2)
	y := 0 - ((p1.Y + p2.Y) / 2)

	shape.transform(x, y)

	if shape.originCovered() {
		if Config.debug {
			fmt.Println("Moving shape to cover origin sucessful.")
		}
	} else {
		if Config.debug {
			fmt.Println("Could not ensure origin cover for shape.")
		}
	}

	for index, _ := range shape.Variants{
		shape.Variants[index].ensureOriginCover()
	}
}

func (shape *Shape) originCovered() bool {
	originRectangle := NewPolyline(&[]Point{{-0.5, -0.5}, {0.5, -0.5}, {0.5, 0.5}, {-0.5, 0.5}}, nil)

	for _, line := range shape.Lines {
		if intersectingPolylines(&line, originRectangle) {
			return true
		}
	}

	return false
}

func (shape *Shape) getSize() (float64, float64) {
	minX, maxX, minY, maxY := shape.getMaxAndMinCoordinates()
	return maxX - minX, maxY - minY
}

func (shape *Shape) getMaxSize() (float64, float64) {
	maxX, maxY := math.MaxFloat64*-1, math.MaxFloat64*-1
	for i := 0.0; i < 360; i += 5 {
		x, y := shape.rotateCopy(i, shape.centroid).getSize()
		if x > maxX {
			maxX = y
		}
		if y > maxY {
			maxX = y
		}
	}

	return maxX, maxY
}

func (shape *Shape) getMaxAndMinCoordinates() (float64, float64, float64, float64) {
	minX := math.MaxFloat64
	maxX := math.MaxFloat64 * -1
	minY := math.MaxFloat64
	maxY := math.MaxFloat64 * -1

	for lineIndex := range shape.Lines {
		for pointIndex := range shape.Lines[lineIndex].points {
			currentPoint := &shape.Lines[lineIndex].points[pointIndex]
			minX = math.Min(currentPoint.X, minX)
			minY = math.Min(currentPoint.Y, minY)
			maxX = math.Max(currentPoint.X, maxX)
			maxY = math.Max(currentPoint.Y, maxY)
		}
	}

	return minX, maxX, minY, maxY
}

func (shape *Shape) generateShapeVariants() {
	//Todo do not rotate circles
	//Test

	if Config.shapeAngleDeviationRange <= 0 {
		shape.Variants = append(shape.Variants, shape.copy())
		return
	}

	currentAngle := 0.0
	for currentAngle <= Config.shapeAngleDeviationRange {
		shape.getVariant(currentAngle)
		currentAngle += Config.shapeAngleDeviationStep
	}
	currentAngle = (0 - Config.shapeAngleDeviationStep)
	for currentAngle >= (0 - Config.shapeAngleDeviationRange) {
		shape.getVariant(currentAngle)
		currentAngle -= Config.shapeAngleDeviationStep
	}
}

func (shape *Shape) getVariant(angle float64) {
	currentVariant := shape.rotateCopy(angle, shape.centroid)
	currentVariant.centerOnOrigin()
	for shapeVariantIndex := range shape.Variants {
		if shape.Variants[shapeVariantIndex].equalTo(currentVariant, 5, true) {
			return
		}
	}

	shape.Variants = append(shape.Variants, *currentVariant)
}

func (shape *Shape) isSingleCircle() bool {
	if len(shape.Lines) != 1 {
		return false
	}

	if shape.Lines[0].originalShape == nil {
		return false
	}

	switch shape.Lines[0].originalShape.(type) {
	default:
		return false
	case Circle:
		return true
	}
}

func (shape *Shape) rotate(angle float64, origin Point) {
	for index := range shape.Lines {
		shape.Lines[index].rotate(angle, origin)
	}
}

func (shape *Shape) rotateCopy(angle float64, origin Point) *Shape {
	copiedShape := shape.copy()
	copiedShape.rotate(angle, origin)

	return &copiedShape
}

func (shape *Shape) transform(x float64, y float64) {
	for index := range shape.Lines {
		shape.Lines[index].transform(x, y)
	}
	shape.calculateCentroid()
}

func (shape *Shape) scale(factor float64) {
	shape.calculateCentroid()

	for index := range shape.Lines {
		shape.Lines[index].scale(factor, shape.centroid)
	}

	shape.calculateCentroid()
}

func (shape *Shape) mmToPixel(dpi float64) {
	for index := range shape.Lines {
		shape.Lines[index].mmToPixel(dpi)
	}
	shape.calculateCentroid()
}

func (shape *Shape) pixelToMM(dpi float64) {
	for index := range shape.Lines {
		shape.Lines[index].pixelToMM(dpi)
	}
	shape.calculateCentroid()
}

func (shape *Shape) scaleCopy(factor float64) *Shape {
	copiedShape := shape.copy()
	copiedShape.scale(factor)

	return &copiedShape
}

func (shape *Shape) transformCopy(x, y float64) *Shape {
	copiedShape := shape.copy()
	copiedShape.transform(x, y)

	return &copiedShape
}

func (shape *Shape) copy() Shape {
	var copiedShape Shape
	for _, line := range shape.Lines {
		copiedShape.Lines = append(copiedShape.Lines, line.copy())
	}

	return copiedShape
}

func (shape *Shape) toSVG(style string) string {
	svg := ""

	if len(shape.Lines) == 0 {
		panic(66)
	}

	for index := range shape.Lines {
		svg += shape.Lines[index].toSVG(style)
		if index != len(shape.Lines)-1 {
			svg += "\n"
		}
	}
	return svg
}

func (shape *Shape) toJSON() any {
	if len(shape.Lines) == 0 {
		return []any{}
	}

	if len(shape.Lines) == 1 {
		return shape.Lines[0].toJSON()
	}

	var polylines []any
	for _, polyline := range shape.Lines {
		polylines = append(polylines, polyline.toJSON())
	}

	return polylines
}
