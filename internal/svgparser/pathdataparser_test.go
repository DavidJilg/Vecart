package svgparser

import (
	"math"
	"testing"

	"github.com/DavidJilg/Vecart/internal/utils"
)

func assertPoint(t *testing.T, pX, pY, expectedX, expectedY float64) {
	t.Helper()
	if math.Abs(pX-expectedX) > 1e-9 || math.Abs(pY-expectedY) > 1e-9 {
		t.Fatalf("unexpected point: got {X:%v, Y:%v}, expected {X:%v, Y:%v}", pX, pY, expectedX, expectedY)
	}
}

func TestParsePathDataString_TriangleWithMixedCommands(t *testing.T) {
	path, err := parsePathDataString("M591.6,648.3h200l-100,200L591.6,648.3z")
	if err != nil {
		t.Fatalf("parsePathDataString failed: %v", err)
	}

	polyline := path.ToPolyline(0.1)
	if len(polyline.Points) != 4 {
		t.Fatalf("expected 4 points, got %d", len(polyline.Points))
	}

	assertPoint(t, polyline.Points[0].X, polyline.Points[0].Y, 591.6, 648.3)
	assertPoint(t, polyline.Points[1].X, polyline.Points[1].Y, 791.6, 648.3)
	assertPoint(t, polyline.Points[2].X, polyline.Points[2].Y, 691.6, 848.3)
	assertPoint(t, polyline.Points[3].X, polyline.Points[3].Y, 591.6, 648.3)
}

func TestParsePathDataString_HVCommandsDoNotAliasPoints(t *testing.T) {
	path, err := parsePathDataString("M10,20h30h40v5v15z")
	if err != nil {
		t.Fatalf("parsePathDataString failed: %v", err)
	}

	polyline := path.ToPolyline(0.1)
	if len(polyline.Points) != 6 {
		t.Fatalf("expected 6 points, got %d", len(polyline.Points))
	}

	assertPoint(t, polyline.Points[0].X, polyline.Points[0].Y, 10, 20)
	assertPoint(t, polyline.Points[1].X, polyline.Points[1].Y, 40, 20)
	assertPoint(t, polyline.Points[2].X, polyline.Points[2].Y, 80, 20)
	assertPoint(t, polyline.Points[3].X, polyline.Points[3].Y, 80, 25)
	assertPoint(t, polyline.Points[4].X, polyline.Points[4].Y, 80, 40)
	assertPoint(t, polyline.Points[5].X, polyline.Points[5].Y, 10, 20)
}

func TestParsePathDataString_PreservesCompoundSubpaths(t *testing.T) {
	path, err := parsePathDataString("M0,0L10,0 M20,0L30,0")
	if err != nil {
		t.Fatalf("parsePathDataString failed: %v", err)
	}

	if path.SubpathCount() != 2 {
		t.Fatalf("expected 2 subpaths, got %d", path.SubpathCount())
	}

	polylines := path.ToPolylines(0.1)
	if len(polylines) != 2 {
		t.Fatalf("expected 2 polylines, got %d", len(polylines))
	}

	assertPoint(t, polylines[0].Points[0].X, polylines[0].Points[0].Y, 0, 0)
	assertPoint(t, polylines[0].Points[1].X, polylines[0].Points[1].Y, 10, 0)
	assertPoint(t, polylines[1].Points[0].X, polylines[1].Points[0].Y, 20, 0)
	assertPoint(t, polylines[1].Points[1].X, polylines[1].Points[1].Y, 30, 0)
}

func TestPathToGCode_DoesNotConnectCompoundSubpaths(t *testing.T) {
	path, err := parsePathDataString("M0,0L10,0 M20,0L30,0")
	if err != nil {
		t.Fatalf("parsePathDataString failed: %v", err)
	}

	gcode := path.ToGCode(nil, nil, 0.1, false, 100)
	if len(gcode) != 4 {
		t.Fatalf("expected 4 gcode lines, got %d", len(gcode))
	}

	if gcode[0] != utils.G00XY(0, 0) {
		t.Fatalf("unexpected first move: %s", gcode[0])
	}
	if gcode[1] != utils.G01XY(10, 0, 100) {
		t.Fatalf("unexpected first cut: %s", gcode[1])
	}
	if gcode[2] != utils.G00XY(20, 0) {
		t.Fatalf("expected rapid move to next subpath, got %s", gcode[2])
	}
	if gcode[3] != utils.G01XY(30, 0, 100) {
		t.Fatalf("unexpected second cut: %s", gcode[3])
	}
}

func TestPathToGCode_KeepsFinalRelativeLineSubpath(t *testing.T) {
	path, err := parsePathDataString("M382,281.4l2-2c0.3-0.3,0.7,0,0.4,0.3l-2.4,2.4 M376.8,286.6l2.8-2.8")
	if err != nil {
		t.Fatalf("parsePathDataString failed: %v", err)
	}

	polylines := path.ToPolylines(0.1)
	if len(polylines) != 2 {
		t.Fatalf("expected 2 polylines, got %d", len(polylines))
	}

	lastPolyline := polylines[len(polylines)-1]
	if len(lastPolyline.Points) != 2 {
		t.Fatalf("expected final polyline to have 2 points, got %d", len(lastPolyline.Points))
	}

	assertPoint(t, lastPolyline.Points[0].X, lastPolyline.Points[0].Y, 376.8, 286.6)
	assertPoint(t, lastPolyline.Points[1].X, lastPolyline.Points[1].Y, 379.6, 283.8)

	gcode := path.ToGCode(nil, nil, 0.1, false, 100)
	if len(gcode) < 2 {
		t.Fatalf("expected gcode for final subpath, got %d lines", len(gcode))
	}

	if gcode[len(gcode)-2] != utils.G00XY(376.8, 286.6) {
		t.Fatalf("unexpected final subpath rapid move: %s", gcode[len(gcode)-2])
	}
	if gcode[len(gcode)-1] != utils.G01XY(379.6, 283.8, 100) {
		t.Fatalf("unexpected final subpath cut: %s", gcode[len(gcode)-1])
	}
}

func TestPathToGCode_FullIllustratorPathEndsWithVisibleSegment(t *testing.T) {
	dString := "M381.9,280.8l3.2-3.2 M379.4,282.5l-3.1,3.1c-0.3,0.3-0.3,1,0.1,0.6l3-3 M380.9,280.4l4.2-4.2 c0.3-0.3,0.3,0.3,0,0.7l-3.5,3.5 M379.4,281.1l-3.1,3.1c-0.3,0.3-0.3,1,0,0.7 l3.1-3.1 M381.1,278.1l-4.7,4.7c-0.3,0.3-0.3,1,0,0.7 l8.6-8.6c0.3-0.3,0.4,0.3,0.1,0.6l-4.9,4.9 M381.9,276.6l2.6-2.6c0.3-0.3,0.6,0.1,0.2,0.5l-2.9,2.9 M379.7,278l-3.4,3.4 c-0.3,0.3-0.3,1,0,0.7l4-4 M381.8,275.3l2.1-2.1c0.3-0.3,0.7,0,0.4,0.3l-2.4,2.4 M379.4,276.9l-3.1,3.1c-0.3,0.3-0.3,1,0,0.7 l3.1-3.1 M380.7,274.9l2.3-2.3c0.3-0.3,0.8-0.1,0.5,0.2l-2.1,2.1 M379.4,275.5l-3.1,3.1c-0.3,0.3-0.3,1,0,0.7l3.1-3.1 M376.1,272.6 l0.2-0.2c0.3-0.3,1-0.3,0.7,0l-0.7,0.7c-0.3,0.3-0.3,1,0,0.7l1.3-1.3c0.3-0.3,1-0.3,0.7,0l-2,2c-0.3,0.3-0.3,1,0,0.7l2.7-2.7 c0.3-0.3,1-0.3,0.7,0l-3.4,3.4c-0.3,0.3-0.3,1,0,0.7l4.1-4.1c0.3-0.3,1-0.3,0.7,0l-4.8,4.8c-0.3,0.3-0.3,1,0,0.7l5.5-5.5 c0.3-0.3,1-0.3,0.7,0L380,275 M382,282.8l2.7-2.7c0.3-0.3,0.6,0.1,0.3,0.4l-5.9,5.9c-0.3,0.3,0.3,0.3,0.7,0l5.4-5.4 c0.3-0.3,0.5,0.2,0.1,0.6l-4.8,4.8c-0.3,0.3,0.3,0.3,0.7,0l4.1-4.1c0.3-0.3,0.3,0.3,0,0.7l-3.4,3.4c-0.3,0.3,0.3,0.3,0.7,0l2.7-2.7 c0.3-0.3,0.2,0.5-0.2,0.9l-2,2 M380.2,283.8l-2.5,2.5c-0.3,0.3,0.3,0.3,0.7,0l2.5-2.5 M382,281.4l2-2c0.3-0.3,0.7,0,0.4,0.3 l-2.4,2.4 M376.8,286.6l2.8-2.8"

	path, err := parsePathDataString(dString)
	if err != nil {
		t.Fatalf("parsePathDataString failed: %v", err)
	}

	gcode := path.ToGCode(nil, nil, 0.1, false, 100)
	if len(gcode) < 2 {
		t.Fatalf("expected gcode lines, got %d", len(gcode))
	}

	if gcode[len(gcode)-2] != utils.G00XY(376.8, 286.6) {
		t.Fatalf("unexpected final rapid move: %s", gcode[len(gcode)-2])
	}
	if gcode[len(gcode)-1] != utils.G01XY(379.6, 283.8, 100) {
		t.Fatalf("unexpected final cut line: %s", gcode[len(gcode)-1])
	}
}

func TestPathToGCode_SkipsMoveOnlySubpaths(t *testing.T) {
	path, err := parsePathDataString("M0,0L10,0 M20,20 M30,30L40,30")
	if err != nil {
		t.Fatalf("parsePathDataString failed: %v", err)
	}

	polylines := path.ToPolylines(0.1)
	if len(polylines) != 2 {
		t.Fatalf("expected 2 drawable polylines, got %d", len(polylines))
	}

	gcode := path.ToGCode(nil, nil, 0.1, false, 100)
	if len(gcode) != 4 {
		t.Fatalf("expected 4 gcode lines, got %d", len(gcode))
	}

	if gcode[0] != utils.G00XY(0, 0) {
		t.Fatalf("unexpected first rapid move: %s", gcode[0])
	}
	if gcode[1] != utils.G01XY(10, 0, 100) {
		t.Fatalf("unexpected first cut line: %s", gcode[1])
	}
	if gcode[2] != utils.G00XY(30, 30) {
		t.Fatalf("unexpected second rapid move: %s", gcode[2])
	}
	if gcode[3] != utils.G01XY(40, 30, 100) {
		t.Fatalf("unexpected second cut line: %s", gcode[3])
	}
}
