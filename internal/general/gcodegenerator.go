package general

import (
	"cmp"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"

	"github.com/DavidJilg/Vecart/internal/shapes"
	"github.com/DavidJilg/Vecart/internal/svgparser"
	"github.com/DavidJilg/Vecart/internal/utils"
)

type gcodeShapeSortEntry struct {
	shape         any
	primary       float64
	secondary     float64
	originalIndex int
}

func ConvertSVGToGCode() string {
	RandSource = rand.New(rand.NewPCG(uint64(Config.RandomSeed), uint64(Config.RandomSeed)))

	svgFile, err := utils.GetFileContentsFromFilePath(Config.InputPath)
	if err != nil {
		fmt.Println("Could not read SVG file from '" + Config.InputPath + "'")
		return ""
	}

	xmlTree, err := svgparser.ParseXMLTree(svgFile)
	if err != nil {
		fmt.Println("Could not parse SVG file from '" + Config.InputPath + "'")
		return ""
	}

	svg := svgparser.ExtractShapes(xmlTree, true, Config.KeepSVGGroups)

	if Config.SortGCode {
		sortSVGForGCode(svg)
	}

	gcodeText := GenerateGCode(svg)

	return gcodeText
}

func sortSVGForGCode(svg *svgparser.SVG) {
	if !Config.StampMode {
		flattenSVGGroupsForGCodeSorting(svg)
	}
	svg.Shapes = serpentineSortShapesForGCode(
		svg.Shapes,
		svgparser.Axis(Config.SortAxis),
		Config.BezierTolerance,
		Config.ReverseShapeOrder,
	)
}

func flattenSVGGroupsForGCodeSorting(svg *svgparser.SVG) {
	svg.Shapes = flattenShapesForGCodeSorting(svg.Shapes)
}

func flattenShapesForGCodeSorting(shapeList []any) []any {
	flattenedShapes := make([]any, 0, len(shapeList))
	for _, shape := range shapeList {
		group, ok := shape.(*shapes.Group)
		if !ok {
			flattenedShapes = append(flattenedShapes, shape)
			continue
		}

		flattenedShapes = append(flattenedShapes, flattenShapesForGCodeSorting(group.Shapes)...)
	}

	return flattenedShapes
}

func serpentineSortShapesForGCode(shapeList []any, primaryAxis svgparser.Axis, bezierTolerance float64, reverse bool) []any {
	if len(shapeList) < 2 {
		return shapeList
	}

	sortEntries := buildGCodeShapeSortEntries(shapeList, primaryAxis, bezierTolerance)
	bandCount := autoGCodeSortBandCount(len(sortEntries))
	if bandCount <= 1 {
		slices.SortFunc(sortEntries, func(a, b gcodeShapeSortEntry) int {
			return compareGCodeShapeSortEntries(a, b, true)
		})
	} else {
		minPrimary, maxPrimary := getPrimaryRange(sortEntries)
		if minPrimary == maxPrimary {
			bandCount = 1
			slices.SortFunc(sortEntries, func(a, b gcodeShapeSortEntry) int {
				return compareGCodeShapeSortEntries(a, b, true)
			})
		} else {
			sortEntries = sortShapesSerpentineByBands(sortEntries, bandCount, minPrimary, maxPrimary)
		}
	}

	if reverse {
		slices.Reverse(sortEntries)
	}

	sortedShapes := make([]any, len(sortEntries))
	for index, entry := range sortEntries {
		sortedShapes[index] = entry.shape
	}

	return sortedShapes
}

func buildGCodeShapeSortEntries(shapeList []any, primaryAxis svgparser.Axis, bezierTolerance float64) []gcodeShapeSortEntry {
	secondaryAxis := svgparser.Axis(svgparser.YAxis)
	if primaryAxis == svgparser.YAxis {
		secondaryAxis = svgparser.XAxis
	}

	sortEntries := make([]gcodeShapeSortEntry, 0, len(shapeList))
	for index, shape := range shapeList {
		sortEntries = append(sortEntries, gcodeShapeSortEntry{
			shape:         shape,
			primary:       svgparser.GetStrokeStart(shape, primaryAxis, bezierTolerance, Config.StampMode),
			secondary:     svgparser.GetStrokeStart(shape, secondaryAxis, bezierTolerance, Config.StampMode),
			originalIndex: index,
		})
	}

	return sortEntries
}

func autoGCodeSortBandCount(shapeCount int) int {
	if shapeCount <= 1 {
		return shapeCount
	}

	bandCount := int(math.Round(math.Sqrt(float64(shapeCount))))
	if bandCount < 4 {
		bandCount = 4
	}
	if bandCount > 24 {
		bandCount = 24
	}
	if bandCount > shapeCount {
		bandCount = shapeCount
	}

	return bandCount
}

func getPrimaryRange(sortEntries []gcodeShapeSortEntry) (float64, float64) {
	minPrimary := sortEntries[0].primary
	maxPrimary := sortEntries[0].primary
	for _, entry := range sortEntries[1:] {
		if entry.primary < minPrimary {
			minPrimary = entry.primary
		}
		if entry.primary > maxPrimary {
			maxPrimary = entry.primary
		}
	}

	return minPrimary, maxPrimary
}

func sortShapesSerpentineByBands(sortEntries []gcodeShapeSortEntry, bandCount int, minPrimary float64, maxPrimary float64) []gcodeShapeSortEntry {
	primarySpan := maxPrimary - minPrimary
	bands := make([][]gcodeShapeSortEntry, bandCount)
	for _, entry := range sortEntries {
		normalizedPrimary := (entry.primary - minPrimary) / primarySpan
		bandIndex := int(normalizedPrimary * float64(bandCount))
		if bandIndex >= bandCount {
			bandIndex = bandCount - 1
		}

		bands[bandIndex] = append(bands[bandIndex], entry)
	}

	orderedEntries := make([]gcodeShapeSortEntry, 0, len(sortEntries))
	for bandIndex, band := range bands {
		ascendingSecondary := bandIndex%2 == 0
		slices.SortFunc(band, func(a, b gcodeShapeSortEntry) int {
			return compareGCodeShapeSortEntries(a, b, ascendingSecondary)
		})

		orderedEntries = append(orderedEntries, band...)
	}

	return orderedEntries
}

func compareGCodeShapeSortEntries(a, b gcodeShapeSortEntry, ascendingSecondary bool) int {
	secondaryComparison := cmp.Compare(a.secondary, b.secondary)
	if !ascendingSecondary {
		secondaryComparison *= -1
	}
	if secondaryComparison != 0 {
		return secondaryComparison
	}

	primaryComparison := cmp.Compare(a.primary, b.primary)
	if primaryComparison != 0 {
		return primaryComparison
	}

	return cmp.Compare(a.originalIndex, b.originalIndex)
}

func GenerateGCode(svg *svgparser.SVG) string {
	svg.PixelToMM(Config.OutputDpi)

	gcodeLines := make([]string, 0)
	gcodeLines = append(gcodeLines, utils.G90())
	gcodeLines = append(gcodeLines, utils.G21())
	gcodeLines = append(gcodeLines, parseCustomGCode(Config.InitialCommands)...)

	if Config.RefillTool {
		gcodeLines = append(gcodeLines, parseCustomGCode(Config.RefillToolCommands)...)
	}
	shapeCount := 0
	for _, shape := range svg.Shapes {
		if Config.RefillTool && shapeCount >= Config.NrOfShapesWithoutRefill {
			shapeCount = 0
			gcodeLines = append(gcodeLines, parseCustomGCode(Config.RefillToolCommands)...)
		}
		switch shapeType := shape.(type) {
		case *shapes.Circle:
			gcodeLines = append(gcodeLines, shape.(*shapes.Circle).ToGCode(
				Config.ActivateToolCommands, Config.DeactivateToolCommands, Config.StampMode, Config.FeedRateXY)...)
		case *shapes.Group:
			gcodeLines = append(gcodeLines, shape.(*shapes.Group).ToGCode(
				Config.ActivateToolCommands, Config.DeactivateToolCommands, Config.BezierTolerance, Config.StampMode, Config.FeedRateXY)...)
		case *shapes.Path:
			gcodeLines = append(gcodeLines, shape.(*shapes.Path).ToGCode(
				Config.ActivateToolCommands, Config.DeactivateToolCommands, Config.BezierTolerance, Config.StampMode, Config.FeedRateXY)...)
		case *shapes.Point:
			gcodeLines = append(gcodeLines, shape.(*shapes.Point).ToGCode(
				Config.ActivateToolCommands, Config.DeactivateToolCommands)...)
		case *shapes.Polygon:
			gcodeLines = append(gcodeLines, shape.(*shapes.Polygon).ToGCode(
				Config.ActivateToolCommands, Config.DeactivateToolCommands, Config.StampMode, Config.FeedRateXY)...)
		case *shapes.Polyline:
			gcodeLines = append(gcodeLines, shape.(*shapes.Polyline).ToGCode(
				Config.ActivateToolCommands, Config.DeactivateToolCommands, Config.StampMode, Config.FeedRateXY)...)
		case *shapes.Shape:
			gcodeLines = append(gcodeLines, shape.(*shapes.Shape).ToGCode(
				Config.ActivateToolCommands, Config.DeactivateToolCommands, Config.StampMode, Config.FeedRateXY)...)
		default:
			if utils.DebugModeEnabled() {
				log.Println("Invalid shape type for svg shapes: ", shapeType)
			}
		}
		shapeCount++
	}
	gcodeLines = append(gcodeLines, parseCustomGCode(Config.FinalCommands)...)

	return strings.Join(gcodeLines, "\n")
}

func parseCustomGCode(customGCode []string) []string {
	parsedLines := make([]string, 0, len(customGCode))
	for _, line := range customGCode {
		parseCustomFunctions := parseCustomFunctions(line)
		parsedLines = append(parsedLines, parseCustomFunctions)
	}
	return parsedLines
}

func parseCustomFunctions(line string) string {
	functionsFound := true
	for functionsFound {
		if strings.Contains(line, "RANDOM_POS_XY") {
			parsedLine, err := parseRandomPosXYFunction(line)
			if err != nil {
				log.Println("Error parsing RANDOM_POS_XY function: ", err)
				return ""
			} else {
				line = parsedLine
			}
			continue
		}

		//Other functions can be added here in the future

		functionsFound = false
	}

	return line
}

func parseRandomPosXYFunction(line string) (string, error) {
	//Example line: 'G0 RANDOM_POS_XY(100,50,200,50,200,100,100,100)' This example should move to a random point within the quadrilateral defined by the four points (100,50), (200,50), (200,100), (100,100)

	const RANDOM_POS_XY = "RANDOM_POS_XY"
	startIndex := strings.Index(line, RANDOM_POS_XY)
	startBracketIndex := strings.Index(line[startIndex:], "(")
	if startBracketIndex == -1 {
		return "", fmt.Errorf("Invalid function format for RANDOM_POS_XY, missing opening bracket")
	}
	endBracketIndex := strings.Index(line[startIndex:], ")")
	if endBracketIndex == -1 {
		return "", fmt.Errorf("Invalid function format for RANDOM_POS_XY, missing closing bracket")
	}
	//Extract the numbers from the brackets
	numbersString := line[startIndex+startBracketIndex+1 : startIndex+endBracketIndex]
	numbers := strings.Split(numbersString, ",")
	if len(numbers) != 8 {
		return "", fmt.Errorf("Invalid number of arguments for RANDOM_POS_XY, expected 8 but got %d", len(numbers))
	}
	//Convert the numbers to float64
	var x1, y1, x2, y2, x3, y3, x4, y4 float64
	var err error

	x1, err = strconv.ParseFloat(numbers[0], 64)
	if err != nil {
		return "", fmt.Errorf("Invalid number format for RANDOM_POS_XY: %s", numbers[0])
	}
	y1, err = strconv.ParseFloat(numbers[1], 64)
	if err != nil {
		return "", fmt.Errorf("Invalid number format for RANDOM_POS_XY: %s", numbers[1])
	}
	x2, err = strconv.ParseFloat(numbers[2], 64)
	if err != nil {
		return "", fmt.Errorf("Invalid number format for RANDOM_POS_XY: %s", numbers[2])
	}
	y2, err = strconv.ParseFloat(numbers[3], 64)
	if err != nil {
		return "", fmt.Errorf("Invalid number format for RANDOM_POS_XY: %s", numbers[3])
	}
	x3, err = strconv.ParseFloat(numbers[4], 64)
	if err != nil {
		return "", fmt.Errorf("Invalid number format for RANDOM_POS_XY: %s", numbers[4])
	}
	y3, err = strconv.ParseFloat(numbers[5], 64)
	if err != nil {
		return "", fmt.Errorf("Invalid number format for RANDOM_POS_XY: %s", numbers[5])
	}
	x4, err = strconv.ParseFloat(numbers[6], 64)
	if err != nil {
		return "", fmt.Errorf("Invalid number format for RANDOM_POS_XY: %s", numbers[6])
	}
	y4, err = strconv.ParseFloat(numbers[7], 64)
	if err != nil {
		return "", fmt.Errorf("Invalid number format for RANDOM_POS_XY: %s", numbers[7])
	}

	var ax, ay, bx, by, cx, cy, dx, dy float64
	if ((x3-x1)*(y2-y1)-(y3-y1)*(x2-x1))*((x3-x1)*(y4-y1)-(y3-y1)*(x4-x1)) <= 0 {
		ax, ay, bx, by, cx, cy, dx, dy = x1, y1, x2, y2, x3, y3, x4, y4
	} else {
		ax, ay, bx, by, cx, cy, dx, dy = x2, y2, x1, y1, x4, y4, x3, y3
	}

	area1 := math.Abs((bx-ax)*(cy-ay) - (by-ay)*(cx-ax))
	area2 := math.Abs((cx-ax)*(dy-ay) - (cy-ay)*(dx-ax))
	totalArea := area1 + area2
	if totalArea == 0 {
		return "", fmt.Errorf("Invalid quadrilateral for RANDOM_POS_XY: area is zero")
	}

	r := RandSource.Float64() * totalArea
	u := r / area1
	if r >= area1 {
		r -= area1
		u = r / area2
		bx, by, cx, cy = cx, cy, dx, dy
	}
	v := RandSource.Float64()
	if u+v > 1 {
		u = 1 - u
		v = 1 - v
	}
	x := ax + u*(bx-ax) + v*(cx-ax)
	y := ay + u*(by-ay) + v*(cy-ay)

	result := make([]byte, 0, len(line)+32)
	result = append(result, line[:startIndex]...)
	result = append(result, 'X')
	result = strconv.AppendFloat(result, x, 'f', utils.GCodePrecision, 64)
	result = append(result, ' ', 'Y')
	result = strconv.AppendFloat(result, y, 'f', utils.GCodePrecision, 64)
	result = append(result, line[startIndex+endBracketIndex+1:]...)
	return string(result), nil
}
