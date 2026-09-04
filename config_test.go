package main

import (
	"testing"

	"github.com/DavidJilg/Vecart/internal/utils"

	"github.com/DavidJilg/Vecart/internal/general"
	"github.com/DavidJilg/Vecart/internal/shapes"
)

func initForTests() {
	utils.InitFlags()
	general.ResetStaticVariables()
	general.InitStaticAssets(StaticAssets, License)
	utils.EnableTestMode()
	general.SetFonts(general.LoadFonts())
	utils.SetVersion(Version)
}

func TestEmptyConfigJson(t *testing.T) {
	initForTests()

	baseConfig := general.NewConfig()

	config := general.NewConfig()
	config.FromJSON("", "")

	if !config.EqualTo(&baseConfig, true) {
		t.Error("Reading Empty Config failed!")
	}
}

func TestAllConfigJson(t *testing.T) {
	initForTests()

	baseConfig := general.NewConfig()
	baseConfig.Mode = general.GCode
	baseConfig.InputPath = "static/configs/proved/some/path/picture.png"
	baseConfig.OutputPath = "static/configs/proved/some/path/art.svg"
	baseConfig.ArtworkWidth = 1
	baseConfig.ArtworkHeight = 2
	baseConfig.QuadrantWidth = 3
	baseConfig.QuadrantHeight = 4
	baseConfig.DarknessThreshold = 5
	baseConfig.ShapeDarknessFactor = 6
	baseConfig.WhitePunishmentBoundry = 7
	baseConfig.WhitePunishmentValue = 8.5
	baseConfig.RandomSeed = 9
	baseConfig.ParallelRoutines = 10
	baseConfig.UpdateFrequency = 10
	baseConfig.ShapeRefinementIterations = 16
	baseConfig.ShapeRefinementPercentage = 16.5
	baseConfig.ShapeRefinement = false
	baseConfig.HighPrecisionShapePositioning = true
	baseConfig.SmoothEdges = false
	baseConfig.CombineShapes = false
	baseConfig.CombineShapesTolerance = 11.5
	baseConfig.CombineShapesIterations = 12
	baseConfig.StrokeWidth = 13.5
	baseConfig.StrokeColor = "red"
	baseConfig.BackgroundColor = "white"
	baseConfig.ReverseShapeOrder = true
	baseConfig.ConfigInOutput = false
	baseConfig.ShortConfig = false
	baseConfig.StatsInOutput = false
	baseConfig.ProcessingDpi = 14
	baseConfig.OutputDpi = 15
	baseConfig.Timeout = 20
	baseConfig.OverwriteExisting = true

	baseConfig.FeedRateXY = 99
	baseConfig.FeedRateZ = 99
	baseConfig.SortGCode = false
	baseConfig.KeepSVGGroups = false
	baseConfig.SortAxis = 0
	baseConfig.BezierTolerance = 0.055

	baseConfig.StampMode = true
	baseConfig.NrOfShapesWithoutRefill = 5
	baseConfig.RefillTool = true
	baseConfig.RefillToolCommands = []string{"G0 X5 Y5", "G0 X5 Y5"}

	baseConfig.ActivateToolCommands = []string{"G1 Z5", "G1 Z5"}
	baseConfig.DeactivateToolCommands = []string{"G1 Z2", "G1 Z2"}
	baseConfig.InitialCommands = []string{"G1 Z1", "G1 Z1"}
	baseConfig.FinalCommands = []string{"G1 Z3", "G1 Z3"}

	baseConfig.ShapeList = nil

	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewLine(shapes.NewPoint(0, 0), shapes.NewPoint(0, 2)))
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewLine(shapes.NewPoint(0, 0), shapes.NewPoint(0, 4)))
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewLine(shapes.NewPoint(0, 0), shapes.NewPoint(0, 8)))

	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewSingleLineShape(*shapes.NewPolyline(&[]shapes.Point{{X: 0, Y: 0}, {X: 0, Y: 2}, {X: 1, Y: 2}}, nil)))

	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewPolygon(&[]shapes.Point{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 2}}).ToShape())
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewPolygon(&[]shapes.Point{{X: 0, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 4}}).ToShape())
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewPolygon(&[]shapes.Point{{X: 0, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 8}}).ToShape())

	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewPolygon(&[]shapes.Point{{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 2}}).ToShape())
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewPolygon(&[]shapes.Point{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}}).ToShape())
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewPolygon(&[]shapes.Point{{X: 0, Y: 0}, {X: 8, Y: 0}, {X: 8, Y: 8}, {X: 0, Y: 8}}).ToShape())

	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewCircle(*shapes.NewPoint(0, 0), 1).ToShape())
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewCircle(*shapes.NewPoint(0, 0), 2).ToShape())
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewCircle(*shapes.NewPoint(0, 0), 4).ToShape())

	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewPolygon(&[]shapes.Point{{X: -1, Y: 0}, {X: -0.5, Y: 1}, {X: 0.5, Y: 1}, {X: 1, Y: 0}}).ToShape())
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewPolygon(&[]shapes.Point{{X: -2, Y: 0}, {X: -1, Y: 2}, {X: 1, Y: 2}, {X: 2, Y: 0}}).ToShape())
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewPolygon(&[]shapes.Point{{X: -3, Y: 0}, {X: -2, Y: 3}, {X: 2, Y: 3}, {X: 3, Y: 0}}).ToShape())

	var Fonts = general.LoadFonts()

	baseConfig.ShapeList = append(baseConfig.ShapeList, Fonts["IBM-Plex-Sans"].GetText("Test", 1, *shapes.NewPoint(0, 0)))
	baseConfig.ShapeList = append(baseConfig.ShapeList, Fonts["IBM-Plex-Sans"].GetText("Test", 2, *shapes.NewPoint(0, 0)))
	baseConfig.ShapeList = append(baseConfig.ShapeList, Fonts["IBM-Plex-Sans"].GetText("Test", 3, *shapes.NewPoint(0, 0)))

	var group []shapes.Shape
	group = append(group, *shapes.NewLine(shapes.NewPoint(0, 0), shapes.NewPoint(0, 2)))
	group = append(group, *shapes.NewSingleLineShape(*shapes.NewPolyline(&[]shapes.Point{{X: 0, Y: 0}, {X: 0, Y: 2}, {X: 1, Y: 2}}, nil)))
	group = append(group, *shapes.NewPolygon(&[]shapes.Point{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 2}}).ToShape())
	group = append(group, *shapes.NewPolygon(&[]shapes.Point{{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 2}}).ToShape())
	group = append(group, *shapes.NewPolygon(&[]shapes.Point{{X: -1, Y: 0}, {X: -0.5, Y: 1}, {X: 0.5, Y: 1}, {X: 1, Y: 0}}).ToShape())
	group = append(group, *shapes.NewCircle(*shapes.NewPoint(0, 0), 1).ToShape())
	baseConfig.ShapeList = append(baseConfig.ShapeList, *general.CombineShapes(&group))

	baseConfig.ShapeAngleDeviationRange = 16
	baseConfig.ShapeAngleDeviationStep = 17.5

	allConfigFile, err := StaticAssets.Open("static/configs/proved/all.json")
	if err != nil {
		t.Error("Reading all.json from static assets failed!")
	}

	content, err := utils.GetFileContentsFromStaticAssets(allConfigFile)

	if err != nil {
		t.Error("Reading content from all.json from static assets failed!")
	}

	allConfig := general.NewConfig()
	allConfig.FromJSON(content, "static/configs/proved/all.json")

	if !allConfig.EqualTo(&baseConfig, true) {
		t.Error("Parsing all.json to config failed!")
	}
}

func TestLineConfigJson(t *testing.T) {
	initForTests()

	baseConfig := general.NewConfig()
	baseConfig.ShapeList = nil
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewLine(shapes.NewPoint(0, 0), shapes.NewPoint(0, 2)))
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewLine(shapes.NewPoint(0, 0), shapes.NewPoint(0, 4)))
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewLine(shapes.NewPoint(0, 0), shapes.NewPoint(0, 8)))

	testConfigFile, err := StaticAssets.Open("static/configs/proved/lines.json")
	if err != nil {
		t.Error("Reading all.json from static assets failed!")
	}

	content, err := utils.GetFileContentsFromStaticAssets(testConfigFile)

	if err != nil {
		t.Error("Reading content from all.json from static assets failed!")
	}

	testConfig := general.NewConfig()
	testConfig.FromJSON(content, "static/configs/proved/lines.json")

	if !shapes.ShapesEqual(&baseConfig.ShapeList, &testConfig.ShapeList, 5, true) {
		t.Error("Parsing line.json to config failed!")
	}
}

func TestPolylineConfigJson(t *testing.T) {
	initForTests()

	baseConfig := general.NewConfig()
	baseConfig.ShapeList = nil
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewSingleLineShape(*shapes.NewPolyline(&[]shapes.Point{{X: 0, Y: 0}, {X: 0, Y: 2}, {X: 1, Y: 2}}, nil)))
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewSingleLineShape(*shapes.NewPolyline(&[]shapes.Point{{X: 0, Y: 0}, {X: 0, Y: 2}, {X: 1, Y: -2}}, nil)))
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewSingleLineShape(*shapes.NewPolyline(&[]shapes.Point{{X: 0, Y: 0}, {X: 0, Y: 4}, {X: 2, Y: 4}}, nil)))
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewSingleLineShape(*shapes.NewPolyline(&[]shapes.Point{{X: 0, Y: 0}, {X: 0, Y: 4}, {X: 2, Y: -4}}, nil)))
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewSingleLineShape(*shapes.NewPolyline(&[]shapes.Point{{X: 0, Y: 0}, {X: 0, Y: 8}, {X: 4, Y: 8}}, nil)))
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewSingleLineShape(*shapes.NewPolyline(&[]shapes.Point{{X: 0, Y: 0}, {X: 0, Y: 8}, {X: 4, Y: -8}}, nil)))

	testConfigFile, err := StaticAssets.Open("static/configs/proved/polylines.json")
	if err != nil {
		t.Error("Reading polylines.json from static assets failed!")
	}

	content, err := utils.GetFileContentsFromStaticAssets(testConfigFile)

	if err != nil {
		t.Error("Reading content from polylines.json from static assets failed!")
	}

	testConfig := general.NewConfig()
	testConfig.FromJSON(content, "static/configs/proved/polylines.json")

	if !shapes.ShapesEqual(&baseConfig.ShapeList, &testConfig.ShapeList, 5, true) {
		t.Error("Parsing polylines.json to config failed!")
	}
}

func TestTriangleConfigJson(t *testing.T) {
	initForTests()

	baseConfig := general.NewConfig()
	baseConfig.ShapeList = nil
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewPolygon(&[]shapes.Point{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 2}}).ToShape())
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewPolygon(&[]shapes.Point{{X: 0, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 4}}).ToShape())
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewPolygon(&[]shapes.Point{{X: 0, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 8}}).ToShape())

	testConfigFile, err := StaticAssets.Open("static/configs/proved/triangles.json")
	if err != nil {
		t.Error("Reading triangles.json from static assets failed!")
	}

	content, err := utils.GetFileContentsFromStaticAssets(testConfigFile)

	if err != nil {
		t.Error("Reading content from triangles.json from static assets failed!")
	}

	testConfig := general.NewConfig()
	testConfig.FromJSON(content, "static/configs/proved/triangles.json")

	if !shapes.ShapesEqual(&baseConfig.ShapeList, &testConfig.ShapeList, 5, true) {
		t.Error("Parsing triangles.json to config failed!")
	}
}

func TestRectanglesConfigJson(t *testing.T) {
	initForTests()

	baseConfig := general.NewConfig()
	baseConfig.ShapeList = nil
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewPolygon(&[]shapes.Point{{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 2}}).ToShape())
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewPolygon(&[]shapes.Point{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}}).ToShape())
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewPolygon(&[]shapes.Point{{X: 0, Y: 0}, {X: 8, Y: 0}, {X: 8, Y: 8}, {X: 0, Y: 8}}).ToShape())

	testConfigFile, err := StaticAssets.Open("static/configs/proved/rectangles.json")
	if err != nil {
		t.Error("Reading rectangles.json from static assets failed!")
	}

	content, err := utils.GetFileContentsFromStaticAssets(testConfigFile)

	if err != nil {
		t.Error("Reading content from rectangles.json from static assets failed!")
	}

	testConfig := general.NewConfig()
	testConfig.FromJSON(content, "static/configs/proved/rectangles.json")

	if !shapes.ShapesEqual(&baseConfig.ShapeList, &testConfig.ShapeList, 5, true) {
		t.Error("Parsing rectangles.json to config failed!")
	}
}

func TestCirclesConfigJson(t *testing.T) {
	initForTests()

	baseConfig := general.NewConfig()
	baseConfig.ShapeList = nil
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewCircle(*shapes.NewPoint(0, 0), 1).ToShape())
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewCircle(*shapes.NewPoint(0, 0), 2).ToShape())
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewCircle(*shapes.NewPoint(0, 0), 4).ToShape())

	testConfigFile, err := StaticAssets.Open("static/configs/proved/circles.json")
	if err != nil {
		t.Error("Reading circles.json from static assets failed!")
	}

	content, err := utils.GetFileContentsFromStaticAssets(testConfigFile)

	if err != nil {
		t.Error("Reading content from circles.json from static assets failed!")
	}

	testConfig := general.NewConfig()
	testConfig.FromJSON(content, "static/configs/proved/circles.json")

	if !shapes.ShapesEqual(&baseConfig.ShapeList, &testConfig.ShapeList, 5, true) {
		t.Error("Parsing circles.json to config failed!")
	}
}

func TestPolygonsConfigJson(t *testing.T) {
	initForTests()

	baseConfig := general.NewConfig()
	baseConfig.ShapeList = nil
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewPolygon(&[]shapes.Point{{X: -1, Y: 0}, {X: -0.5, Y: 1}, {X: 0.5, Y: 1}, {X: 1, Y: 0}}).ToShape())
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewPolygon(&[]shapes.Point{{X: -2, Y: 0}, {X: -1, Y: 2}, {X: 1, Y: 2}, {X: 2, Y: 0}}).ToShape())
	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewPolygon(&[]shapes.Point{{X: -3, Y: 0}, {X: -2, Y: 3}, {X: 2, Y: 3}, {X: 3, Y: 0}}).ToShape())

	testConfigFile, err := StaticAssets.Open("static/configs/proved/polygons.json")
	if err != nil {
		t.Error("Reading polygons.json from static assets failed!")
	}

	content, err := utils.GetFileContentsFromStaticAssets(testConfigFile)

	if err != nil {
		t.Error("Reading content from polygons.json from static assets failed!")
	}

	testConfig := general.NewConfig()
	testConfig.FromJSON(content, "static/configs/proved/polygons.json")

	if !shapes.ShapesEqual(&baseConfig.ShapeList, &testConfig.ShapeList, 5, true) {
		t.Error("Parsing polygons.json to config failed!")
	}
}

func TestGroupConfigJson(t *testing.T) {
	initForTests()

	baseConfig := general.NewConfig()
	baseConfig.ShapeList = nil

	var lines []shapes.Polyline
	lines = append(lines, *shapes.NewPolyline(&[]shapes.Point{{X: 0, Y: 0}, {X: 0, Y: 2}}, nil))                                               //Line
	lines = append(lines, *shapes.NewPolyline(&[]shapes.Point{{X: 0, Y: 0}, {X: 0, Y: 2}, {X: 1, Y: 2}}, nil))                                 //Polyline
	lines = append(lines, shapes.NewPolygon(&[]shapes.Point{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 2}}).ToShape().Lines[0])                     //Triangle
	lines = append(lines, shapes.NewPolygon(&[]shapes.Point{{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 2}}).ToShape().Lines[0])       //Rectangle
	lines = append(lines, shapes.NewPolygon(&[]shapes.Point{{X: -1, Y: 0}, {X: -0.5, Y: 1}, {X: 0.5, Y: 1}, {X: 1, Y: 0}}).ToShape().Lines[0]) //Polygon
	lines = append(lines, shapes.NewCircle(*shapes.NewPoint(0, 0), 1).ToShape().Lines[0])                                                      //Circle

	baseConfig.ShapeList = append(baseConfig.ShapeList, *shapes.NewShape(lines))

	testConfigFile, err := StaticAssets.Open("static/configs/proved/group.json")
	if err != nil {
		t.Error("Reading group.json from static assets failed!")
	}

	content, err := utils.GetFileContentsFromStaticAssets(testConfigFile)

	if err != nil {
		t.Error("Reading content from group.json from static assets failed!")
	}

	testConfig := general.NewConfig()
	testConfig.FromJSON(content, "static/configs/proved/group.json")

	if !shapes.ShapesEqual(&baseConfig.ShapeList, &testConfig.ShapeList, 5, true) {
		t.Error("Parsing group.json to config failed!")
	}
}
