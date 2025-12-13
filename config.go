package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

const (
	Shapes = iota
	Mosaic
	SingleLine
)

var errorsOcurred bool
var occuredErrors []string

type VecartConfig struct {
	mode uint

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
	updateFrequency        int

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
	backgroundColor               string
	reverseShapeOrder             bool
	configInOutput                bool
	shortConfig                   bool
	statsInOutput                 bool
	processingDpi                 float64
	outputDpi                     float64
	overwriteExisting             bool
	timeout                       int

	shapes                   []Shape
	shapeAngleDeviationRange float64
	shapeAngleDeviationStep  float64
}

func NewConfig() VecartConfig {
	var config VecartConfig

	config.mode = Shapes
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
	config.updateFrequency = 2

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
	config.backgroundColor = "NONE"
	config.reverseShapeOrder = false
	config.configInOutput = true
	config.shortConfig = true
	config.statsInOutput = true
	config.processingDpi = 25
	config.outputDpi = 72
	config.overwriteExisting = false
	config.timeout = 60

	config.shapes = append(Config.shapes, *NewShape([]Polyline{{[]Point{{0, 0}, {0, 2}}, nil}}))
	config.shapes = append(Config.shapes, *NewShape([]Polyline{{[]Point{{0, 0}, {0, 4}}, nil}}))
	config.shapes = append(Config.shapes, *NewShape([]Polyline{{[]Point{{0, 0}, {0, 8}}, nil}}))
	config.shapeAngleDeviationRange = 180
	config.shapeAngleDeviationStep = 10

	return config
}

func (config *VecartConfig) setByName(key string, value any) {
	switch key {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid config field key '"+key+"'")
	case "mode":
		config.mode = value.(uint)
	case "inputPath":
		config.inputPath = value.(string)
	case "outputPath":
		config.outputPath = value.(string)
	case "artworkWidth":
		config.artworkWidth = value.(int)
	case "artworkHeight":
		config.artworkHeight = value.(int)
	case "quadrantWidth":
		config.quadrantWidth = value.(int)
	case "quadrantHeight":
		config.quadrantHeight = value.(int)
	case "darknessThreshold":
		config.darknessThreshold = value.(float64)
	case "shapeDarknessFactor":
		config.shapeDarknessFactor = value.(float64)
	case "whitePunishmentBoundry":
		config.whitePunishmentBoundry = value.(int)
	case "whitePunishmentValue":
		config.whitePunishmentValue = value.(float64)
	case "randomSeed":
		config.randomSeed = value.(int)
	case "parallelRoutines":
		config.parallelRoutines = value.(int)
	case "updateFrequency":
		config.updateFrequency = value.(int)
	case "highPrecisionShapePositioning":
		config.highPrecisionShapePositioning = value.(bool)
	case "shapeRefinement":
		config.shapeRefinement = value.(bool)
	case "shapeRefinementIterations":
		config.shapeRefinementIterations = value.(int)
	case "shapeRefinementPercentage":
		config.shapeRefinementPercentage = value.(float64)
	case "smoothEdges":
		config.smoothEdges = value.(bool)
	case "combineShapes":
		config.combineShapes = value.(bool)
	case "combineShapesTolerance":
		config.combineShapesTolerance = value.(float64)
	case "combineShapesIterations":
		config.combineShapesIterations = value.(int)
	case "strokeWidth":
		config.strokeWidth = value.(float64)
	case "strokeColor":
		config.strokeColor = value.(string)
	case "backgroundColor":
		config.backgroundColor = value.(string)
	case "reverseShapeOrder":
		config.reverseShapeOrder = value.(bool)
	case "configInOutput":
		config.configInOutput = value.(bool)
	case "shortConfig":
		config.shortConfig = value.(bool)
	case "statsInOutput":
		config.statsInOutput = value.(bool)
	case "processingDpi":
		config.processingDpi = value.(float64)
	case "outputDpi":
		config.outputDpi = value.(float64)
	case "overwriteExisting":
		config.overwriteExisting = value.(bool)
	case "timeout":
		config.timeout = value.(int)
	case "shapes":
		config.shapes = value.([]Shape)
	case "shapeAngleDeviationRange":
		config.shapeAngleDeviationRange = value.(float64)
	case "shapeAngleDeviationStep":
		config.shapeAngleDeviationStep = value.(float64)
	}
}

func modeFromString(mode string) (uint, bool) {
	switch strings.ToLower(mode) {
	case "shapes":
		return Shapes, true
	case "mosaic":
		return Mosaic, true
	case "singleline":
		return SingleLine, true
	default:
		return Shapes, false
	}
}

func modeStringFromInt(mode uint) (string, bool) {
	switch mode {
	case Shapes:
		return "Shapes", true
	case Mosaic:
		return "Mosaic", true
	case SingleLine:
		return "SingleLine", true
	default:
		return "Shapes", false
	}
}

func (config *VecartConfig) equalTo(otherConfig *VecartConfig, printDifference bool) bool {
	if config.mode != otherConfig.mode {
		if printDifference {
			fmt.Printf("Mode differs: %d != %d\n", config.mode, otherConfig.mode)
		}
		return false
	}
	if config.inputPath != otherConfig.inputPath {
		if printDifference {
			fmt.Printf("InputPath differs: %s != %s\n", config.inputPath, otherConfig.inputPath)
		}
		return false
	}
	if config.outputPath != otherConfig.outputPath {
		if printDifference {
			fmt.Printf("OutputPath differs: %s != %s\n", config.outputPath, otherConfig.outputPath)
		}
		return false
	}
	if config.artworkWidth != otherConfig.artworkWidth {
		if printDifference {
			fmt.Printf("ArtworkWidth differs: %d != %d\n", config.artworkWidth, otherConfig.artworkWidth)
		}
		return false
	}
	if config.artworkHeight != otherConfig.artworkHeight {
		if printDifference {
			fmt.Printf("ArtworkHeight differs: %d != %d\n", config.artworkHeight, otherConfig.artworkHeight)
		}
		return false
	}
	if config.quadrantWidth != otherConfig.quadrantWidth {
		if printDifference {
			fmt.Printf("QuadrantWidth differs: %d != %d\n", config.quadrantWidth, otherConfig.quadrantWidth)
		}
		return false
	}
	if config.quadrantHeight != otherConfig.quadrantHeight {
		if printDifference {
			fmt.Printf("QuadrantHeight differs: %d != %d\n", config.quadrantHeight, otherConfig.quadrantHeight)
		}
		return false
	}
	if config.darknessThreshold != otherConfig.darknessThreshold {
		if printDifference {
			fmt.Printf("DarknessThreshold differs: %f != %f\n", config.darknessThreshold, otherConfig.darknessThreshold)
		}
		return false
	}
	if config.shapeDarknessFactor != otherConfig.shapeDarknessFactor {
		if printDifference {
			fmt.Printf("ShapeDarknessFactor differs: %f != %f\n", config.shapeDarknessFactor, otherConfig.shapeDarknessFactor)
		}
		return false
	}
	if config.whitePunishmentBoundry != otherConfig.whitePunishmentBoundry {
		if printDifference {
			fmt.Printf("WhitePunishmentBoundry differs: %d != %d\n", config.whitePunishmentBoundry, otherConfig.whitePunishmentBoundry)
		}
		return false
	}
	if config.whitePunishmentValue != otherConfig.whitePunishmentValue {
		if printDifference {
			fmt.Printf("WhitePunishmentValue differs: %f != %f\n", config.whitePunishmentValue, otherConfig.whitePunishmentValue)
		}
		return false
	}
	if config.randomSeed != otherConfig.randomSeed {
		if printDifference {
			fmt.Printf("RandomSeed differs: %d != %d\n", config.randomSeed, otherConfig.randomSeed)
		}
		return false
	}
	if config.parallelRoutines != otherConfig.parallelRoutines {
		if printDifference {
			fmt.Printf("ParallelRoutines differs: %d != %d\n", config.parallelRoutines, otherConfig.parallelRoutines)
		}
		return false
	}
	if config.updateFrequency != otherConfig.updateFrequency {
		if printDifference {
			fmt.Printf("UpdateFrequency differs: %d != %d\n", config.updateFrequency, otherConfig.updateFrequency)
		}
		return false
	}
	if config.highPrecisionShapePositioning != otherConfig.highPrecisionShapePositioning {
		if printDifference {
			fmt.Printf("HighPrecisionShapePositioning differs: %t != %t\n", config.highPrecisionShapePositioning, otherConfig.highPrecisionShapePositioning)
		}
		return false
	}
	if config.shapeRefinement != otherConfig.shapeRefinement {
		if printDifference {
			fmt.Printf("ShapeRefinement differs: %t != %t\n", config.shapeRefinement, otherConfig.shapeRefinement)
		}
		return false
	}
	if config.shapeRefinementIterations != otherConfig.shapeRefinementIterations {
		if printDifference {
			fmt.Printf("ShapeRefinementIterations differs: %d != %d\n", config.shapeRefinementIterations, otherConfig.shapeRefinementIterations)
		}
		return false
	}
	if config.shapeRefinementPercentage != otherConfig.shapeRefinementPercentage {
		if printDifference {
			fmt.Printf("ShapeRefinementPercentage differs: %f != %f\n", config.shapeRefinementPercentage, otherConfig.shapeRefinementPercentage)
		}
		return false
	}
	if config.smoothEdges != otherConfig.smoothEdges {
		if printDifference {
			fmt.Printf("SmoothEdges differs: %t != %t\n", config.smoothEdges, otherConfig.smoothEdges)
		}
		return false
	}
	if config.combineShapes != otherConfig.combineShapes {
		if printDifference {
			fmt.Printf("CombineShapes differs: %t != %t\n", config.combineShapes, otherConfig.combineShapes)
		}
		return false
	}
	if config.combineShapesTolerance != otherConfig.combineShapesTolerance {
		if printDifference {
			fmt.Printf("CombineShapesTolerance differs: %f != %f\n", config.combineShapesTolerance, otherConfig.combineShapesTolerance)
		}
		return false
	}
	if config.combineShapesIterations != otherConfig.combineShapesIterations {
		if printDifference {
			fmt.Printf("CombineShapesIterations differs: %d != %d\n", config.combineShapesIterations, otherConfig.combineShapesIterations)
		}
		return false
	}
	if config.strokeWidth != otherConfig.strokeWidth {
		if printDifference {
			fmt.Printf("StrokeWidth differs: %f != %f\n", config.strokeWidth, otherConfig.strokeWidth)
		}
		return false
	}
	if config.strokeColor != otherConfig.strokeColor {
		if printDifference {
			fmt.Printf("StrokeColor differs: %s != %s\n", config.strokeColor, otherConfig.strokeColor)
		}
		return false
	}
	if config.backgroundColor != otherConfig.backgroundColor {
		if printDifference {
			fmt.Printf("BackgroundColor differs: %s != %s\n", config.backgroundColor, otherConfig.backgroundColor)
		}
		return false
	}
	if config.reverseShapeOrder != otherConfig.reverseShapeOrder {
		if printDifference {
			fmt.Printf("ReverseShapeOrder differs: %t != %t\n", config.reverseShapeOrder, otherConfig.reverseShapeOrder)
		}
		return false
	}
	if config.configInOutput != otherConfig.configInOutput {
		if printDifference {
			fmt.Printf("ConfigInOutput differs: %t != %t\n", config.configInOutput, otherConfig.configInOutput)
		}
		return false
	}
	if config.shortConfig != otherConfig.shortConfig {
		if printDifference {
			fmt.Printf("ShortConfig differs: %t != %t\n", config.shortConfig, otherConfig.shortConfig)
		}
		return false
	}
	if config.statsInOutput != otherConfig.statsInOutput {
		if printDifference {
			fmt.Printf("StatsInOutput differs: %t != %t\n", config.statsInOutput, otherConfig.statsInOutput)
		}
		return false
	}
	if config.processingDpi != otherConfig.processingDpi {
		if printDifference {
			fmt.Printf("ProcessingDpi differs: %f != %f\n", config.processingDpi, otherConfig.processingDpi)
		}
		return false
	}
	if config.outputDpi != otherConfig.outputDpi {
		if printDifference {
			fmt.Printf("OutputDpi differs: %f != %f\n", config.outputDpi, otherConfig.outputDpi)
		}
		return false
	}
	if config.timeout != otherConfig.timeout {
		if printDifference {
			fmt.Printf("Timeout differs: %d != %d\n", config.timeout, otherConfig.timeout)
		}
		return false
	}
	if config.overwriteExisting != otherConfig.overwriteExisting {
		if printDifference {
			fmt.Printf("OverwriteExisting differs: %t != %t\n", config.overwriteExisting, otherConfig.overwriteExisting)
		}
		return false
	}

	if !shapesEqual(&config.shapes, &otherConfig.shapes, 10, false) {
		if printDifference {
			fmt.Printf("Shapes differ\n")
		}
		return false
	}

	if config.shapeAngleDeviationRange != otherConfig.shapeAngleDeviationRange {
		if printDifference {
			fmt.Printf("ShapeAngleDeviationRange differs: %f != %f\n", config.shapeAngleDeviationRange, otherConfig.shapeAngleDeviationRange)
		}
		return false
	}
	if config.shapeAngleDeviationStep != otherConfig.shapeAngleDeviationStep {
		if printDifference {
			fmt.Printf("ShapeAngleDeviationStep differs: %f != %f\n", config.shapeAngleDeviationStep, otherConfig.shapeAngleDeviationStep)
		}
		return false
	}

	return true
}

func (config *VecartConfig) copy() VecartConfig {
	copiedConfig := NewConfig()

	copiedConfig.mode = config.mode
	copiedConfig.inputPath = config.inputPath
	copiedConfig.outputPath = config.outputPath
	copiedConfig.artworkWidth = config.artworkWidth
	copiedConfig.artworkHeight = config.artworkHeight
	copiedConfig.quadrantWidth = config.quadrantWidth
	copiedConfig.quadrantHeight = config.quadrantHeight
	copiedConfig.darknessThreshold = config.darknessThreshold
	copiedConfig.shapeDarknessFactor = config.shapeDarknessFactor
	copiedConfig.whitePunishmentBoundry = config.whitePunishmentBoundry
	copiedConfig.whitePunishmentValue = config.whitePunishmentValue
	copiedConfig.randomSeed = config.randomSeed
	copiedConfig.parallelRoutines = config.parallelRoutines
	copiedConfig.updateFrequency = config.updateFrequency
	copiedConfig.highPrecisionShapePositioning = config.highPrecisionShapePositioning
	copiedConfig.shapeRefinement = config.shapeRefinement
	copiedConfig.shapeRefinementIterations = config.shapeRefinementIterations
	copiedConfig.shapeRefinementPercentage = config.shapeRefinementPercentage
	copiedConfig.smoothEdges = config.smoothEdges
	copiedConfig.combineShapes = config.combineShapes
	copiedConfig.combineShapesTolerance = config.combineShapesTolerance
	copiedConfig.combineShapesIterations = config.combineShapesIterations
	copiedConfig.strokeWidth = config.strokeWidth
	copiedConfig.strokeColor = config.strokeColor
	copiedConfig.backgroundColor = config.backgroundColor
	copiedConfig.reverseShapeOrder = config.reverseShapeOrder
	copiedConfig.configInOutput = config.configInOutput
	copiedConfig.shortConfig = config.shortConfig
	copiedConfig.statsInOutput = config.statsInOutput
	copiedConfig.processingDpi = config.processingDpi
	copiedConfig.outputDpi = config.outputDpi
	copiedConfig.overwriteExisting = config.overwriteExisting
	copiedConfig.timeout = config.timeout
	copiedConfig.shapeAngleDeviationRange = config.shapeAngleDeviationRange
	copiedConfig.shapeAngleDeviationStep = config.shapeAngleDeviationStep
	copiedConfig.shapes = copyShapes(config.shapes)

	return copiedConfig
}

func copyShapes(shapes []Shape) []Shape {
	copiedShapes := []Shape{}
	for _, shape := range shapes {
		copiedShapes = append(copiedShapes, shape.copy())
	}
	return copiedShapes
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

func (config *VecartConfig) fromJSON(jsonString string) ([]string, any, []*VecartConfig, bool) {
	allConfigs := []*VecartConfig{config}

	var jsonData map[string]any
	err := json.Unmarshal([]byte(jsonString), &jsonData)
	if err != nil {
		fmt.Printf("Could not parse config file: '%e'\n", err)
		return []string{}, nil, []*VecartConfig{}, false
	}

	errorsOcurred = false

	var userKeys []string
	for key := range jsonData {
		userKeys = append(userKeys, key)
	}

	mode, _ := modeStringFromInt(Shapes)
	getString(jsonData, "mode", &mode)
	modeInt, ok := modeFromString(mode)
	config.mode = modeInt
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected mode '"+mode+"'")
	}

	getString(jsonData, "outputPath", &config.outputPath)
	userShapes := getShapes(jsonData, "shapes", &config.shapes)

	getStringDynamic(jsonData, "inputPath", &allConfigs)
	getIntDynamic(jsonData, "artworkWidth", &allConfigs)
	getIntDynamic(jsonData, "artworkHeight", &allConfigs)
	getIntDynamic(jsonData, "quadrantWidth", &allConfigs)
	getIntDynamic(jsonData, "quadrantHeight", &allConfigs)
	getIntDynamic(jsonData, "whitePunishmentBoundry", &allConfigs)
	getFloatDynamic(jsonData, "whitePunishmentValue", &allConfigs)
	getIntDynamic(jsonData, "randomSeed", &allConfigs)
	getIntDynamic(jsonData, "parallelRoutines", &allConfigs)
	getIntDynamic(jsonData, "updateFrequency", &allConfigs)
	getBoolDynamic(jsonData, "highPrecisionShapePositioning", &allConfigs)
	getBoolDynamic(jsonData, "shapeRefinement", &allConfigs)
	getIntDynamic(jsonData, "shapeRefinementIterations", &allConfigs)
	getFloatDynamic(jsonData, "shapeRefinementPercentage", &allConfigs)
	getBoolDynamic(jsonData, "smoothEdges", &allConfigs)
	getBoolDynamic(jsonData, "combineShapes", &allConfigs)
	getFloatDynamic(jsonData, "combineShapesTolerance", &allConfigs)
	getIntDynamic(jsonData, "combineShapesIterations", &allConfigs)
	getFloatDynamic(jsonData, "strokeWidth", &allConfigs)
	getStringDynamic(jsonData, "strokeColor", &allConfigs)
	getStringDynamic(jsonData, "backgroundColor", &allConfigs)
	getBoolDynamic(jsonData, "reverseShapeOrder", &allConfigs)
	getBoolDynamic(jsonData, "configInOutput", &allConfigs)
	getBoolDynamic(jsonData, "shortConfig", &allConfigs)
	getBoolDynamic(jsonData, "statsInOutput", &allConfigs)
	getFloatDynamic(jsonData, "processingDpi", &allConfigs)
	getFloatDynamic(jsonData, "outputDpi", &allConfigs)
	getIntDynamic(jsonData, "timeout", &allConfigs)
	getBoolDynamic(jsonData, "overwriteExisting", &allConfigs)
	getFloatDynamic(jsonData, "shapeAngleDeviationRange", &allConfigs)
	getFloatDynamic(jsonData, "shapeAngleDeviationStep", &allConfigs)
	getFloatDynamic(jsonData, "darknessThreshold", &allConfigs)
	getFloatDynamic(jsonData, "shapeDarknessFactor", &allConfigs)

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
				occuredErrors = append(occuredErrors, "Unkown Key in Config '"+key+"'")
			} else {
				occuredErrors = append(occuredErrors, "Unkown Key in Config '"+key+"'. Did you mean '"+bestCorrectKey+"'?")
			}

		}
	}

	for _, curConfig := range allConfigs {
		if !curConfig.validate() {
			errorsOcurred = true
		}
	}

	if Log {
		for _, errorString := range occuredErrors {
			log.Println(errorString)
		}
	}

	if len(allConfigs) > 1 {
		return userKeys, userShapes, allConfigs[1:], !errorsOcurred
	}

	return userKeys, userShapes, []*VecartConfig{}, !errorsOcurred
}

func getFloatDynamic(jsonData map[string]any, key string, allConfigs *[]*VecartConfig) {
	if isArray(jsonData, key) {
		floatValues, ok := getFloatArray(jsonData, key)
		if !ok || len(floatValues) == 0 {
			errorsOcurred = true
			occuredErrors = append(occuredErrors, key+" should be a float or an array of floats for a grid search")
		} else {
			for _, currentConfig := range *allConfigs {
				currentConfig.setByName(key, floatValues[0])
				for i := 1; i < len(floatValues); i++ {
					newConfig := currentConfig.copy()
					newConfig.setByName(key, floatValues[i])
					*allConfigs = append(*allConfigs, &newConfig)
				}
			}
		}
	} else {
		value, ok := getFloatWithReturn(jsonData, key)
		if !ok {
			return
		}
		for _, currentConfig := range *allConfigs {
			currentConfig.setByName(key, value)
		}
	}
}

func getIntDynamic(jsonData map[string]any, key string, allConfigs *[]*VecartConfig) {
	if isArray(jsonData, key) {
		values, ok := getIntArray(jsonData, key)
		if !ok || len(values) == 0 {
			errorsOcurred = true
			occuredErrors = append(occuredErrors, key+" should be an integer or an array of integers for a grid search")
		} else {
			for _, currentConfig := range *allConfigs {
				currentConfig.setByName(key, values[0])
				for i := 1; i < len(values); i++ {
					newConfig := currentConfig.copy()
					newConfig.setByName(key, values[i])
					*allConfigs = append(*allConfigs, &newConfig)
				}
			}
		}
	} else {
		value, ok := getIntWithReturn(jsonData, key)
		if !ok {
			return
		}
		for _, currentConfig := range *allConfigs {
			currentConfig.setByName(key, value)
		}
	}
}

func getBoolDynamic(jsonData map[string]any, key string, allConfigs *[]*VecartConfig) {
	if isArray(jsonData, key) {
		values, ok := getBoolArray(jsonData, key)
		if !ok || len(values) == 0 {
			errorsOcurred = true
			occuredErrors = append(occuredErrors, key+" should be a integer or an array of integers for a grid search")
		} else {
			for _, currentConfig := range *allConfigs {
				currentConfig.setByName(key, values[0])
				for i := 1; i < len(values); i++ {
					newConfig := currentConfig.copy()
					newConfig.setByName(key, values[i])
					*allConfigs = append(*allConfigs, &newConfig)
				}
			}
		}
	} else {
		value, ok := getBoolWithReturn(jsonData, key)
		if !ok {
			return
		}
		for _, currentConfig := range *allConfigs {
			currentConfig.setByName(key, value)
		}
	}
}

func getStringDynamic(jsonData map[string]any, key string, allConfigs *[]*VecartConfig) {
	if isArray(jsonData, key) {
		values, ok := getStringArray(jsonData, key)
		if !ok || len(values) == 0 {
			errorsOcurred = true
			occuredErrors = append(occuredErrors, key+" should be a string or an array of strings for a grid search")
		} else {
			for _, currentConfig := range *allConfigs {
				currentConfig.setByName(key, values[0])
				for i := 1; i < len(values); i++ {
					newConfig := currentConfig.copy()
					newConfig.setByName(key, values[i])
					*allConfigs = append(*allConfigs, &newConfig)
				}
			}
		}
	} else {
		value, ok := getStringWithReturn(jsonData, key)
		if !ok {
			return
		}
		for _, currentConfig := range *allConfigs {
			currentConfig.setByName(key, value)
		}
	}
}

func (config *VecartConfig) validate() bool {
	valid := true

	validModes := []uint{Shapes, Mosaic, SingleLine}
	if !slices.Contains(validModes, config.mode) {
		valid = false
		occuredErrors = append(occuredErrors, "Invalid mode '"+strconv.Itoa(int(config.mode))+"'")
	}

	if config.inputPath != "" {
		if !pathValid(config.inputPath) {
			valid = false
			occuredErrors = append(occuredErrors, "Input path '"+config.inputPath+"' is not a valid!")
		}
	}

	if !pathValid(filepath.Dir(config.outputPath)) {
		valid = false
		occuredErrors = append(occuredErrors, "Output path '"+config.outputPath+"' is not a valid!")

	}

	if config.artworkHeight == 0 && config.artworkWidth == 0 {
		valid = false
		occuredErrors = append(occuredErrors, "Both artworkWidth and artworkHeight parameters are 0. This is not allowed!")

	}

	if config.quadrantWidth == 0 {
		valid = false
		occuredErrors = append(occuredErrors, "A quadrantWidth of 0 is invalid!")

	}

	if config.quadrantHeight == 0 {
		valid = false
		occuredErrors = append(occuredErrors, "A quadrantHeight below 0 is invalid!")

	}

	if config.darknessThreshold < 0 {
		valid = false
		occuredErrors = append(occuredErrors, "A darknessThreshold below 0 is invalid!")

	}

	if config.shapeDarknessFactor <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "A shapeDarknessFactor below or equal to 0 is invalid!")

	}

	if config.whitePunishmentBoundry < 0 {
		valid = false
		occuredErrors = append(occuredErrors, "A whitePunishmentBoundry below 0 is invalid!")

	}

	if config.parallelRoutines < 1 {
		valid = false
		occuredErrors = append(occuredErrors, "parallelRoutines must be greater than 0!")

	}

	if config.updateFrequency < 1 {
		valid = false
		occuredErrors = append(occuredErrors, "updateFrequency must be greater than 0!")
	}

	if config.shapeRefinementIterations <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "shapeRefinementIterations must be greater than 0!")

	}

	if config.shapeRefinementPercentage <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "shapeRefinementPercentage must be greater than 0!")

	}

	if config.combineShapesTolerance <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "combineShapesTolerance must be greater than 0!")

	}

	if config.combineShapesIterations <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "combineShapesIterations must be greater than 0!")

	}

	if config.strokeWidth < 0 {
		valid = false
		occuredErrors = append(occuredErrors, "strokeWidth must be greater or equal to 0!")

	}

	if config.processingDpi <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "processingDpi must be greater than 0!")

	}

	if config.outputDpi <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "outputDpi must be greater than 0!")

	}

	if config.timeout <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "timeout must be greater than 0!")

	}

	if config.shapeAngleDeviationRange < 0 {
		valid = false
		occuredErrors = append(occuredErrors, "shapeAngleDeviationRange must be greater or equal to 0!")

	}

	if config.shapeAngleDeviationStep <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "shapeAngleDeviationStep must be greater than 0!")

	}

	return valid
}

func pathValid(path string) bool {
	pwd, err := os.Getwd()
	if err != nil {
		occuredErrors = append(occuredErrors, "Could not get current working directory to validate relative config paths!")
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

func (config *VecartConfig) toShortJson(userConfigKeys []string, userShapes any) string {
	fullJsonMap := config.toMap()
	shortJsonMap := make(map[string]any)
	for key, _ := range fullJsonMap {
		if slices.Contains(userConfigKeys, key) {
			shortJsonMap[key] = fullJsonMap[key]
		}
	}
	shortJsonMap["shapes"] = userShapes

	jsonBytes, err := json.MarshalIndent(shortJsonMap, "", "    ")
	if err == nil {
		return string(jsonBytes)
	}

	fmt.Printf("Error ocurred while parsing config to short json %T\n", err)

	return ""
}

func (config *VecartConfig) toMap() map[string]any {
	jsonData := make(map[string]any)

	mode, _ := modeStringFromInt(config.mode)
	jsonData["mode"] = mode
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
	jsonData["updateFrequency"] = config.updateFrequency

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
	jsonData["backgroundColor"] = config.backgroundColor
	jsonData["reverseShapeOrder"] = config.reverseShapeOrder
	jsonData["configInOutput"] = config.configInOutput
	jsonData["shortConfig"] = config.shortConfig
	jsonData["statsInOutput"] = config.statsInOutput
	jsonData["processingDpi"] = config.processingDpi
	jsonData["outputDpi"] = config.outputDpi
	jsonData["timeout"] = config.timeout
	jsonData["overwriteExisting"] = config.overwriteExisting

	jsonData["shapeAngleDeviationRange"] = config.shapeAngleDeviationRange
	jsonData["shapeAngleDeviationStep"] = config.shapeAngleDeviationStep

	var shapes []any
	for _, shape := range config.shapes {
		shapes = append(shapes, shape.toJSON())
	}
	jsonData["shapes"] = shapes

	return jsonData
}

func getShapes(jsonData map[string]any, key string, configOption *[]Shape) any {
	shapeArray, success := getArray(jsonData, key)
	if !success {
		return jsonData[key]
	}

	shapes := &[]Shape{}
	for _, shape := range shapeArray {
		getShape(shape, shapes)
	}

	if len(*shapes) != 0 {
		*configOption = *shapes
	}

	return jsonData[key]
}

func getShape(shape any, targetArray *[]Shape) {
	switch shape := shape.(type) {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected type for shape definition | expected object or array of polylines")
		return
	case map[string]any:
		parseShape(shape, targetArray)
	}
}

func parseShape(shapeParameters map[string]any, targetArray *[]Shape) {
	shapeType, ok := shapeParameters["type"]
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Missing 'type' attribute for shape definition!")
		return
	}

	switch value := shapeType.(type) {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected type for Shape type | expected string")
		return
	case string:
		switch strings.ToLower(value) {
		default:
			errorsOcurred = true
			occuredErrors = append(occuredErrors, "Invalid shape type")
			return
		case "line":
			getLine(shapeParameters, targetArray)
		case "rectangle":
			getRectangle(shapeParameters, targetArray)
		case "triangle":
			getTriangle(shapeParameters, targetArray)
		case "circle":
			getCircle(shapeParameters, targetArray)
		case "polyline":
			getPolyline(shapeParameters, targetArray)
		case "polygon":
			getPolygon(shapeParameters, targetArray)
		case "heart":
			getHeart(shapeParameters, targetArray)
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
		occuredErrors = append(occuredErrors, "Missing 'p1' attribute for line definition!")
		return
	}

	p2, ok := shapeParameters["p2"]
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Missing 'p2' attribute for line definition!")
		return
	}
	var points []Point
	getPoint(p1, &points)
	getPoint(p2, &points)
	if len(points) != 2 {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid point definition for line definition!")
		return
	}
	*targetArray = append(*targetArray, *NewLine(&points[0], &points[1]))
}

func getRectangle(shapeParameters map[string]any, targetArray *[]Shape) {
	topLeftArray, ok := getArray(shapeParameters, "topLeft")
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Missing 'topLeft' attribute for rectangle definition!")
		return
	}

	if len(topLeftArray) != 2 {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid topLeft definition for rectangle definition!")
		return
	}

	topLeftX, ok := getFloatFromAny(topLeftArray[0])
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid topLeft definition for rectangle definition!")
		return
	}
	topLeftY, ok := getFloatFromAny(topLeftArray[1])
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid topLeft definition for rectangle definition!")
		return
	}

	width, ok := shapeParameters["width"]
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Missing 'width' attribute for rectangle definition!")
		return
	}

	height, ok := shapeParameters["width"]
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Missing 'height' attribute for rectangle definition!")
		return
	}

	widthValue, ok := getFloatFromAny(width)
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid width value for rectangle definition!")
		return
	}
	heigthValue, ok := getFloatFromAny(height)
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid width value for rectangle definition!")
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
		occuredErrors = append(occuredErrors, "Missing 'p1' attribute for rectangle definition!")
		return
	}

	p2, ok := shapeParameters["p2"]
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Missing 'p2' attribute for rectangle definition!")
		return
	}

	p3, ok := shapeParameters["p3"]
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Missing 'p3' attribute for rectangle definition!")
		return
	}

	var points []Point
	getPoint(p1, &points)
	getPoint(p2, &points)
	getPoint(p3, &points)
	if len(points) != 3 {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid point definition for triangle definition!")
		return
	}

	*targetArray = append(*targetArray, *NewPolygon(&points).toShape())
}
func getCircle(shapeParameters map[string]any, targetArray *[]Shape) {
	center, ok := getArray(shapeParameters, "center")
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Missing 'center' attribute for circle definition!")
		return
	}

	radius, ok := getFloatFromAny(shapeParameters["radius"])
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Missing 'radius' attribute for circle definition!")
		return
	}

	if len(center) < 2 {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid center definition for circle definition!")
		return
	}

	centerX, ok := getFloatFromAny(center[0])
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid center value for circle definition!")
		return
	}
	centerY, ok := getFloatFromAny(center[1])
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid center value for circle definition!")
		return
	}
	radiusFloat, ok := getFloatFromAny(radius)
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid radius value for circle definition!")
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
		occuredErrors = append(occuredErrors, "Missing 'points' attribute or value is not an array for polyline definition!")
		return
	}

	var points []Point
	for _, point := range pointsAny {
		getPoint(point, &points)
	}

	if len(points) < 3 {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid point definition for polyline definition!")
		return
	}

	*targetArray = append(*targetArray, *NewSingleLineShape(*NewPolyline(&points, nil)))
}

func getPolygon(shapeParameters map[string]any, targetArray *[]Shape) {
	pointsAny, ok := getArray(shapeParameters, "points")
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Missing 'points' attribute or value is not an array for polygon definition!")
		return
	}

	var points []Point
	for _, point := range pointsAny {
		getPoint(point, &points)
	}

	if len(points) < 3 {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid point definition for polygon definition!")
		return
	}

	*targetArray = append(*targetArray, *NewPolygon(&points).toShape())
}

func getHeart(shapeParameters map[string]any, targetArray *[]Shape) {
	size, ok := getFloatFromAny(shapeParameters["size"])
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Missing 'size' attribute for heart definition!")
		return
	}

	*targetArray = append(*targetArray, *NewHeart(size))
}

func getText(shapeParameters map[string]any, targetArray *[]Shape) {
	lineHeightAny, ok := shapeParameters["lineHeight"]
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Missing 'lineHeight' attribute for text definition!")
		return
	}
	lineHeight, ok := getFloatFromAny(lineHeightAny)
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid value for 'lineHeight' attribute for text definition")
		return
	}

	center, ok := getArray(shapeParameters, "center")
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Missing or invalid center value for text definition")
		return
	}

	if len(center) < 2 {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid center definition for text definition!")
		return
	}

	centerX, ok := getFloatFromAny(center[0])
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid center value for text definition!")
		return
	}
	centerY, ok := getFloatFromAny(center[1])
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid center value for text definition!")
		return
	}

	textAny, ok := shapeParameters["text"]
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Missing 'text' attribute for text definition!")
		return
	}
	text, ok := getStringFromAny(textAny)
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid value for 'text' attribute for text definition")
		return
	}

	font := Fonts["IBM-Plex-Sans"]

	fontAny, ok := shapeParameters["font"]
	if ok {
		fontString, ok := getStringFromAny(fontAny)
		if !ok {
			errorsOcurred = true
			occuredErrors = append(occuredErrors, "Invalid value for 'font' attribute for text definition")
			return
		}
		font, ok = Fonts[fontString]
		if !ok {
			errorsOcurred = true
			occuredErrors = append(occuredErrors, "Unkown font '"+fontString+"' attribute for text definition")
			return
		}
	}

	*targetArray = append(*targetArray, font.getText(text, lineHeight, *NewPoint(centerX, centerY)))
}

func getGroup(shapeParameters map[string]any, targetArray *[]Shape) {
	shapesAny, ok := getArray(shapeParameters, "shapes")
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Missing 'shapes' attribute or value is not an array for group definition!")
		return
	}

	shapes := &[]Shape{}
	for _, shape := range shapesAny {
		getShape(shape, shapes)
	}

	if len(*shapes) < 1 {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid shape definition for group definition!")
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
	switch value := value.(type) {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected type for expected string value")
		return "", false
	case string:
		return value, true
	}
}

func getFloatFromAny(value any) (float64, bool) {
	switch value := value.(type) {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected type for expected float value")
		return 0, false
	case float64:
		return value, true
	case int:
		return float64(value), true
	}
}

func getIntFromAny(value any) (int, bool) {
	switch value := value.(type) {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected type for expected int value")
		return 0, false
	case int:
		return value, true
	case float64:
		return int(value), true
	}
}

func getPoint(point any, points *[]Point) {
	switch point.(type) {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected type for Point | expected array")
		return
	case []any:
		var currentPoint Point
		currentPoint.X = math.SmallestNonzeroFloat64
		currentPoint.Y = math.SmallestNonzeroFloat64

		pointArray := point.([]any)
		for index, pointValue := range pointArray {
			switch pointValue := pointValue.(type) {
			default:
				errorsOcurred = true
				occuredErrors = append(occuredErrors, "Unexpected type for Point x or y value | expected float")
				return
			case float64:
				if index == 0 {
					currentPoint.X = pointValue
				} else {
					currentPoint.Y = pointValue
				}
			case int:
				if index == 0 {
					currentPoint.X = float64(pointValue)
				} else {
					currentPoint.Y = float64(pointValue)
				}
			}
		}

		if currentPoint.X != math.SmallestNonzeroFloat64 && currentPoint.Y != math.SmallestNonzeroFloat64 {
			*points = append(*points, currentPoint)
		}
	}
}

func getBool(jsonData map[string]any, key string, configOption *bool) {
	if _, ok := jsonData[key]; !ok {
		return
	}

	switch jsonData[key].(type) {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected type for '"+key+"' | expected bool")
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
		occuredErrors = append(occuredErrors, "Unexpected type for '"+key+"' | expected string")
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
		occuredErrors = append(occuredErrors, "Unexpected type for '"+key+"' | expected int")
	case int:
		*configOption = jsonData[key].(int)
	case float64:
		*configOption = int(jsonData[key].(float64))
	}
}

func getFloatWithReturn(jsonData map[string]any, key string) (float64, bool) {
	if _, ok := jsonData[key]; !ok {
		return 0.0, false
	}

	switch jsonData[key].(type) {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected type for '"+key+"' | expected float", key)
		return 0.0, false
	case int:
		return float64(jsonData[key].(int)), true
	case float64:
		return jsonData[key].(float64), true
	}
}

func getIntWithReturn(jsonData map[string]any, key string) (int, bool) {
	if _, ok := jsonData[key]; !ok {
		return 0, false
	}

	switch jsonData[key].(type) {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected type for '"+key+"' | expected int", key)
		return 0, false
	case int:
		return jsonData[key].(int), true
	case float64:
		return int(jsonData[key].(float64)), true
	}
}

func getStringWithReturn(jsonData map[string]any, key string) (string, bool) {
	if _, ok := jsonData[key]; !ok {
		return "", false
	}

	switch jsonData[key].(type) {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected type for '"+key+"' | expected string", key)
		return "", false
	case string:
		return jsonData[key].(string), true
	}
}

func getBoolWithReturn(jsonData map[string]any, key string) (bool, bool) {
	if _, ok := jsonData[key]; !ok {
		return false, false
	}

	switch jsonData[key].(type) {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected type for '"+key+"' | expected float", key)
		return false, false
	case bool:
		return jsonData[key].(bool), true
	}
}

func getFloat(jsonData map[string]any, key string, configOption *float64) {
	if _, ok := jsonData[key]; !ok {
		return
	}

	switch jsonData[key].(type) {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected type for '"+key+"' | expected float", key)
	case int:
		*configOption = float64(jsonData[key].(int))
	case float64:
		*configOption = jsonData[key].(float64)
	}
}

func getArray(jsonData map[string]any, key string) ([]any, bool) {
	if _, ok := jsonData[key]; !ok {
		return nil, false
	}

	switch jsonData[key].(type) {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected type for '"+key+"' | expected array")
		return nil, false
	case []any:
		return jsonData[key].([]any), true
	}
}

func getFloatArray(jsonData map[string]any, key string) ([]float64, bool) {
	if _, ok := jsonData[key]; !ok {
		return nil, false
	}
	switch jsonData[key].(type) {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected type for '"+key+"' | expected array of floats")
		return nil, false
	case []any:
		var floatArray []float64
		for _, value := range jsonData[key].([]any) {
			floatValue, ok := getFloatFromAny(value)
			if !ok {
				errorsOcurred = true
				occuredErrors = append(occuredErrors, "Unexpected type for '"+key+"' | expected array of floats")
				return nil, false
			}
			floatArray = append(floatArray, floatValue)
		}
		return floatArray, true
	}
}

func getIntArray(jsonData map[string]any, key string) ([]int, bool) {
	if _, ok := jsonData[key]; !ok {
		return nil, false
	}
	switch jsonData[key].(type) {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected type for '"+key+"' | expected int or array of ints")
		return nil, false
	case []any:
		var intArray []int
		for _, value := range jsonData[key].([]any) {
			intValue, ok := getIntFromAny(value)
			if !ok {
				errorsOcurred = true
				occuredErrors = append(occuredErrors, "Unexpected type for '"+key+"' | expected array of ints")
				return nil, false
			}
			intArray = append(intArray, intValue)
		}
		return intArray, true
	}
}

func getStringArray(jsonData map[string]any, key string) ([]string, bool) {
	if _, ok := jsonData[key]; !ok {
		return nil, false
	}
	switch jsonData[key].(type) {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected type for '"+key+"' | expected array of strings")
		return nil, false
	case []any:
		var stringArray []string
		for _, value := range jsonData[key].([]any) {
			stringValue, ok := getStringFromAny(value)
			if !ok {
				errorsOcurred = true
				occuredErrors = append(occuredErrors, "Unexpected type for '"+key+"' | expected array of strings")
				return nil, false
			}
			stringArray = append(stringArray, stringValue)
		}
		return stringArray, true
	}
}

func getBoolArray(jsonData map[string]any, key string) ([]bool, bool) {
	if _, ok := jsonData[key]; !ok {
		return nil, false
	}
	switch jsonData[key].(type) {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected type for '"+key+"' | expected array of bools")
		return nil, false
	case []any:
		var boolArray []bool
		for _, value := range jsonData[key].([]any) {
			switch value := value.(type) {
			default:
				errorsOcurred = true
				occuredErrors = append(occuredErrors, "Unexpected type for '"+key+"' | expected array of bools")
				return nil, false
			case bool:
				boolArray = append(boolArray, value)
			}
		}
		return boolArray, true
	}
}

func isArray(jsonData map[string]any, key string) bool {
	if _, ok := jsonData[key]; !ok {
		return false
	}
	switch jsonData[key].(type) {
	default:
		return false
	case []any:
		return true
	}
}
