package svgparser

import (
	"testing"

	"github.com/DavidJilg/Vecart/internal/utils"
)

func init() {
	utils.InitFlags()
	utils.DisableDebugMode()
}

func TestExtractShapesIgnoreHiddenFalseKeepsHiddenShapes(t *testing.T) {
	svgText := `<?xml version="1.0" encoding="utf-8"?>
<svg version="1.1" xmlns="http://www.w3.org/2000/svg">
	<style type="text/css">
		.st0{display:none;fill:none;stroke:#E30613;stroke-miterlimit:10;}
		.st1{fill:none;stroke:#E30613;stroke-miterlimit:10;}
	</style>
	<rect class="st0" width="1847.5" height="1425.8"/>
	<rect style="display:none" x="852.9" y="642" class="st1" width="141.7" height="141.7"/>
</svg>`

	root, err := ParseXMLTree(svgText)
	if err != nil {
		t.Fatalf("ParseXMLTree failed: %v", err)
	}

	svg := ExtractShapes(root, false, false)
	if len(svg.Shapes) != 2 {
		t.Fatalf("expected 2 shapes when ignoreHidden=false, got %d", len(svg.Shapes))
	}
}

func TestExtractShapesIgnoreHiddenTrueSkipsHiddenShapes(t *testing.T) {
	svgText := `<?xml version="1.0" encoding="utf-8"?>
<svg version="1.1" xmlns="http://www.w3.org/2000/svg">
	<style type="text/css">
		.st0{display:none;fill:none;stroke:#E30613;stroke-miterlimit:10;}
		.st1{fill:none;stroke:#E30613;stroke-miterlimit:10;}
	</style>
	<rect class="st0" width="1847.5" height="1425.8"/>
	<rect style="display:none" x="852.9" y="642" class="st1" width="141.7" height="141.7"/>
	<rect class="st1" x="10" y="20" width="30" height="40"/>
</svg>`

	root, err := ParseXMLTree(svgText)
	if err != nil {
		t.Fatalf("ParseXMLTree failed: %v", err)
	}

	svg := ExtractShapes(root, true, false)
	if len(svg.Shapes) != 1 {
		t.Fatalf("expected 1 visible shape when ignoreHidden=true, got %d", len(svg.Shapes))
	}
}

func TestExtractShapesIgnoreHiddenTrueSkipsHiddenGroup(t *testing.T) {
	svgText := `<?xml version="1.0" encoding="utf-8"?>
<svg version="1.1" xmlns="http://www.w3.org/2000/svg">
	<g style="display:none">
		<rect x="0" y="0" width="10" height="10"/>
	</g>
	<rect x="20" y="20" width="5" height="5"/>
</svg>`

	root, err := ParseXMLTree(svgText)
	if err != nil {
		t.Fatalf("ParseXMLTree failed: %v", err)
	}

	svg := ExtractShapes(root, true, false)
	if len(svg.Shapes) != 1 {
		t.Fatalf("expected 1 visible shape when hidden group is ignored, got %d", len(svg.Shapes))
	}
}
