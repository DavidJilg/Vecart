package general

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"math/rand/v2"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DavidJilg/Vecart/internal/utils"

	"github.com/DavidJilg/Vecart/internal/shapes"
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
var lastAjustedDarknessChange int

var finishQuadrantsMutex sync.Mutex
var finishQuadrantsStop bool

var canvasBorder shapes.Polyline

var halfTimeoutReached bool

func resetStaticShapeArtVariables() {
	quadrants = []*Quadrant{}
	ShapeCount = 0
	ShapeCountMutex = nil
	spinnerFrames = []string{}
	spinnerFinishedFrame = ""
	currentSpinnerFrame = 0
	spinnerUpdateFrequency = 0
	stopSpinnerBool = false
	stopSpinnerMutex = sync.Mutex{}
	lastAjustedDarkness = 0.0
	lastAjustedDarknessChange = 0
	finishQuadrantsMutex = sync.Mutex{}
	finishQuadrantsStop = false
	canvasBorder = shapes.Polyline{}
	halfTimeoutReached = false
}

func initializeQuadrants(image *image.Gray, neighborRange int) {
	quadrantsPerRow := (*image).Bounds().Max.X / Config.QuadrantWidth
	quadrantsPerColumn := (*image).Bounds().Max.Y / Config.QuadrantHeight

	nrOfQuadrants := quadrantsPerRow * quadrantsPerColumn

	for quadrantId := 0; quadrantId < nrOfQuadrants; quadrantId++ {
		quadrants = append(quadrants, NewQuadrant(image, uint(quadrantId), uint(nrOfQuadrants),
			uint(quadrantsPerRow), uint(quadrantsPerColumn), uint(neighborRange)))
	}

	calculateNeighbors(quadrantsPerRow, neighborRange)
}

func generateShapeVariants(shape *shapes.Shape) {
	//Todo do not rotate circles
	//Test

	if Config.ShapeAngleDeviationRange <= 0 {
		shape.Variants = append(shape.Variants, shape.Copy())
		return
	}

	currentAngle := 0.0
	for currentAngle <= Config.ShapeAngleDeviationRange {
		shape.GetVariant(currentAngle)
		currentAngle += Config.ShapeAngleDeviationStep
	}
	currentAngle = (0 - Config.ShapeAngleDeviationStep)
	for currentAngle >= (0 - Config.ShapeAngleDeviationRange) {
		shape.GetVariant(currentAngle)
		currentAngle -= Config.ShapeAngleDeviationStep
	}
}

func calculateNeighborRange() int {
	maxSize := math.MaxFloat64 * -1

	for index := range Config.ShapeList {
		maxX, maxY := Config.ShapeList[index].GetMaxSize()
		if maxX > maxSize {
			maxSize = maxX
		}
		if maxY > maxSize {
			maxSize = maxY
		}
	}

	return int(math.Ceil((maxSize / 2) / math.Min(float64(Config.QuadrantHeight), float64(Config.QuadrantWidth))))
}

func calculateNeighbors(quadrantsPerRow int, neighborRange int) {
	var quadrantGrid [][]*Quadrant
	quandrantCoordinates := make(map[*Quadrant]shapes.Point)
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
		quandrantCoordinates[quadrants[index]] = shapes.Point{X: float64(currentRow), Y: float64(currentColumn)}
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
	for index := range Config.ShapeList {
		Config.ShapeList[index].MMToPixel(Config.ProcessingDpi)
		Config.ShapeList[index].CenterOnOrigin()
		generateShapeVariants(&Config.ShapeList[index])
		//Config.shapes[index].ensureOriginCover() //TODO reinstate
	}
}

func initializeForShapeGeneration(image *image.Gray, neighborRange int) {
	resetStaticShapeArtVariables()
	initializeQuadrants(image, neighborRange)
	canvasBorder = getCanvasBorder()
	initializeShapes()
}

func getCanvasBorder() shapes.Polyline {
	canvasBorder := shapes.NewPolygon(&[]shapes.Point{
		{X: 0, Y: 0},
		{X: float64(Config.ArtworkWidth), Y: 0},
		{X: float64(Config.ArtworkWidth), Y: float64(Config.ArtworkWidth)},
		{X: 0, Y: float64(Config.ArtworkWidth)},
	}).ToPolyline()
	return *canvasBorder
}

func StartShapeArtGeneration() string {
	RandSource = rand.New(rand.NewPCG(uint64(Config.RandomSeed), uint64(Config.RandomSeed)))

	var img image.Image
	var err error
	if Config.InputPath == "" {
		ellieFile, err := StaticAssets.Open("static/ellie.png")
		if err != nil {
			fmt.Println("Can not open image 'ellie.png' from static Vecart ressources!")
			fmt.Println(err)
			return ""
		}
		defer ellieFile.Close()

		img, _, err = image.Decode(ellieFile)
		if err != nil {
			fmt.Println("Can not decode image 'ellie.png' from static Vecart ressources!")
			fmt.Println(err)
			return ""
		}
	} else {
		img, err = utils.GetImageFromFilePath(Config.InputPath)
		if err != nil {
			fmt.Printf("Can not decode image '%s'!\n", Config.InputPath)
			fmt.Println(err)
			return ""
		}
	}

	if Config.ArtworkWidth != 0 && Config.ArtworkHeight != 0 {
		imageWidthPixel := int(math.Round(utils.MMToPixel(float64(Config.ArtworkWidth), Config.ProcessingDpi)))
		imageHeightPixel := int(math.Round(utils.MMToPixel(float64(Config.ArtworkHeight), Config.ProcessingDpi)))

		imageWidth := imageWidthPixel - (imageWidthPixel % Config.QuadrantWidth)
		imageHeight := imageHeightPixel - (imageHeightPixel % Config.QuadrantHeight)

		img = ResizeImage(img, imageWidth, imageHeight)
	} else if Config.ArtworkWidth != 0 || Config.ArtworkHeight != 0 {
		imageWidthPixel := int(math.Round(utils.MMToPixel(float64(Config.ArtworkWidth), Config.ProcessingDpi)))
		imageHeightPixel := int(math.Round(utils.MMToPixel(float64(Config.ArtworkHeight), Config.ProcessingDpi)))
		img = ResizeImage(img, imageWidthPixel, imageHeightPixel)

		imageWidth := img.Bounds().Max.X - (img.Bounds().Max.X % Config.QuadrantWidth)
		imageHeight := img.Bounds().Max.Y - (img.Bounds().Max.Y % Config.QuadrantHeight)

		img = ResizeImage(img, imageWidth, imageHeight)
	}

	greyscaleImg := image.NewGray(img.Bounds())
	draw.Draw(greyscaleImg, greyscaleImg.Bounds(), img, img.Bounds().Min, draw.Src)

	initializeForShapeGeneration(greyscaleImg, calculateNeighborRange())

	return generateShapeArt(greyscaleImg.Bounds().Max.X, greyscaleImg.Bounds().Max.Y)
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

func scoreShape(quadrant *Quadrant, shape *shapes.Shape) float64 {
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

	if len(intersectedPixels) == 0 && utils.DebugModeEnabled() {
		log.Println("Placed shape intersects no pixels!")
	}

	for _, pixel := range intersectedPixels {
		if pixel.Darkness <= Config.WhitePunishmentBoundry {
			punishment += Config.WhitePunishmentValue
		}
	}

	neighborhoodDarknessBefore /= nrOfQuadrants
	neighborhoodDarknessAfter /= nrOfQuadrants

	//Todo Remove
	return ((neighborhoodDarknessBefore - neighborhoodDarknessAfter) / float64(len(intersectedPixels))) - (punishment / float64(len(intersectedPixels)))
}

func finishQuadrants(wg *sync.WaitGroup, randSource *rand.Rand) {
	defer wg.Done()

	currentQuadrant := getUnfinishedQuadrant(randSource)

	for currentQuadrant != nil {
		finishQuadrantsMutex.Lock()
		if finishQuadrantsStop {
			if utils.DebugModeEnabled() {
				log.Println("\nfinishQuadrants routine ending early since it was requested by the montoring routine.")
			}
			finishQuadrantsMutex.Unlock()
			return
		}
		finishQuadrantsMutex.Unlock()

		darkestPixelMidpoint := currentQuadrant.getAdjustedDarkestPixel().midpoint
		var pixelMidpoints []*shapes.Point
		if Config.HighPrecisionShapePositioning {
			for pixelIndex := range currentQuadrant.FlattenPixels {
				pixelMidpoints = append(pixelMidpoints, &currentQuadrant.FlattenPixels[pixelIndex].midpoint)
			}
		}

		var shapeCopies []*shapes.Shape
		currentQuadrant.accessMutex.Lock() //Why lock here?
		for shapeIndex := range Config.ShapeList {
			for shapeVariantIndex := range Config.ShapeList[shapeIndex].Variants {
				if !Config.HighPrecisionShapePositioning {
					//shapeCopy := Config.ShapeList[shapeIndex].Variants[shapeVariantIndex].TransformCopy(darkestPixelMidpoint.X, darkestPixelMidpoint.Y)
					//if !currentQuadrant.isInNeighborRangeOfCanvasBorder && !shapeCopy.Intersects(&canvasBorder) {
					shapeCopies = append(shapeCopies, Config.ShapeList[shapeIndex].Variants[shapeVariantIndex].TransformCopy(darkestPixelMidpoint.X, darkestPixelMidpoint.Y))
					//}
					continue
				}

				for midPointIndex := range pixelMidpoints {
					//shapeCopy := Config.ShapeList[shapeIndex].Variants[shapeVariantIndex].TransformCopy(pixelMidpoints[midPointIndex].X, pixelMidpoints[midPointIndex].Y)
					//if !currentQuadrant.isInNeighborRangeOfCanvasBorder && !shapeCopy.Intersects(&canvasBorder) {
					shapeCopies = append(shapeCopies, Config.ShapeList[shapeIndex].Variants[shapeVariantIndex].TransformCopy(pixelMidpoints[midPointIndex].X, pixelMidpoints[midPointIndex].Y))
					//}
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

		var bestShapes []*shapes.Shape
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
			if utils.DebugModeEnabled() {
				log.Println("No Best Shape")
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
	if utils.DebugModeEnabled() {
		log.Println("Shapes | Unfinished Quadrants | Avg. adjusted darkness")
		log.Println("------------------------------------------------------")
	}
	nrOfQuadrants := float64(len(quadrants))
	nrOfNotAlreadyFinishedQuadrants := nrOfQuadrants - alreadyFinishedQuadrants
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
			if !utils.DebugModeEnabled() {
				fmt.Printf("\r%s %s %3.2f %s", spinnerFinishedFrame, message, 100.0, "%")
			}
			return
		}

		adjustedDarkness = adjustedDarkness / float64(len(quadrants))
		if adjustedDarkness == lastAjustedDarkness {
			lastAjustedDarknessChange++
		} else {
			lastAjustedDarkness = adjustedDarkness
			lastAjustedDarknessChange = 0
		}

		if !halfTimeoutReached && !Config.HighPrecisionShapePositioning && Config.UpdateFrequency*lastAjustedDarknessChange > int(Config.Timeout/2) {
			Config.HighPrecisionShapePositioning = true
			halfTimeoutReached = true
			if utils.DebugModeEnabled() {
				percentage := 100 * ((nrOfNotAlreadyFinishedQuadrants - nrOfUnfinishedQuadrants) / nrOfNotAlreadyFinishedQuadrants)
				ShapeCountMutex.Lock()
				nrOfShapes := ShapeCount
				ShapeCountMutex.Unlock()
				log.Printf("%05d | %5.0f | %03s", nrOfShapes, nrOfUnfinishedQuadrants, strconv.FormatFloat(adjustedDarkness, 'f', 2, 64))
				log.Printf(" || Half Timeout: No change for %d seconds. Activating high precision shape placement (%3.2f %s).\n", int(Config.Timeout/2), percentage, "s")
			}
		}
		if Config.UpdateFrequency*lastAjustedDarknessChange > Config.Timeout {
			percentage := 100 * ((nrOfNotAlreadyFinishedQuadrants - nrOfUnfinishedQuadrants) / nrOfNotAlreadyFinishedQuadrants)
			if !utils.DebugModeEnabled() {
				fmt.Printf("\r%s %s %3.2f %s", spinnerFinishedFrame, message, percentage, "%")
				fmt.Printf(" || Timeout: No change for %d seconds. Prematurly ending shape placement (%3.2f %s).", Config.Timeout, percentage, "s")
			} else {
				log.Printf("\nNo adjusted darkness change for %d seconds. Stopped at %3.2f %s. Ending line placement early\n", Config.UpdateFrequency*lastAjustedDarknessChange, percentage, "%")
			}
			endFinishQuadrantsRoutines()
			return
		}

		if utils.DebugModeEnabled() {
			ShapeCountMutex.Lock()
			nrOfShapes := ShapeCount
			ShapeCountMutex.Unlock()

			log.Printf("%05d | %5.0f | %03s\n", nrOfShapes, nrOfUnfinishedQuadrants, strconv.FormatFloat(adjustedDarkness, 'f', 2, 64))
			time.Sleep(time.Duration(Config.UpdateFrequency) * time.Second)
			continue
		}
		percentage := 100 * ((nrOfNotAlreadyFinishedQuadrants - nrOfUnfinishedQuadrants) / nrOfNotAlreadyFinishedQuadrants)

		start := time.Now()
		for time.Since(start) < time.Duration(Config.UpdateFrequency)*time.Second {
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
	nrOfShapesToBeRemoved := int(math.Floor(float64(countShapes()) * Config.ShapeRefinementPercentage))
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

func generateShapeArt(artworkWidth, artworkHeight int) string {
	spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerFinishedFrame = "✓"
	currentSpinnerFrame = 0
	spinnerUpdateFrequency = time.Millisecond * 100
	stopSpinnerMutex = sync.Mutex{}
	start := time.Now()

	wg := sync.WaitGroup{}
	ShapeCount = 0
	ShapeCountMutex = &sync.Mutex{}
	alreadyFinishedQuadrants := float64(len(quadrants)) - float64(countUnfinishedQuadrants(&quadrants))

	finishQuadrantsStop = false
	for i := 0; i < Config.ParallelRoutines; i++ {
		wg.Add(1)
		go finishQuadrants(&wg, rand.New(rand.NewPCG(RandSource.Uint64(), RandSource.Uint64())))
	}
	wg.Add(1)
	go monitorQuadrants(&wg, alreadyFinishedQuadrants, "Placing Shapes")
	wg.Wait()

	if halfTimeoutReached {
		halfTimeoutReached = false
		Config.HighPrecisionShapePositioning = false
	}

	if Config.ShapeRefinement {
		for i := 0; i < Config.ShapeRefinementIterations; i++ {
			wg.Add(1)
			stopSpinnerBool = false
			go startSpinner("Removing Worst Shapes", &wg, &stopSpinnerBool, &stopSpinnerMutex)
			removeWorstShapes()
			stopSpinner()
			wg.Wait()

			alreadyFinishedQuadrants = float64(len(quadrants)) - float64(countUnfinishedQuadrants(&quadrants))
			if !utils.DebugModeEnabled() {
				wg.Add(1)
				go monitorQuadrants(&wg, alreadyFinishedQuadrants, "Refining Shapes ("+strconv.FormatInt(int64(i+1), 10)+")")
			}

			finishQuadrantsStop = false

			for range Config.ParallelRoutines {
				wg.Add(1)
				go finishQuadrants(&wg, rand.New(rand.NewPCG(RandSource.Uint64(), RandSource.Uint64())))
			}

			wg.Wait()
		}

	}

	if !utils.DebugModeEnabled() {
		wg.Add(1)
		stopSpinnerBool = false
		go startSpinner("Removing Unecessary Shapes", &wg, &stopSpinnerBool, &stopSpinnerMutex)
		removeWorthlessShapes()
		stopSpinner()
		wg.Wait()
	} else {
		log.Println("\nRemoving Unecessary Shapes")
		removeWorthlessShapes()
	}

	if Config.SmoothEdges {
		if !utils.DebugModeEnabled() {
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
			quadrants[quadrantIndex].Shapes[shapeIndex].PixelToMM(Config.ProcessingDpi)
		}
	}

	if Config.CombineShapes {
		if utils.DebugModeEnabled() {
			log.Println("\nCombining Shapes")
			log.Println("   Before: ", countTotalLines())
		} else {
			wg.Add(1)
			stopSpinnerBool = false
			go startSpinner("Combining Shapes", &wg, &stopSpinnerBool, &stopSpinnerMutex)
		}
		combineLines(Config.CombineShapesIterations)
		stopSpinner()
		if utils.DebugModeEnabled() {
			log.Println("   After: ", countTotalLines())
			log.Println()
		}
	}

	for quadrantIndex := range quadrants {
		for shapeIndex := range quadrants[quadrantIndex].Shapes {
			quadrants[quadrantIndex].Shapes[shapeIndex].MMToPixel(Config.OutputDpi)
		}
	}

	wg.Wait()

	if utils.DebugModeEnabled() {
		log.Println("Generating SVG")
		return generateSVG(artworkWidth, artworkHeight, collectShapes(), start)
	}

	wg.Add(1)
	stopSpinnerBool = false
	go startSpinner("Generating SVG File", &wg, &stopSpinnerBool, &stopSpinnerMutex)
	svg := generateSVG(artworkWidth, artworkHeight, collectShapes(), start)
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

func canvasContains(point *shapes.Point, canvasWidth, canvasHeight float64) bool {
	return !(point.X < 0 || point.X > canvasWidth || point.Y < 0 || point.Y > canvasHeight)
}

func smoothEdges(artworkWidth, artworkHeight float64) {
	for quadrantIndex := range quadrants {
		for shapeIndex := range quadrants[quadrantIndex].Shapes {

			currentShape := &quadrants[quadrantIndex].Shapes[shapeIndex]
			var smoothedLines []shapes.Polyline

			for lineIndex := range currentShape.Lines {
				lineCut := false
				for pointIndex := range currentShape.Lines[lineIndex].Points {
					if !canvasContains(&currentShape.Lines[lineIndex].Points[pointIndex], artworkWidth, artworkHeight) {
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

func isCombinable(line *shapes.Polyline) bool {
	if line.OriginalShape == nil {
		return true
	}

	switch line.OriginalShape.(type) {
	default:
		return false
	case shapes.Polyline, shapes.Polygon:
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
				var notCombinedLines []shapes.Polyline
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
			if utils.DebugModeEnabled() {
				if currentInteration == 1 {
					log.Println("       Combining Shapes finished after 1 Iteration because no shapes were combined in the first iteration.")
				} else {
					log.Printf("       Combining Shapes finished after %d Iterations because no more shapes were combined in the last iteration.\n", currentInteration)
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

func tryToCombineWithNeighborsLines(line *shapes.Polyline, neighbors []*Quadrant) bool {
	for index := range neighbors {
		if tryToCombineWithSpecificNeighborLines(line, neighbors[index]) {
			return true
		}
	}

	return false
}

func tryToCombineWithSpecificNeighborLines(line *shapes.Polyline, neighbor *Quadrant) bool {
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
func canCombineLines(line1, line2 *shapes.Polyline) bool {
	var combinedPoints []shapes.Point

	if line1.Points[0].DistanceTo(&line2.Points[0]) < Config.CombineShapesTolerance {
		for index := len(line1.Points) - 1; index >= 0; index-- {
			combinedPoints = append(combinedPoints, line1.Points[index])
		}
		combinedPoints = append(combinedPoints, line2.Points...)
		line2.Points = combinedPoints
		return true
	}

	if line1.Points[0].DistanceTo(&line2.Points[len(line2.Points)-1]) < Config.CombineShapesTolerance {
		combinedPoints = append(combinedPoints, line2.Points...)
		combinedPoints = append(combinedPoints, line1.Points...)
		line2.Points = combinedPoints
		return true
	}

	if line1.Points[len(line1.Points)-1].DistanceTo(&line2.Points[len(line2.Points)-1]) < Config.CombineShapesTolerance {
		combinedPoints = append(combinedPoints, line1.Points...)
		for index := len(line2.Points) - 1; index >= 0; index-- {
			combinedPoints = append(combinedPoints, line2.Points[index])
		}
		line2.Points = combinedPoints
		return true
	}

	if line1.Points[len(line1.Points)-1].DistanceTo(&line2.Points[0]) < Config.CombineShapesTolerance {
		combinedPoints = append(combinedPoints, line1.Points...)
		combinedPoints = append(combinedPoints, line2.Points...)
		line2.Points = combinedPoints
		return true
	}

	return false
}

func cutLineExcess(line *shapes.Polyline, canvasWidth, canvasHeight float64) []shapes.Polyline {
	var lineSegments []shapes.Polyline

	if line.OriginalShape != nil {
		switch value := line.OriginalShape.(type) {
		default:
			lineSegments = line.GetLineSegments()
			if utils.DebugModeEnabled() {
				log.Println("Invalid original shape type: ", value)
			}
		case *shapes.Circle:
			circle, _ := line.OriginalShape.(*shapes.Circle)
			lineSegments = circle.ToPolyline(120).GetLineSegments()
		case *shapes.Polygon:
			polygon, _ := line.OriginalShape.(*shapes.Polygon)
			lineSegments = polygon.ToPolyline().GetLineSegments()
		}
	} else {
		lineSegments = line.GetLineSegments()
	}

	segmentIndex := 0
	for segmentIndex < len(lineSegments) {
		currentSegment := &lineSegments[segmentIndex]
		p1, p2 := &currentSegment.Points[0], &currentSegment.Points[1]
		p1OutOfBounds := !canvasContains(p1, canvasWidth, canvasHeight)
		p2OutOfBounds := !canvasContains(p2, canvasWidth, canvasHeight)

		if p1OutOfBounds && p2OutOfBounds {
			lineSegments = removePolylineFromSlice(lineSegments, segmentIndex)
			continue
		}

		if p1OutOfBounds && !p2OutOfBounds {
			intersectionPoint, intersection := getCanvasBorderIntersect(p1, p2, canvasWidth, canvasHeight)
			if intersection {
				currentSegment.Points[0] = intersectionPoint
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
				currentSegment.Points[1] = intersectionPoint
			}
			segmentIndex++
			continue
		}

	}

	segmentIndex = 1
	for segmentIndex < len(lineSegments) {
		p1 := &lineSegments[segmentIndex-1].Points[len(lineSegments[segmentIndex-1].Points)-1]
		p2 := &lineSegments[segmentIndex].Points[0]
		if p1.EqualTo(p2, 10) {
			lineSegments[segmentIndex-1].Points = append(lineSegments[segmentIndex-1].Points, lineSegments[segmentIndex].Points...)
			lineSegments = removePolylineFromSlice(lineSegments, segmentIndex)
			continue
		}
		segmentIndex++
	}

	return lineSegments
}

func removePolylineFromSlice(slice []shapes.Polyline, elementIndex int) []shapes.Polyline {
	return append(slice[:elementIndex], slice[elementIndex+1:]...)
}

func getCanvasBorderIntersect(p1, p2 *shapes.Point, canvasWidth, canvasHeight float64) (shapes.Point, bool) {
	topLeft := shapes.NewPoint(0, 0)
	topRight := shapes.NewPoint(canvasWidth, 0)
	bottomRight := shapes.NewPoint(canvasWidth, canvasHeight)
	bottomLeft := shapes.NewPoint(0, canvasHeight)

	intersectionPoint, intersection := shapes.CalculateLineIntersection(p1, p2, topLeft, topRight)
	if !intersection {
		intersectionPoint, intersection = shapes.CalculateLineIntersection(p1, p2, topRight, bottomRight)
	}
	if !intersection {
		intersectionPoint, intersection = shapes.CalculateLineIntersection(p1, p2, bottomRight, bottomLeft)
	}

	if !intersection {
		intersectionPoint, intersection = shapes.CalculateLineIntersection(p1, p2, bottomLeft, topLeft)
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

/* For Debugging purposes only
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
	WriteStringToFile(svg, filepath)
}*/

func collectShapes() []*shapes.Shape {
	var shapes []*shapes.Shape

	if Config.ReverseShapeOrder {
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

func generateSVG(artworkWidth, artworkHeight int, shapeList []*shapes.Shape, start time.Time) string {
	artworkHeightPixel := utils.MMToPixel(utils.PixelToMM(float64(artworkHeight), Config.ProcessingDpi), Config.OutputDpi)
	artworkWidthPixel := utils.MMToPixel(utils.PixelToMM(float64(artworkWidth), Config.ProcessingDpi), Config.OutputDpi)
	artworkHeightPixelString := strconv.FormatFloat(artworkHeightPixel, 'f', 2, 64)
	artworkWidthPixelString := strconv.FormatFloat(artworkWidthPixel, 'f', 2, 64)

	var svgLines []string
	svgLines = append(svgLines, "<?xml version=\"1.0\"?>")
	svgLines = append(svgLines, "<!-- Generated by Vecart v. "+utils.GetVersion())
	svgLines = append(svgLines, "     https://github.com/DavidJilg/Vecart")
	svgLines = append(svgLines, "     https://david-jilg.com/vecart")
	if Config.ConfigInOutput {
		svgLines = append(svgLines, "\nConfig:")
		if Config.ShortConfig {
			svgLines = append(svgLines, CurrentConfigEntry.toShortJson())
		} else {
			svgLines = append(svgLines, CurrentConfigEntry.originalConfig.ToJson())
		}

	}

	if Config.StatsInOutput {
		svgLines = append(svgLines, "\nStats:")
		svgLines = append(svgLines, "    Nr. of Shapes: "+strconv.Itoa(len(shapeList)))

		nrOfLines := 0
		for _, shape := range shapeList {
			nrOfLines += len(shape.Lines)
		}
		svgLines = append(svgLines, "    Nr. of Lines: "+strconv.Itoa(nrOfLines))

		nrOfLineSegments := 0
		for _, shape := range shapeList {
			for _, line := range shape.Lines {
				nrOfLineSegments += len(line.Points) - 1
			}
		}
		svgLines = append(svgLines, "    Nr. of Line Segments: "+strconv.Itoa(nrOfLineSegments))

		var end time.Time
		if utils.TestModeEnabled() {
			start = time.Unix(2942992800, 0)
			end = time.Unix(2942996400, 0)
		} else {
			end = time.Now()
		}

		svgLines = append(svgLines, "\nTime: ")
		svgLines = append(svgLines, fmt.Sprintf("    Start: %s", start.Format(time.RFC822)))
		svgLines = append(svgLines, fmt.Sprintf("    End: %s", end.Format(time.RFC822)))
		svgLines = append(svgLines, fmt.Sprintf("    Duration: %s", end.Sub(start).Round(time.Second)))

	}

	svgLines = append(svgLines, "-->")
	svgLines = append(svgLines, "<svg viewBox=\"0 0 "+artworkWidthPixelString+" "+artworkHeightPixelString+"\" xmlns=\"http://www.w3.org/2000/svg\">")

	style := "stroke:" + Config.StrokeColor + "; fill:none; stroke-width: " + strconv.FormatFloat(Config.StrokeWidth, 'f', 2, 64) + "px"

	if strings.ToLower(Config.BackgroundColor) != "none" {
		background := shapes.NewPolygon(&[]shapes.Point{
			{X: 0, Y: 0},
			{X: float64(artworkWidthPixel), Y: 0},
			{X: float64(artworkWidthPixel), Y: float64(artworkHeightPixel)},
			{X: 0, Y: float64(artworkHeightPixel)}})
		svgLines = append(svgLines, background.ToSVG("fill: "+Config.BackgroundColor+"; stroke:none;"))
	}

	/*
		for _, quadrant := range quadrants {
			line := quadrant.border
			line.MMToPixel(Config.OutputDpi)
			if quadrant.isInNeighborRangeOfCanvasBorder {
				svgLines = append(svgLines, line.ToSVG("stroke:green; fill:none; stroke-width: 1px"))
				continue
			}
			svgLines = append(svgLines, line.ToSVG("stroke:red; fill:none; stroke-width: 1px"))
		}*/

	for _, shape := range shapeList {
		svgLines = append(svgLines, shape.ToSVG(style))
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
