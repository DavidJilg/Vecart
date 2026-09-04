package general

import (
	"testing"

	"github.com/DavidJilg/Vecart/internal/shapes"
	"github.com/DavidJilg/Vecart/internal/svgparser"
)

func TestSortSVGForGCodeUsesSerpentineBandsOnXAxis(t *testing.T) {
	oldConfig := Config
	Config = NewConfig()
	Config.SortAxis = svgparser.XAxis
	Config.ReverseShapeOrder = false
	Config.BezierTolerance = 0.1
	t.Cleanup(func() {
		Config = oldConfig
	})

	svgText := `<?xml version="1.0" encoding="utf-8"?>
<svg version="1.1" xmlns="http://www.w3.org/2000/svg">
	<line x1="20" y1="20" x2="21" y2="20"/>
	<line x1="0" y1="10" x2="1" y2="10"/>
	<line x1="30" y1="0" x2="31" y2="0"/>
	<line x1="10" y1="30" x2="11" y2="30"/>
	<line x1="20" y1="30" x2="21" y2="30"/>
	<line x1="0" y1="30" x2="1" y2="30"/>
	<line x1="30" y1="20" x2="31" y2="20"/>
	<line x1="10" y1="0" x2="11" y2="0"/>
	<line x1="20" y1="0" x2="21" y2="0"/>
	<line x1="0" y1="20" x2="1" y2="20"/>
	<line x1="30" y1="30" x2="31" y2="30"/>
	<line x1="10" y1="20" x2="11" y2="20"/>
	<line x1="20" y1="10" x2="21" y2="10"/>
	<line x1="0" y1="0" x2="1" y2="0"/>
	<line x1="30" y1="10" x2="31" y2="10"/>
	<line x1="10" y1="10" x2="11" y2="10"/>
</svg>`

	svg := extractSVGForGCodeTest(t, svgText, false)

	sortSVGForGCode(svg)

	assertStrokeStarts(t, svg.Shapes, Config.BezierTolerance, []shapes.Point{
		{X: 0, Y: 0},
		{X: 0, Y: 10},
		{X: 0, Y: 20},
		{X: 0, Y: 30},
		{X: 10, Y: 30},
		{X: 10, Y: 20},
		{X: 10, Y: 10},
		{X: 10, Y: 0},
		{X: 20, Y: 0},
		{X: 20, Y: 10},
		{X: 20, Y: 20},
		{X: 20, Y: 30},
		{X: 30, Y: 30},
		{X: 30, Y: 20},
		{X: 30, Y: 10},
		{X: 30, Y: 0},
	})
}

func TestSortSVGForGCodeReverseReversesSerpentinePath(t *testing.T) {
	oldConfig := Config
	Config = NewConfig()
	Config.SortAxis = svgparser.XAxis
	Config.ReverseShapeOrder = true
	Config.BezierTolerance = 0.1
	t.Cleanup(func() {
		Config = oldConfig
	})

	svgText := `<?xml version="1.0" encoding="utf-8"?>
<svg version="1.1" xmlns="http://www.w3.org/2000/svg">
	<line x1="20" y1="20" x2="21" y2="20"/>
	<line x1="0" y1="10" x2="1" y2="10"/>
	<line x1="30" y1="0" x2="31" y2="0"/>
	<line x1="10" y1="30" x2="11" y2="30"/>
	<line x1="20" y1="30" x2="21" y2="30"/>
	<line x1="0" y1="30" x2="1" y2="30"/>
	<line x1="30" y1="20" x2="31" y2="20"/>
	<line x1="10" y1="0" x2="11" y2="0"/>
	<line x1="20" y1="0" x2="21" y2="0"/>
	<line x1="0" y1="20" x2="1" y2="20"/>
	<line x1="30" y1="30" x2="31" y2="30"/>
	<line x1="10" y1="20" x2="11" y2="20"/>
	<line x1="20" y1="10" x2="21" y2="10"/>
	<line x1="0" y1="0" x2="1" y2="0"/>
	<line x1="30" y1="10" x2="31" y2="10"/>
	<line x1="10" y1="10" x2="11" y2="10"/>
</svg>`

	svg := extractSVGForGCodeTest(t, svgText, false)

	sortSVGForGCode(svg)

	assertStrokeStarts(t, svg.Shapes, Config.BezierTolerance, []shapes.Point{
		{X: 30, Y: 0},
		{X: 30, Y: 10},
		{X: 30, Y: 20},
		{X: 30, Y: 30},
		{X: 20, Y: 30},
		{X: 20, Y: 20},
		{X: 20, Y: 10},
		{X: 20, Y: 0},
		{X: 10, Y: 0},
		{X: 10, Y: 10},
		{X: 10, Y: 20},
		{X: 10, Y: 30},
		{X: 0, Y: 30},
		{X: 0, Y: 20},
		{X: 0, Y: 10},
		{X: 0, Y: 0},
	})
}

func TestSortSVGForGCodeExplodesNestedGroupsBeforeSerpentineSort(t *testing.T) {
	oldConfig := Config
	Config = NewConfig()
	Config.SortAxis = svgparser.YAxis
	Config.ReverseShapeOrder = false
	Config.BezierTolerance = 0.1
	t.Cleanup(func() {
		Config = oldConfig
	})

	svgText := `<?xml version="1.0" encoding="utf-8"?>
<svg version="1.1" xmlns="http://www.w3.org/2000/svg">
	<g id="group-a">
		<line x1="0" y1="30" x2="10" y2="30"/>
		<g>
			<line x1="0" y1="10" x2="10" y2="10"/>
		</g>
	</g>
	<g id="group-b">
		<g>
			<line x1="0" y1="20" x2="10" y2="20"/>
		</g>
		<line x1="0" y1="40" x2="10" y2="40"/>
	</g>
</svg>`

	svg := extractSVGForGCodeTest(t, svgText, true)
	if len(svg.Shapes) != 2 {
		t.Fatalf("expected 2 top-level groups before sorting, got %d", len(svg.Shapes))
	}

	sortSVGForGCode(svg)

	if len(svg.Shapes) != 4 {
		t.Fatalf("expected 4 individual shapes after flattening, got %d", len(svg.Shapes))
	}

	for _, shape := range svg.Shapes {
		if _, ok := shape.(*shapes.Group); ok {
			t.Fatalf("expected flattened shapes to contain no groups, found %T", shape)
		}
	}

	assertStrokeStarts(t, svg.Shapes, Config.BezierTolerance, []shapes.Point{
		{X: 0, Y: 10},
		{X: 0, Y: 20},
		{X: 0, Y: 30},
		{X: 0, Y: 40},
	})
}

func extractSVGForGCodeTest(t *testing.T, svgText string, keepGroups bool) *svgparser.SVG {
	t.Helper()

	root, err := svgparser.ParseXMLTree(svgText)
	if err != nil {
		t.Fatalf("ParseXMLTree failed: %v", err)
	}

	return svgparser.ExtractShapes(root, false, keepGroups)
}

func assertStrokeStarts(t *testing.T, shapeList []any, bezierTolerance float64, expected []shapes.Point) {
	t.Helper()

	if len(shapeList) != len(expected) {
		t.Fatalf("expected %d shapes, got %d", len(expected), len(shapeList))
	}

	for index, expectedPoint := range expected {
		actualX := svgparser.GetStrokeStart(shapeList[index], svgparser.XAxis, bezierTolerance, Config.StampMode)
		actualY := svgparser.GetStrokeStart(shapeList[index], svgparser.YAxis, bezierTolerance, Config.StampMode)
		if actualX != expectedPoint.X || actualY != expectedPoint.Y {
			t.Fatalf(
				"expected shape %d to start at (%v, %v), got (%v, %v)",
				index,
				expectedPoint.X,
				expectedPoint.Y,
				actualX,
				actualY,
			)
		}
	}
}
