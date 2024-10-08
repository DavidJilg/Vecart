package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

var errorsOcurred bool
var errors []string
var debug bool

type VecartConfig struct {
	inputPath     string
	outputPath    string
	artworkWidth  int
	artworkHeight int

	quadrantWidth          int
	quadrantHeight         int
	darknessThreshold      float64
	shapeDarknessFactor    float64
	whitePunishmentBoundry int
	whitePunishmentValue   float64
	randomSeed             int
	parallelRoutines       int

	highPrecisionShapePositioning bool
	shapeRefinement               bool
	shapeRefinementIterations     int
	shapeRefinementPercentage     float64
	smoothEdges                   bool
	combineShapes                 bool
	combineShapesTolerance        float64
	combineShapesIterations       int
	strokeWidth                   float64
	strokeColor                   string
	reverseShapeOrder             bool
	configInOutput                bool
	processingDpi                 float64
	outputDpi                     float64
	debug                         bool
	timeout                       int

	shapes                   []Shape
	shapeAngleDeviationRange float64
	shapeAngleDeviationStep  float64
}

func NewConfig() VecartConfig {
	var config VecartConfig

	config.inputPath = ""
	config.outputPath = "output.svg"
	config.artworkWidth = 255
	config.artworkHeight = 370

	config.quadrantWidth = 5
	config.quadrantHeight = 5
	config.darknessThreshold = 18
	config.shapeDarknessFactor = 40
	config.whitePunishmentBoundry = 5
	config.whitePunishmentValue = 0.85
	config.randomSeed = 1701
	config.parallelRoutines = 5

	config.highPrecisionShapePositioning = false
	config.shapeRefinement = true
	config.shapeRefinementIterations = 1
	config.shapeRefinementPercentage = 0.2
	config.smoothEdges = true
	config.combineShapes = true
	config.combineShapesTolerance = 0.5
	config.combineShapesIterations = 5
	config.strokeWidth = 0.75
	config.strokeColor = "black"
	config.reverseShapeOrder = false
	config.configInOutput = true
	config.processingDpi = 25
	config.outputDpi = 72
	config.debug = false
	config.timeout = 30

	config.shapes = append(Config.shapes, *NewShape([]Polyline{{[]Point{{0, 0}, {0, 2}}, nil}}))
	config.shapes = append(Config.shapes, *NewShape([]Polyline{{[]Point{{0, 0}, {0, 4}}, nil}}))
	config.shapes = append(Config.shapes, *NewShape([]Polyline{{[]Point{{0, 0}, {0, 8}}, nil}}))
	config.shapeAngleDeviationRange = 180
	config.shapeAngleDeviationStep = 10

	return config
}

func (config *VecartConfig) equalTo(otherConfig *VecartConfig) bool {
	if config.inputPath != otherConfig.inputPath {
		return false
	}
	if config.outputPath != otherConfig.outputPath {
		return false
	}
	if config.artworkWidth != otherConfig.artworkWidth {
		return false
	}
	if config.artworkHeight != otherConfig.artworkHeight {
		return false
	}
	if config.quadrantWidth != otherConfig.quadrantWidth {
		return false
	}
	if config.quadrantHeight != otherConfig.quadrantHeight {
		return false
	}
	if config.darknessThreshold != otherConfig.darknessThreshold {
		return false
	}
	if config.shapeDarknessFactor != otherConfig.shapeDarknessFactor {
		return false
	}
	if config.whitePunishmentBoundry != otherConfig.whitePunishmentBoundry {
		return false
	}
	if config.whitePunishmentValue != otherConfig.whitePunishmentValue {
		return false
	}
	if config.randomSeed != otherConfig.randomSeed {
		return false
	}
	if config.parallelRoutines != otherConfig.parallelRoutines {
		return false
	}
	if config.highPrecisionShapePositioning != otherConfig.highPrecisionShapePositioning {
		return false
	}
	if config.shapeRefinement != otherConfig.shapeRefinement {
		return false
	}
	if config.shapeRefinementIterations != otherConfig.shapeRefinementIterations {
		return false
	}
	if config.shapeRefinementPercentage != otherConfig.shapeRefinementPercentage {
		return false
	}
	if config.smoothEdges != otherConfig.smoothEdges {
		return false
	}
	if config.combineShapes != otherConfig.combineShapes {
		return false
	}
	if config.combineShapesTolerance != otherConfig.combineShapesTolerance {
		return false
	}
	if config.combineShapesIterations != otherConfig.combineShapesIterations {
		return false
	}
	if config.strokeWidth != otherConfig.strokeWidth {
		return false
	}
	if config.strokeColor != otherConfig.strokeColor {
		return false
	}
	if config.reverseShapeOrder != otherConfig.reverseShapeOrder {
		return false
	}
	if config.configInOutput != otherConfig.configInOutput {
		return false
	}
	if config.processingDpi != otherConfig.processingDpi {
		return false
	}
	if config.outputDpi != otherConfig.outputDpi {
		return false
	}
	if config.timeout != otherConfig.timeout {
		return false
	}
	if config.debug != otherConfig.debug {
		return false
	}

	if !shapesEqual(&config.shapes, &otherConfig.shapes, 10, false) {
		return false
	}

	if config.shapeAngleDeviationRange != otherConfig.shapeAngleDeviationRange {
		return false
	}
	if config.shapeAngleDeviationStep != otherConfig.shapeAngleDeviationStep {
		return false
	}

	return true
}

func shapesEqual(shapes *[]Shape, otherShapes *[]Shape, precision int, ignoreVariants bool) bool {
	if shapes == nil && otherShapes == nil {
		return true
	}

	if !(shapes != nil && otherShapes != nil) {
		return false
	}

	if len(*shapes) != len(*otherShapes) {
		return false
	}

	var shapePointers, otherShapePointers []*Shape

	for index := range *shapes {
		shapePointers = append(shapePointers, &((*shapes)[index]))
		otherShapePointers = append(otherShapePointers, &((*otherShapes)[index]))
	}

	for i := 0; i < len(shapePointers); i++ {
		currentShape := shapePointers[i]
		foundEqualShape := false
		for j := 0; j < len(otherShapePointers); j++ {
			currentOtherShape := otherShapePointers[j]
			if currentOtherShape.equalTo(currentShape, precision, ignoreVariants) {
				foundEqualShape = true
				removeShape(otherShapePointers, j)
				break
			}
		}

		if !foundEqualShape {
			return false
		}
	}

	return true
}

func removeShape(s []*Shape, index int) []*Shape {
	return append(s[:index], s[index+1:]...)
}

func (config *VecartConfig) fromJSON(jsonString string) bool {
	var jsonData map[string]any
	err := json.Unmarshal([]byte(jsonString), &jsonData)
	if err != nil {
		fmt.Printf("Could not parse config file: '%e'\n", err)
		return false
	}

	errorsOcurred = false

	getBool(jsonData, "debug", &config.debug)
	debug = config.debug

	getString(jsonData, "inputPath", &config.inputPath)
	getString(jsonData, "outputPath", &config.outputPath)
	getInt(jsonData, "artworkWidth", &config.artworkWidth)
	getInt(jsonData, "artworkHeight", &config.artworkHeight)

	getInt(jsonData, "quadrantWidth", &config.quadrantWidth)
	getInt(jsonData, "quadrantHeight", &config.quadrantHeight)
	getFloat(jsonData, "darknessThreshold", &config.darknessThreshold)
	getFloat(jsonData, "shapeDarknessFactor", &config.shapeDarknessFactor)
	getInt(jsonData, "whitePunishmentBoundry", &config.whitePunishmentBoundry)
	getFloat(jsonData, "whitePunishmentValue", &config.whitePunishmentValue)

	getInt(jsonData, "randomSeed", &config.randomSeed)
	getInt(jsonData, "parallelRoutines", &config.parallelRoutines)

	getBool(jsonData, "highPrecisionShapePositioning", &config.highPrecisionShapePositioning)
	getBool(jsonData, "shapeRefinement", &config.shapeRefinement)
	getInt(jsonData, "shapeRefinementIterations", &config.shapeRefinementIterations)
	getFloat(jsonData, "shapeRefinementPercentage", &config.shapeRefinementPercentage)
	getBool(jsonData, "smoothEdges", &config.smoothEdges)
	getBool(jsonData, "combineShapes", &config.combineShapes)
	getFloat(jsonData, "combineShapesTolerance", &config.combineShapesTolerance)
	getInt(jsonData, "combineShapesIterations", &config.combineShapesIterations)
	getFloat(jsonData, "strokeWidth", &config.strokeWidth)
	getString(jsonData, "strokeColor", &config.strokeColor)
	getBool(jsonData, "reverseShapeOrder", &config.reverseShapeOrder)
	getBool(jsonData, "configInOutput", &config.configInOutput)
	getFloat(jsonData, "processingDpi", &config.processingDpi)
	getFloat(jsonData, "outputDpi", &config.outputDpi)
	getInt(jsonData, "timeout", &config.timeout)

	getFloat(jsonData, "shapeAngleDeviationRange", &config.shapeAngleDeviationRange)
	getFloat(jsonData, "shapeAngleDeviationStep", &config.shapeAngleDeviationStep)

	getShapes(jsonData, "shapes", &config.shapes)

	configMap := config.toMap()

	for key := range jsonData {
		_, ok := configMap[key]
		if !ok && key != "shapes" {
			errorsOcurred = true
			bestDistance := math.MaxInt
			bestCorrectKey := ""
			for correctKey := range configMap {
				distance := levenshteinDistance(key, correctKey)
				if distance < bestDistance {
					bestDistance = distance
					bestCorrectKey = correctKey
				}
			}

			if bestCorrectKey == "" {
				errors = append(errors, "Unkown Key in Config '"+key+"'")
			} else {
				errors = append(errors, "Unkown Key in Config '"+key+"'. Did you mean '"+bestCorrectKey+"'?")
			}

		}
	}

	if !config.validate() {
		errorsOcurred = true
	}

	if debug {
		for _, errorString := range errors {
			fmt.Println(errorString)
		}
	}

	return !errorsOcurred
}

func (config *VecartConfig) validate() bool {
	valid := true

	if config.inputPath != "" {
		if !pathValid(config.inputPath) {
			valid = false
			errors = append(errors, "Input path '"+config.inputPath+"' is not a valid!")
		}
	}

	if !pathValid(filepath.Dir(config.outputPath)) {
		valid = false
		errors = append(errors, "Output path '"+config.outputPath+"' is not a valid!")

	}

	if config.artworkHeight == 0 && config.artworkWidth == 0 {
		valid = false
		errors = append(errors, "Both artworkWidth and artworkHeight parameters are 0. This is not allowed!")

	}

	if config.quadrantWidth == 0 {
		valid = false
		errors = append(errors, "A quadrantWidth of 0 is invalid!")

	}

	if config.quadrantHeight == 0 {
		valid = false
		errors = append(errors, "A quadrantHeight below 0 is invalid!")

	}

	if config.darknessThreshold < 0 {
		valid = false
		errors = append(errors, "A darknessThreshold below 0 is invalid!")

	}

	if config.shapeDarknessFactor <= 0 {
		valid = false
		errors = append(errors, "A shapeDarknessFactor below or equal to 0 is invalid!")

	}

	if config.whitePunishmentBoundry < 0 {
		valid = false
		errors = append(errors, "A whitePunishmentBoundry below 0 is invalid!")

	}

	if config.parallelRoutines < 1 {
		valid = false
		errors = append(errors, "parallelRoutines must be greater than 0!")

	}

	if config.shapeRefinementIterations <= 0 {
		valid = false
		errors = append(errors, "shapeRefinementIterations must be greater than 0!")

	}

	if config.shapeRefinementPercentage <= 0 {
		valid = false
		errors = append(errors, "shapeRefinementPercentage must be greater than 0!")

	}

	if config.combineShapesTolerance <= 0 {
		valid = false
		errors = append(errors, "combineShapesTolerance must be greater than 0!")

	}

	if config.combineShapesIterations <= 0 {
		valid = false
		errors = append(errors, "combineShapesIterations must be greater than 0!")

	}

	if config.strokeWidth < 0 {
		valid = false
		errors = append(errors, "strokeWidth must be greater or equal to 0!")

	}

	if config.processingDpi <= 0 {
		valid = false
		errors = append(errors, "processingDpi must be greater than 0!")

	}

	if config.outputDpi <= 0 {
		valid = false
		errors = append(errors, "outputDpi must be greater than 0!")

	}

	if config.timeout <= 0 {
		valid = false
		errors = append(errors, "timeout must be greater than 0!")

	}

	if config.shapeAngleDeviationRange < 0 {
		valid = false
		errors = append(errors, "shapeAngleDeviationRange must be greater or equal to 0!")

	}

	if config.shapeAngleDeviationStep <= 0 {
		valid = false
		errors = append(errors, "shapeAngleDeviationStep must be greater than 0!")

	}

	return valid
}

func pathValid(path string) bool {
	pwd, err := os.Getwd()
	if err != nil {
		errors = append(errors, "Could not get current working directory to validate relative config paths!")
		return false
	}

	if _, err := os.Stat(pwd + path); os.IsNotExist(err) {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return false
		}
	}

	return true
}

func (config *VecartConfig) toJson() string {
	jsonBytes, err := json.MarshalIndent(config.toMap(), "", "    ")
	if err == nil {
		return string(jsonBytes)
	}

	fmt.Printf("Error ocurred while parsing config to json %T\n", err)

	return ""
}

func (config *VecartConfig) toMap() map[string]any {
	jsonData := make(map[string]any)

	jsonData["inputPath"] = config.inputPath
	jsonData["inputPath"] = config.inputPath
	jsonData["outputPath"] = config.outputPath
	jsonData["artworkWidth"] = config.artworkWidth
	jsonData["artworkHeight"] = config.artworkHeight

	jsonData["quadrantWidth"] = config.quadrantWidth
	jsonData["quadrantHeight"] = config.quadrantHeight
	jsonData["darknessThreshold"] = config.darknessThreshold
	jsonData["shapeDarknessFactor"] = config.shapeDarknessFactor
	jsonData["whitePunishmentBoundry"] = config.whitePunishmentBoundry
	jsonData["whitePunishmentValue"] = config.whitePunishmentValue
	jsonData["randomSeed"] = config.randomSeed
	jsonData["parallelRoutines"] = config.parallelRoutines

	jsonData["highPrecisionShapePositioning"] = config.highPrecisionShapePositioning
	jsonData["shapeRefinement"] = config.shapeRefinement
	jsonData["shapeRefinementIterations"] = config.shapeRefinementIterations
	jsonData["shapeRefinementPercentage"] = config.shapeRefinementPercentage
	jsonData["smoothEdges"] = config.smoothEdges
	jsonData["combineShapes"] = config.combineShapes
	jsonData["combineShapesTolerance"] = config.combineShapesTolerance
	jsonData["combineShapesIterations"] = config.combineShapesIterations
	jsonData["strokeWidth"] = config.strokeWidth
	jsonData["strokeColor"] = config.strokeColor
	jsonData["reverseShapeOrder"] = config.reverseShapeOrder
	jsonData["configInOutput"] = config.configInOutput
	jsonData["processingDpi"] = config.processingDpi
	jsonData["outputDpi"] = config.outputDpi
	jsonData["timeout"] = config.timeout
	jsonData["debug"] = config.debug

	jsonData["shapeAngleDeviationRange"] = config.shapeAngleDeviationRange
	jsonData["shapeAngleDeviationStep"] = config.shapeAngleDeviationStep

	var shapes []any
	for _, shape := range config.shapes {
		shapes = append(shapes, shape.toJSON())
	}
	jsonData["shapes"] = shapes

	return jsonData
}

func getShapes(jsonData map[string]any, key string, configOption *[]Shape) {
	shapeArray, success := getArray(jsonData, key)
	if !success {
		return
	}

	shapes := &[]Shape{}
	for _, shape := range shapeArray {
		getShape(shape, shapes)
	}

	if len(*shapes) != 0 {
		*configOption = *shapes
	}
}

func getShape(shape any, targetArray *[]Shape) {
	switch shape.(type) {
	default:
		errorsOcurred = true
		errors = append(errors, "Unexpected type for shape definition | expected object or array of polylines")
		return
	case map[string]any:
		parseShape(shape.(map[string]any), targetArray)
	}
}

func parseShape(shapeParameters map[string]any, targetArray *[]Shape) {
	shapeType, ok := shapeParameters["type"]
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Missing 'type' attribute for shape definition!")
		return
	}

	switch value := shapeType.(type) {
	default:
		errorsOcurred = true
		errors = append(errors, "Unexpected type for Shape type | expected string")
		return
	case string:
		switch value {
		default:
			errorsOcurred = true
			errors = append(errors, "Invalid shape type")
			return
		case "line", "Line", "LINE":
			getLine(shapeParameters, targetArray)
		case "rectangle", "Rectangle", "RECTANGLE":
			getRectangle(shapeParameters, targetArray)
		case "triangle", "Triangle", "TRIANGLE":
			getTriangle(shapeParameters, targetArray)
		case "circle", "Circle", "CIRCLE":
			getCircle(shapeParameters, targetArray)
		case "polyline", "Polyline", "POLYLINE":
			getPolyline(shapeParameters, targetArray)
		case "polygon", "Polygon", "POLYGON":
			getPolygon(shapeParameters, targetArray)
		case "text", "Text", "TEXT":
			getText(shapeParameters, targetArray)
		case "group", "Group", "GROUP":
			getGroup(shapeParameters, targetArray)
		}

	}
}

func getLine(shapeParameters map[string]any, targetArray *[]Shape) {
	p1, ok := shapeParameters["p1"]
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Missing 'p1' attribute for line definition!")
		return
	}

	p2, ok := shapeParameters["p2"]
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Missing 'p2' attribute for line definition!")
		return
	}
	var points []Point
	getPoint(p1, &points)
	getPoint(p2, &points)
	if len(points) != 2 {
		errorsOcurred = true
		errors = append(errors, "Invalid point definition for line definition!")
		return
	}
	*targetArray = append(*targetArray, *NewLine(&points[0], &points[1]))
}

func getRectangle(shapeParameters map[string]any, targetArray *[]Shape) {
	topLeftArray, ok := getArray(shapeParameters, "topLeft")
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Missing 'topLeft' attribute for rectangle definition!")
		return
	}

	if len(topLeftArray) != 2 {
		errorsOcurred = true
		errors = append(errors, "Invalid topLeft definition for rectangle definition!")
		return
	}

	topLeftX, ok := getFloatFromAny(topLeftArray[0])
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Invalid topLeft definition for rectangle definition!")
		return
	}
	topLeftY, ok := getFloatFromAny(topLeftArray[1])
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Invalid topLeft definition for rectangle definition!")
		return
	}

	width, ok := shapeParameters["width"]
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Missing 'width' attribute for rectangle definition!")
		return
	}

	height, ok := shapeParameters["width"]
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Missing 'height' attribute for rectangle definition!")
		return
	}

	widthValue, ok := getFloatFromAny(width)
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Invalid width value for rectangle definition!")
		return
	}
	heigthValue, ok := getFloatFromAny(height)
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Invalid width value for rectangle definition!")
		return
	}

	topLeft := NewPoint(topLeftX, topLeftY)
	topRight := NewPoint(topLeft.X+widthValue, topLeft.Y)
	bottomRight := NewPoint(topLeft.X+widthValue, topLeft.Y+heigthValue)
	bottomLeft := NewPoint(topLeft.X, topLeft.Y+heigthValue)

	*targetArray = append(*targetArray, *NewPolygon(&[]Point{*topLeft, *topRight, *bottomRight, *bottomLeft}).toShape())

}

func getTriangle(shapeParameters map[string]any, targetArray *[]Shape) {
	p1, ok := shapeParameters["p1"]
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Missing 'p1' attribute for rectangle definition!")
		return
	}

	p2, ok := shapeParameters["p2"]
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Missing 'p2' attribute for rectangle definition!")
		return
	}

	p3, ok := shapeParameters["p3"]
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Missing 'p3' attribute for rectangle definition!")
		return
	}

	var points []Point
	getPoint(p1, &points)
	getPoint(p2, &points)
	getPoint(p3, &points)
	if len(points) != 3 {
		errorsOcurred = true
		errors = append(errors, "Invalid point definition for triangle definition!")
		return
	}

	*targetArray = append(*targetArray, *NewPolygon(&points).toShape())
}
func getCircle(shapeParameters map[string]any, targetArray *[]Shape) {
	center, ok := getArray(shapeParameters, "center")
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Missing 'center' attribute for circle definition!")
		return
	}

	radius, ok := getFloatFromAny(shapeParameters["radius"])
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Missing 'radius' attribute for circle definition!")
		return
	}

	if len(center) < 2 {
		errorsOcurred = true
		errors = append(errors, "Invalid center definition for circle definition!")
		return
	}

	centerX, ok := getFloatFromAny(center[0])
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Invalid center value for circle definition!")
		return
	}
	centerY, ok := getFloatFromAny(center[1])
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Invalid center value for circle definition!")
		return
	}
	radiusFloat, ok := getFloatFromAny(radius)
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Invalid radius value for circle definition!")
		return
	}

	if centerX != math.SmallestNonzeroFloat64 && centerY != math.SmallestNonzeroFloat64 && radiusFloat != math.SmallestNonzeroFloat64 {
		*targetArray = append(*targetArray, *NewCircle(*NewPoint(centerX, centerY), radiusFloat).toShape())
	}
}

func getPolyline(shapeParameters map[string]any, targetArray *[]Shape) {
	pointsAny, ok := getArray(shapeParameters, "points")
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Missing 'points' attribute or value is not an array for polyline definition!")
		return
	}

	var points []Point
	for _, point := range pointsAny {
		getPoint(point, &points)
	}

	if len(points) < 3 {
		errorsOcurred = true
		errors = append(errors, "Invalid point definition for polyline definition!")
		return
	}

	*targetArray = append(*targetArray, *NewSingleLineShape(*NewPolyline(&points, nil)))
}

func getPolygon(shapeParameters map[string]any, targetArray *[]Shape) {
	pointsAny, ok := getArray(shapeParameters, "points")
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Missing 'points' attribute or value is not an array for polygon definition!")
		return
	}

	var points []Point
	for _, point := range pointsAny {
		getPoint(point, &points)
	}

	if len(points) < 3 {
		errorsOcurred = true
		errors = append(errors, "Invalid point definition for polygon definition!")
		return
	}

	*targetArray = append(*targetArray, *NewPolygon(&points).toShape())
}

func getText(shapeParameters map[string]any, targetArray *[]Shape) {
	lineHeightAny, ok := shapeParameters["lineHeight"]
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Missing 'lineHeight' attribute for text definition!")
		return
	}
	lineHeight, ok := getFloatFromAny(lineHeightAny)
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Invalid value for 'lineHeight' attribute for text definition")
		return
	}

	center, ok := getArray(shapeParameters, "center")
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Missing or invalid center value for text definition")
		return
	}

	if len(center) < 2 {
		errorsOcurred = true
		errors = append(errors, "Invalid center definition for text definition!")
		return
	}

	centerX, ok := getFloatFromAny(center[0])
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Invalid center value for text definition!")
		return
	}
	centerY, ok := getFloatFromAny(center[1])
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Invalid center value for text definition!")
		return
	}

	textAny, ok := shapeParameters["text"]
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Missing 'text' attribute for text definition!")
		return
	}
	text, ok := getStringFromAny(textAny)
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Invalid value for 'text' attribute for text definition")
		return
	}

	font := Fonts["IBM-Plex-Sans"]

	fontAny, ok := shapeParameters["font"]
	if ok {
		fontString, ok := getStringFromAny(fontAny)
		if !ok {
			errorsOcurred = true
			errors = append(errors, "Invalid value for 'font' attribute for text definition")
			return
		}
		font, ok = Fonts[fontString]
		if !ok {
			errorsOcurred = true
			errors = append(errors, "Unkown font '"+fontString+"' attribute for text definition")
			return
		}
	}

	*targetArray = append(*targetArray, font.getText(text, lineHeight, *NewPoint(centerX, centerY)))
}

func getGroup(shapeParameters map[string]any, targetArray *[]Shape) {
	shapesAny, ok := getArray(shapeParameters, "shapes")
	if !ok {
		errorsOcurred = true
		errors = append(errors, "Missing 'shapes' attribute or value is not an array for group definition!")
		return
	}

	shapes := &[]Shape{}
	for _, shape := range shapesAny {
		getShape(shape, shapes)
	}

	if len(*shapes) < 1 {
		errorsOcurred = true
		errors = append(errors, "Invalid shape definition for group definition!")
		return
	}

	*targetArray = append(*targetArray, *combineShapes(shapes))
}

func combineShapes(shapes *[]Shape) *Shape {
	var lines []Polyline

	for _, shape := range *shapes {
		lines = append(lines, shape.Lines...)
	}

	return NewShape(lines)
}

func getStringFromAny(value any) (string, bool) {
	switch value.(type) {
	default:
		errorsOcurred = true
		errors = append(errors, "Unexpected type for expected string value")
		return "", false
	case string:
		return value.(string), true
	}
}

func getFloatFromAny(value any) (float64, bool) {
	switch value.(type) {
	default:
		errorsOcurred = true
		errors = append(errors, "Unexpected type for expected float value")
		return 0, false
	case float64:
		return value.(float64), true
	case int:
		return float64(value.(int)), true
	}
}

func getPoint(point any, points *[]Point) {
	switch point.(type) {
	default:
		errorsOcurred = true
		errors = append(errors, "Unexpected type for Point | expected array")
		return
	case []any:
		var currentPoint Point
		currentPoint.X = math.SmallestNonzeroFloat64
		currentPoint.Y = math.SmallestNonzeroFloat64

		pointArray := point.([]any)
		for index, pointValue := range pointArray {
			switch pointValue.(type) {
			default:
				errorsOcurred = true
				errors = append(errors, "Unexpected type for Point x or y value | expected float")
				return
			case float64:
				if index == 0 {
					currentPoint.X = pointValue.(float64)
				} else {
					currentPoint.Y = pointValue.(float64)
				}
			case int:
				if index == 0 {
					currentPoint.X = float64(pointValue.(int))
				} else {
					currentPoint.Y = float64(pointValue.(int))
				}
			}
		}

		if currentPoint.X != math.SmallestNonzeroFloat64 && currentPoint.Y != math.SmallestNonzeroFloat64 {
			*points = append(*points, currentPoint)
		}
	}
}

func getArray(jsonData map[string]any, key string) ([]any, bool) {
	if _, ok := jsonData[key]; !ok {
		return nil, false
	}

	switch jsonData[key].(type) {
	default:
		errorsOcurred = true
		errors = append(errors, "Unexpected type for '"+key+"' | expected array")
		return nil, false
	case []any:
		return jsonData[key].([]any), true
	}
}

func getBool(jsonData map[string]any, key string, configOption *bool) {
	if _, ok := jsonData[key]; !ok {
		return
	}

	switch jsonData[key].(type) {
	default:
		errorsOcurred = true
		errors = append(errors, "Unexpected type for '"+key+"' | expected bool")
	case bool:
		*configOption = jsonData[key].(bool)
	}
}

func getString(jsonData map[string]any, key string, configOption *string) {
	if _, ok := jsonData[key]; !ok {
		return
	}
	switch jsonData[key].(type) {
	default:
		errorsOcurred = true
		errors = append(errors, "Unexpected type for '"+key+"' | expected string")
	case string:
		*configOption = jsonData[key].(string)
	}
}

func getInt(jsonData map[string]any, key string, configOption *int) {
	if _, ok := jsonData[key]; !ok {
		return
	}

	switch jsonData[key].(type) {
	default:
		errorsOcurred = true
		errors = append(errors, "Unexpected type for '"+key+"' | expected int")
	case int:
		*configOption = jsonData[key].(int)
	case float64:
		*configOption = int(jsonData[key].(float64))
	}
}

func getFloat(jsonData map[string]any, key string, configOption *float64) {
	if _, ok := jsonData[key]; !ok {
		return
	}

	switch jsonData[key].(type) {
	default:
		errorsOcurred = true
		errors = append(errors, "Unexpected type for '"+key+"' | expected float", key)
	case int:
		*configOption = float64(jsonData[key].(int))
	case float64:
		*configOption = jsonData[key].(float64)
	}
}
