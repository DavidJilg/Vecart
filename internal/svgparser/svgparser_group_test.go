package svgparser

import (
	"testing"

	"github.com/DavidJilg/Vecart/internal/shapes"
)

func TestExtractShapesKeepGroupsPreservesNestedGroups(t *testing.T) {
	svgText := `<?xml version="1.0" encoding="utf-8"?>
<svg version="1.1" xmlns="http://www.w3.org/2000/svg">
	<g>
		<g>
			<line x1="0" y1="0" x2="10" y2="0"/>
		</g>
		<rect x="20" y="0" width="5" height="5"/>
	</g>
	<g>
		<line x1="0" y1="20" x2="10" y2="20"/>
	</g>
</svg>`

	root, err := ParseXMLTree(svgText)
	if err != nil {
		t.Fatalf("ParseXMLTree failed: %v", err)
	}

	svg := ExtractShapes(root, false, true)
	if len(svg.Shapes) != 2 {
		t.Fatalf("expected 2 top-level shapes, got %d", len(svg.Shapes))
	}

	firstGroup, ok := svg.Shapes[0].(*shapes.Group)
	if !ok {
		t.Fatalf("expected first top-level shape to be *shapes.Group, got %T", svg.Shapes[0])
	}

	if len(firstGroup.Shapes) != 2 {
		t.Fatalf("expected first group to contain 2 shapes, got %d", len(firstGroup.Shapes))
	}

	if _, ok := firstGroup.Shapes[0].(*shapes.Group); !ok {
		t.Fatalf("expected first child to be nested *shapes.Group, got %T", firstGroup.Shapes[0])
	}
}

func TestSVGSortSortsByGroupStrokeStart(t *testing.T) {
	svgText := `<?xml version="1.0" encoding="utf-8"?>
<svg version="1.1" xmlns="http://www.w3.org/2000/svg">
	<g id="group-a">
		<line x1="0" y1="100" x2="10" y2="100"/>
		<line x1="0" y1="0" x2="10" y2="0"/>
	</g>
	<g id="group-b">
		<line x1="0" y1="50" x2="10" y2="50"/>
	</g>
</svg>`

	root, err := ParseXMLTree(svgText)
	if err != nil {
		t.Fatalf("ParseXMLTree failed: %v", err)
	}

	svg := ExtractShapes(root, false, true)
	if len(svg.Shapes) != 2 {
		t.Fatalf("expected 2 top-level groups, got %d", len(svg.Shapes))
	}

	svg.Sort(YAxis, false, 0.1, false)

	firstGroup, ok := svg.Shapes[0].(*shapes.Group)
	if !ok {
		t.Fatalf("expected first sorted shape to be *shapes.Group, got %T", svg.Shapes[0])
	}

	if len(firstGroup.Shapes) != 1 {
		t.Fatalf("expected first sorted group to have 1 child, got %d", len(firstGroup.Shapes))
	}

	firstStrokeStart := firstGroup.GetStrokeStart(0.1, false)
	assertPoint(t, firstStrokeStart.X, firstStrokeStart.Y, 0, 50)

	secondGroup, ok := svg.Shapes[1].(*shapes.Group)
	if !ok {
		t.Fatalf("expected second sorted shape to be *shapes.Group, got %T", svg.Shapes[1])
	}

	if len(secondGroup.Shapes) != 2 {
		t.Fatalf("expected second sorted group to keep both children, got %d", len(secondGroup.Shapes))
	}

	secondStrokeStart := secondGroup.GetStrokeStart(0.1, false)
	assertPoint(t, secondStrokeStart.X, secondStrokeStart.Y, 0, 100)
}
