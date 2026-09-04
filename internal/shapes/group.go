package shapes

import (
	"log"
	"reflect"

	"github.com/DavidJilg/Vecart/internal/utils"
)

type Group struct {
	Shapes []any
}

func NewGroup(shapes []any) *Group {
	return &Group{Shapes: shapes}
}

func (group *Group) PixelToMM(dpi float64) {
	for index := range group.Shapes {
		switch currentShape := group.Shapes[index].(type) {
		case *Circle:
			currentShape.PixelToMM(dpi)
		case *Group:
			currentShape.PixelToMM(dpi)
		case *Path:
			currentShape.PixelToMM(dpi)
		case *Point:
			currentShape.PixelToMM(dpi)
		case *Polygon:
			currentShape.PixelToMM(dpi)
		case *Polyline:
			currentShape.PixelToMM(dpi)
		case *Shape:
			currentShape.PixelToMM(dpi)
		default:
			if utils.DebugModeEnabled() {
				log.Println("Invalid group element: ", reflect.TypeOf(currentShape))
			}
		}
	}
}

func (group *Group) GetBoundries(bezierTolerance float64) (float64, float64, float64, float64) {
	if len(group.Shapes) == 0 {
		return 0, 0, 0, 0
	}

	minX, minY, maxX, maxY := groupShapeBoundries(group.Shapes[0], bezierTolerance)
	for index := 1; index < len(group.Shapes); index++ {
		currentMinX, currentMinY, currentMaxX, currentMaxY := groupShapeBoundries(group.Shapes[index], bezierTolerance)
		if currentMinX < minX {
			minX = currentMinX
		}
		if currentMinY < minY {
			minY = currentMinY
		}
		if currentMaxX > maxX {
			maxX = currentMaxX
		}
		if currentMaxY > maxY {
			maxY = currentMaxY
		}
	}

	return minX, minY, maxX, maxY
}

func (group *Group) GetStrokeStart(bezierTolerance float64, stampMode bool) Point {
	if len(group.Shapes) == 0 {
		return Point{}
	}

	return groupShapeStrokeStart(group.Shapes[0], bezierTolerance, stampMode)
}

func (group *Group) GetMidpoint(bezierTolerance float64) Point {
	minX, minY, maxX, maxY := group.GetBoundries(bezierTolerance)
	return Point{
		X: minX + (maxX - minX) / 2,
		Y: minY + (maxY - minY) / 2,
	}
}

func anyListToShapeList(shapeList []any, bezierTolerance float64) *[]Shape {
	var shapes []Shape
	for _, shape := range shapeList {
		shapes = append(shapes, *anyToShape(shape, bezierTolerance))
	}
	return &shapes
}

func anyToShape(shape any, bezierTolerance float64) *Shape {
	switch shapeType := shape.(type) {
	case Circle:
		shapePointer := shape.(*Circle)
		return shapePointer.ToShape()
	case Path:
		shapePointer := shape.(*Path)
		return shapePointer.ToPolyline(bezierTolerance).ToShape(bezierTolerance)
	case Polygon:
		shapePointer := shape.(*Polygon)
		return shapePointer.ToShape()
	case Polyline:
		shapePointer := shape.(*Polyline)
		return shapePointer.ToShape(bezierTolerance)
	case Shape:
		shapePointer := shape.(*Shape)
		return shapePointer
	case Group:
		shapePointer := shape.(*Group)
		return shapePointer.ToShape(bezierTolerance)
	default:
		if utils.DebugModeEnabled() {
			log.Println("Invalid shape type for svg shapes: ", shapeType)
		}
	}
	return nil
}

func combineShapes(shapeList *[]Shape) *Shape {
	var lines []Polyline

	for _, shape := range *shapeList {
		lines = append(lines, shape.Lines...)
	}

	return NewShape(lines)
}

func (group *Group) ToShape(bezierTolerance float64) *Shape {
	return combineShapes(anyListToShapeList(group.Shapes, bezierTolerance))
}

func (group *Group) ToGCode(activateTool []string, deactivateTool []string, bezierTolerance float64, stampMode bool, feedRateXY int) []string {
	lines := make([]string, 0)

	if stampMode {
		center := group.GetMidpoint(bezierTolerance)
		lines = append(lines, utils.G00XY(center.X, center.Y))
		lines = append(lines, activateTool...)
		lines = append(lines, deactivateTool...)
		return lines
	}

	for _, shape := range group.Shapes {
		lines = append(lines, groupShapeToGCode(shape, activateTool, deactivateTool, bezierTolerance, false, feedRateXY)...)
	}

	return lines
}

func groupShapeBoundries(shape any, bezierTolerance float64) (float64, float64, float64, float64) {
	switch currentShape := shape.(type) {
	case *Circle:
		return currentShape.GetBoundries()
	case *Group:
		return currentShape.GetBoundries(bezierTolerance)
	case *Path:
		return currentShape.GetBoundries(bezierTolerance)
	case *Polygon:
		return currentShape.GetBoundries()
	case *Polyline:
		return currentShape.GetBoundries()
	case *Shape:
		return currentShape.GetBoundries()
	default:
		if utils.DebugModeEnabled() {
			log.Println("GetBoundries: Unknown group shape type ", reflect.TypeOf(currentShape))
		}
		return 0, 0, 0, 0
	}
}

func groupShapeStrokeStart(shape any, bezierTolerance float64, stampMode bool) Point {
	switch currentShape := shape.(type) {
	case *Circle:
		return currentShape.GetStrokeStart(stampMode)
	case *Group:
		return currentShape.GetStrokeStart(bezierTolerance, stampMode)
	case *Path:
		return currentShape.GetStrokeStart(bezierTolerance, stampMode)
	case *Point:
		return *currentShape
	case *Polygon:
		return currentShape.GetStrokeStart(bezierTolerance, stampMode)
	case *Polyline:
		return currentShape.GetStrokeStart(bezierTolerance, stampMode)
	case *Shape:
		return currentShape.GetStrokeStart(bezierTolerance, stampMode)
	default:
		if utils.DebugModeEnabled() {
			log.Println("GetStrokeStart: Unknown group shape type ", reflect.TypeOf(currentShape))
		}
		return Point{}
	}
}

func groupShapeToGCode(shape any, activateTool []string, deactivateTool []string, bezierTolerance float64, stampMode bool, feedRateXY int) []string {
	switch currentShape := shape.(type) {
	case *Circle:
		return currentShape.ToGCode(activateTool, deactivateTool, stampMode, feedRateXY)
	case *Group:
		return currentShape.ToGCode(activateTool, deactivateTool, bezierTolerance, stampMode, feedRateXY)
	case *Path:
		return currentShape.ToGCode(activateTool, deactivateTool, bezierTolerance, stampMode, feedRateXY)
	case *Point:
		return currentShape.ToGCode(activateTool, deactivateTool)
	case *Polygon:
		return currentShape.ToGCode(activateTool, deactivateTool, stampMode, feedRateXY)
	case *Polyline:
		return currentShape.ToGCode(activateTool, deactivateTool, stampMode, feedRateXY)
	case *Shape:
		return currentShape.ToGCode(activateTool, deactivateTool, stampMode, feedRateXY)
	default:
		if utils.DebugModeEnabled() {
			log.Println("Invalid group shape type for gcode: ", reflect.TypeOf(currentShape))
		}
		return nil
	}
}
