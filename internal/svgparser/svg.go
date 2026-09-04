package svgparser

import (
	"cmp"
	"log"
	"reflect"
	"slices"

	"github.com/DavidJilg/Vecart/internal/shapes"
	"github.com/DavidJilg/Vecart/internal/utils"
)

type SVG struct {
	Shapes []any
}

func (svg *SVG) GetShapes(bezierTolerance float64) []shapes.Shape {
	var shapeList []shapes.Shape
	for _, shape := range svg.Shapes {
		shapeList = append(shapeList,
			*AnyToShape(shape, bezierTolerance))
	}

	return shapeList
}

func AnyToShape(shape any, bezierTolerance float64) *shapes.Shape {
	switch shapeType := shape.(type) {
	case shapes.Circle:
		shapePointer := shape.(shapes.Circle)
		return shapePointer.ToShape()
	case *shapes.Circle:
		shapePointer := shape.(*shapes.Circle)
		return shapePointer.ToShape()
	case shapes.Path:
		shapePointer := shape.(shapes.Path)
		polylinePointers := shapePointer.ToPolylines(bezierTolerance)
		polylines := make([]shapes.Polyline, len(polylinePointers))
		for i, polyline := range polylinePointers {
			polylines[i] = *polyline
		}
		return shapes.NewShape(polylines)
	case *shapes.Path:
		shapePointer := shape.(*shapes.Path)
		polylinePointers := shapePointer.ToPolylines(bezierTolerance)
		polylines := make([]shapes.Polyline, len(polylinePointers))
		for i, polyline := range polylinePointers {
			polylines[i] = *polyline
		}
		return shapes.NewShape(polylines)
	case shapes.Polygon:
		shapePointer := shape.(shapes.Polygon)
		return shapePointer.ToShape()
	case *shapes.Polygon:
		shapePointer := shape.(*shapes.Polygon)
		return shapePointer.ToShape()
	case shapes.Polyline:
		shapePointer := shape.(shapes.Polyline)
		return shapePointer.ToShape(bezierTolerance)
	case *shapes.Polyline:
		shapePointer := shape.(*shapes.Polyline)
		return shapePointer.ToShape(bezierTolerance)
	case shapes.Shape:
		shapePointer := shape.(shapes.Shape)
		return &shapePointer
	case *shapes.Shape:
		shapePointer := shape.(*shapes.Shape)
		return shapePointer
	case shapes.Group:
		shapePointer := shape.(shapes.Group)
		return shapePointer.ToShape(bezierTolerance)
	case *shapes.Group:
		shapePointer := shape.(*shapes.Group)
		return shapePointer.ToShape(bezierTolerance)
	default:
		if utils.DebugModeEnabled() {
			log.Println("Invalid shape type for svg shapes: ", shapeType)
		}
	}
	return nil
}

type Axis int

const (
	XAxis = iota
	YAxis
)

func AxisStringFromInt(axis int) (string, bool) {
	if axis == 0 {
		return "XAxis", true
	} else if axis == 1 {
		return "YAxis", true
	} else {
		return "YAxis", false
	}
}

func AxisFromString(axis string) (int, bool) {
	if axis == "XAxis" {
		return XAxis, true
	} else if axis == "YAxis" {
		return YAxis, true
	} else {
		return YAxis, false
	}
}

type Anchor int

const (
	Left = iota
	Top
	Right
	Bottom
)

func (svg *SVG) PixelToMM(dpi float64) {
	for i := 0; i < len(svg.Shapes); i++ {
		switch svg.Shapes[i].(type) {
		case *shapes.Circle:
			svg.Shapes[i].(*shapes.Circle).PixelToMM(dpi)
		case *shapes.Group:
			svg.Shapes[i].(*shapes.Group).PixelToMM(dpi)
		case *shapes.Polygon:
			svg.Shapes[i].(*shapes.Polygon).PixelToMM(dpi)
		case *shapes.Polyline:
			svg.Shapes[i].(*shapes.Polyline).PixelToMM(dpi)
		case *shapes.Shape:
			svg.Shapes[i].(*shapes.Shape).PixelToMM(dpi)
		case *shapes.Path:
			svg.Shapes[i].(*shapes.Path).PixelToMM(dpi)
		default:
			if utils.DebugModeEnabled() {
				log.Println("Invalid svg element: ", reflect.TypeOf(svg.Shapes[i]))
			}
			return
		}
	}
}

func (svg *SVG) Sort(axis Axis, reverse bool, bezierTolerance float64, stampMode bool) {
	slices.SortFunc(svg.Shapes,
		func(a, b any) int {
			return cmp.Compare(GetStrokeStart(a, axis, bezierTolerance, stampMode), GetStrokeStart(b, axis, bezierTolerance, stampMode))
		})

	if reverse {
		slices.Reverse(svg.Shapes)
	}
}

func GetBoundry(element any, index int, bezierTolerance float64) float64 {
	a, b, c, d := GetBoundries(element, bezierTolerance)
	boundries := [4]float64{a, b, c, d}
	return boundries[index]
}

func GetBoundries(element any, bezierTolerance float64) (float64, float64, float64, float64) {
	switch elementType := element.(type) {
	case *shapes.Circle:
		return elementType.GetBoundries()
	case *shapes.Group:
		return elementType.GetBoundries(bezierTolerance)
	case *shapes.Polygon:
		return elementType.GetBoundries()
	case *shapes.Polyline:
		return elementType.GetBoundries()
	case *shapes.Shape:
		return elementType.GetBoundries()
	case *shapes.Path:
		return elementType.GetBoundries(bezierTolerance)
	default:
		if utils.DebugModeEnabled() {
			log.Println("GetBoundries: Unknown shape type ", reflect.TypeOf(elementType))
		}
		return 0, 0, 0, 0
	}
}

func GetStrokeStart(element any, axis Axis, bezierTolerance float64, stampMode bool) float64 {
	strokeStart := shapes.Point{}
	switch elementType := element.(type) {
	case *shapes.Circle:
		strokeStart = elementType.GetStrokeStart(stampMode)
	case *shapes.Group:
		strokeStart = elementType.GetStrokeStart(bezierTolerance, stampMode)
	case *shapes.Polygon:
		strokeStart = elementType.GetStrokeStart(bezierTolerance, stampMode)
	case *shapes.Polyline:
		strokeStart = elementType.GetStrokeStart(bezierTolerance, stampMode)
	case *shapes.Shape:
		strokeStart = elementType.GetStrokeStart(bezierTolerance, stampMode)
	case *shapes.Path:
		strokeStart = elementType.GetStrokeStart(bezierTolerance, stampMode)
	default:
		if utils.DebugModeEnabled() {
			log.Println("GetStrokeStart: Unknown shape type ", reflect.TypeOf(elementType))
		}
		return 0.0
	}

	if axis == XAxis {
		return strokeStart.X
	}
	return strokeStart.Y
}
