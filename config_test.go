package main

import "testing"

func TestEmptyConfigJson(t *testing.T) {
	baseConfig := NewConfig()

	config := NewConfig()
	config.fromJSON("")

	if !config.equalTo(&baseConfig) {
		t.Error("Reading Empty Config failed!")
	}
}

func TestAllConfigJson(t *testing.T) {
	baseConfig := NewConfig()
	baseConfig.inputPath = "/some/path/picture.png"
	baseConfig.outputPath = "/some/path/art.svg"
	baseConfig.artworkWidth = 1
	baseConfig.artworkHeight = 2
	baseConfig.quadrantWidth = 3
	baseConfig.quadrantHeight = 4
	baseConfig.darknessThreshold = 5
	baseConfig.shapeDarknessFactor = 6
	baseConfig.whitePunishmentBoundry = 7
	baseConfig.whitePunishmentValue = 8.5
	baseConfig.randomSeed = 9
	baseConfig.parallelRoutines = 10
	baseConfig.shapeRefinementIterations = 16
	baseConfig.shapeRefinementPercentage = 16.5
	baseConfig.shapeRefinement = false
	baseConfig.highPrecisionShapePositioning = true
	baseConfig.smoothEdges = false
	baseConfig.combineShapes = false
	baseConfig.combineShapesTolerance = 11.5
	baseConfig.combineShapesIterations = 12
	baseConfig.strokeWidth = 13.5
	baseConfig.strokeColor = "red"
	baseConfig.reverseShapeOrder = true
	baseConfig.configInOutput = false
	baseConfig.processingDpi = 14
	baseConfig.outputDpi = 15
	baseConfig.timeout = 20
	baseConfig.debug = true
	baseConfig.shapes = nil

	baseConfig.shapes = append(baseConfig.shapes, *NewLine(NewPoint(0, 0), NewPoint(0, 2)))
	baseConfig.shapes = append(baseConfig.shapes, *NewLine(NewPoint(0, 0), NewPoint(0, 4)))
	baseConfig.shapes = append(baseConfig.shapes, *NewLine(NewPoint(0, 0), NewPoint(0, 8)))

	baseConfig.shapes = append(baseConfig.shapes, *NewSingleLineShape(*NewPolyline(&[]Point{{0, 0}, {0, 2}, {1, 2}}, nil)))

	baseConfig.shapes = append(baseConfig.shapes, *NewPolygon(&[]Point{{0, 0}, {1, 1}, {0, 2}}).toShape())
	baseConfig.shapes = append(baseConfig.shapes, *NewPolygon(&[]Point{{0, 0}, {2, 2}, {0, 4}}).toShape())
	baseConfig.shapes = append(baseConfig.shapes, *NewPolygon(&[]Point{{0, 0}, {4, 4}, {0, 8}}).toShape())

	baseConfig.shapes = append(baseConfig.shapes, *NewPolygon(&[]Point{{0, 0}, {2, 0}, {2, 2}, {0, 2}}).toShape())
	baseConfig.shapes = append(baseConfig.shapes, *NewPolygon(&[]Point{{0, 0}, {4, 0}, {4, 4}, {0, 4}}).toShape())
	baseConfig.shapes = append(baseConfig.shapes, *NewPolygon(&[]Point{{0, 0}, {8, 0}, {8, 8}, {0, 8}}).toShape())

	baseConfig.shapes = append(baseConfig.shapes, *NewCircle(*NewPoint(0, 0), 1).toShape())
	baseConfig.shapes = append(baseConfig.shapes, *NewCircle(*NewPoint(0, 0), 2).toShape())
	baseConfig.shapes = append(baseConfig.shapes, *NewCircle(*NewPoint(0, 0), 4).toShape())

	baseConfig.shapes = append(baseConfig.shapes, *NewPolygon(&[]Point{{-1, 0}, {-0.5, 1}, {0.5, 1}, {1, 0}}).toShape())
	baseConfig.shapes = append(baseConfig.shapes, *NewPolygon(&[]Point{{-2, 0}, {-1, 2}, {1, 2}, {2, 0}}).toShape())
	baseConfig.shapes = append(baseConfig.shapes, *NewPolygon(&[]Point{{-3, 0}, {-2, 3}, {2, 3}, {3, 0}}).toShape())

	Fonts = loadFonts()

	baseConfig.shapes = append(baseConfig.shapes, Fonts["IBM-Plex-Sans"].getText("Test", 1, *NewPoint(0, 0)))
	baseConfig.shapes = append(baseConfig.shapes, Fonts["IBM-Plex-Sans"].getText("Test", 2, *NewPoint(0, 0)))
	baseConfig.shapes = append(baseConfig.shapes, Fonts["IBM-Plex-Sans"].getText("Test", 3, *NewPoint(0, 0)))

	var group []Shape
	group = append(group, *NewLine(NewPoint(0, 0), NewPoint(0, 2)))
	group = append(group, *NewSingleLineShape(*NewPolyline(&[]Point{{0, 0}, {0, 2}, {1, 2}}, nil)))
	group = append(group, *NewPolygon(&[]Point{{0, 0}, {1, 1}, {0, 2}}).toShape())
	group = append(group, *NewPolygon(&[]Point{{0, 0}, {2, 0}, {2, 2}, {0, 2}}).toShape())
	group = append(group, *NewPolygon(&[]Point{{-1, 0}, {-0.5, 1}, {0.5, 1}, {1, 0}}).toShape())
	group = append(group, *NewCircle(*NewPoint(0, 0), 1).toShape())
	baseConfig.shapes = append(baseConfig.shapes, *combineShapes(&group))

	baseConfig.shapeAngleDeviationRange = 16
	baseConfig.shapeAngleDeviationStep = 17.5

	allConfigFile, err := StaticAssets.Open("static/configs/proved/all.json")
	if err != nil {
		t.Error("Reading all.json from static assets failed!")
	}

	content, err := getFileContentsFromStaticAssets(allConfigFile)

	if err != nil {
		t.Error("Reading content from all.json from static assets failed!")
	}

	allConfig := NewConfig()
	allConfig.fromJSON(content)

	if !allConfig.equalTo(&baseConfig) {
		t.Error("Parsing all.json to config failed!")
	}
}

func TestLineConfigJson(t *testing.T) {
	baseConfig := NewConfig()
	baseConfig.shapes = nil
	baseConfig.shapes = append(baseConfig.shapes, *NewLine(NewPoint(0, 0), NewPoint(0, 2)))
	baseConfig.shapes = append(baseConfig.shapes, *NewLine(NewPoint(0, 0), NewPoint(0, 4)))
	baseConfig.shapes = append(baseConfig.shapes, *NewLine(NewPoint(0, 0), NewPoint(0, 8)))

	testConfigFile, err := StaticAssets.Open("static/configs/proved/lines.json")
	if err != nil {
		t.Error("Reading all.json from static assets failed!")
	}

	content, err := getFileContentsFromStaticAssets(testConfigFile)

	if err != nil {
		t.Error("Reading content from all.json from static assets failed!")
	}

	testConfig := NewConfig()
	testConfig.fromJSON(content)

	if !shapesEqual(&baseConfig.shapes, &testConfig.shapes, 5, true) {
		t.Error("Parsing line.json to config failed!")
	}
}

func TestPolylineConfigJson(t *testing.T) {
	baseConfig := NewConfig()
	baseConfig.shapes = nil
	baseConfig.shapes = append(baseConfig.shapes, *NewSingleLineShape(*NewPolyline(&[]Point{{0, 0}, {0, 2}, {1, 2}}, nil)))
	baseConfig.shapes = append(baseConfig.shapes, *NewSingleLineShape(*NewPolyline(&[]Point{{0, 0}, {0, 2}, {1, -2}}, nil)))
	baseConfig.shapes = append(baseConfig.shapes, *NewSingleLineShape(*NewPolyline(&[]Point{{0, 0}, {0, 4}, {2, 4}}, nil)))
	baseConfig.shapes = append(baseConfig.shapes, *NewSingleLineShape(*NewPolyline(&[]Point{{0, 0}, {0, 4}, {2, -4}}, nil)))
	baseConfig.shapes = append(baseConfig.shapes, *NewSingleLineShape(*NewPolyline(&[]Point{{0, 0}, {0, 8}, {4, 8}}, nil)))
	baseConfig.shapes = append(baseConfig.shapes, *NewSingleLineShape(*NewPolyline(&[]Point{{0, 0}, {0, 8}, {4, -8}}, nil)))

	testConfigFile, err := StaticAssets.Open("static/configs/proved/polylines.json")
	if err != nil {
		t.Error("Reading polylines.json from static assets failed!")
	}

	content, err := getFileContentsFromStaticAssets(testConfigFile)

	if err != nil {
		t.Error("Reading content from polylines.json from static assets failed!")
	}

	testConfig := NewConfig()
	testConfig.fromJSON(content)

	if !shapesEqual(&baseConfig.shapes, &testConfig.shapes, 5, true) {
		t.Error("Parsing polylines.json to config failed!")
	}
}

func TestTriangleConfigJson(t *testing.T) {
	baseConfig := NewConfig()
	baseConfig.shapes = nil
	baseConfig.shapes = append(baseConfig.shapes, *NewPolygon(&[]Point{{0, 0}, {1, 1}, {0, 2}}).toShape())
	baseConfig.shapes = append(baseConfig.shapes, *NewPolygon(&[]Point{{0, 0}, {2, 2}, {0, 4}}).toShape())
	baseConfig.shapes = append(baseConfig.shapes, *NewPolygon(&[]Point{{0, 0}, {4, 4}, {0, 8}}).toShape())

	testConfigFile, err := StaticAssets.Open("static/configs/proved/triangles.json")
	if err != nil {
		t.Error("Reading triangles.json from static assets failed!")
	}

	content, err := getFileContentsFromStaticAssets(testConfigFile)

	if err != nil {
		t.Error("Reading content from triangles.json from static assets failed!")
	}

	testConfig := NewConfig()
	testConfig.fromJSON(content)

	if !shapesEqual(&baseConfig.shapes, &testConfig.shapes, 5, true) {
		t.Error("Parsing triangles.json to config failed!")
	}
}

func TestRectanglesConfigJson(t *testing.T) {
	baseConfig := NewConfig()
	baseConfig.shapes = nil
	baseConfig.shapes = append(baseConfig.shapes, *NewPolygon(&[]Point{{0, 0}, {2, 0}, {2, 2}, {0, 2}}).toShape())
	baseConfig.shapes = append(baseConfig.shapes, *NewPolygon(&[]Point{{0, 0}, {4, 0}, {4, 4}, {0, 4}}).toShape())
	baseConfig.shapes = append(baseConfig.shapes, *NewPolygon(&[]Point{{0, 0}, {8, 0}, {8, 8}, {0, 8}}).toShape())

	testConfigFile, err := StaticAssets.Open("static/configs/proved/rectangles.json")
	if err != nil {
		t.Error("Reading rectangles.json from static assets failed!")
	}

	content, err := getFileContentsFromStaticAssets(testConfigFile)

	if err != nil {
		t.Error("Reading content from rectangles.json from static assets failed!")
	}

	testConfig := NewConfig()
	testConfig.fromJSON(content)

	if !shapesEqual(&baseConfig.shapes, &testConfig.shapes, 5, true) {
		t.Error("Parsing rectangles.json to config failed!")
	}
}

func TestCirclesConfigJson(t *testing.T) {
	baseConfig := NewConfig()
	baseConfig.shapes = nil
	baseConfig.shapes = append(baseConfig.shapes, *NewCircle(*NewPoint(0, 0), 1).toShape())
	baseConfig.shapes = append(baseConfig.shapes, *NewCircle(*NewPoint(0, 0), 2).toShape())
	baseConfig.shapes = append(baseConfig.shapes, *NewCircle(*NewPoint(0, 0), 4).toShape())

	testConfigFile, err := StaticAssets.Open("static/configs/proved/circles.json")
	if err != nil {
		t.Error("Reading circles.json from static assets failed!")
	}

	content, err := getFileContentsFromStaticAssets(testConfigFile)

	if err != nil {
		t.Error("Reading content from circles.json from static assets failed!")
	}

	testConfig := NewConfig()
	testConfig.fromJSON(content)

	if !shapesEqual(&baseConfig.shapes, &testConfig.shapes, 5, true) {
		t.Error("Parsing circles.json to config failed!")
	}
}

func TestPolygonsConfigJson(t *testing.T) {
	baseConfig := NewConfig()
	baseConfig.shapes = nil
	baseConfig.shapes = append(baseConfig.shapes, *NewPolygon(&[]Point{{-1, 0}, {-0.5, 1}, {0.5, 1}, {1, 0}}).toShape())
	baseConfig.shapes = append(baseConfig.shapes, *NewPolygon(&[]Point{{-2, 0}, {-1, 2}, {1, 2}, {2, 0}}).toShape())
	baseConfig.shapes = append(baseConfig.shapes, *NewPolygon(&[]Point{{-3, 0}, {-2, 3}, {2, 3}, {3, 0}}).toShape())

	testConfigFile, err := StaticAssets.Open("static/configs/proved/polygons.json")
	if err != nil {
		t.Error("Reading polygons.json from static assets failed!")
	}

	content, err := getFileContentsFromStaticAssets(testConfigFile)

	if err != nil {
		t.Error("Reading content from polygons.json from static assets failed!")
	}

	testConfig := NewConfig()
	testConfig.fromJSON(content)

	if !shapesEqual(&baseConfig.shapes, &testConfig.shapes, 5, true) {
		t.Error("Parsing polygons.json to config failed!")
	}
}

func TestGroupConfigJson(t *testing.T) {
	baseConfig := NewConfig()
	baseConfig.shapes = nil

	var lines []Polyline
	lines = append(lines, *NewPolyline(&[]Point{{0, 0}, {0, 2}}, nil))                                   //Line
	lines = append(lines, *NewPolyline(&[]Point{{0, 0}, {0, 2}, {1, 2}}, nil))                           //Polyline
	lines = append(lines, NewPolygon(&[]Point{{0, 0}, {1, 1}, {0, 2}}).toShape().Lines[0])               //Triangle
	lines = append(lines, NewPolygon(&[]Point{{0, 0}, {2, 0}, {2, 2}, {0, 2}}).toShape().Lines[0])       //Rectangle
	lines = append(lines, NewPolygon(&[]Point{{-1, 0}, {-0.5, 1}, {0.5, 1}, {1, 0}}).toShape().Lines[0]) //Polygon
	lines = append(lines, NewCircle(*NewPoint(0, 0), 1).toShape().Lines[0])                              //Circle

	baseConfig.shapes = append(baseConfig.shapes, *NewShape(lines))

	testConfigFile, err := StaticAssets.Open("static/configs/proved/group.json")
	if err != nil {
		t.Error("Reading group.json from static assets failed!")
	}

	content, err := getFileContentsFromStaticAssets(testConfigFile)

	if err != nil {
		t.Error("Reading content from group.json from static assets failed!")
	}

	testConfig := NewConfig()
	testConfig.fromJSON(content)

	if !shapesEqual(&baseConfig.shapes, &testConfig.shapes, 5, true) {
		t.Error("Parsing group.json to config failed!")
	}
}
