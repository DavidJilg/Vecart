package shapes

import (
	"log"
	"reflect"
	"github.com/DavidJilg/Vecart/internal/utils"
)

type Path struct {
	subpaths [][]any // Elements can be *Point, *CubicBezier or *QuadraticBezier
}

func NewPath(elements []any) Path {
	if len(elements) == 0 {
		return Path{}
	}

	return Path{subpaths: [][]any{elements}}
}

func NewCompoundPath(subpaths [][]any) Path {
	return Path{subpaths: subpaths}
}

func NewEllipse(center Point, rx, ry float64) *Path {
	const kappa = 0.5522847498307936

	elements := make([]any, 0, 5)

	start := Point{X: center.X + rx, Y: center.Y}
	elements = append(elements, &start)

	elements = append(elements, &CubicBezier{
		Start:    start,
		Control1: Point{X: center.X + rx, Y: center.Y + kappa*ry},
		Control2: Point{X: center.X + kappa*rx, Y: center.Y + ry},
		End:      Point{X: center.X, Y: center.Y + ry},
	})

	elements = append(elements, &CubicBezier{
		Start:    Point{X: center.X, Y: center.Y + ry},
		Control1: Point{X: center.X - kappa*rx, Y: center.Y + ry},
		Control2: Point{X: center.X - rx, Y: center.Y + kappa*ry},
		End:      Point{X: center.X - rx, Y: center.Y},
	})

	elements = append(elements, &CubicBezier{
		Start:    Point{X: center.X - rx, Y: center.Y},
		Control1: Point{X: center.X - rx, Y: center.Y - kappa*ry},
		Control2: Point{X: center.X - kappa*rx, Y: center.Y - ry},
		End:      Point{X: center.X, Y: center.Y - ry},
	})

	elements = append(elements, &CubicBezier{
		Start:    Point{X: center.X, Y: center.Y - ry},
		Control1: Point{X: center.X + kappa*rx, Y: center.Y - ry},
		Control2: Point{X: center.X + rx, Y: center.Y - kappa*ry},
		End:      Point{X: center.X + rx, Y: center.Y},
	})

	newPath := NewPath(elements)
	return &newPath
}

func NewRoundedRect(x, y, width, height, rx, ry float64) (*Path, bool) {
	if rx < 0 {
		rx = 0
	}
	if ry < 0 {
		ry = 0
	}

	if rx > width*0.5 {
		rx = width * 0.5
	}
	if ry > height*0.5 {
		ry = height * 0.5
	}

	if rx == 0 || ry == 0 {
		if utils.DebugModeEnabled() {
			log.Println("NewRoundedRect called with radius of 0")
		}
		return &Path{}, false
	}

	const kappa = 0.5522847498307936

	elements := make([]any, 0, 9)
	start := Point{X: x + rx, Y: y}
	elements = append(elements, &start)

	elements = append(elements, &Point{X: x + width - rx, Y: y})
	elements = append(elements, &CubicBezier{
		Start:    Point{X: x + width - rx, Y: y},
		Control1: Point{X: x + width - rx + kappa*rx, Y: y},
		Control2: Point{X: x + width, Y: y + ry - kappa*ry},
		End:      Point{X: x + width, Y: y + ry},
	})

	elements = append(elements, &Point{X: x + width, Y: y + height - ry})
	elements = append(elements, &CubicBezier{
		Start:    Point{X: x + width, Y: y + height - ry},
		Control1: Point{X: x + width, Y: y + height - ry + kappa*ry},
		Control2: Point{X: x + width - rx + kappa*rx, Y: y + height},
		End:      Point{X: x + width - rx, Y: y + height},
	})

	elements = append(elements, &Point{X: x + rx, Y: y + height})
	elements = append(elements, &CubicBezier{
		Start:    Point{X: x + rx, Y: y + height},
		Control1: Point{X: x + rx - kappa*rx, Y: y + height},
		Control2: Point{X: x, Y: y + height - ry + kappa*ry},
		End:      Point{X: x, Y: y + height - ry},
	})

	elements = append(elements, &Point{X: x, Y: y + ry})
	elements = append(elements, &CubicBezier{
		Start:    Point{X: x, Y: y + ry},
		Control1: Point{X: x, Y: y + ry - kappa*ry},
		Control2: Point{X: x + rx - kappa*rx, Y: y},
		End:      Point{X: x + rx, Y: y},
	})

	newPath := NewPath(elements)
	return &newPath, true
}

func (path *Path) Move(x, y float64) {
	for subpathIndex := range path.subpaths {
		for elementIndex, elem := range path.subpaths[subpathIndex] {
			switch elementType := elem.(type) {
			case *Point:
				path.subpaths[subpathIndex][elementIndex].(*Point).Move(x, y)
			case *CubicBezier:
				path.subpaths[subpathIndex][elementIndex].(*CubicBezier).Move(x, y)
			case *QuadraticBezier:
				path.subpaths[subpathIndex][elementIndex].(*QuadraticBezier).Move(x, y)
			default:
				if utils.DebugModeEnabled() {
					log.Println("Invalid Type for original Shape ", elementType)
				}
			}
		}
	}
}

func (path *Path) Transform(a, b, c, d, e, f float64) {
	for subpathIndex := range path.subpaths {
		for elementIndex, elem := range path.subpaths[subpathIndex] {
			switch elementType := elem.(type) {
			case *Point:
				path.subpaths[subpathIndex][elementIndex].(*Point).Transform(a, b, c, d, e, f)
			case *CubicBezier:
				path.subpaths[subpathIndex][elementIndex].(*CubicBezier).Transform(a, b, c, d, e, f)
			case *QuadraticBezier:
				path.subpaths[subpathIndex][elementIndex].(*QuadraticBezier).Transform(a, b, c, d, e, f)
			default:
				if utils.DebugModeEnabled() {
					log.Println("Invalid Type for original Shape ", elementType)
				}
			}
		}
	}
}

func (path *Path) Scale(factor float64, centroid Point) {
	for subpathIndex := range path.subpaths {
		for elementIndex, elem := range path.subpaths[subpathIndex] {
			switch elem.(type) {
			case *Point:
				path.subpaths[subpathIndex][elementIndex].(*Point).Scale(factor, centroid)
			case *CubicBezier:
				path.subpaths[subpathIndex][elementIndex].(*CubicBezier).Scale(factor, centroid)
			case *QuadraticBezier:
				path.subpaths[subpathIndex][elementIndex].(*QuadraticBezier).Scale(factor, centroid)
			default:
				if utils.DebugModeEnabled() {
					log.Println("Invalid type in Path elements:", elem)
				}
			}
		}
	}
}

func (path *Path) Rotate(angle float64, origin Point) {
	for subpathIndex := range path.subpaths {
		for elementIndex, elem := range path.subpaths[subpathIndex] {
			switch elementType := elem.(type) {
			case *Point:
				path.subpaths[subpathIndex][elementIndex].(*Point).Rotate(angle, origin)
			case *CubicBezier:
				path.subpaths[subpathIndex][elementIndex].(*CubicBezier).Rotate(angle, origin)
			case *QuadraticBezier:
				path.subpaths[subpathIndex][elementIndex].(*QuadraticBezier).Rotate(angle, origin)
			default:
				if utils.DebugModeEnabled() {
					log.Println("Invalid Type for original Shape ", elementType)
				}
			}
		}
	}
}

func (path *Path) PixelToMM(dpi float64) {
	for subpathIndex := range path.subpaths {
		for elementIndex := range path.subpaths[subpathIndex] {
			switch path.subpaths[subpathIndex][elementIndex].(type) {
			case *Point:
				path.subpaths[subpathIndex][elementIndex].(*Point).PixelToMM(dpi)
			case *CubicBezier:
				path.subpaths[subpathIndex][elementIndex].(*CubicBezier).PixelToMM(dpi)
			case *QuadraticBezier:
				path.subpaths[subpathIndex][elementIndex].(*QuadraticBezier).PixelToMM(dpi)
			default:
				if utils.DebugModeEnabled() {
					log.Println("Invalid Type for path element ", reflect.TypeOf(path.subpaths[subpathIndex][elementIndex]))
				}
			}
		}
	}
}

func (path *Path) GetBoundries(bezierTolerance float64) (float64, float64, float64, float64) {
	polylines := path.ToPolylines(bezierTolerance)
	if len(polylines) == 0 {
		return 0, 0, 0, 0
	}

	minX, minY, maxX, maxY := polylines[0].GetBoundries()
	for index := 1; index < len(polylines); index++ {
		currentMinX, currentMinY, currentMaxX, currentMaxY := polylines[index].GetBoundries()
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

func (path *Path) ToPolylines(bezierTolerance float64) []*Polyline {
	polylines := make([]*Polyline, 0, len(path.subpaths))

	for _, subpath := range path.subpaths {
		points := make([]Point, 0)

		for _, element := range subpath {
			switch elementType := element.(type) {
			case *Point:
				points = append(points, *element.(*Point))
			case *CubicBezier:
				polyline := element.(*CubicBezier).ToPolyline(bezierTolerance)
				if len(points) == 0 {
					points = append(points, polyline.Points...)
				} else {
					points = append(points, polyline.Points[1:len(polyline.Points)]...)
				}
			case *QuadraticBezier:
				polyline := element.(*QuadraticBezier).ToPolyline(bezierTolerance)
				if len(points) == 0 {
					points = append(points, polyline.Points...)
				} else {
					points = append(points, polyline.Points[1:len(polyline.Points)]...)
				}
			default:
				if utils.DebugModeEnabled() {
					log.Println("Invalid Type for path element ", reflect.TypeOf(elementType))
				}
			}
		}

		if len(points) == 0 {
			continue
		}

		polyline := &Polyline{Points: points}
		if !polyline.HasDrawableSegments() {
			continue
		}

		polylines = append(polylines, polyline)
	}

	return polylines
}

func (path *Path) ToPolyline(bezierTolerance float64) *Polyline {
	polylines := path.ToPolylines(bezierTolerance)
	if len(polylines) == 0 {
		return &Polyline{}
	}

	if len(polylines) == 1 {
		polylines[0].OriginalShape = path
		return polylines[0]
	}

	points := make([]Point, 0)
	for _, polyline := range polylines {
		points = append(points, polyline.Points...)
	}

	return &Polyline{Points: points, OriginalShape: path}
}

func (path *Path) FirstPoint(bezierTolerance float64) (Point, bool) {
	polylines := path.ToPolylines(bezierTolerance)
	if len(polylines) == 0 || len(polylines[0].Points) == 0 {
		return Point{}, false
	}

	return polylines[0].Points[0], true
}

func (path *Path) IsCompound() bool {
	return len(path.subpaths) > 1
}

func (path *Path) SubpathCount() int {
	return len(path.subpaths)
}

func (path *Path) ElementCount() int {
	count := 0
	for _, subpath := range path.subpaths {
		count += len(subpath)
	}

	return count
}

func (path *Path) HasGeometry() bool {
	for _, subpath := range path.subpaths {
		if len(subpath) > 0 {
			return true
		}
	}

	return false
}

func (path *Path) GetMidpoint(bezierTolerance float64) Point {
	minX, minY, maxX, maxY := path.GetBoundries(bezierTolerance)
	return Point{
		X: minX + (maxX - minX) / 2,
		Y: minY + (maxY - minY) / 2,
	}
}

func (path Path) GetStrokeStart(bezierTolerance float64, stampMode bool) Point {
	if(stampMode) {
		return path.GetMidpoint(bezierTolerance)
	}

	firstPoint, found := path.FirstPoint(bezierTolerance)
	if !found {
		return Point{}
	}

	return firstPoint
}

func (path Path) ToGCode(activateTool []string, deactivateTool []string, bezierTolerance float64, stampMode bool, feedrateXY int) []string {
	lines := make([]string, 0)

	if stampMode {
		polyline := path.ToPolyline(bezierTolerance)
		polyline.OriginalShape = nil
		lines = append(lines, polyline.ToGCode(activateTool, deactivateTool, stampMode, feedrateXY)...)
		return lines
	}

	polylines := path.ToPolylines(bezierTolerance)
	for _, polyline := range polylines {
		polyline.OriginalShape = nil
		lines = append(lines, polyline.ToGCode(activateTool, deactivateTool, false, feedrateXY)...)
	}

	return lines
}
