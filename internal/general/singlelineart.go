package general

import (
	"fmt"
	"image"
	"image/draw"
	"math"
	"math/rand/v2"
	"sort"
	"sync"
	"time"

	"github.com/DavidJilg/Vecart/internal/utils"

	"github.com/DavidJilg/Vecart/internal/shapes"
)

var nrOfFinishedQuadrants int

func StartSingleLineArtGeneration() string {
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

	initializeQuadrants(greyscaleImg, calculateNeighborRange())

	return generateSingleLineArt(greyscaleImg.Bounds().Max.X, greyscaleImg.Bounds().Max.Y)
}

func generateSingleLineArt(artworkWidth, artworkHeight int) string {
	spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerFinishedFrame = "✓"
	currentSpinnerFrame = 0
	spinnerUpdateFrequency = time.Millisecond * 100
	stopSpinnerMutex = sync.Mutex{}
	start := time.Now()
	//wg := sync.WaitGroup{}

	//alreadyFinishedQuadrants := float64(len(quadrants)) - float64(countUnfinishedQuadrants(&quadrants))
	finishedQuadrants := make(map[int]bool)

	for index := range quadrants {
		finishedQuadrants[index] = false
		if quadrants[index].isDone() {
			finishedQuadrants[index] = true
		} else {
			finishedQuadrants[index] = false
		}
	}

	//wg.Add(1)
	//go monitorQuadrants(&wg, alreadyFinishedQuadrants, "Drawing Line")

	singleLine := shapes.Polyline{}
	currentQuadrant := getUnfinishedQuadrant(RandSource)
	singleLine.Points = append(singleLine.Points, currentQuadrant.getAdjustedDarkestPixel().midpoint)

	unfinishedQuadrantsByDistance := *getUnfinishedQuadrantsByDistance(&singleLine.Points[0], finishedQuadrants)

	OPTIONS := 10
	iterations := 0
	for !allQuadrantsDone(&finishedQuadrants) && iterations < 1000 {
		fmt.Printf("%s %d\n", "Iteration", iterations)
		iterations++

		var lineOptions []shapes.Polyline
		for i := 0; i < OPTIONS; i++ {
			lineOptions = append(lineOptions, *shapes.NewPolyline(&[]shapes.Point{singleLine.Points[len(singleLine.Points)-1],
				unfinishedQuadrantsByDistance[i].Key.getAdjustedDarkestPixel().midpoint}, nil))
		}

		var lineScores []float64
		for shapeIndex := range lineOptions {
			lineScores = append(lineScores, scoreLine(lineOptions[shapeIndex]))
		}

		bestLine := lineOptions[0]
		bestScore := lineScores[0]
		for scoreIndex := range lineScores {
			if lineScores[scoreIndex] > bestScore {
				bestScore = lineScores[scoreIndex]
				bestLine = lineOptions[scoreIndex]
			}
		}
		singleLine.Points = append(singleLine.Points, bestLine.Points[1])
		addLine(&bestLine)
	}

	return generateSVG(artworkWidth, artworkHeight, []*shapes.Shape{shapes.NewShape([]shapes.Polyline{singleLine})}, start)
}

func addLine(line *shapes.Polyline) {
	for index := range quadrants {
		if quadrants[index].isIntersectedByLine(line) {
			fmt.Println("Adding Line")
			quadrants[index].updateLineIntersects(shapes.NewShape([]shapes.Polyline{*line}), true)
		}
	}
}

func scoreLine(line shapes.Polyline) float64 {
	adjustedDarknessBefore := 0.0
	adjustedDarknessAfter := 0.0
	punishment := 0.0
	var intersectedQuadrants []*Quadrant
	var intersectedPixels []*Pixel

	for index := range quadrants {
		if quadrants[index].isIntersectedByLine(&line) {
			intersectedQuadrants = append(intersectedQuadrants, quadrants[index])
			adjustedDarknessBefore += quadrants[index].getAdjustedDarkness()
			intersectedPixels = append(intersectedPixels,
				quadrants[index].updateLineIntersects(shapes.NewShape([]shapes.Polyline{line}), true)...)
			adjustedDarknessAfter += quadrants[index].getAdjustedDarkness()
			quadrants[index].updateLineIntersects(shapes.NewShape([]shapes.Polyline{line}), false)
		}
	}

	for _, pixel := range intersectedPixels {
		if pixel.Darkness <= Config.WhitePunishmentBoundry {
			punishment += Config.WhitePunishmentValue
		}
	}

	return (((adjustedDarknessBefore - adjustedDarknessAfter) / float64(len(intersectedPixels))) -
		(punishment / float64(len(intersectedPixels)))) + line.Points[0].DistanceTo(&line.Points[1])/float64(Config.ArtworkHeight)
}

type kv struct {
	Key   *Quadrant
	Value float64
}

func getUnfinishedQuadrantsByDistance(from *shapes.Point, finishedQuadrants map[int]bool) *[]kv {
	var unfinishedQuadrants []kv
	var distance float64
	for index, quadrant := range quadrants {
		if finishedQuadrants[index] {
			continue
		}
		distance = from.DistanceTo(&quadrant.getAdjustedDarkestPixel().midpoint)
		if distance > 0 {
			unfinishedQuadrants = append(unfinishedQuadrants, kv{Key: quadrants[index], Value: distance})
		}
	}
	sort.Slice(unfinishedQuadrants, func(i, j int) bool {
		return unfinishedQuadrants[i].Value < unfinishedQuadrants[j].Value
	})

	return &unfinishedQuadrants
}

func allQuadrantsDone(finishedQuadrants *map[int]bool) bool {
	for _, done := range *finishedQuadrants {
		if !done {
			return false
		}
	}
	return true
}
