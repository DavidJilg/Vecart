package general

import (
	"image"
	"image/color"
	"log"
	"math"
	"slices"
	"sync"

	"github.com/DavidJilg/Vecart/internal/utils"

	"github.com/DavidJilg/Vecart/internal/shapes"
)

const (
	TopLeftCorner = iota
	TopRightCorner
	BottomLeftCorner
	BottomRightCorner
	Middle
	LeftSide
	RightSide
	TopSide
	BottomSide
)

type Quadrant struct {
	Id                              uint
	X1, Y1, X2, Y2                  int
	QuadrantType                    int
	Pixels                          [][]Pixel
	FlattenPixels                   []*Pixel
	Neighbors                       []*Quadrant
	Shapes                          []shapes.Shape
	accessMutex                     sync.Mutex
	processingMutex                 sync.Mutex
	border                          shapes.Polyline
	isInNeighborRangeOfCanvasBorder bool
}

func NewQuadrant(img *image.Gray, quadrantId uint, nrOfQuadrants uint, quadrantsPerRow uint, quadrantsPerColumn uint,
	neighborRange uint) *Quadrant {
	xPosition := quadrantId % quadrantsPerRow
	yPosition := uint(math.Floor(float64(quadrantId) / float64(quadrantsPerRow)))

	currentQuadrant := Quadrant{}
	currentQuadrant.Id = uint(quadrantId)
	currentQuadrant.QuadrantType = getQuadrantType(uint(quadrantId), uint(nrOfQuadrants), uint(quadrantsPerRow), uint(quadrantsPerColumn))
	currentQuadrant.Pixels = *getQuadrantPixels(img, int(xPosition), int(yPosition))

	currentQuadrant.X1 = currentQuadrant.getTopLeftPixel().X1
	currentQuadrant.Y1 = currentQuadrant.getTopLeftPixel().Y1

	currentQuadrant.X2 = currentQuadrant.getBottomRightPixel().X2
	currentQuadrant.Y2 = currentQuadrant.getBottomRightPixel().Y2

	currentQuadrant.border = *shapes.NewPolyline(&[]shapes.Point{
		{X: float64(currentQuadrant.X1), Y: float64(currentQuadrant.Y1)},
		{X: float64(currentQuadrant.X2), Y: float64(currentQuadrant.Y1)},
		{X: float64(currentQuadrant.X2), Y: float64(currentQuadrant.Y2)},
		{X: float64(currentQuadrant.X1), Y: float64(currentQuadrant.Y2)},
		{X: float64(currentQuadrant.X1), Y: float64(currentQuadrant.Y1)}}, nil)

	for index := range currentQuadrant.Pixels {
		for index2 := range currentQuadrant.Pixels[index] {
			currentQuadrant.FlattenPixels = append(currentQuadrant.FlattenPixels, &currentQuadrant.Pixels[index][index2])
		}
	}

	if xPosition <= neighborRange || xPosition >= (quadrantsPerRow-uint(neighborRange)-1) {
		currentQuadrant.isInNeighborRangeOfCanvasBorder = true
	}

	if yPosition <= neighborRange || yPosition >= (quadrantsPerColumn-uint(neighborRange)-1) {
		currentQuadrant.isInNeighborRangeOfCanvasBorder = true
	}

	return &currentQuadrant
}

func (quadrant *Quadrant) getIntersectedPixels(shape *shapes.Shape) []*Pixel {
	var intersectedPixels []*Pixel
	for pixelIndex := range quadrant.FlattenPixels {
		for lineIndex := range shape.Lines {
			for i := 0; i < shapes.CountLineIntersections(&shape.Lines[lineIndex], &quadrant.FlattenPixels[pixelIndex].Border); i++ {
				intersectedPixels = append(intersectedPixels, quadrant.FlattenPixels[pixelIndex])
			}
		}
	}
	return intersectedPixels
}

func (quadrant *Quadrant) updateLineIntersects(shape *shapes.Shape, shapeAdded bool) []*Pixel {
	var intersectedPixels []*Pixel

	for pixelIndex := range quadrant.FlattenPixels {
		currentPixel := quadrant.FlattenPixels[pixelIndex]
		for lineIndex := range shape.Lines {
			currentLineIntersections := shapes.CountLineIntersections(&shape.Lines[lineIndex], &quadrant.FlattenPixels[pixelIndex].Border)
			if shapeAdded {
				currentPixel.LineIntersects += currentLineIntersections
			} else {
				currentPixel.LineIntersects -= currentLineIntersections
			}
			for i := 0; i < currentLineIntersections; i++ {
				intersectedPixels = append(intersectedPixels, quadrant.FlattenPixels[pixelIndex])
			}
		}
		currentPixel.AdjustedDarkness = math.Max(0,
			float64(currentPixel.Darkness)-(float64(currentPixel.LineIntersects)*Config.ShapeDarknessFactor))
	}
	return intersectedPixels
}

func (quadrant *Quadrant) addShapeWithoutNeighbors(shape *shapes.Shape) []*Pixel {
	quadrant.Shapes = append(quadrant.Shapes, *shape)
	intersectedPixels := quadrant.updateLineIntersects(shape, true)

	return intersectedPixels
}

func (quadrant *Quadrant) scoreShapes() []*ShapeScore {
	var shapeScores []*ShapeScore

	for index := range quadrant.Shapes {
		currentShape := quadrant.Shapes[index]

		neighborhoodDarknessBefore := quadrant.getAdjustedNeighborhoodDarkness()
		intersectedPixels := quadrant.getIntersectedPixels(&quadrant.Shapes[index])
		for index := range quadrant.Neighbors {
			intersectedPixels = append(intersectedPixels, quadrant.Neighbors[index].getIntersectedPixels(&currentShape)...)
		}
		quadrant.removeShape(index)
		neighborhoodDarknessAfter := quadrant.getAdjustedNeighborhoodDarkness()
		quadrant.addShapeAtIndex(&currentShape, index)
		//Todo test different score types for example with white punishemnt
		score := ((neighborhoodDarknessAfter - neighborhoodDarknessBefore) / float64(len(intersectedPixels))) //- (punishment / float64(len(intersectedPixels)))
		shapeScores = append(shapeScores, &ShapeScore{int(quadrant.Id), index, score})
	}

	return shapeScores
}

func (quadrant *Quadrant) addShapeAtIndex(shape *shapes.Shape, index int) {
	quadrant.accessMutex.Lock()
	quadrant.Shapes = slices.Insert(quadrant.Shapes, index, *shape)
	quadrant.updateLineIntersects(shape, true)
	quadrant.accessMutex.Unlock()

	for index := range quadrant.Neighbors {
		quadrant.Neighbors[index].accessMutex.Lock()
		quadrant.Neighbors[index].updateLineIntersects(shape, true)
		quadrant.Neighbors[index].accessMutex.Unlock()
	}
}

func (quadrant *Quadrant) addShape(shape *shapes.Shape) {
	quadrant.accessMutex.Lock()
	quadrant.Shapes = append(quadrant.Shapes, *shape)
	quadrant.updateLineIntersects(shape, true)
	quadrant.accessMutex.Unlock()

	for index := range quadrant.Neighbors {
		quadrant.Neighbors[index].accessMutex.Lock()
		quadrant.Neighbors[index].updateLineIntersects(shape, true)
		quadrant.Neighbors[index].accessMutex.Unlock()
	}
}

func (quadrant *Quadrant) removeShapeWithoutNeighbors(shapeIndex int) {
	quadrant.updateLineIntersects(&quadrant.Shapes[shapeIndex], false)

	if shapeIndex != len(quadrant.Shapes)-1 {
		quadrant.Shapes[shapeIndex] = quadrant.Shapes[len(quadrant.Shapes)-1]
	}
	quadrant.Shapes = quadrant.Shapes[:len(quadrant.Shapes)-1]
}

func (quadrant *Quadrant) removeShape(shapeIndex int) {
	if shapeIndex >= len(quadrant.Shapes) {
		if utils.DebugModeEnabled() {
			log.Printf("Quadrant.removeShape called with shapeIndex %d while len(Shapes) == %d\n", shapeIndex, len(quadrant.Shapes))
		}
		return
	}

	for index := range quadrant.Neighbors {
		quadrant.Neighbors[index].accessMutex.Lock()
		quadrant.Neighbors[index].updateLineIntersects(&quadrant.Shapes[shapeIndex], false)
		quadrant.Neighbors[index].accessMutex.Unlock()
	}

	quadrant.accessMutex.Lock()
	defer quadrant.accessMutex.Unlock()

	quadrant.updateLineIntersects(&quadrant.Shapes[shapeIndex], false)

	if shapeIndex != len(quadrant.Shapes)-1 {
		quadrant.Shapes[shapeIndex] = quadrant.Shapes[len(quadrant.Shapes)-1]
	}
	quadrant.Shapes = quadrant.Shapes[:len(quadrant.Shapes)-1]
}

func (quadrant *Quadrant) getAdjustedDarkness() float64 {
	darkness := 0.0
	for _, pixel := range quadrant.FlattenPixels {
		darkness += pixel.AdjustedDarkness
	}

	return darkness / float64(len(quadrant.FlattenPixels))
}

func (quadrant *Quadrant) getAdjustedNeighborhoodDarkness() float64 {
	neighborhoodDarkness := quadrant.getAdjustedDarkness()

	for index := range quadrant.Neighbors {
		currentNeighbor := quadrant.Neighbors[index]

		currentNeighbor.accessMutex.Lock()
		neighborhoodDarkness += currentNeighbor.getAdjustedDarkness()
		currentNeighbor.accessMutex.Unlock()
	}

	return neighborhoodDarkness / float64(len(quadrant.Neighbors))
}

func (quadrant *Quadrant) getAdjustedDarkestPixel() *Pixel {
	quadrant.accessMutex.Lock()
	defer quadrant.accessMutex.Unlock()
	maxDarkness := math.MaxFloat64 * -1
	maxDarknessPixel := quadrant.FlattenPixels[0]
	for index := range quadrant.FlattenPixels {
		currentPixel := quadrant.FlattenPixels[index]
		darkness := currentPixel.AdjustedDarkness
		if darkness > maxDarkness {
			maxDarkness = darkness
			maxDarknessPixel = currentPixel
		}
	}

	return maxDarknessPixel
}

func (quadrant *Quadrant) isDone() bool {
	quadrant.accessMutex.Lock()
	defer quadrant.accessMutex.Unlock()
	return quadrant.getAdjustedDarkness() <= Config.DarknessThreshold
}

func (quadrant *Quadrant) getTopLeftPixel() *Pixel {
	return &quadrant.Pixels[0][0]
}

func (quadrant *Quadrant) getBottomRightPixel() *Pixel {
	return &quadrant.Pixels[len(quadrant.Pixels)-1][len(quadrant.Pixels[len(quadrant.Pixels)-1])-1]
}

func getQuadrantType(quadrantId uint, nrOfQuadrants uint, quadrantsPerRow uint, quadrantsPerColumn uint) int {
	if quadrantId == 0 {
		return TopLeftCorner
	}
	if quadrantId == nrOfQuadrants-1 {
		return BottomRightCorner
	}
	if quadrantId == quadrantsPerColumn-1 {
		return TopRightCorner
	}
	if quadrantId == ((quadrantsPerRow) * (quadrantsPerColumn - 1)) {
		return BottomLeftCorner
	}
	if (int(quadrantId) - int(quadrantsPerRow)) < 0 {
		return TopSide
	}
	if quadrantId+quadrantsPerRow > nrOfQuadrants-1 {
		return BottomSide
	}
	if (quadrantId % quadrantsPerRow) == 0 {
		return LeftSide
	}
	if ((quadrantId + 1) % quadrantsPerRow) == 0 {
		return RightSide
	}

	return Middle
}

func getQuadrantPixels(img *image.Gray, xPosition int, yPosition int) *[][]Pixel {
	xStart := Config.QuadrantWidth * xPosition
	yStart := Config.QuadrantHeight * yPosition
	xEnd := xStart + Config.QuadrantWidth - 1
	yEnd := yStart + Config.QuadrantHeight - 1

	var pixels [][]Pixel

	for i := xStart; i <= xEnd; i++ {
		var currentRow []Pixel
		for j := yStart; j <= yEnd; j++ {
			color := int(255 - (*img).At(int(i), int(j)).(color.Gray).Y)
			currentRow = append(currentRow, *NewPixel(int(i), int(j), int(i+1), int(j+1), color))
		}
		pixels = append(pixels, currentRow)
	}

	return &pixels
}

func (quadrant *Quadrant) isIntersectedByLine(line *shapes.Polyline) bool {
	return shapes.IntersectingPolylines(line, &quadrant.border)
}
