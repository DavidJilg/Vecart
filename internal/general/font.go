package general

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/DavidJilg/Vecart/internal/utils"

	"github.com/DavidJilg/Vecart/internal/shapes"
)

type Font struct {
	name             string
	characters       map[string]*Character
	spaceWidth       float64
	characterSpacing float64
}

func (font *Font) fromXML(xmlString string) error {
	lineheight := 1.0
	font.spaceWidth = 0.5
	font.characterSpacing = 0.2
	font.characters = make(map[string]*Character)

	root, err := parseXMLTree(xmlString)
	if err != nil {
		return err
	}

	for index := range root.Nodes {
		currentNode := &root.Nodes[index]
		if currentNode.XMLName.Local != "g" {
			continue
		}

		id, found := currentNode.getAttributeByName("id")
		if !found {
			if utils.DebugModeEnabled() {
				log.Println("Font Group has no ID")
			}
			continue
		}

		if strings.Contains(id, "Lineheight") || strings.Contains(id, "lineheight") {
			lineheightShape, _, _ := currentNode.getShape()
			lineheightShape.CenterOnOrigin()
			_, lineheight = lineheightShape.GetSize()
			continue
		}

		var symbol string
		if strings.HasPrefix(id, "ASCII") {
			asciiValueString, _ := strings.CutPrefix(id, "ASCII")
			asciiValue, err := strconv.ParseInt(asciiValueString, 10, 8)
			if err != nil {
				if utils.DebugModeEnabled() {
					log.Printf("Invalid ASCII value %s\n", asciiValueString)
				}
				continue
			}
			symbol = string(rune(asciiValue))
		}

		if strings.HasPrefix(id, "UTF16") {
			utfValueString, _ := strings.CutPrefix(id, "UTF16")
			utfValue, err := strconv.ParseInt(utfValueString, 10, 16)
			if err != nil {
				if utils.DebugModeEnabled() {
					log.Printf("Invalid UTF16 value %s\n", utfValueString)
				}
				continue
			}
			symbol = string(rune(utfValue))
		}

		var currentChar Character
		currentChar.symbol = symbol

		currentChar.shape, currentChar.alignment, currentChar.offset = currentNode.getShape()
		currentChar.shape.CenterOnOrigin()
		currentChar.width, currentChar.height = currentChar.shape.GetSize()

		if len(currentChar.shape.Lines) > 0 {
			font.characters[currentChar.symbol] = &currentChar
		} else {
			if utils.DebugModeEnabled() {
				log.Printf("Character definition for symbol %s (%s) has no valid shape definition.\n", currentChar.symbol, id)
			}
		}

	}

	if len(font.characters) < 1 {
		if utils.DebugModeEnabled() {
			log.Printf("Font '%s' has no characters\n", font.name)
		}
		return nil
	}

	font.normalize(lineheight)

	return nil
}

func LoadFonts() map[string]*Font {
	fonts := make(map[string]*Font)
	svgFile, err := StaticAssets.Open("static/fonts/IBM-Plex-Sans.svg")
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Println("Can not open font file 'IBM-Plex-Sans.svg' from static Vecart ressources!")
			log.Println(err)
		}
		return fonts
	}
	defer svgFile.Close()

	xmlString, err := utils.GetFileContentsFromStaticAssets(svgFile)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Println("Can not read font file 'IBM-Plex-Sans.svg' from static Vecart ressources!")
			log.Println(err)
		}
		return fonts
	}

	font := Font{}
	font.name = "IBM-Plex-Sans"
	err = font.fromXML(xmlString)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Println("Can not parse font file 'IBM-Plex-Sans.svg' from static Vecart ressources!")
			log.Println(err)
		}
		return fonts
	}

	fonts["IBM-Plex-Sans"] = &font

	return fonts
}

func (font *Font) GetText(text string, lineheight float64, center shapes.Point) shapes.Shape {
	var lines []shapes.Polyline
	currentX := 0.0

	for _, char := range text {
		if string(char) == " " {
			currentX += font.spaceWidth * lineheight
			continue
		}

		currentCharacter, found := font.characters[string(char)]
		if !found {
			if utils.DebugModeEnabled() {
				log.Printf("Font '%s' does not support character '%c'\n", font.name, char)
			}
			continue
		}

		scaledCharacterCopy := currentCharacter.shape.ScaleCopy(lineheight)
		width, height := scaledCharacterCopy.GetSize()
		yTransform := 0.0
		switch currentCharacter.alignment {
		default:
			fmt.Printf(" Invalid charactet alignment '%d'\n", currentCharacter.alignment)
			panic(15)
		case Top:
			scaledCharacterCopy.TopLeftOnOrigin()
		case Middle:
			scaledCharacterCopy.CenterLeftOnOrigin()
			yTransform = (lineheight / 2)
		case Bottom:
			scaledCharacterCopy.BottomLeftOnOrigin()
			yTransform = lineheight
		}
		scaledCharacterCopy.Move(currentX, yTransform+(currentCharacter.offset*height))

		lines = append(lines, scaledCharacterCopy.Lines...)

		currentX += (font.characterSpacing * lineheight) + width
	}

	shape := shapes.NewShape(lines)
	shape.CenterOnPoint(center)

	return *shape
}

func (font *Font) normalize(lineheight float64) {
	scaleFactor := 1.0 / lineheight

	for index := range font.characters {
		font.characters[index].scale(scaleFactor)
		font.characters[index].shape.CenterOnOrigin()
	}
}

func (node *Node) getShape() (shapes.Shape, int, float64) {
	alignment := Bottom
	var offset float64
	var lines []shapes.Polyline

	for index := range node.Nodes {
		currentChildNode := node.Nodes[index]
		switch currentChildNode.XMLName.Local {
		default:
			if utils.DebugModeEnabled() {
				log.Printf(" Unsupported XMLTag for Shape definition '%s'\n", currentChildNode.XMLName.Local)
			}
			continue
		case "line":
			currentChildNode.getLine(&lines)
		case "polyline":
			currentChildNode.getPolyline(&lines)
		case "polygon":
			currentChildNode.getPolygon(&lines)
		case "circle":
			currentChildNode.getCircle(&lines)
		case "text":
			alignment, offset = getPositionInformation(string(currentChildNode.Content), alignment, offset)
		}
	}

	return *shapes.NewShape(lines), alignment, offset
}

func getPositionInformation(text string, alignment int, offset float64) (int, float64) {
	parts := strings.Split(text, " ")
	if len(parts) != 2 {
		if utils.DebugModeEnabled() {
			log.Printf(" Invalid alignment or offset text '%s'\n", text)
		}
		return alignment, offset
	}

	switch parts[0] {
	default:
		if utils.DebugModeEnabled() {
			log.Printf(" Unsupported position information '%s'\n", parts[0])
		}
		return alignment, offset

	case "offset", "ofset", "Offset", "Ofset":
		offsetValue, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			if utils.DebugModeEnabled() {
				log.Printf(" Invalid offset value '%s'\n", parts[1])
			}
			return alignment, offset
		}
		offset = offsetValue

	case "alignment":
		switch parts[1] {
		default:
			if utils.DebugModeEnabled() {
				log.Printf(" Unsupported alignment position '%s'\n", parts[1])
			}
			return alignment, offset
		case "top":
			return Top, offset
		case "middle":
			return Middle, offset
		case "bottom":
			return Bottom, offset
		}
	}

	return alignment, offset
}

func (node *Node) getLine(lines *[]shapes.Polyline) {
	x1String, found := node.getAttributeByName("x1")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Line definition for character '%s' is missing x1 attribute\n", node.XMLName.Local)
		}
		return
	}
	x2String, found := node.getAttributeByName("x2")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Line definition for character '%s' is missing x2 attribute\n", node.XMLName.Local)
		}
		return
	}
	y1String, found := node.getAttributeByName("y1")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Line definition for character '%s' is missing y1 attribute\n", node.XMLName.Local)
		}
		return
	}
	y2String, found := node.getAttributeByName("y2")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Line definition for character '%s' is missing y2 attribute\n", node.XMLName.Local)
		}
		return
	}

	x1, err := strconv.ParseFloat(x1String, 64)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Line definition for character '%s' has an invalid x1 value\n", node.XMLName.Local)
		}
		return
	}

	x2, err := strconv.ParseFloat(x2String, 64)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Line definition for character '%s' has an invalid x2 value\n", node.XMLName.Local)
		}
		return
	}

	y1, err := strconv.ParseFloat(y1String, 64)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Line definition for character '%s' has an invalid y1 value\n", node.XMLName.Local)
		}
		return
	}

	y2, err := strconv.ParseFloat(y2String, 64)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Line definition for character '%s' has an invalid y2 value\n", node.XMLName.Local)
		}
		return
	}

	*lines = append(*lines, *shapes.NewPolyline(&[]shapes.Point{{X: x1, Y: y1}, {X: x2, Y: y2}}, nil))
}

func (node *Node) getCircle(lines *[]shapes.Polyline) {
	cxString, found := node.getAttributeByName("cx")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Circle definition for character '%s' is missing cx attribute\n", node.XMLName.Local)
		}
		return
	}
	cyString, found := node.getAttributeByName("cy")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Circle definition for character '%s' is missing cy attribute\n", node.XMLName.Local)
		}
		return
	}
	rString, found := node.getAttributeByName("r")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Circle definition for character '%s' is missing r attribute\n", node.XMLName.Local)
		}
		return
	}

	x, err := strconv.ParseFloat(cxString, 64)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Circle definition for character '%s' has an invalid cx value\n", node.XMLName.Local)
		}
		return
	}

	y, err := strconv.ParseFloat(cyString, 64)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Circle definition for character '%s' has an invalid cy value\n", node.XMLName.Local)
		}
		return
	}

	r, err := strconv.ParseFloat(rString, 64)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Circle definition for character '%s' has an invalid r value\n", node.XMLName.Local)
		}
		return
	}

	*lines = append(*lines, *shapes.NewCircle(shapes.Point{X: x, Y: y}, r).ToPolyline(6))
}

func (node *Node) getPolyline(lines *[]shapes.Polyline) {
	pointsString, found := node.getAttributeByName("points")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Polyline definition for character '%s' is missing points attribute\n", node.XMLName.Local)
		}
		return
	}
	points := parsePointsString(pointsString)
	if len(points) < 2 {
		if utils.DebugModeEnabled() {
			log.Printf("Polyline definition for character '%s' has no valid points definition\n", node.XMLName.Local)
		}
		return
	}
	*lines = append(*lines, *shapes.NewPolyline(&points, nil))
}

func (node *Node) getPolygon(lines *[]shapes.Polyline) {
	pointsString, found := node.getAttributeByName("points")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Polygon definition for character '%s' is missing points attribute\n", node.XMLName.Local)
		}
		return
	}
	points := parsePointsString(pointsString)
	if len(points) < 2 {
		if utils.DebugModeEnabled() {
			log.Printf("Polygon definition for character '%s' has no valid points definition\n", node.XMLName.Local)
		}
		return
	}
	*lines = append(*lines, *shapes.NewPolygon(&points).ToPolyline())
}

func CleanString(value string) string {
	value = strings.ReplaceAll(value, "\\", "/")
	value = strings.ReplaceAll(value, "\a", "")
	value = strings.ReplaceAll(value, "\b", "")
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\t", "")
	value = strings.ReplaceAll(value, "\v", "")
	value = strings.ReplaceAll(value, "\n", "")
	return value
}

func parsePointsString(pointsString string) []shapes.Point {
	pointsString = CleanString(pointsString)
	pointsString = strings.TrimSpace(pointsString)

	var points []shapes.Point
	pointStrings := strings.Split(pointsString, " ")
	if strings.Contains(pointsString, ",") {
		for _, pointString := range pointStrings {
			coordinates := strings.Split(pointString, ",")
			if len(coordinates) != 2 {
				if utils.DebugModeEnabled() {
					log.Printf("Point definition '%s' is invalid\n", pointString)
				}
				continue
			}

			x, err := strconv.ParseFloat(coordinates[0], 64)
			if err != nil {
				if utils.DebugModeEnabled() {
					log.Printf("Point definition '%s' is invalid\n", pointString)
				}
				continue
			}
			y, err := strconv.ParseFloat(coordinates[1], 64)
			if err != nil {
				if utils.DebugModeEnabled() {
					log.Printf("Point definition '%s' is invalid\n", pointString)
				}
				continue
			}

			points = append(points, *shapes.NewPoint(x, y))
		}
	} else {
		if len(pointStrings) < 2 {
			if utils.DebugModeEnabled() {
				log.Printf("Point definition '%s' is invalid\n", pointsString)
			}
			return points
		}

		for i := 1; i < len(pointStrings); i *= 2 {
			x, err := strconv.ParseFloat(pointStrings[i-1], 64)
			if err != nil {
				if utils.DebugModeEnabled() {
					log.Printf("Point  definition has an invalid X value '%s'\n", pointStrings[i-1])
				}
				continue
			}
			y, err := strconv.ParseFloat(pointStrings[i], 64)
			if err != nil {
				if utils.DebugModeEnabled() {
					log.Printf("Point  definition has an invalid X value '%s'\n", pointStrings[i])
				}
				continue
			}

			points = append(points, *shapes.NewPoint(x, y))
		}
	}
	return points
}

func (node *Node) getAttributeByName(name string) (string, bool) {
	for index := range node.Attrs {
		if node.Attrs[index].Name.Local == name {
			return node.Attrs[index].Value, true
		}
	}

	return "", false
}

const (
	Top = iota
	Bottom
)

type Character struct {
	width     float64
	height    float64
	symbol    string
	shape     shapes.Shape
	alignment int
	offset    float64
}

func (character *Character) scale(factor float64) {
	character.shape.Scale(factor)
}
