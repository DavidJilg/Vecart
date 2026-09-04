package general

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

	"github.com/DavidJilg/Vecart/internal/shapes"
	"github.com/DavidJilg/Vecart/internal/svgparser"
	"github.com/DavidJilg/Vecart/internal/utils"
)

const (
	ShapeArt = iota
	Mosaic
	SingleLine
	GCode
)

func ModeFromString(mode string) (uint, bool) {
	switch mode {
	case "ShapeArt":
		return ShapeArt, true
	case "Mosaic":
		return Mosaic, true
	case "SingleLine":
		return SingleLine, true
	case "GCode":
		return GCode, true
	default:
		if utils.DebugModeEnabled() {
			log.Println("Invalid Mode ", mode)
		}
		return ShapeArt, false
	}
}

func ModeStringFromInt(mode uint) (string, bool) {
	switch mode {
	case ShapeArt:
		return "ShapeArt", true
	case Mosaic:
		return "Mosaic", true
	case SingleLine:
		return "SingleLine", true
	case GCode:
		return "GCode", true
	default:
		if utils.DebugModeEnabled() {
			log.Println("Invalid Mode ", mode)
		}
		return "ShapeArt", false
	}
}

var errorsOcurred bool
var occuredErrors []string

type VecartConfig struct {
	Mode uint

	InputPath     string
	OutputPath    string
	ArtworkWidth  int
	ArtworkHeight int

	QuadrantWidth          int
	QuadrantHeight         int
	DarknessThreshold      float64
	ShapeDarknessFactor    float64
	WhitePunishmentBoundry int
	WhitePunishmentValue   float64
	RandomSeed             int
	ParallelRoutines       int
	UpdateFrequency        int

	HighPrecisionShapePositioning bool
	ShapeRefinement               bool
	ShapeRefinementIterations     int
	ShapeRefinementPercentage     float64
	SmoothEdges                   bool
	CombineShapes                 bool
	CombineShapesTolerance        float64
	CombineShapesIterations       int
	StrokeWidth                   float64
	StrokeColor                   string
	BackgroundColor               string
	ReverseShapeOrder             bool
	ConfigInOutput                bool
	ShortConfig                   bool
	StatsInOutput                 bool
	ProcessingDpi                 float64
	OutputDpi                     float64
	OverwriteExisting             bool
	Timeout                       int
	FeedRateXY                    int
	FeedRateZ                     int
	SortGCode                     bool
	KeepSVGGroups                 bool
	SortAxis                      int
	BezierTolerance               float64
	StampMode                     bool
	RefillTool                    bool
	NrOfShapesWithoutRefill       int
	RefillToolCommands            []string
	ActivateToolCommands          []string
	DeactivateToolCommands        []string
	InitialCommands               []string
	FinalCommands                 []string
	ShapeList                     []shapes.Shape
	ShapeAngleDeviationRange      float64
	ShapeAngleDeviationStep       float64
}

func NewConfig() VecartConfig {
	var config VecartConfig

	config.Mode = ShapeArt
	config.InputPath = ""
	config.OutputPath = "output.svg"
	config.ArtworkWidth = 255
	config.ArtworkHeight = 370

	config.QuadrantWidth = 5
	config.QuadrantHeight = 5
	config.DarknessThreshold = 18
	config.ShapeDarknessFactor = 40
	config.WhitePunishmentBoundry = 5
	config.WhitePunishmentValue = 0.85
	config.RandomSeed = 1701
	config.ParallelRoutines = 5
	config.UpdateFrequency = 2

	config.HighPrecisionShapePositioning = false
	config.ShapeRefinement = true
	config.ShapeRefinementIterations = 1
	config.ShapeRefinementPercentage = 0.2
	config.SmoothEdges = true
	config.CombineShapes = true
	config.CombineShapesTolerance = 0.5
	config.CombineShapesIterations = 5
	config.StrokeWidth = 0.75
	config.StrokeColor = "black"
	config.BackgroundColor = "NONE"
	config.ReverseShapeOrder = false
	config.ConfigInOutput = true
	config.ShortConfig = true
	config.StatsInOutput = true
	config.ProcessingDpi = 25
	config.OutputDpi = 72
	config.OverwriteExisting = false
	config.Timeout = 60
	config.FeedRateXY = 100
	config.FeedRateZ = 100
	config.SortGCode = true
	config.KeepSVGGroups = true
	config.SortAxis = 1
	config.StampMode = false
	config.NrOfShapesWithoutRefill = 1
	config.BezierTolerance = 0.01
	config.RefillTool = false
	config.RefillToolCommands = []string{}
	config.ActivateToolCommands = []string{}
	config.DeactivateToolCommands = []string{}
	config.InitialCommands = []string{}
	config.FinalCommands = []string{}

	config.ShapeList = []shapes.Shape{}
	config.ShapeList = append(config.ShapeList, *shapes.NewShape([]shapes.Polyline{*shapes.NewPolyline(&[]shapes.Point{{X: 0, Y: 0}, {X: 0, Y: 2}}, nil)}))
	config.ShapeList = append(config.ShapeList, *shapes.NewShape([]shapes.Polyline{*shapes.NewPolyline(&[]shapes.Point{{X: 0, Y: 0}, {X: 0, Y: 4}}, nil)}))
	config.ShapeList = append(config.ShapeList, *shapes.NewShape([]shapes.Polyline{*shapes.NewPolyline(&[]shapes.Point{{X: 0, Y: 0}, {X: 0, Y: 8}}, nil)}))
	config.ShapeAngleDeviationRange = 180
	config.ShapeAngleDeviationStep = 10

	return config
}

func (config *VecartConfig) SetByName(key string, value any) {
	switch key {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid config field key '"+key+"'")
	case "mode":
		mode, ok := ModeFromString(value.(string))
		if !ok {
			errorsOcurred = true
			occuredErrors = append(occuredErrors, "Invalid mode '"+value.(string)+"'")
		} else {
			config.Mode = mode
		}
	case "inputPath":
		config.InputPath = value.(string)
	case "outputPath":
		config.OutputPath = value.(string)
	case "artworkWidth":
		config.ArtworkWidth = value.(int)
	case "artworkHeight":
		config.ArtworkHeight = value.(int)
	case "quadrantWidth":
		config.QuadrantWidth = value.(int)
	case "quadrantHeight":
		config.QuadrantHeight = value.(int)
	case "darknessThreshold":
		config.DarknessThreshold = value.(float64)
	case "shapeDarknessFactor":
		config.ShapeDarknessFactor = value.(float64)
	case "whitePunishmentBoundry":
		config.WhitePunishmentBoundry = value.(int)
	case "whitePunishmentValue":
		config.WhitePunishmentValue = value.(float64)
	case "randomSeed":
		config.RandomSeed = value.(int)
	case "parallelRoutines":
		config.ParallelRoutines = value.(int)
	case "updateFrequency":
		config.UpdateFrequency = value.(int)
	case "highPrecisionShapePositioning":
		config.HighPrecisionShapePositioning = value.(bool)
	case "shapeRefinement":
		config.ShapeRefinement = value.(bool)
	case "shapeRefinementIterations":
		config.ShapeRefinementIterations = value.(int)
	case "shapeRefinementPercentage":
		config.ShapeRefinementPercentage = value.(float64)
	case "smoothEdges":
		config.SmoothEdges = value.(bool)
	case "combineShapes":
		config.CombineShapes = value.(bool)
	case "combineShapesTolerance":
		config.CombineShapesTolerance = value.(float64)
	case "combineShapesIterations":
		config.CombineShapesIterations = value.(int)
	case "strokeWidth":
		config.StrokeWidth = value.(float64)
	case "strokeColor":
		config.StrokeColor = value.(string)
	case "backgroundColor":
		config.BackgroundColor = value.(string)
	case "reverseShapeOrder":
		config.ReverseShapeOrder = value.(bool)
	case "configInOutput":
		config.ConfigInOutput = value.(bool)
	case "shortConfig":
		config.ShortConfig = value.(bool)
	case "statsInOutput":
		config.StatsInOutput = value.(bool)
	case "processingDpi":
		config.ProcessingDpi = value.(float64)
	case "outputDpi":
		config.OutputDpi = value.(float64)
	case "overwriteExisting":
		config.OverwriteExisting = value.(bool)
	case "timeout":
		config.Timeout = value.(int)
	case "feedRateXY":
		config.FeedRateXY = value.(int)
	case "feedRateZ":
		config.FeedRateZ = value.(int)
	case "sortGCode":
		config.SortGCode = value.(bool)
	case "keepSVGGroups":
		config.KeepSVGGroups = value.(bool)
	case "sortAxis":
		config.SortAxis = value.(int)
	case "bezierTolerance":
		config.BezierTolerance = value.(float64)
	case "refillTool":
		config.RefillTool = value.(bool)
	case "stampMode":
		config.StampMode = value.(bool)
	case "nrOfShapesWithoutRefill":
		config.NrOfShapesWithoutRefill = value.(int)
	case "refillToolCommands":
		config.RefillToolCommands = value.([]string)
	case "activateToolCommands":
		config.ActivateToolCommands = value.([]string)
	case "deactivateToolCommands":
		config.DeactivateToolCommands = value.([]string)
	case "initialCommands":
		config.InitialCommands = value.([]string)
	case "finalCommands":
		config.FinalCommands = value.([]string)
	case "shapes":
		config.ShapeList = value.([]shapes.Shape)
	case "shapeAngleDeviationRange":
		config.ShapeAngleDeviationRange = value.(float64)
	case "shapeAngleDeviationStep":
		config.ShapeAngleDeviationStep = value.(float64)
	}
}

func (config *VecartConfig) EqualTo(otherConfig *VecartConfig, printDifference bool) bool {
	if config.Mode != otherConfig.Mode {
		if printDifference {
			fmt.Printf("Mode differs: %d != %d\n", config.Mode, otherConfig.Mode)
		}
		return false
	}
	if config.InputPath != otherConfig.InputPath {
		if printDifference {
			fmt.Printf("InputPath differs: %s != %s\n", config.InputPath, otherConfig.InputPath)
		}
		return false
	}
	if config.OutputPath != otherConfig.OutputPath {
		if printDifference {
			fmt.Printf("OutputPath differs: %s != %s\n", config.OutputPath, otherConfig.OutputPath)
		}
		return false
	}
	if config.ArtworkWidth != otherConfig.ArtworkWidth {
		if printDifference {
			fmt.Printf("ArtworkWidth differs: %d != %d\n", config.ArtworkWidth, otherConfig.ArtworkWidth)
		}
		return false
	}
	if config.ArtworkHeight != otherConfig.ArtworkHeight {
		if printDifference {
			fmt.Printf("ArtworkHeight differs: %d != %d\n", config.ArtworkHeight, otherConfig.ArtworkHeight)
		}
		return false
	}
	if config.QuadrantWidth != otherConfig.QuadrantWidth {
		if printDifference {
			fmt.Printf("QuadrantWidth differs: %d != %d\n", config.QuadrantWidth, otherConfig.QuadrantWidth)
		}
		return false
	}
	if config.QuadrantHeight != otherConfig.QuadrantHeight {
		if printDifference {
			fmt.Printf("QuadrantHeight differs: %d != %d\n", config.QuadrantHeight, otherConfig.QuadrantHeight)
		}
		return false
	}
	if config.DarknessThreshold != otherConfig.DarknessThreshold {
		if printDifference {
			fmt.Printf("DarknessThreshold differs: %f != %f\n", config.DarknessThreshold, otherConfig.DarknessThreshold)
		}
		return false
	}
	if config.ShapeDarknessFactor != otherConfig.ShapeDarknessFactor {
		if printDifference {
			fmt.Printf("ShapeDarknessFactor differs: %f != %f\n", config.ShapeDarknessFactor, otherConfig.ShapeDarknessFactor)
		}
		return false
	}
	if config.WhitePunishmentBoundry != otherConfig.WhitePunishmentBoundry {
		if printDifference {
			fmt.Printf("WhitePunishmentBoundry differs: %d != %d\n", config.WhitePunishmentBoundry, otherConfig.WhitePunishmentBoundry)
		}
		return false
	}
	if config.WhitePunishmentValue != otherConfig.WhitePunishmentValue {
		if printDifference {
			fmt.Printf("WhitePunishmentValue differs: %f != %f\n", config.WhitePunishmentValue, otherConfig.WhitePunishmentValue)
		}
		return false
	}
	if config.RandomSeed != otherConfig.RandomSeed {
		if printDifference {
			fmt.Printf("RandomSeed differs: %d != %d\n", config.RandomSeed, otherConfig.RandomSeed)
		}
		return false
	}
	if config.ParallelRoutines != otherConfig.ParallelRoutines {
		if printDifference {
			fmt.Printf("ParallelRoutines differs: %d != %d\n", config.ParallelRoutines, otherConfig.ParallelRoutines)
		}
		return false
	}
	if config.UpdateFrequency != otherConfig.UpdateFrequency {
		if printDifference {
			fmt.Printf("UpdateFrequency differs: %d != %d\n", config.UpdateFrequency, otherConfig.UpdateFrequency)
		}
		return false
	}
	if config.HighPrecisionShapePositioning != otherConfig.HighPrecisionShapePositioning {
		if printDifference {
			fmt.Printf("HighPrecisionShapePositioning differs: %t != %t\n", config.HighPrecisionShapePositioning, otherConfig.HighPrecisionShapePositioning)
		}
		return false
	}
	if config.ShapeRefinement != otherConfig.ShapeRefinement {
		if printDifference {
			fmt.Printf("ShapeRefinement differs: %t != %t\n", config.ShapeRefinement, otherConfig.ShapeRefinement)
		}
		return false
	}
	if config.ShapeRefinementIterations != otherConfig.ShapeRefinementIterations {
		if printDifference {
			fmt.Printf("ShapeRefinementIterations differs: %d != %d\n", config.ShapeRefinementIterations, otherConfig.ShapeRefinementIterations)
		}
		return false
	}
	if config.ShapeRefinementPercentage != otherConfig.ShapeRefinementPercentage {
		if printDifference {
			fmt.Printf("ShapeRefinementPercentage differs: %f != %f\n", config.ShapeRefinementPercentage, otherConfig.ShapeRefinementPercentage)
		}
		return false
	}
	if config.SmoothEdges != otherConfig.SmoothEdges {
		if printDifference {
			fmt.Printf("SmoothEdges differs: %t != %t\n", config.SmoothEdges, otherConfig.SmoothEdges)
		}
		return false
	}
	if config.CombineShapes != otherConfig.CombineShapes {
		if printDifference {
			fmt.Printf("CombineShapes differs: %t != %t\n", config.CombineShapes, otherConfig.CombineShapes)
		}
		return false
	}
	if config.CombineShapesTolerance != otherConfig.CombineShapesTolerance {
		if printDifference {
			fmt.Printf("CombineShapesTolerance differs: %f != %f\n", config.CombineShapesTolerance, otherConfig.CombineShapesTolerance)
		}
		return false
	}
	if config.CombineShapesIterations != otherConfig.CombineShapesIterations {
		if printDifference {
			fmt.Printf("CombineShapesIterations differs: %d != %d\n", config.CombineShapesIterations, otherConfig.CombineShapesIterations)
		}
		return false
	}
	if config.StrokeWidth != otherConfig.StrokeWidth {
		if printDifference {
			fmt.Printf("StrokeWidth differs: %f != %f\n", config.StrokeWidth, otherConfig.StrokeWidth)
		}
		return false
	}
	if config.StrokeColor != otherConfig.StrokeColor {
		if printDifference {
			fmt.Printf("StrokeColor differs: %s != %s\n", config.StrokeColor, otherConfig.StrokeColor)
		}
		return false
	}
	if config.BackgroundColor != otherConfig.BackgroundColor {
		if printDifference {
			fmt.Printf("BackgroundColor differs: %s != %s\n", config.BackgroundColor, otherConfig.BackgroundColor)
		}
		return false
	}
	if config.ReverseShapeOrder != otherConfig.ReverseShapeOrder {
		if printDifference {
			fmt.Printf("ReverseShapeOrder differs: %t != %t\n", config.ReverseShapeOrder, otherConfig.ReverseShapeOrder)
		}
		return false
	}
	if config.ConfigInOutput != otherConfig.ConfigInOutput {
		if printDifference {
			fmt.Printf("ConfigInOutput differs: %t != %t\n", config.ConfigInOutput, otherConfig.ConfigInOutput)
		}
		return false
	}
	if config.ShortConfig != otherConfig.ShortConfig {
		if printDifference {
			fmt.Printf("ShortConfig differs: %t != %t\n", config.ShortConfig, otherConfig.ShortConfig)
		}
		return false
	}
	if config.StatsInOutput != otherConfig.StatsInOutput {
		if printDifference {
			fmt.Printf("StatsInOutput differs: %t != %t\n", config.StatsInOutput, otherConfig.StatsInOutput)
		}
		return false
	}
	if config.ProcessingDpi != otherConfig.ProcessingDpi {
		if printDifference {
			fmt.Printf("ProcessingDpi differs: %f != %f\n", config.ProcessingDpi, otherConfig.ProcessingDpi)
		}
		return false
	}
	if config.OutputDpi != otherConfig.OutputDpi {
		if printDifference {
			fmt.Printf("OutputDpi differs: %f != %f\n", config.OutputDpi, otherConfig.OutputDpi)
		}
		return false
	}
	if config.Timeout != otherConfig.Timeout {
		if printDifference {
			fmt.Printf("Timeout differs: %d != %d\n", config.Timeout, otherConfig.Timeout)
		}
		return false
	}
	if config.FeedRateXY != otherConfig.FeedRateXY {
		if printDifference {
			fmt.Printf("FeedRateXY differs: %d != %d\n", config.FeedRateXY, otherConfig.FeedRateXY)
		}
		return false
	}
	if config.FeedRateZ != otherConfig.FeedRateZ {
		if printDifference {
			fmt.Printf("FeedRateZ differs: %d != %d\n", config.FeedRateZ, otherConfig.FeedRateZ)
		}
		return false
	}
	if config.SortGCode != otherConfig.SortGCode {
		if printDifference {
			fmt.Printf("SortGCode differs: %t != %t\n", config.SortGCode, otherConfig.SortGCode)
		}
		return false
	}
	if config.KeepSVGGroups != otherConfig.KeepSVGGroups {
		if printDifference {
			fmt.Printf("KeepSVGGroups differs: %t != %t\n", config.KeepSVGGroups, otherConfig.KeepSVGGroups)
		}
		return false
	}
	if config.SortAxis != otherConfig.SortAxis {
		if printDifference {
			fmt.Printf("SortAxis differs: %d != %d\n", config.SortAxis, otherConfig.SortAxis)
		}
		return false
	}
	if config.BezierTolerance != otherConfig.BezierTolerance {
		if printDifference {
			fmt.Printf("BezierTolerance differs: %f != %f\n", config.BezierTolerance, otherConfig.BezierTolerance)
		}
		return false
	}
	if config.RefillTool != otherConfig.RefillTool {
		if printDifference {
			fmt.Printf("RefillTool differs: %t != %t\n", config.RefillTool, otherConfig.RefillTool)
		}
		return false
	}
	if config.StampMode != otherConfig.StampMode {
		if printDifference {
			fmt.Printf("StampMode differs: %t != %t\n", config.StampMode, otherConfig.StampMode)
		}
		return false
	}
	if config.NrOfShapesWithoutRefill != otherConfig.NrOfShapesWithoutRefill {
		if printDifference {
			fmt.Printf("NrOfShapesWithoutRefill differs: %d != %d\n", config.NrOfShapesWithoutRefill, otherConfig.NrOfShapesWithoutRefill)
		}
		return false
	}
	if !utils.GCodeSlicesEqual(config.RefillToolCommands, otherConfig.RefillToolCommands) {
		if printDifference {
			fmt.Printf("RefillToolCommands differs: %s != %s\n", strings.Join(config.RefillToolCommands, "\n"), strings.Join(otherConfig.RefillToolCommands, "\n"))
		}
		return false
	}
	if !utils.GCodeSlicesEqual(config.ActivateToolCommands, otherConfig.ActivateToolCommands) {
		if printDifference {
			fmt.Printf("ActivateToolCommands differs: %s != %s\n", strings.Join(config.ActivateToolCommands, "\n"), strings.Join(otherConfig.ActivateToolCommands, "\n"))
		}
		return false
	}
	if !utils.GCodeSlicesEqual(config.DeactivateToolCommands, otherConfig.DeactivateToolCommands) {
		if printDifference {
			fmt.Printf("DeactivateToolCommands differs: %s != %s\n", strings.Join(config.DeactivateToolCommands, "\n"), strings.Join(otherConfig.DeactivateToolCommands, "\n"))
		}
		return false
	}
	if !utils.GCodeSlicesEqual(config.InitialCommands, otherConfig.InitialCommands) {
		if printDifference {
			fmt.Printf("InitialCommands differs: %s != %s\n", strings.Join(config.InitialCommands, "\n"), strings.Join(otherConfig.InitialCommands, "\n"))
		}
		return false
	}
	if !utils.GCodeSlicesEqual(config.FinalCommands, otherConfig.FinalCommands) {
		if printDifference {
			fmt.Printf("FinalCommands differs: %s != %s\n", strings.Join(config.FinalCommands, "\n"), strings.Join(otherConfig.FinalCommands, "\n"))
		}
		return false
	}
	if config.OverwriteExisting != otherConfig.OverwriteExisting {
		if printDifference {
			fmt.Printf("OverwriteExisting differs: %t != %t\n", config.OverwriteExisting, otherConfig.OverwriteExisting)
		}
		return false
	}

	if !shapes.ShapesEqual(&config.ShapeList, &otherConfig.ShapeList, 10, false) {
		if printDifference {
			fmt.Printf("Shapes differ\n")
		}
		return false
	}

	if config.ShapeAngleDeviationRange != otherConfig.ShapeAngleDeviationRange {
		if printDifference {
			fmt.Printf("ShapeAngleDeviationRange differs: %f != %f\n", config.ShapeAngleDeviationRange, otherConfig.ShapeAngleDeviationRange)
		}
		return false
	}
	if config.ShapeAngleDeviationStep != otherConfig.ShapeAngleDeviationStep {
		if printDifference {
			fmt.Printf("ShapeAngleDeviationStep differs: %f != %f\n", config.ShapeAngleDeviationStep, otherConfig.ShapeAngleDeviationStep)
		}
		return false
	}

	return true
}

func (config *VecartConfig) Copy() VecartConfig {
	copiedConfig := NewConfig()

	copiedConfig.Mode = config.Mode
	copiedConfig.InputPath = config.InputPath
	copiedConfig.OutputPath = config.OutputPath
	copiedConfig.ArtworkWidth = config.ArtworkWidth
	copiedConfig.ArtworkHeight = config.ArtworkHeight
	copiedConfig.QuadrantWidth = config.QuadrantWidth
	copiedConfig.QuadrantHeight = config.QuadrantHeight
	copiedConfig.DarknessThreshold = config.DarknessThreshold
	copiedConfig.ShapeDarknessFactor = config.ShapeDarknessFactor
	copiedConfig.WhitePunishmentBoundry = config.WhitePunishmentBoundry
	copiedConfig.WhitePunishmentValue = config.WhitePunishmentValue
	copiedConfig.RandomSeed = config.RandomSeed
	copiedConfig.ParallelRoutines = config.ParallelRoutines
	copiedConfig.UpdateFrequency = config.UpdateFrequency
	copiedConfig.HighPrecisionShapePositioning = config.HighPrecisionShapePositioning
	copiedConfig.ShapeRefinement = config.ShapeRefinement
	copiedConfig.ShapeRefinementIterations = config.ShapeRefinementIterations
	copiedConfig.ShapeRefinementPercentage = config.ShapeRefinementPercentage
	copiedConfig.SmoothEdges = config.SmoothEdges
	copiedConfig.CombineShapes = config.CombineShapes
	copiedConfig.CombineShapesTolerance = config.CombineShapesTolerance
	copiedConfig.CombineShapesIterations = config.CombineShapesIterations
	copiedConfig.StrokeWidth = config.StrokeWidth
	copiedConfig.StrokeColor = config.StrokeColor
	copiedConfig.BackgroundColor = config.BackgroundColor
	copiedConfig.ReverseShapeOrder = config.ReverseShapeOrder
	copiedConfig.ConfigInOutput = config.ConfigInOutput
	copiedConfig.ShortConfig = config.ShortConfig
	copiedConfig.StatsInOutput = config.StatsInOutput
	copiedConfig.ProcessingDpi = config.ProcessingDpi
	copiedConfig.OutputDpi = config.OutputDpi
	copiedConfig.OverwriteExisting = config.OverwriteExisting
	copiedConfig.Timeout = config.Timeout
	copiedConfig.FeedRateXY = config.FeedRateXY
	copiedConfig.FeedRateZ = config.FeedRateZ
	copiedConfig.SortGCode = config.SortGCode
	copiedConfig.KeepSVGGroups = config.KeepSVGGroups
	copiedConfig.SortAxis = config.SortAxis
	copiedConfig.BezierTolerance = config.BezierTolerance
	copiedConfig.RefillTool = config.RefillTool
	copiedConfig.StampMode = config.StampMode
	copiedConfig.NrOfShapesWithoutRefill = config.NrOfShapesWithoutRefill
	copiedConfig.RefillToolCommands = config.RefillToolCommands
	copiedConfig.ActivateToolCommands = config.ActivateToolCommands
	copiedConfig.DeactivateToolCommands = config.DeactivateToolCommands
	copiedConfig.InitialCommands = config.InitialCommands
	copiedConfig.FinalCommands = config.FinalCommands
	copiedConfig.ShapeAngleDeviationRange = config.ShapeAngleDeviationRange
	copiedConfig.ShapeAngleDeviationStep = config.ShapeAngleDeviationStep
	copiedConfig.ShapeList = copyShapes(config.ShapeList)

	return copiedConfig
}

func copyShapes(shapeList []shapes.Shape) []shapes.Shape {
	copiedShapes := []shapes.Shape{}
	for _, shape := range shapeList {
		copiedShapes = append(copiedShapes, shape.Copy())
	}
	return copiedShapes
}

func (config *VecartConfig) FromJSON(jsonString string, configPath string) ([]string, any, []*VecartConfig, bool) {
	configPathFixed := strings.ReplaceAll(configPath, "\\", "/")
	configDirectory := configPathFixed[:strings.LastIndex(configPathFixed, "/")+1]

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

	mode, _ := ModeStringFromInt(ShapeArt)
	getString(jsonData, "mode", &mode)
	modeInt, ok := ModeFromString(mode)
	config.Mode = modeInt
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected mode '"+mode+"'")
	}

	getString(jsonData, "outputPath", &config.OutputPath)
	if config.OutputPath != "" && !filepath.IsAbs(config.OutputPath) {
		config.OutputPath = configDirectory + config.OutputPath
		config.OutputPath = strings.ReplaceAll(config.OutputPath, "//", "/")
	}

	activateToolArray, ok := getStringArray(jsonData, "activateToolCommands")
	if ok {
		config.ActivateToolCommands = activateToolArray
	}

	deactivateToolArray, ok := getStringArray(jsonData, "deactivateToolCommands")
	if ok {
		config.DeactivateToolCommands = deactivateToolArray
	}

	initialCommandsArray, ok := getStringArray(jsonData, "initialCommands")
	if ok {
		config.InitialCommands = initialCommandsArray
	}

	finalCommandsArray, ok := getStringArray(jsonData, "finalCommands")
	if ok {
		config.FinalCommands = finalCommandsArray
	}

	refillCommandsArray, ok := getStringArray(jsonData, "refillToolCommands")
	if ok {
		config.RefillToolCommands = refillCommandsArray
	}

	userShapes := getShapes(jsonData, "shapes", &config.ShapeList, configDirectory)

	getStringDynamic(jsonData, "inputPath", &allConfigs)
	for i := 0; i < len(allConfigs); i++ {
		if allConfigs[i].InputPath == "" && !filepath.IsAbs(allConfigs[i].InputPath) {
			continue
		}
		allConfigs[i].InputPath = configDirectory + allConfigs[i].InputPath
		allConfigs[i].InputPath = strings.ReplaceAll(allConfigs[i].InputPath, "//", "/")
	}
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

	getIntDynamic(jsonData, "feedRateXY", &allConfigs)
	getIntDynamic(jsonData, "feedRateZ", &allConfigs)
	getBoolDynamic(jsonData, "sortGCode", &allConfigs)
	getBoolDynamic(jsonData, "keepSVGGroups", &allConfigs)
	getIntDynamic(jsonData, "sortAxis", &allConfigs)
	getBoolDynamic(jsonData, "reverseSort", &allConfigs)
	getFloatDynamic(jsonData, "bezierTolerance", &allConfigs)
	getBoolDynamic(jsonData, "refillTool", &allConfigs)
	getBoolDynamic(jsonData, "stampMode", &allConfigs)
	getIntDynamic(jsonData, "nrOfShapesWithoutRefill", &allConfigs)

	configMap := config.ToMap()

	for key := range jsonData {
		_, ok := configMap[key]
		if !ok && key != "shapes" {
			errorsOcurred = true
			bestDistance := math.MaxInt
			bestCorrectKey := ""
			for correctKey := range configMap {
				distance := utils.LevenshteinDistance(key, correctKey)
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
		if !curConfig.Validate() {
			errorsOcurred = true
		}
	}

	if utils.DebugModeEnabled() {
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
				currentConfig.SetByName(key, floatValues[0])
				for i := 1; i < len(floatValues); i++ {
					newConfig := currentConfig.Copy()
					newConfig.SetByName(key, floatValues[i])
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
			currentConfig.SetByName(key, value)
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
				currentConfig.SetByName(key, values[0])
				for i := 1; i < len(values); i++ {
					newConfig := currentConfig.Copy()
					newConfig.SetByName(key, values[i])
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
			currentConfig.SetByName(key, value)
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
				currentConfig.SetByName(key, values[0])
				for i := 1; i < len(values); i++ {
					newConfig := currentConfig.Copy()
					newConfig.SetByName(key, values[i])
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
			currentConfig.SetByName(key, value)
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
				currentConfig.SetByName(key, values[0])
				for i := 1; i < len(values); i++ {
					newConfig := currentConfig.Copy()
					newConfig.SetByName(key, values[i])
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
			currentConfig.SetByName(key, value)
		}
	}
}

func (config *VecartConfig) Validate() bool {
	valid := true

	validModes := []uint{ShapeArt, Mosaic, SingleLine, GCode}
	if !slices.Contains(validModes, config.Mode) {
		valid = false
		occuredErrors = append(occuredErrors, "Invalid mode '"+strconv.Itoa(int(config.Mode))+"'")
	}

	if config.InputPath != "" {
		if !pathValid(config.InputPath) {
			valid = false
			occuredErrors = append(occuredErrors, "Input path '"+config.InputPath+"' is not a valid!")
		}
	}

	if !pathValid(filepath.Dir(config.OutputPath)) {
		valid = false
		occuredErrors = append(occuredErrors, "Output path '"+config.OutputPath+"' is not a valid!")

	}

	if config.ArtworkHeight == 0 && config.ArtworkWidth == 0 {
		valid = false
		occuredErrors = append(occuredErrors, "Both artworkWidth and artworkHeight parameters are 0. This is not allowed!")

	}

	if config.QuadrantWidth == 0 {
		valid = false
		occuredErrors = append(occuredErrors, "A quadrantWidth of 0 is invalid!")

	}

	if config.QuadrantHeight == 0 {
		valid = false
		occuredErrors = append(occuredErrors, "A quadrantHeight below 0 is invalid!")

	}

	if config.DarknessThreshold < 0 {
		valid = false
		occuredErrors = append(occuredErrors, "A darknessThreshold below 0 is invalid!")

	}

	if config.ShapeDarknessFactor <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "A shapeDarknessFactor below or equal to 0 is invalid!")

	}

	if config.WhitePunishmentBoundry < 0 {
		valid = false
		occuredErrors = append(occuredErrors, "A whitePunishmentBoundry below 0 is invalid!")

	}

	if config.ParallelRoutines < 1 {
		valid = false
		occuredErrors = append(occuredErrors, "parallelRoutines must be greater than 0!")

	}

	if config.UpdateFrequency < 1 {
		valid = false
		occuredErrors = append(occuredErrors, "updateFrequency must be greater than 0!")
	}

	if config.ShapeRefinementIterations <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "shapeRefinementIterations must be greater than 0!")

	}

	if config.ShapeRefinementPercentage <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "shapeRefinementPercentage must be greater than 0!")

	}

	if config.CombineShapesTolerance <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "combineShapesTolerance must be greater than 0!")

	}

	if config.CombineShapesIterations <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "combineShapesIterations must be greater than 0!")

	}

	if config.StrokeWidth < 0 {
		valid = false
		occuredErrors = append(occuredErrors, "strokeWidth must be greater or equal to 0!")

	}

	if config.ProcessingDpi <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "processingDpi must be greater than 0!")

	}

	if config.OutputDpi <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "outputDpi must be greater than 0!")

	}

	if config.Timeout <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "timeout must be greater than 0!")

	}

	if config.FeedRateXY <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "feedRateXY must be greater than 0!")

	}

	if config.FeedRateZ <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "feedRateZ must be greater than 0!")

	}

	if config.BezierTolerance <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "bezierTolerance must be greater than 0!")

	}

	if config.NrOfShapesWithoutRefill <= 0 {
		valid = false
		occuredErrors = append(occuredErrors, "NrOfShapesWithoutRefill must be greater than 0!")

	}

	if config.ShapeAngleDeviationRange < 0 {
		valid = false
		occuredErrors = append(occuredErrors, "shapeAngleDeviationRange must be greater or equal to 0!")

	}

	if config.ShapeAngleDeviationStep <= 0 {
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

func (config *VecartConfig) ToJson() string {
	jsonBytes, err := json.MarshalIndent(config.ToMap(), "", "    ")
	if err == nil {
		return string(jsonBytes)
	}

	fmt.Printf("Error ocurred while parsing config to json %T\n", err)

	return ""
}

func (config *VecartConfig) ToShortJson(userConfigKeys []string, userShapes any) string {
	fullJsonMap := config.ToMap()
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

func (config *VecartConfig) ToMap() map[string]any {
	jsonData := make(map[string]any)

	mode, _ := ModeStringFromInt(config.Mode)
	jsonData["mode"] = mode
	jsonData["inputPath"] = config.InputPath
	jsonData["outputPath"] = config.OutputPath
	jsonData["artworkWidth"] = config.ArtworkWidth
	jsonData["artworkHeight"] = config.ArtworkHeight

	jsonData["quadrantWidth"] = config.QuadrantWidth
	jsonData["quadrantHeight"] = config.QuadrantHeight
	jsonData["darknessThreshold"] = config.DarknessThreshold
	jsonData["shapeDarknessFactor"] = config.ShapeDarknessFactor
	jsonData["whitePunishmentBoundry"] = config.WhitePunishmentBoundry
	jsonData["whitePunishmentValue"] = config.WhitePunishmentValue
	jsonData["randomSeed"] = config.RandomSeed
	jsonData["parallelRoutines"] = config.ParallelRoutines
	jsonData["updateFrequency"] = config.UpdateFrequency

	jsonData["highPrecisionShapePositioning"] = config.HighPrecisionShapePositioning
	jsonData["shapeRefinement"] = config.ShapeRefinement
	jsonData["shapeRefinementIterations"] = config.ShapeRefinementIterations
	jsonData["shapeRefinementPercentage"] = config.ShapeRefinementPercentage
	jsonData["smoothEdges"] = config.SmoothEdges
	jsonData["combineShapes"] = config.CombineShapes
	jsonData["combineShapesTolerance"] = config.CombineShapesTolerance
	jsonData["combineShapesIterations"] = config.CombineShapesIterations
	jsonData["strokeWidth"] = config.StrokeWidth
	jsonData["strokeColor"] = config.StrokeColor
	jsonData["backgroundColor"] = config.BackgroundColor
	jsonData["reverseShapeOrder"] = config.ReverseShapeOrder
	jsonData["configInOutput"] = config.ConfigInOutput
	jsonData["shortConfig"] = config.ShortConfig
	jsonData["statsInOutput"] = config.StatsInOutput
	jsonData["processingDpi"] = config.ProcessingDpi
	jsonData["outputDpi"] = config.OutputDpi
	jsonData["timeout"] = config.Timeout
	jsonData["feedRateXY"] = config.FeedRateXY
	jsonData["feedRateZ"] = config.FeedRateZ
	jsonData["sortGCode"] = config.SortGCode
	jsonData["keepSVGGroups"] = config.KeepSVGGroups
	jsonData["sortAxis"] = config.SortAxis
	jsonData["bezierTolerance"] = config.BezierTolerance
	jsonData["refillTool"] = config.RefillTool
	jsonData["stampMode"] = config.StampMode
	jsonData["nrOfShapesWithoutRefill"] = config.NrOfShapesWithoutRefill
	jsonData["refillToolCommands"] = config.RefillToolCommands
	jsonData["activateToolCommands"] = config.ActivateToolCommands
	jsonData["deactivateToolCommands"] = config.DeactivateToolCommands
	jsonData["initialCommands"] = config.InitialCommands
	jsonData["finalCommands"] = config.FinalCommands
	jsonData["overwriteExisting"] = config.OverwriteExisting

	jsonData["shapeAngleDeviationRange"] = config.ShapeAngleDeviationRange
	jsonData["shapeAngleDeviationStep"] = config.ShapeAngleDeviationStep

	var shapes []any
	for _, shape := range config.ShapeList {
		shapes = append(shapes, shape.ToJSON())
	}
	jsonData["shapes"] = shapes

	return jsonData
}

func getShapes(jsonData map[string]any, key string, configOption *[]shapes.Shape, configDirectory string) any {
	shapeArray, success := getArray(jsonData, key)
	if !success {
		return jsonData[key]
	}

	shapes := &[]shapes.Shape{}
	for _, shape := range shapeArray {
		getShape(shape, shapes, configDirectory)
	}

	if len(*shapes) != 0 {
		*configOption = *shapes
	}

	return jsonData[key]
}

func getShape(shape any, targetArray *[]shapes.Shape, configDirectory string) {
	switch shape := shape.(type) {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected type for shape definition | expected object or array of polylines")
		return
	case map[string]any:
		parseShape(shape, targetArray, configDirectory)
	}
}

func parseShape(shapeParameters map[string]any, targetArray *[]shapes.Shape, configDirectory string) {
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
			getGroup(shapeParameters, targetArray, configDirectory)
		case "svg", "Svg", "SVG":
			getSVG(shapeParameters, targetArray, configDirectory)
		}

	}
}

func getLine(shapeParameters map[string]any, targetArray *[]shapes.Shape) {
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
	var points []shapes.Point
	getPoint(p1, &points)
	getPoint(p2, &points)
	if len(points) != 2 {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid point definition for line definition!")
		return
	}
	*targetArray = append(*targetArray, *shapes.NewLine(&points[0], &points[1]))
}

func getRectangle(shapeParameters map[string]any, targetArray *[]shapes.Shape) {
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

	height, ok := shapeParameters["height"]
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

	topLeft := shapes.NewPoint(topLeftX, topLeftY)
	topRight := shapes.NewPoint(topLeft.X+widthValue, topLeft.Y)
	bottomRight := shapes.NewPoint(topLeft.X+widthValue, topLeft.Y+heigthValue)
	bottomLeft := shapes.NewPoint(topLeft.X, topLeft.Y+heigthValue)

	*targetArray = append(*targetArray, *shapes.NewPolygon(&[]shapes.Point{*topLeft, *topRight, *bottomRight, *bottomLeft}).ToShape())

}

func getTriangle(shapeParameters map[string]any, targetArray *[]shapes.Shape) {
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

	var points []shapes.Point
	getPoint(p1, &points)
	getPoint(p2, &points)
	getPoint(p3, &points)
	if len(points) != 3 {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid point definition for triangle definition!")
		return
	}

	*targetArray = append(*targetArray, *shapes.NewPolygon(&points).ToShape())
}
func getCircle(shapeParameters map[string]any, targetArray *[]shapes.Shape) {
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
		*targetArray = append(*targetArray, *shapes.NewCircle(*shapes.NewPoint(centerX, centerY), radiusFloat).ToShape())
	}
}

func getPolyline(shapeParameters map[string]any, targetArray *[]shapes.Shape) {
	pointsAny, ok := getArray(shapeParameters, "points")
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Missing 'points' attribute or value is not an array for polyline definition!")
		return
	}

	var points []shapes.Point
	for _, point := range pointsAny {
		getPoint(point, &points)
	}

	if len(points) < 3 {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid point definition for polyline definition!")
		return
	}

	*targetArray = append(*targetArray, *shapes.NewSingleLineShape(*shapes.NewPolyline(&points, nil)))
}

func getPolygon(shapeParameters map[string]any, targetArray *[]shapes.Shape) {
	pointsAny, ok := getArray(shapeParameters, "points")
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Missing 'points' attribute or value is not an array for polygon definition!")
		return
	}

	var points []shapes.Point
	for _, point := range pointsAny {
		getPoint(point, &points)
	}

	if len(points) < 3 {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid point definition for polygon definition!")
		return
	}

	*targetArray = append(*targetArray, *shapes.NewPolygon(&points).ToShape())
}

func getHeart(shapeParameters map[string]any, targetArray *[]shapes.Shape) {
	size, ok := getFloatFromAny(shapeParameters["size"])
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Missing 'size' attribute for heart definition!")
		return
	}

	*targetArray = append(*targetArray, *shapes.NewHeart(size))
}

func getText(shapeParameters map[string]any, targetArray *[]shapes.Shape) {
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

	*targetArray = append(*targetArray, font.GetText(text, lineHeight, *shapes.NewPoint(centerX, centerY)))
}

func getGroup(shapeParameters map[string]any, targetArray *[]shapes.Shape, configDirectory string) {
	shapesAny, ok := getArray(shapeParameters, "shapes")
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Missing 'shapes' attribute or value is not an array for group definition!")
		return
	}

	shapes := &[]shapes.Shape{}
	for _, shape := range shapesAny {
		getShape(shape, shapes, configDirectory)
	}

	if len(*shapes) < 1 {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid shape definition for group definition!")
		return
	}

	*targetArray = append(*targetArray, *CombineShapes(shapes))
}

func getSVG(shapeParameters map[string]any, targetArray *[]shapes.Shape, configDirectory string) {
	filepath, ok := getStringFromAny(shapeParameters["filePath"])
	if !ok {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Missing 'filePath' attribute or value is not an array for group definition!")
		return
	}
	size, ok := getFloatFromAny(shapeParameters["size"])
	if !ok {
		size = 1.0
	}
	if len(filepath) > 0 && filepath[0] == '/' {
		filepath = filepath[1:]
	}
	files := utils.GetFiles(configDirectory + filepath)
	if len(files) != 1 {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Invalid shape definition for type svg: multiple files or no file referenced!")
		return
	}

	xmlTree, err := svgparser.ParseXMLTree(files[0])
	if err != nil {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Could not parse SVG file from '"+filepath+"'")
		return
	}
	var shapes []shapes.Shape
	svg := svgparser.ExtractShapes(xmlTree, false, false)
	shapes = svg.GetShapes(NewConfig().BezierTolerance)
	if len(shapes) == 0 {
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Could not extract any shapes from SVG file '"+filepath+"'")
		return
	}
	
	if len(shapes) != 1 {
		shape := *CombineShapes(&shapes)
		shape.ScaleToSize(size)
		*targetArray = append(*targetArray, shape)
	} else {
		*targetArray = append(*targetArray, shapes[0])
	}
}

func CombineShapes(shapeList *[]shapes.Shape) *shapes.Shape {
	var lines []shapes.Polyline

	for _, shape := range *shapeList {
		lines = append(lines, shape.Lines...)
	}

	return shapes.NewShape(lines)
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

func getPoint(point any, points *[]shapes.Point) {
	switch point.(type) {
	default:
		errorsOcurred = true
		occuredErrors = append(occuredErrors, "Unexpected type for Point | expected array")
		return
	case []any:
		var currentPoint shapes.Point
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
