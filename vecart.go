package main

import (
	"fmt"
	"image"
	"math"
	"math/rand/v2"
	"sort"
	"strconv"
	"sync"
	"time"
)

var quadrants []*Quadrant
var ShapeCount int
var ShapeCountMutex *sync.Mutex

var spinnerFrames []string
var spinnerFinishedFrame string
var currentSpinnerFrame int
var spinnerUpdateFrequency time.Duration

var stopSpinnerBool bool
var stopSpinnerMutex sync.Mutex

var lastAjustedDarkness float64
var lastAjustedDarknessChange float64

var finishQuadrantsMutex sync.Mutex
var finishQuadrantsStop bool

func initializeQuadrants(image *image.Gray, neighborRange int) {
	quadrantsPerRow := (*image).Bounds().Max.X / Config.quadrantWidth
	quadrantsPerColumn := (*image).Bounds().Max.Y / Config.quadrantHeight

	nrOfQuadrants := quadrantsPerRow * quadrantsPerColumn

	for quadrantId := 0; quadrantId < nrOfQuadrants; quadrantId++ {
		quadrants = append(quadrants, NewQuadrant(image, uint(quadrantId), uint(nrOfQuadrants),
			uint(quadrantsPerRow), uint(quadrantsPerColumn)))
	}

	calculateNeighbors(quadrantsPerRow, neighborRange)
}

func calculateNeighborRange() int {
	maxSize := math.MaxFloat64 * -1

	for index := range Config.shapes {
		maxX, maxY := Config.shapes[index].getMaxSize()
		if maxX > maxSize {
			maxSize = maxX
		}
		if maxY > maxSize {
			maxSize = maxY
		}
	}

	return int(math.Ceil((maxSize / 2) / math.Min(float64(Config.quadrantHeight), float64(Config.quadrantWidth))))
}

func calculateNeighbors(quadrantsPerRow int, neighborRange int) {
	var quadrantGrid [][]*Quadrant
	quandrantCoordinates := make(map[*Quadrant]Point)
	currentRow := 0
	currentColumn := 0
	for i := 0; i < len(quadrants); i++ {
		quadrantGrid = append(quadrantGrid, []*Quadrant{})
	}
	for index := range quadrants {
		if (index%quadrantsPerRow) == 0 && index != 0 {
			currentRow++
			currentColumn = 0
		}
		quandrantCoordinates[quadrants[index]] = Point{float64(currentRow), float64(currentColumn)}
		quadrantGrid[currentRow] = append(quadrantGrid[currentRow], quadrants[index])
		currentColumn++
	}

	for index := range quadrants {
		for i := 0 - neighborRange; i <= neighborRange; i++ {
			for j := 0 - neighborRange; j <= neighborRange; j++ {
				if i == 0 && j == 0 {
					continue
				}
				if int(quandrantCoordinates[quadrants[index]].X)+i >= len(quadrantGrid) {
					continue
				}
				if int(quandrantCoordinates[quadrants[index]].X)+i < 0 {
					continue
				}
				if int(quandrantCoordinates[quadrants[index]].Y)+j >= len(quadrantGrid[int(quandrantCoordinates[quadrants[index]].X)+i]) {
					continue
				}
				if int(quandrantCoordinates[quadrants[index]].Y)+j < 0 {
					continue
				}
				(quadrants)[index].Neighbors = append((quadrants)[index].Neighbors, quadrantGrid[int(quandrantCoordinates[quadrants[index]].X)+i][int(quandrantCoordinates[quadrants[index]].Y)+j])
			}
		}
	}
}

func initializeShapes() {
	for index := range Config.shapes {
		Config.shapes[index].mmToPixel(Config.processingDpi)
		Config.shapes[index].centerOnOrigin()
		Config.shapes[index].generateShapeVariants()
		//Config.shapes[index].ensureOriginCover() //TODO reinstate
	}
}

func initialize(image *image.Gray, neighborRange int) {
	initializeQuadrants(image, neighborRange)
	initializeShapes()
}

func countUnfinishedQuadrants(quadrantList *[]*Quadrant) int {
	remainingQuadrants := 0
	for index := range *quadrantList {
		if !(*quadrantList)[index].isDone() {
			remainingQuadrants++
		}
	}
	return remainingQuadrants
}

func getUnfinishedQuadrant(randSource *rand.Rand) *Quadrant {
	startIndex := randSource.IntN(len(quadrants))

	for index := startIndex; index < len(quadrants); index++ {
		if !quadrants[index].processingMutex.TryLock() {
			continue
		}
		if !quadrants[index].isDone() {
			return quadrants[index]
		}
		quadrants[index].processingMutex.Unlock()
	}

	for index := startIndex - 1; index >= 0; index-- {
		if !quadrants[index].processingMutex.TryLock() {
			continue
		}
		if !quadrants[index].isDone() {
			return quadrants[index]
		}
		quadrants[index].processingMutex.Unlock()
	}

	return nil
}

func scoreShape(quadrant *Quadrant, shape *Shape) float64 {
	neighborhoodDarknessBefore := 0.0
	neighborhoodDarknessAfter := 0.0
	punishment := 0.0
	var intersectedPixels []*Pixel

	quadrant.accessMutex.Lock()
	nrOfQuadrants := float64(len(quadrant.Neighbors))
	neighborhoodDarknessBefore += quadrant.getAdjustedDarkness()
	intersectedPixels = append(intersectedPixels, quadrant.addShapeWithoutNeighbors(shape)...)
	neighborhoodDarknessAfter += quadrant.getAdjustedDarkness()
	quadrant.removeShapeWithoutNeighbors(len(quadrant.Shapes) - 1)
	quadrant.accessMutex.Unlock()

	for index := range quadrant.Neighbors {
		currentNeighbor := quadrant.Neighbors[index]

		currentNeighbor.accessMutex.Lock()
		neighborhoodDarknessBefore += currentNeighbor.getAdjustedDarkness()
		intersectedPixels = append(intersectedPixels, currentNeighbor.addShapeWithoutNeighbors(shape)...)
		neighborhoodDarknessAfter += currentNeighbor.getAdjustedDarkness()
		currentNeighbor.removeShapeWithoutNeighbors(len(currentNeighbor.Shapes) - 1)
		currentNeighbor.accessMutex.Unlock()
	}

	if len(intersectedPixels) == 0 && Config.debug{
		fmt.Println("Placed shape intersects no pixels!")
	}

	for _, pixel := range intersectedPixels {
		if pixel.Darkness <= Config.whitePunishmentBoundry {
			punishment += Config.whitePunishmentValue
		}
	}

	neighborhoodDarknessBefore /= nrOfQuadrants
	neighborhoodDarknessAfter /= nrOfQuadrants

	//Todo Remove
	score := ((neighborhoodDarknessBefore - neighborhoodDarknessAfter) / float64(len(intersectedPixels))) - (punishment / float64(len(intersectedPixels)))
	return score
}

func finishQuadrants(wg *sync.WaitGroup, randSource *rand.Rand) {
	defer wg.Done()

	currentQuadrant := getUnfinishedQuadrant(randSource)

	for currentQuadrant != nil {
		finishQuadrantsMutex.Lock()
		if finishQuadrantsStop {
			if Config.debug {
				fmt.Println("\nfinishQuadrants routine ending early since it was requested by the montoring routine.")
			}
			finishQuadrantsMutex.Unlock()
			return
		}
		finishQuadrantsMutex.Unlock()

		darkestPixelMidpoint := currentQuadrant.getAdjustedDarkestPixel().midpoint
		var pixelMidpoints []*Point
		if Config.highPrecisionShapePositioning {
			for pixelIndex := range currentQuadrant.FlattenPixels {
				pixelMidpoints = append(pixelMidpoints, &currentQuadrant.FlattenPixels[pixelIndex].midpoint)
			}
		}

		var shapeCopies []*Shape
		currentQuadrant.accessMutex.Lock()
		for shapeIndex := range Config.shapes {
			for shapeVariantIndex := range Config.shapes[shapeIndex].Variants {
				if !Config.highPrecisionShapePositioning {
					shapeCopies = append(shapeCopies, Config.shapes[shapeIndex].Variants[shapeVariantIndex].transformCopy(darkestPixelMidpoint.X, darkestPixelMidpoint.Y))
					continue
				}

				for midPointIndex := range pixelMidpoints {
					shapeCopies = append(shapeCopies, Config.shapes[shapeIndex].Variants[shapeVariantIndex].transformCopy(pixelMidpoints[midPointIndex].X, pixelMidpoints[midPointIndex].Y))
				}
			}
		}

		currentQuadrant.accessMutex.Unlock()

		var shapeScores []float64
		for shapeIndex := range shapeCopies {
			shapeScores = append(shapeScores, scoreShape(currentQuadrant, shapeCopies[shapeIndex]))
		}

		bestShapeScore := math.MaxFloat64 * -1
		for _, score := range shapeScores {
			if score > bestShapeScore {
				bestShapeScore = score
			}
		}

		var bestShapes []*Shape
		for index, score := range shapeScores {
			if score == bestShapeScore {
				bestShapes = append(bestShapes, shapeCopies[index])
			}
		}

		if len(bestShapes) != 0 {
			ShapeCountMutex.Lock()
			ShapeCount++
			ShapeCountMutex.Unlock()
			currentQuadrant.addShape(bestShapes[randSource.IntN(len(bestShapes))])

		} else {
			if Config.debug {
				fmt.Println("No Best Shape")
			}
			currentQuadrant.processingMutex.Unlock()
			currentQuadrant = getUnfinishedQuadrant(randSource)
			continue
		}

		if currentQuadrant.isDone() {
			currentQuadrant.processingMutex.Unlock()
			currentQuadrant = getUnfinishedQuadrant(randSource)
		}
	}
}

func endFinishQuadrantsRoutines() {
	finishQuadrantsMutex.Lock()
	finishQuadrantsStop = true
	finishQuadrantsMutex.Unlock()
}

func monitorQuadrants(wg *sync.WaitGroup, alreadyFinishedQuadrants float64, message string) {
	defer wg.Done()
	fmt.Println()
	if Config.debug {
		fmt.Println("Shapes | Unfinished Quadrants | Avg. adjusted darkness")
		fmt.Println("------------------------------------------------------")
	}
	nrOfQuadrants := float64(len(quadrants))
	nrOfNotAlreadyFinishedQuadrants := nrOfQuadrants - alreadyFinishedQuadrants
	updateFrequency := time.Second * 2
	finishedQuadrants := make(map[int]bool)
	for index := range quadrants {
		finishedQuadrants[index] = false
	}

	lastAjustedDarkness = math.MaxFloat64
	lastAjustedDarknessChange = 0
	for {
		nrOfUnfinishedQuadrants := 0.0
		adjustedDarkness := 0.0
		for index := range quadrants {
			if !finishedQuadrants[index] {
				if quadrants[index].isDone() {
					finishedQuadrants[index] = true
				} else {
					nrOfUnfinishedQuadrants++
				}

				quadrants[index].accessMutex.Lock()
				adjustedDarkness += quadrants[index].getAdjustedDarkness()
				quadrants[index].accessMutex.Unlock()

			}
		}

		if nrOfUnfinishedQuadrants == 0 {
			if !Config.debug {
				fmt.Printf("\r%s %s %3.2f %s", spinnerFinishedFrame, message, 100.0, "%")
			}
			return
		}

		adjustedDarkness = adjustedDarkness / float64(len(quadrants))
		if adjustedDarkness == lastAjustedDarkness {
			lastAjustedDarknessChange++
			if int(updateFrequency.Seconds()*lastAjustedDarknessChange) > Config.timeout {
				if !Config.debug {
					fmt.Printf("\r%s %s %3.2f %s", spinnerFinishedFrame, message, 100.0, "%")
				} else {
					fmt.Printf("\nNo adjusted darkness change for %d seconds. Ending line placement early\n", int(updateFrequency.Seconds()*lastAjustedDarknessChange))
				}
				endFinishQuadrantsRoutines()
				return
			}
		} else {
			lastAjustedDarkness = adjustedDarkness
			lastAjustedDarknessChange = 0
		}

		if Config.debug {
			ShapeCountMutex.Lock()
			nrOfShapes := ShapeCount
			ShapeCountMutex.Unlock()

			fmt.Printf("%05d | %5.0f | %03s\n", nrOfShapes, nrOfUnfinishedQuadrants, strconv.FormatFloat(adjustedDarkness, 'f', 2, 64))
			time.Sleep(updateFrequency)
			continue
		}
		percentage := 100 * ((nrOfNotAlreadyFinishedQuadrants - nrOfUnfinishedQuadrants) / nrOfNotAlreadyFinishedQuadrants)

		start := time.Now()
		for time.Since(start) < updateFrequency {
			fmt.Printf("\r%s %s %3.2f %s", spinnerFrames[currentSpinnerFrame], message, percentage, "%")

			currentSpinnerFrame++
			if currentSpinnerFrame > len(spinnerFrames)-1 {
				currentSpinnerFrame = 0
			}

			time.Sleep(spinnerUpdateFrequency)
		}
	}

}

func countShapes() int {
	nrOfShapes := 0

	for index := range quadrants {
		nrOfShapes += len(quadrants[index].Shapes)
	}

	return nrOfShapes
}

type ShapeScore struct {
	quadrantIndex int
	shapeIndex    int
	score         float64
}

func removeWorstShapes() {
	nrOfShapesToBeRemoved := int(math.Floor(float64(countShapes()) * Config.shapeRefinementPercentage))
	var shapeScores []*ShapeScore

	for index := range quadrants {
		shapeScores = append(shapeScores, quadrants[index].scoreShapes()...)
	}

	sort.Slice(shapeScores, func(i, j int) bool {
		return shapeScores[i].score < shapeScores[j].score
	})

	var toBeRemoved []*ShapeScore
	for i := 0; i < nrOfShapesToBeRemoved; i++ {
		toBeRemoved = append(toBeRemoved, shapeScores[i])
	}

	removeShapes(toBeRemoved)
}

func removeWorthlessShapes() {
	var shapeScores []*ShapeScore

	for index := range quadrants {
		shapeScores = append(shapeScores, quadrants[index].scoreShapes()...)
	}

	var toBeRemoved []*ShapeScore
	for index := range shapeScores {
		if shapeScores[index].score == 0 {
			toBeRemoved = append(toBeRemoved, shapeScores[index])
		}
	}

	removeShapes(toBeRemoved)
}

func removeShapes(toBeRemoved []*ShapeScore) {
	for quadrantIndex := 0; quadrantIndex < len(quadrants); quadrantIndex++ {
		var currentShapes []*ShapeScore
		for index := range toBeRemoved {
			if toBeRemoved[index].quadrantIndex == quadrantIndex {
				currentShapes = append(currentShapes, toBeRemoved[index])
			}
		}
		if len(currentShapes) == 0 {
			continue
		}
		sort.Slice(currentShapes, func(i, j int) bool {
			return currentShapes[i].shapeIndex < currentShapes[j].shapeIndex
		})
		for shapeIndex := len(currentShapes) - 1; shapeIndex >= 0; shapeIndex-- {
			quadrants[quadrantIndex].removeShape(currentShapes[shapeIndex].shapeIndex)
		}

	}
}

func generateVectorArt(artworkWidth, artworkHeight int) string {
	spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerFinishedFrame = "✓"
	currentSpinnerFrame = 0
	spinnerUpdateFrequency = time.Millisecond * 100
	stopSpinnerMutex = sync.Mutex{}

	wg := sync.WaitGroup{}
	ShapeCount = 0
	ShapeCountMutex = &sync.Mutex{}
	alreadyFinishedQuadrants := float64(len(quadrants)) - float64(countUnfinishedQuadrants(&quadrants))

	finishQuadrantsStop = false
	for i := 0; i < Config.parallelRoutines; i++ {
		wg.Add(1)
		go finishQuadrants(&wg, rand.New(rand.NewPCG(RandSource.Uint64(), RandSource.Uint64())))
	}
	wg.Add(1)
	go monitorQuadrants(&wg, alreadyFinishedQuadrants, "Placing Shapes")
	wg.Wait()

	if Config.shapeRefinement {
		for i := 0; i < Config.shapeRefinementIterations; i++ {
			wg.Add(1)
			stopSpinnerBool = false
			go startSpinner("Removing Worst Shapes", &wg, &stopSpinnerBool, &stopSpinnerMutex)
			removeWorstShapes()
			stopSpinner()
			wg.Wait()

			alreadyFinishedQuadrants = float64(len(quadrants)) - float64(countUnfinishedQuadrants(&quadrants))
			if !Config.debug {
				wg.Add(1)
				go monitorQuadrants(&wg, alreadyFinishedQuadrants, "Refining Shapes ("+strconv.FormatInt(int64(i+1), 10)+")")
			}

			finishQuadrantsStop = false

			for range Config.parallelRoutines {
				wg.Add(1)
				go finishQuadrants(&wg, rand.New(rand.NewPCG(RandSource.Uint64(), RandSource.Uint64())))
			}

			wg.Wait()
		}

	}

	if !Config.debug {
		wg.Add(1)
		stopSpinnerBool = false
		go startSpinner("Removing Unecessary Shapes", &wg, &stopSpinnerBool, &stopSpinnerMutex)
		removeWorthlessShapes()
		stopSpinner()
		wg.Wait()
	} else {
		fmt.Println("\nRemoving Unecessary Shapes")
		removeWorthlessShapes()
	}

	if Config.smoothEdges {
		if !Config.debug {
			wg.Add(1)
			stopSpinnerBool = false
			go startSpinner("Smoothing Edges", &wg, &stopSpinnerBool, &stopSpinnerMutex)
		}
		smoothEdges(float64(artworkWidth), float64(artworkHeight))
		stopSpinner()

	}
	wg.Wait()

	for quadrantIndex := range quadrants {
		for shapeIndex := range quadrants[quadrantIndex].Shapes {
			quadrants[quadrantIndex].Shapes[shapeIndex].pixelToMM(Config.processingDpi)
		}
	}

	if Config.combineShapes {
		if Config.debug {
			fmt.Println("\nCombining Shapes")
			fmt.Println("   Before: ", countTotalLines())
		} else {
			wg.Add(1)
			stopSpinnerBool = false
			go startSpinner("Combining Shapes", &wg, &stopSpinnerBool, &stopSpinnerMutex)
		}
		combineLines(Config.combineShapesIterations)
		stopSpinner()
		if Config.debug {
			fmt.Println("   After: ", countTotalLines())
			fmt.Println()
		}
	}

	for quadrantIndex := range quadrants {
		for shapeIndex := range quadrants[quadrantIndex].Shapes {
			quadrants[quadrantIndex].Shapes[shapeIndex].mmToPixel(Config.outputDpi)
		}
	}

	wg.Wait()

	
	if Config.debug {
		fmt.Println("Generating SVG")
		return generateSVG(artworkWidth, artworkHeight, collectShapes())
	}

	wg.Add(1)
	stopSpinnerBool = false
	go startSpinner("Generating SVG File", &wg, &stopSpinnerBool, &stopSpinnerMutex)
	svg := generateSVG(artworkWidth, artworkHeight, collectShapes())
	stopSpinner()
	wg.Wait()

	return svg
}

func startSpinner(message string, wg *sync.WaitGroup, stopSpinner *bool, stopSpinnerMutex *sync.Mutex) {
	fmt.Println()
	defer wg.Done()

	for {
		stopSpinnerMutex.Lock()
		if *stopSpinner {
			stopSpinnerMutex.Unlock()
			fmt.Printf("\r%s %s", spinnerFinishedFrame, message)
			return
		}
		stopSpinnerMutex.Unlock()

		fmt.Printf("\r%s %s", spinnerFrames[currentSpinnerFrame], message)

		currentSpinnerFrame++
		if currentSpinnerFrame > len(spinnerFrames)-1 {
			currentSpinnerFrame = 0
		}

		time.Sleep(spinnerUpdateFrequency)
	}
}

func stopSpinner() {
	stopSpinnerMutex.Lock()
	stopSpinnerBool = true
	stopSpinnerMutex.Unlock()
}

func canvasContains(point *Point, canvasWidth, canvasHeight float64) bool {
	return !(point.X < 0 || point.X > canvasWidth || point.Y < 0 || point.Y > canvasHeight)
}

func smoothEdges(artworkWidth, artworkHeight float64) {
	for quadrantIndex := range quadrants {
		for shapeIndex := range quadrants[quadrantIndex].Shapes {

			currentShape := &quadrants[quadrantIndex].Shapes[shapeIndex]
			var smoothedLines []Polyline

			for lineIndex := range currentShape.Lines {
				lineCut := false
				for pointIndex := range currentShape.Lines[lineIndex].points {
					if !canvasContains(&currentShape.Lines[lineIndex].points[pointIndex], artworkWidth, artworkHeight) {
						smoothedLines = append(smoothedLines, cutLineExcess(&currentShape.Lines[lineIndex], artworkWidth, artworkHeight)...)
						lineCut = true
						break
					}
				}
				if !lineCut {
					smoothedLines = append(smoothedLines, currentShape.Lines[lineIndex])
				}
			}
			currentShape.Lines = smoothedLines
		}
	}
}

func isCombinable(line *Polyline) bool {
	if line.originalShape == nil {
		return true
	}

	switch line.originalShape.(type) {
	default:
		return false
	case Polyline, Polygon:
		return true
	}
}

func combineLines(iterations int) {
	defer removeEmptyShapes()

	currentInteration := 0
	for currentInteration < iterations {
		linesCombined := false
		currentInteration++

		for quadrantIndex := range quadrants {
			for shapeIndex := range quadrants[quadrantIndex].Shapes {
				currentShapeLines := &quadrants[quadrantIndex].Shapes[shapeIndex].Lines
				var notCombinedLines []Polyline
				for lineIndex := range *currentShapeLines {
					if !isCombinable(&(*currentShapeLines)[lineIndex]) {
						notCombinedLines = append(notCombinedLines, (*currentShapeLines)[lineIndex])
						continue
					}
					if !tryToCombineWithNeighborsLines(&(*currentShapeLines)[lineIndex], quadrants[quadrantIndex].Neighbors) {
						notCombinedLines = append(notCombinedLines, (*currentShapeLines)[lineIndex])
					} else {
						linesCombined = true
					}
				}
				quadrants[quadrantIndex].Shapes[shapeIndex].Lines = notCombinedLines
			}
		}

		if !linesCombined {
			if Config.debug {
				if currentInteration == 1 {
					fmt.Println("       Combining Shapes finished after 1 Iteration because no shapes were combined in the first iteration.")
				} else {
					fmt.Printf("       Combining Shapes finished after %d Iterations because no more shapes were combined in the last iteration.\n", currentInteration)
				}
			}
			return
		}
	}
}

func removeEmptyShapes() {
	var toBeRemoved []*ShapeScore

	for quadrantIndex := 0; quadrantIndex < len(quadrants); quadrantIndex++ {
		for shapeIndex := range quadrants[quadrantIndex].Shapes {
			if len(quadrants[quadrantIndex].Shapes[shapeIndex].Lines) == 0 {
				toBeRemoved = append(toBeRemoved, &ShapeScore{quadrantIndex, shapeIndex, 0})
			}
		}
	}

	removeShapes(toBeRemoved)
}

func tryToCombineWithNeighborsLines(line *Polyline, neighbors []*Quadrant) bool {
	for index := range neighbors {
		if tryToCombineWithSpecificNeighborLines(line, neighbors[index]) {
			return true
		}
	}

	return false
}

func tryToCombineWithSpecificNeighborLines(line *Polyline, neighbor *Quadrant) bool {
	for shapeIndex := range neighbor.Shapes {
		for lineIndex := range neighbor.Shapes[shapeIndex].Lines {
			if canCombineLines(line, &neighbor.Shapes[shapeIndex].Lines[lineIndex]) {
				return true
			}
		}
	}

	return false
}

// Tries to combine lines by adding points from line1 to line2
func canCombineLines(line1, line2 *Polyline) bool {
	var combinedPoints []Point

	if line1.points[0].distanceTo(&line2.points[0]) < Config.combineShapesTolerance {
		for index := len(line1.points) - 1; index >= 0; index-- {
			combinedPoints = append(combinedPoints, line1.points[index])
		}
		combinedPoints = append(combinedPoints, line2.points...)
		line2.points = combinedPoints
		return true
	}

	if line1.points[0].distanceTo(&line2.points[len(line2.points)-1]) < Config.combineShapesTolerance {
		combinedPoints = append(combinedPoints, line2.points...)
		combinedPoints = append(combinedPoints, line1.points...)
		line2.points = combinedPoints
		return true
	}

	if line1.points[len(line1.points)-1].distanceTo(&line2.points[len(line2.points)-1]) < Config.combineShapesTolerance {
		combinedPoints = append(combinedPoints, line1.points...)
		for index := len(line2.points) - 1; index >= 0; index-- {
			combinedPoints = append(combinedPoints, line2.points[index])
		}
		line2.points = combinedPoints
		return true
	}

	if line1.points[len(line1.points)-1].distanceTo(&line2.points[0]) < Config.combineShapesTolerance {
		combinedPoints = append(combinedPoints, line1.points...)
		combinedPoints = append(combinedPoints, line2.points...)
		line2.points = combinedPoints
		return true
	}

	return false
}

func cutLineExcess(line *Polyline, canvasWidth, canvasHeight float64) []Polyline {
	var lineSegments []Polyline

	if line.originalShape != nil {
		switch value := line.originalShape.(type) {
		default:
			lineSegments = line.getLineSegments()
			if Config.debug {
				fmt.Println("Invalid original shape type: ", value)
			}
		case *Circle:
			circle, _ := line.originalShape.(*Circle)
			lineSegments = circle.toPolyline(120).getLineSegments()
		case *Polygon:
			polygon, _ := line.originalShape.(*Polygon)
			lineSegments = polygon.toPolyline().getLineSegments()
		}
	} else {
		lineSegments = line.getLineSegments()
	}

	segmentIndex := 0
	for segmentIndex < len(lineSegments) {
		currentSegment := &lineSegments[segmentIndex]
		p1, p2 := &currentSegment.points[0], &currentSegment.points[1]
		p1OutOfBounds := !canvasContains(p1, canvasWidth, canvasHeight)
		p2OutOfBounds := !canvasContains(p2, canvasWidth, canvasHeight)

		if p1OutOfBounds && p2OutOfBounds {
			lineSegments = removePolylineFromSlice(lineSegments, segmentIndex)
			continue
		}

		if p1OutOfBounds && !p2OutOfBounds {
			intersectionPoint, intersection := getCanvasBorderIntersect(p1, p2, canvasWidth, canvasHeight)
			if intersection {
				currentSegment.points[0] = intersectionPoint
			}
			segmentIndex++
			continue
		}

		if !p1OutOfBounds && !p2OutOfBounds {
			segmentIndex++
			continue
		}
		if !p1OutOfBounds && p2OutOfBounds {
			intersectionPoint, intersection := getCanvasBorderIntersect(p1, p2, canvasWidth, canvasHeight)
			if intersection {
				currentSegment.points[1] = intersectionPoint
			}
			segmentIndex++
			continue
		}

	}

	segmentIndex = 1
	for segmentIndex < len(lineSegments) {
		p1 := &lineSegments[segmentIndex-1].points[len(lineSegments[segmentIndex-1].points)-1]
		p2 := &lineSegments[segmentIndex].points[0]
		if p1.equalTo(p2, 10) {
			lineSegments[segmentIndex-1].points = append(lineSegments[segmentIndex-1].points, lineSegments[segmentIndex].points...)
			lineSegments = removePolylineFromSlice(lineSegments, segmentIndex)
			continue
		}
		segmentIndex++
	}

	return lineSegments
}

func removePolylineFromSlice(slice []Polyline, elementIndex int) []Polyline {
	return append(slice[:elementIndex], slice[elementIndex+1:]...)
}

func getCanvasBorderIntersect(p1, p2 *Point, canvasWidth, canvasHeight float64) (Point, bool) {
	topLeft := NewPoint(0, 0)
	topRight := NewPoint(canvasWidth, 0)
	bottomRight := NewPoint(canvasWidth, canvasHeight)
	bottomLeft := NewPoint(0, canvasHeight)

	intersectionPoint, intersection := calculateLineIntersection(p1, p2, topLeft, topRight)
	if !intersection {
		intersectionPoint, intersection = calculateLineIntersection(p1, p2, topRight, bottomRight)
	}
	if !intersection {
		intersectionPoint, intersection = calculateLineIntersection(p1, p2, bottomRight, bottomLeft)
	}

	if !intersection {
		intersectionPoint, intersection = calculateLineIntersection(p1, p2, bottomLeft, topLeft)
	}

	return intersectionPoint, intersection
}

func countTotalLines() int {
	nrOfLines := 0
	for quadrantIndex := range quadrants {
		for shapeIndex := range quadrants[quadrantIndex].Shapes {
			nrOfLines += len(quadrants[quadrantIndex].Shapes[shapeIndex].Lines)
		}
	}
	return nrOfLines
}

func mmToPixel(millimeters float64, dpi float64) float64 {
	return (millimeters * dpi) / 25.4
}

func PixelToMM(pixels float64, dpi float64) float64 {
	return (pixels * 25.4) / dpi
}

// For Debugging purposes only
func ShapeToSVGFile(shape Shape, filepath string) {
	var svgLines []string
	svgLines = append(svgLines, "<?xml version=\"1.0\"?>")
	svgLines = append(svgLines, "<!-- Generated by Vecart v. "+Version)
	svgLines = append(svgLines, "     https://github.com/DavidJilg/Vecart")
	svgLines = append(svgLines, "     https://david-jilg.com/vecart")
	svgLines = append(svgLines, "\nDEBUGGING SHAPE")
	svgLines = append(svgLines, "-->")
	svgLines = append(svgLines, "<svg viewBox=\"0 0 500 500\" xmlns=\"http://www.w3.org/2000/svg\">")

	style := "stroke:" + Config.strokeColor + "; fill:none; stroke-width: " + strconv.FormatFloat(Config.strokeWidth, 'f', 2, 64) + "px"

	svgLines = append(svgLines, shape.toSVG(style))

	svgLines = append(svgLines, "</svg>")

	svg := ""
	for index := range svgLines {
		svg += svgLines[index]
		if index != len(svgLines)-1 {
			svg += "\n"
		}
	}
	writeStringToFile(svg, filepath)
}

func collectShapes() []*Shape{
	var shapes []*Shape

	if Config.reverseShapeOrder {
		for quadrantIndex := len(quadrants) - 1; quadrantIndex >= 0; quadrantIndex-- {
			for shapeIndex := range quadrants[quadrantIndex].Shapes {
				shapes = append(shapes, &quadrants[quadrantIndex].Shapes[shapeIndex])
			}
		}
	} else {
		for quadrantIndex := range quadrants {
			for shapeIndex := range quadrants[quadrantIndex].Shapes {
				shapes = append(shapes, &quadrants[quadrantIndex].Shapes[shapeIndex])
			}
		}
	}

	return shapes
}

func generateSVG(artworkWidth, artworkHeight int, shapes []*Shape) string {
	artworkHeightPixel := strconv.FormatFloat(mmToPixel(PixelToMM(float64(artworkHeight), Config.processingDpi), Config.outputDpi), 'f', 2, 64)
	artworkWidthPixel := strconv.FormatFloat(mmToPixel(PixelToMM(float64(artworkWidth), Config.processingDpi), Config.outputDpi), 'f', 2, 64)

	var svgLines []string
	svgLines = append(svgLines, "<?xml version=\"1.0\"?>")
	svgLines = append(svgLines, "<!-- Generated by Vecart v. "+Version)
	svgLines = append(svgLines, "     https://github.com/DavidJilg/Vecart")
	svgLines = append(svgLines, "     https://david-jilg.com/vecart")
	if Config.configInOutput {
		svgLines = append(svgLines, "\nConfig:")
		svgLines = append(svgLines, UserConfig)
	}
	svgLines = append(svgLines, "-->")
	svgLines = append(svgLines, "<svg viewBox=\"0 0 "+artworkWidthPixel+" "+artworkHeightPixel+"\" xmlns=\"http://www.w3.org/2000/svg\">")

	style := "stroke:" + Config.strokeColor + "; fill:none; stroke-width: " + strconv.FormatFloat(Config.strokeWidth, 'f', 2, 64) + "px"

	
	for _, shape := range shapes {
		svgLines = append(svgLines, shape.toSVG(style))
	}
	
	svgLines = append(svgLines, "</svg>")

	svg := ""
	for index := range svgLines {
		svg += svgLines[index]
		if index != len(svgLines)-1 {
			svg += "\n"
		}
	}
	return svg
}
