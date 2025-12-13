package main

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

var nrOfFinishedQuadrants int

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

	singleLine := Polyline{}
	currentQuadrant := getUnfinishedQuadrant(RandSource)
	singleLine.points = append(singleLine.points, currentQuadrant.getAdjustedDarkestPixel().midpoint)

	unfinishedQuadrantsByDistance := *getUnfinishedQuadrantsByDistance(&singleLine.points[0], finishedQuadrants)

	OPTIONS := 10
	iterations := 0
	for !allQuadrantsDone(&finishedQuadrants) && iterations < 1000 {
		fmt.Printf("%s %d\n", "Iteration", iterations)
		iterations++

		var lineOptions []Polyline
		for i := 0; i < OPTIONS; i++ {
			lineOptions = append(lineOptions, *NewPolyline(&[]Point{singleLine.points[len(singleLine.points)-1],
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
		singleLine.points = append(singleLine.points, bestLine.points[1])
		addLine(&bestLine)
	}

	return generateSVG(artworkWidth, artworkHeight, []*Shape{NewShape([]Polyline{singleLine})}, start)
}

func addLine(line *Polyline) {
	for index := range quadrants {
		if quadrants[index].isIntersectedByLine(line) {
			fmt.Println("Adding Line")
			quadrants[index].updateLineIntersects(NewShape([]Polyline{*line}), true)
		}
	}
}

func scoreLine(line Polyline) float64 {
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
				quadrants[index].updateLineIntersects(NewShape([]Polyline{line}), true)...)
			adjustedDarknessAfter += quadrants[index].getAdjustedDarkness()
			quadrants[index].updateLineIntersects(NewShape([]Polyline{line}), false)
		}
	}

	for _, pixel := range intersectedPixels {
		if pixel.Darkness <= Config.whitePunishmentBoundry {
			punishment += Config.whitePunishmentValue
		}
	}

	return (((adjustedDarknessBefore - adjustedDarknessAfter) / float64(len(intersectedPixels))) -
		(punishment / float64(len(intersectedPixels)))) + line.points[0].distanceTo(&line.points[1])/float64(Config.artworkHeight)
}

type kv struct {
	Key   *Quadrant
	Value float64
}

func getUnfinishedQuadrantsByDistance(from *Point, finishedQuadrants map[int]bool) *[]kv {
	var unfinishedQuadrants []kv
	var distance float64
	for index, quadrant := range quadrants {
		if finishedQuadrants[index] {
			continue
		}
		distance = from.distanceTo(&quadrant.getAdjustedDarkestPixel().midpoint)
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
