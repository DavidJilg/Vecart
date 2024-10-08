package main

import (
	"fmt"
	"strconv"
	"strings"
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
			if Config.debug {
				fmt.Println("Font Group has no ID")
			}
			continue
		}

		if strings.Contains(id, "Lineheight") || strings.Contains(id, "lineheight") {
			lineheightShape, _, _ := currentNode.getShape()
			lineheightShape.centerOnOrigin()
			_, lineheight = lineheightShape.getSize()
			continue
		}

		var symbol string
		if strings.HasPrefix(id, "ASCII") {
			asciiValueString, _ := strings.CutPrefix(id, "ASCII")
			asciiValue, err := strconv.ParseInt(asciiValueString, 10, 8)
			if err != nil {
				if Config.debug {
					fmt.Printf("Invalid ASCII value %s\n", asciiValueString)
				}
				continue
			}
			symbol = string(rune(asciiValue))
		}

		if strings.HasPrefix(id, "UTF16") {
			utfValueString, _ := strings.CutPrefix(id, "UTF16")
			utfValue, err := strconv.ParseInt(utfValueString, 10, 16)
			if err != nil {
				if Config.debug {
					fmt.Printf("Invalid UTF16 value %s\n", utfValueString)
				}
				continue
			}
			symbol = string(rune(utfValue))
		}

		var currentChar Character
		currentChar.symbol = symbol

		currentChar.shape, currentChar.alignment, currentChar.offset = currentNode.getShape()
		currentChar.shape.centerOnOrigin()
		currentChar.width, currentChar.height = currentChar.shape.getSize()

		if len(currentChar.shape.Lines) > 0 {
			font.characters[currentChar.symbol] = &currentChar
		} else {
			if Config.debug {
				fmt.Printf("Character definition for symbol %s (%s) has no valid shape definition.\n", currentChar.symbol, id)
			}
		}

	}

	if len(font.characters) < 1 {
		if Config.debug {
			fmt.Printf("Font '%s' has no characters\n", font.name)
		}
		return nil
	}

	font.normalize(lineheight)

	return nil
}

func (font *Font) getText(text string, lineheight float64, center Point) Shape {
	var lines []Polyline
	currentX := 0.0

	for _, char := range text {
		if string(char) == " " {
			currentX += font.spaceWidth * lineheight
			continue
		}

		currentCharacter, found := font.characters[string(char)]
		if !found {
			if Config.debug {
				fmt.Printf("Font '%s' does not support character '%c'\n", font.name, char)
			}
			continue
		}

		scaledCharacterCopy := currentCharacter.shape.scaleCopy(lineheight)
		width, height := scaledCharacterCopy.getSize()
		yTransform := 0.0
		switch currentCharacter.alignment {
		default:
			fmt.Printf(" Invalid charactet alignment '%d'\n", currentCharacter.alignment)
			panic(15)
		case Top:
			scaledCharacterCopy.topLeftOnOrigin()
		case Middle:
			scaledCharacterCopy.centerLeftOnOrigin()
			yTransform = (lineheight / 2)
		case Bottom:
			scaledCharacterCopy.bottomLeftOnOrigin()
			yTransform = lineheight
		}
		scaledCharacterCopy.transform(currentX, yTransform+(currentCharacter.offset*height))

		lines = append(lines, scaledCharacterCopy.Lines...)

		currentX += (font.characterSpacing * lineheight) + width
	}

	shape := NewShape(lines)
	shape.centerOnPoint(center)

	return *shape
}

func (font *Font) normalize(lineheight float64) {
	scaleFactor := 1.0 / lineheight

	for index := range font.characters {
		font.characters[index].scale(scaleFactor)
		font.characters[index].shape.centerOnOrigin()
	}
}

func (node *Node) getShape() (Shape, int, float64) {
	alignment := Bottom
	var offset float64
	var lines []Polyline

	for index := range node.Nodes {
		currentChildNode := node.Nodes[index]
		switch currentChildNode.XMLName.Local {
		default:
			if Config.debug {
				fmt.Printf(" Unsupported XMLTag for Shape definition '%s'\n", currentChildNode.XMLName.Local)
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

	return *NewShape(lines), alignment, offset
}

func getPositionInformation(text string, alignment int, offset float64) (int, float64) {
	parts := strings.Split(text, " ")
	if len(parts) != 2 {
		if Config.debug {
			fmt.Printf(" Invalid alignment or offset text '%s'\n", text)
		}
		return alignment, offset
	}

	switch parts[0] {
	default:
		if Config.debug {
			fmt.Printf(" Unsupported position information '%s'\n", parts[0])
		}
		return alignment, offset

	case "offset", "ofset", "Offset", "Ofset":
		offsetValue, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			if Config.debug {
				fmt.Printf(" Invalid offset value '%s'\n", parts[1])
			}
			return alignment, offset
		}
		offset = offsetValue

	case "alignment":
		switch parts[1] {
		default:
			if Config.debug {
				fmt.Printf(" Unsupported alignment position '%s'\n", parts[1])
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

func (node *Node) getLine(lines *[]Polyline) {
	x1String, found := node.getAttributeByName("x1")
	if !found {
		if Config.debug {
			fmt.Printf("Line definition for character '%s' is missing x1 attribute\n", node.XMLName.Local)
		}
		return
	}
	x2String, found := node.getAttributeByName("x2")
	if !found {
		if Config.debug {
			fmt.Printf("Line definition for character '%s' is missing x2 attribute\n", node.XMLName.Local)
		}
		return
	}
	y1String, found := node.getAttributeByName("y1")
	if !found {
		if Config.debug {
			fmt.Printf("Line definition for character '%s' is missing y1 attribute\n", node.XMLName.Local)
		}
		return
	}
	y2String, found := node.getAttributeByName("y2")
	if !found {
		if Config.debug {
			fmt.Printf("Line definition for character '%s' is missing y2 attribute\n", node.XMLName.Local)
		}
		return
	}

	x1, err := strconv.ParseFloat(x1String, 64)
	if err != nil {
		if Config.debug {
			fmt.Printf("Line definition for character '%s' has an invalid x1 value\n", node.XMLName.Local)
		}
		return
	}

	x2, err := strconv.ParseFloat(x2String, 64)
	if err != nil {
		if Config.debug {
			fmt.Printf("Line definition for character '%s' has an invalid x2 value\n", node.XMLName.Local)
		}
		return
	}

	y1, err := strconv.ParseFloat(y1String, 64)
	if err != nil {
		if Config.debug {
			fmt.Printf("Line definition for character '%s' has an invalid y1 value\n", node.XMLName.Local)
		}
		return
	}

	y2, err := strconv.ParseFloat(y2String, 64)
	if err != nil {
		if Config.debug {
			fmt.Printf("Line definition for character '%s' has an invalid y2 value\n", node.XMLName.Local)
		}
		return
	}

	*lines = append(*lines, *NewPolyline(&[]Point{{x1, y1}, {x2, y2}}, nil))
}

func (node *Node) getCircle(lines *[]Polyline) {
	cxString, found := node.getAttributeByName("cx")
	if !found {
		if Config.debug {
			fmt.Printf("Circle definition for character '%s' is missing cx attribute\n", node.XMLName.Local)
		}
		return
	}
	cyString, found := node.getAttributeByName("cy")
	if !found {
		if Config.debug {
			fmt.Printf("Circle definition for character '%s' is missing cy attribute\n", node.XMLName.Local)
		}
		return
	}
	rString, found := node.getAttributeByName("r")
	if !found {
		if Config.debug {
			fmt.Printf("Circle definition for character '%s' is missing r attribute\n", node.XMLName.Local)
		}
		return
	}

	x, err := strconv.ParseFloat(cxString, 64)
	if err != nil {
		if Config.debug {
			fmt.Printf("Circle definition for character '%s' has an invalid cx value\n", node.XMLName.Local)
		}
		return
	}

	y, err := strconv.ParseFloat(cyString, 64)
	if err != nil {
		if Config.debug {
			fmt.Printf("Circle definition for character '%s' has an invalid cy value\n", node.XMLName.Local)
		}
		return
	}

	r, err := strconv.ParseFloat(rString, 64)
	if err != nil {
		if Config.debug {
			fmt.Printf("Circle definition for character '%s' has an invalid r value\n", node.XMLName.Local)
		}
		return
	}

	*lines = append(*lines, *NewCircle(Point{x, y}, r).toPolyline(6))
}

func (node *Node) getPolyline(lines *[]Polyline) {
	pointsString, found := node.getAttributeByName("points")
	if !found {
		if Config.debug {
			fmt.Printf("Polyline definition for character '%s' is missing points attribute\n", node.XMLName.Local)
		}
		return
	}
	points := parsePointsString(pointsString)
	if len(points) < 2 {
		if Config.debug {
			fmt.Printf("Polyline definition for character '%s' has no valid points definition\n", node.XMLName.Local)
		}
		return
	}
	*lines = append(*lines, *NewPolyline(&points, nil))
}

func (node *Node) getPolygon(lines *[]Polyline) {
	pointsString, found := node.getAttributeByName("points")
	if !found {
		if Config.debug {
			fmt.Printf("Polygon definition for character '%s' is missing points attribute\n", node.XMLName.Local)
		}
		return
	}
	points := parsePointsString(pointsString)
	if len(points) < 2 {
		if Config.debug {
			fmt.Printf("Polygon definition for character '%s' has no valid points definition\n", node.XMLName.Local)
		}
		return
	}
	*lines = append(*lines, *NewPolygon(&points).toPolyline())
}

func cleanString(value string) string {
	value = strings.ReplaceAll(value, "\\", "/")
	value = strings.ReplaceAll(value, "\a", "")
	value = strings.ReplaceAll(value, "\b", "")
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\t", "")
	value = strings.ReplaceAll(value, "\v", "")
	value = strings.ReplaceAll(value, "\n", "")
	return value
}

func parsePointsString(pointsString string) []Point {
	pointsString = cleanString(pointsString)
	pointsString = strings.TrimSpace(pointsString)

	if strings.Contains(pointsString, "86,13.685.5,") {
		fmt.Println(pointsString[0])
		fmt.Println("")
	}

	var points []Point
	pointStrings := strings.Split(pointsString, " ")
	if strings.Contains(pointsString, ",") {
		for _, pointString := range pointStrings {
			coordinates := strings.Split(pointString, ",")
			if len(coordinates) != 2 {
				if Config.debug {
					fmt.Printf("Point definition '%s' is invalid\n", pointString)
				}
				continue
			}

			x, err := strconv.ParseFloat(coordinates[0], 64)
			if err != nil {
				if Config.debug {
					fmt.Printf("Point definition '%s' is invalid\n", pointString)
				}
				continue
			}
			y, err := strconv.ParseFloat(coordinates[1], 64)
			if err != nil {
				if Config.debug {
					fmt.Printf("Point definition '%s' is invalid\n", pointString)
				}
				continue
			}

			points = append(points, *NewPoint(x, y))
		}
	} else {
		if len(pointStrings) < 2 {
			if Config.debug {
				fmt.Printf("Point definition '%s' is invalid\n", pointsString)
			}
			return points
		}

		for i := 1; i < len(pointStrings); i *= 2 {
			x, err := strconv.ParseFloat(pointStrings[i-1], 64)
			if err != nil {
				if Config.debug {
					fmt.Printf("Point  definition has an invalid X value '%s'\n", pointStrings[i-1])
				}
				continue
			}
			y, err := strconv.ParseFloat(pointStrings[i], 64)
			if err != nil {
				if Config.debug {
					fmt.Printf("Point  definition has an invalid X value '%s'\n", pointStrings[i])
				}
				continue
			}

			points = append(points, *NewPoint(x, y))
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
	shape     Shape
	alignment int
	offset    float64
}

func (character *Character) scale(factor float64) {
	character.shape.scale(factor)
}
