package svgparser

import (
	"testing"

	"github.com/DavidJilg/Vecart/internal/shapes"
)

func TestExtractShapesAppliesTranslateToPath(t *testing.T) {
	svgText := `<?xml version="1.0" encoding="utf-8"?>
<svg version="1.1" xmlns="http://www.w3.org/2000/svg">
	<path transform="translate(5.5,-1.25)" d="M1,2h3v4z"/>
</svg>`

	root, err := ParseXMLTree(svgText)
	if err != nil {
		t.Fatalf("ParseXMLTree failed: %v", err)
	}

	svg := ExtractShapes(root, false, false)
	if len(svg.Shapes) != 1 {
		t.Fatalf("expected 1 shape, got %d", len(svg.Shapes))
	}

	path, ok := svg.Shapes[0].(*shapes.Path)
	if !ok {
		t.Fatalf("expected *shapes.Path, got %T", svg.Shapes[0])
	}

	polyline := path.ToPolyline(0.1)
	if len(polyline.Points) != 4 {
		t.Fatalf("expected 4 points, got %d", len(polyline.Points))
	}

	assertPoint(t, polyline.Points[0].X, polyline.Points[0].Y, 6.5, 0.75)
	assertPoint(t, polyline.Points[1].X, polyline.Points[1].Y, 9.5, 0.75)
	assertPoint(t, polyline.Points[2].X, polyline.Points[2].Y, 9.5, 4.75)
	assertPoint(t, polyline.Points[3].X, polyline.Points[3].Y, 6.5, 0.75)
}

func TestExtractShapesAppliesNestedGroupTranslate(t *testing.T) {
	svgText := `<?xml version="1.0" encoding="utf-8"?>
<svg version="1.1" xmlns="http://www.w3.org/2000/svg">
	<g transform="translate(5,-1)">
		<g transform="translate(2,3)">
			<rect x="1" y="2" width="4" height="6"/>
		</g>
	</g>
</svg>`

	root, err := ParseXMLTree(svgText)
	if err != nil {
		t.Fatalf("ParseXMLTree failed: %v", err)
	}

	svg := ExtractShapes(root, false, false)
	if len(svg.Shapes) != 1 {
		t.Fatalf("expected 1 shape, got %d", len(svg.Shapes))
	}

	polygon, ok := svg.Shapes[0].(*shapes.Polygon)
	if !ok {
		t.Fatalf("expected *shapes.Polygon, got %T", svg.Shapes[0])
	}

	if len(polygon.Points) != 4 {
		t.Fatalf("expected 4 points, got %d", len(polygon.Points))
	}

	assertPoint(t, polygon.Points[0].X, polygon.Points[0].Y, 8, 4)
	assertPoint(t, polygon.Points[1].X, polygon.Points[1].Y, 12, 4)
	assertPoint(t, polygon.Points[2].X, polygon.Points[2].Y, 12, 10)
	assertPoint(t, polygon.Points[3].X, polygon.Points[3].Y, 8, 10)
}

func TestExtractShapesAppliesMatrixToRect(t *testing.T) {
	svgText := `<?xml version="1.0" encoding="utf-8"?>
<svg version="1.1" xmlns="http://www.w3.org/2000/svg">
	<rect x="1" y="2" width="4" height="6" transform="matrix(2,0,0,3,5,-7)"/>
</svg>`

	root, err := ParseXMLTree(svgText)
	if err != nil {
		t.Fatalf("ParseXMLTree failed: %v", err)
	}

	svg := ExtractShapes(root, false, false)
	if len(svg.Shapes) != 1 {
		t.Fatalf("expected 1 shape, got %d", len(svg.Shapes))
	}

	polygon, ok := svg.Shapes[0].(*shapes.Polygon)
	if !ok {
		t.Fatalf("expected *shapes.Polygon, got %T", svg.Shapes[0])
	}

	assertPoint(t, polygon.Points[0].X, polygon.Points[0].Y, 7, -1)
	assertPoint(t, polygon.Points[1].X, polygon.Points[1].Y, 15, -1)
	assertPoint(t, polygon.Points[2].X, polygon.Points[2].Y, 15, 17)
	assertPoint(t, polygon.Points[3].X, polygon.Points[3].Y, 7, 17)
}

func TestExtractShapesAppliesNestedMatrixAndTranslateUsedInPrintSVG(t *testing.T) {
	svgText := `<?xml version="1.0" encoding="utf-8"?>
<svg version="1.1" xmlns="http://www.w3.org/2000/svg">
	<g transform="matrix(0.26458333,0,0,0.26458333,-56.256758,10.834922)">
		<g transform="translate(5.1713131,-0.86188552)">
			<path d="M404.1,276v3l0.1,0.1h1.9c0.2-0.1,0.4-0.3,0.5-0.5v-2c-0.1-0.5-0.3-0.7-0.7-0.7L404.1,276z"/>
		</g>
	</g>
</svg>`

	root, err := ParseXMLTree(svgText)
	if err != nil {
		t.Fatalf("ParseXMLTree failed: %v", err)
	}

	svg := ExtractShapes(root, false, false)
	if len(svg.Shapes) != 1 {
		t.Fatalf("expected 1 shape, got %d", len(svg.Shapes))
	}

	path, ok := svg.Shapes[0].(*shapes.Path)
	if !ok {
		t.Fatalf("expected *shapes.Path, got %T", svg.Shapes[0])
	}

	polyline := path.ToPolyline(0.1)
	expectedX := 0.26458333*(404.1+5.1713131) - 56.256758
	expectedY := 0.26458333*(276-0.86188552) + 10.834922
	assertPoint(t, polyline.Points[0].X, polyline.Points[0].Y, expectedX, expectedY)
}

func TestExtractShapesAppliesScaleToRect(t *testing.T) {
	svgText := `<?xml version="1.0" encoding="utf-8"?>
<svg version="1.1" xmlns="http://www.w3.org/2000/svg">
	<rect x="1" y="2" width="4" height="6" transform="scale(2,3)"/>
</svg>`

	root, err := ParseXMLTree(svgText)
	if err != nil {
		t.Fatalf("ParseXMLTree failed: %v", err)
	}

	svg := ExtractShapes(root, false, false)
	if len(svg.Shapes) != 1 {
		t.Fatalf("expected 1 shape, got %d", len(svg.Shapes))
	}

	polygon, ok := svg.Shapes[0].(*shapes.Polygon)
	if !ok {
		t.Fatalf("expected *shapes.Polygon, got %T", svg.Shapes[0])
	}

	assertPoint(t, polygon.Points[0].X, polygon.Points[0].Y, 2, 6)
	assertPoint(t, polygon.Points[1].X, polygon.Points[1].Y, 10, 6)
	assertPoint(t, polygon.Points[2].X, polygon.Points[2].Y, 10, 24)
	assertPoint(t, polygon.Points[3].X, polygon.Points[3].Y, 2, 24)
}

func TestExtractShapesAppliesRotateAroundCenter(t *testing.T) {
	svgText := `<?xml version="1.0" encoding="utf-8"?>
<svg version="1.1" xmlns="http://www.w3.org/2000/svg">
	<rect x="1" y="2" width="4" height="6" transform="rotate(90,1,2)"/>
</svg>`

	root, err := ParseXMLTree(svgText)
	if err != nil {
		t.Fatalf("ParseXMLTree failed: %v", err)
	}

	svg := ExtractShapes(root, false, false)
	if len(svg.Shapes) != 1 {
		t.Fatalf("expected 1 shape, got %d", len(svg.Shapes))
	}

	polygon, ok := svg.Shapes[0].(*shapes.Polygon)
	if !ok {
		t.Fatalf("expected *shapes.Polygon, got %T", svg.Shapes[0])
	}

	assertPoint(t, polygon.Points[0].X, polygon.Points[0].Y, 1, 2)
	assertPoint(t, polygon.Points[1].X, polygon.Points[1].Y, 1, 6)
	assertPoint(t, polygon.Points[2].X, polygon.Points[2].Y, -5, 6)
	assertPoint(t, polygon.Points[3].X, polygon.Points[3].Y, -5, 2)
}

func TestExtractShapesAppliesSkewXToRect(t *testing.T) {
	svgText := `<?xml version="1.0" encoding="utf-8"?>
<svg version="1.1" xmlns="http://www.w3.org/2000/svg">
	<rect x="1" y="2" width="4" height="6" transform="skewX(45)"/>
</svg>`

	root, err := ParseXMLTree(svgText)
	if err != nil {
		t.Fatalf("ParseXMLTree failed: %v", err)
	}

	svg := ExtractShapes(root, false, false)
	if len(svg.Shapes) != 1 {
		t.Fatalf("expected 1 shape, got %d", len(svg.Shapes))
	}

	polygon, ok := svg.Shapes[0].(*shapes.Polygon)
	if !ok {
		t.Fatalf("expected *shapes.Polygon, got %T", svg.Shapes[0])
	}

	assertPoint(t, polygon.Points[0].X, polygon.Points[0].Y, 3, 2)
	assertPoint(t, polygon.Points[1].X, polygon.Points[1].Y, 7, 2)
	assertPoint(t, polygon.Points[2].X, polygon.Points[2].Y, 13, 8)
	assertPoint(t, polygon.Points[3].X, polygon.Points[3].Y, 9, 8)
}

func TestExtractShapesAppliesSkewYToRect(t *testing.T) {
	svgText := `<?xml version="1.0" encoding="utf-8"?>
<svg version="1.1" xmlns="http://www.w3.org/2000/svg">
	<rect x="1" y="2" width="4" height="6" transform="skewY(45)"/>
</svg>`

	root, err := ParseXMLTree(svgText)
	if err != nil {
		t.Fatalf("ParseXMLTree failed: %v", err)
	}

	svg := ExtractShapes(root, false, false)
	if len(svg.Shapes) != 1 {
		t.Fatalf("expected 1 shape, got %d", len(svg.Shapes))
	}

	polygon, ok := svg.Shapes[0].(*shapes.Polygon)
	if !ok {
		t.Fatalf("expected *shapes.Polygon, got %T", svg.Shapes[0])
	}

	assertPoint(t, polygon.Points[0].X, polygon.Points[0].Y, 1, 3)
	assertPoint(t, polygon.Points[1].X, polygon.Points[1].Y, 5, 7)
	assertPoint(t, polygon.Points[2].X, polygon.Points[2].Y, 5, 13)
	assertPoint(t, polygon.Points[3].X, polygon.Points[3].Y, 1, 9)
}

func TestExtractShapesComposesTransformListInOrder(t *testing.T) {
	svgText := `<?xml version="1.0" encoding="utf-8"?>
<svg version="1.1" xmlns="http://www.w3.org/2000/svg">
	<line x1="1" y1="0" x2="3" y2="0" transform="translate(5,0) scale(2)"/>
</svg>`

	root, err := ParseXMLTree(svgText)
	if err != nil {
		t.Fatalf("ParseXMLTree failed: %v", err)
	}

	svg := ExtractShapes(root, false, false)
	if len(svg.Shapes) != 1 {
		t.Fatalf("expected 1 shape, got %d", len(svg.Shapes))
	}

	polyline, ok := svg.Shapes[0].(*shapes.Polyline)
	if !ok {
		t.Fatalf("expected *shapes.Polyline, got %T", svg.Shapes[0])
	}

	assertPoint(t, polyline.Points[0].X, polyline.Points[0].Y, 12, 0)
	assertPoint(t, polyline.Points[1].X, polyline.Points[1].Y, 16, 0)
}
