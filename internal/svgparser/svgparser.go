package svgparser

import (
	"encoding/xml"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/DavidJilg/Vecart/internal/shapes"
	"github.com/DavidJilg/Vecart/internal/utils"
)

var cssClassSelectorRegex = regexp.MustCompile(`\.([a-zA-Z0-9_-]+)`)
var cssCommentRegex = regexp.MustCompile(`(?s)/\*.*?\*/`)
var transformFunctionRegex = regexp.MustCompile(`([a-zA-Z]+)\s*\(([^)]*)\)`)
var transformNumberRegex = regexp.MustCompile(`[+-]?(?:\d+\.?\d*|\.\d+)(?:[eE][+-]?\d+)?`)

type affineTransform struct {
	A float64
	B float64
	C float64
	D float64
	E float64
	F float64
}

type XMLNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content []byte     `xml:",innerxml"`
	Nodes   []XMLNode  `xml:",any"`
}

func ParseXMLTree(xmlString string) (*XMLNode, error) {
	data := []byte(xmlString)
	var root XMLNode

	err := xml.Unmarshal(data, &root)
	if err != nil {
		return nil, err
	}

	return &root, nil
}

func ExtractShapes(root *XMLNode, ignoreHidden bool, keepGroups bool) *SVG {
	svg := &SVG{}

	/*
		viewbox, found := root.getAttributeByName("viewBox")
		if !found {
			if utils.DebugModeEnabled() {
				log.Println("No viewBox defined in SVG file")
			}
		} else {
			viewboxValues := strings.Split(viewbox, " ")
			if len(viewboxValues) != 4 {
				if utils.DebugModeEnabled() {
					log.Println("Unsuported viewBox definition in SVG file")
				}
			} else {
				viewBoxWidth, err := strconv.ParseFloat(viewboxValues[2], 64)
				if err != nil {
					if utils.DebugModeEnabled() {
						log.Printf("Viewbox definition has an invalid value '%s'\n", viewboxValues[2])
					}
				} else {
					svg.Width = viewBoxWidth
				}

				viewBoxHeight, err := strconv.ParseFloat(viewboxValues[3], 64)
				if err != nil {
					if utils.DebugModeEnabled() {
						log.Printf("Viewbox definition has an invalid value '%s'\n", viewboxValues[3])
					}
				} else {
					svg.Height = viewBoxHeight
				}

			}
		}*/

	hiddenClasses := map[string]bool{}
	if ignoreHidden {
		hiddenClasses = root.getHiddenClasses()
	}

	initialTransform := identityTransform()
	rootTransform, ok := root.getTransform()
	if ok {
		initialTransform = rootTransform
	}

	svg.Shapes = root.getShapes(ignoreHidden, hiddenClasses, initialTransform, keepGroups)

	return svg
}

func (node *XMLNode) getShapes(ignoreHidden bool, hiddenClasses map[string]bool, inheritedTransform affineTransform, keepGroups bool) []any {
	shapeList := []any{}

	for index := range node.Nodes {
		currentNode := &node.Nodes[index]
		if ignoreHidden && currentNode.isHidden(hiddenClasses) {
			continue
		}

		currentTransform := inheritedTransform
		nodeTransform, ok := currentNode.getTransform()
		if ok {
			currentTransform = inheritedTransform.multiply(nodeTransform)
		}

		switch currentNode.XMLName.Local {
		case "rect":
			rectangle, ok := currentNode.getRectangle()
			if ok {
				shapeList = append(shapeList, applyTransform(rectangle, currentTransform))
			}
		case "circle":
			circle, ok := currentNode.getCircle()
			if ok {
				shapeList = append(shapeList, applyTransform(circle, currentTransform))
			}
		case "ellipse":
			elipse, ok := currentNode.getEllipse()
			if ok {
				shapeList = append(shapeList, applyTransform(elipse, currentTransform))
			}
		case "line":
			line, ok := currentNode.getLine()
			if ok {
				shapeList = append(shapeList, applyTransform(line, currentTransform))
			}
		case "polyline":
			polyline, ok := currentNode.getPolyline()
			if ok {
				shapeList = append(shapeList, applyTransform(polyline, currentTransform))
			}
		case "polygon":
			polygon, ok := currentNode.getPolygon()
			if ok {
				shapeList = append(shapeList, applyTransform(polygon, currentTransform))
			}
		case "path":
			path, ok := currentNode.getPath()
			if ok {
				shapeList = append(shapeList, applyTransform(path, currentTransform))
			}
		case "g":
			subShapes := currentNode.getShapes(ignoreHidden, hiddenClasses, currentTransform, keepGroups)
			if keepGroups {
				if len(subShapes) > 0 {
					shapeList = append(shapeList, shapes.NewGroup(subShapes))
				}
			} else {
				shapeList = append(shapeList, subShapes...)
			}

		default:
			if utils.DebugModeEnabled() {
				log.Printf("Unsupported SVG element is ignored: %s\n", currentNode.XMLName.Local)
			}
		}

	}

	return shapeList
}

func (node *XMLNode) getTransform() (affineTransform, bool) {
	transform, found := node.getAttributeByName("transform")
	if !found {
		return identityTransform(), false
	}

	matches := transformFunctionRegex.FindAllStringSubmatch(transform, -1)
	if len(matches) == 0 {
		if utils.DebugModeEnabled() {
			log.Printf("Unsupported SVG transform is ignored: %s\n", transform)
		}
		return identityTransform(), false
	}

	currentTransform := identityTransform()
	for _, match := range matches {
		if len(match) != 3 {
			if utils.DebugModeEnabled() {
				log.Printf("Unsupported SVG transform is ignored: %s\n", transform)
			}
			return identityTransform(), false
		}

		transformName := strings.ToLower(strings.TrimSpace(match[1]))
		values := parseTransformNumbers(match[2])
		transformMatrix, ok := transformMatrixFromValues(transformName, values)
		if !ok {
			if utils.DebugModeEnabled() {
				log.Printf("Unsupported SVG transform is ignored: %s\n", transform)
			}
			return identityTransform(), false
		}

		currentTransform = transformMatrix.multiply(currentTransform)
	}

	return currentTransform, true
}

func parseTransformNumbers(transformParameters string) []float64 {
	numberStrings := transformNumberRegex.FindAllString(transformParameters, -1)
	values := make([]float64, 0, len(numberStrings))

	for _, numberString := range numberStrings {
		value, err := strconv.ParseFloat(numberString, 64)
		if err != nil {
			continue
		}
		values = append(values, value)
	}

	return values
}

func transformMatrixFromValues(transformName string, values []float64) (affineTransform, bool) {
	switch transformName {
	case "translate":
		if len(values) == 0 {
			return identityTransform(), false
		}
		transform := identityTransform()
		transform.E = values[0]
		if len(values) > 1 {
			transform.F = values[1]
		}
		return transform, true
	case "scale":
		if len(values) == 0 || len(values) > 2 {
			return identityTransform(), false
		}

		scaleX := values[0]
		scaleY := scaleX
		if len(values) == 2 {
			scaleY = values[1]
		}

		return affineTransform{
			A: scaleX,
			D: scaleY,
		}, true
	case "rotate":
		if len(values) != 1 && len(values) != 3 {
			return identityTransform(), false
		}

		angleRadians := values[0] * math.Pi / 180
		sinAngle, cosAngle := math.Sincos(angleRadians)
		rotationTransform := affineTransform{
			A: cosAngle,
			B: sinAngle,
			C: -sinAngle,
			D: cosAngle,
		}

		if len(values) == 1 {
			return rotationTransform, true
		}

		centerX := values[1]
		centerY := values[2]
		return newTranslateTransform(centerX, centerY).
			multiply(rotationTransform).
			multiply(newTranslateTransform(-centerX, -centerY)), true
	case "skewx":
		if len(values) != 1 {
			return identityTransform(), false
		}

		return affineTransform{
			A: 1,
			C: math.Tan(values[0] * math.Pi / 180),
			D: 1,
		}, true
	case "skewy":
		if len(values) != 1 {
			return identityTransform(), false
		}

		return affineTransform{
			A: 1,
			B: math.Tan(values[0] * math.Pi / 180),
			D: 1,
		}, true
	case "matrix":
		if len(values) != 6 {
			return identityTransform(), false
		}
		return affineTransform{
			A: values[0],
			B: values[1],
			C: values[2],
			D: values[3],
			E: values[4],
			F: values[5],
		}, true
	default:
		return identityTransform(), false
	}
}

func identityTransform() affineTransform {
	return affineTransform{
		A: 1,
		D: 1,
	}
}

func newTranslateTransform(x, y float64) affineTransform {
	return affineTransform{
		A: 1,
		D: 1,
		E: x,
		F: y,
	}
}

func (transform affineTransform) multiply(other affineTransform) affineTransform {
	return affineTransform{
		A: transform.A*other.A + transform.C*other.B,
		B: transform.B*other.A + transform.D*other.B,
		C: transform.A*other.C + transform.C*other.D,
		D: transform.B*other.C + transform.D*other.D,
		E: transform.A*other.E + transform.C*other.F + transform.E,
		F: transform.B*other.E + transform.D*other.F + transform.F,
	}
}

func (transform affineTransform) isIdentity() bool {
	return transform.A == 1 &&
		transform.B == 0 &&
		transform.C == 0 &&
		transform.D == 1 &&
		transform.E == 0 &&
		transform.F == 0
}

func applyTransform(shape any, transform affineTransform) any {
	if transform.isIdentity() {
		return shape
	}

	switch currentShape := shape.(type) {
	case *shapes.Circle:
		if _, ok := transform.circleScaleFactor(); ok {
			currentShape.Transform(transform.A, transform.B, transform.C, transform.D, transform.E, transform.F)
			return currentShape
		}

		path := shapes.NewEllipse(currentShape.Center, currentShape.Radius, currentShape.Radius)
		path.Transform(transform.A, transform.B, transform.C, transform.D, transform.E, transform.F)
		return path
	case *shapes.Path:
		currentShape.Transform(transform.A, transform.B, transform.C, transform.D, transform.E, transform.F)
	case *shapes.Point:
		currentShape.Transform(transform.A, transform.B, transform.C, transform.D, transform.E, transform.F)
	case *shapes.Polygon:
		currentShape.Transform(transform.A, transform.B, transform.C, transform.D, transform.E, transform.F)
	case *shapes.Polyline:
		currentShape.Transform(transform.A, transform.B, transform.C, transform.D, transform.E, transform.F)
	case *shapes.Shape:
		currentShape.Transform(transform.A, transform.B, transform.C, transform.D, transform.E, transform.F)
	}

	return shape
}

func (transform affineTransform) circleScaleFactor() (float64, bool) {
	const epsilon = 1e-9

	scaleX := transform.A*transform.A + transform.B*transform.B
	scaleY := transform.C*transform.C + transform.D*transform.D
	dotProduct := transform.A*transform.C + transform.B*transform.D

	if math.Abs(scaleX-scaleY) > epsilon || math.Abs(dotProduct) > epsilon {
		return 0, false
	}

	return math.Sqrt(scaleX), true
}

func (node *XMLNode) isHidden(hiddenClasses map[string]bool) bool {
	display, found := node.getAttributeByName("display")
	if found && isDisplayNone(display) {
		return true
	}

	style, found := node.getAttributeByName("style")
	if found && styleContainsDisplayNone(style) {
		return true
	}

	classList, found := node.getAttributeByName("class")
	if !found {
		return false
	}

	for _, className := range strings.Fields(classList) {
		if hiddenClasses[className] {
			return true
		}
	}

	return false
}

func (node *XMLNode) getHiddenClasses() map[string]bool {
	hiddenClasses := map[string]bool{}
	node.collectHiddenClasses(hiddenClasses)
	return hiddenClasses
}

func (node *XMLNode) collectHiddenClasses(hiddenClasses map[string]bool) {
	if node.XMLName.Local == "style" {
		extractHiddenClassesFromStyle(string(node.Content), hiddenClasses)
	}

	for i := range node.Nodes {
		node.Nodes[i].collectHiddenClasses(hiddenClasses)
	}
}

func extractHiddenClassesFromStyle(styleContent string, hiddenClasses map[string]bool) {
	cleanedStyleContent := cssCommentRegex.ReplaceAllString(styleContent, "")
	rules := strings.Split(cleanedStyleContent, "}")

	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		ruleParts := strings.SplitN(rule, "{", 2)
		if len(ruleParts) != 2 {
			continue
		}

		selectors := ruleParts[0]
		declarations := ruleParts[1]
		if !styleContainsDisplayNone(declarations) {
			continue
		}

		selectorParts := strings.Split(selectors, ",")
		for _, selector := range selectorParts {
			matches := cssClassSelectorRegex.FindAllStringSubmatch(selector, -1)
			for _, match := range matches {
				if len(match) == 2 {
					hiddenClasses[match[1]] = true
				}
			}
		}
	}
}

func styleContainsDisplayNone(styleText string) bool {
	declarations := strings.Split(styleText, ";")
	for _, declaration := range declarations {
		keyValue := strings.SplitN(declaration, ":", 2)
		if len(keyValue) != 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(keyValue[0]))
		if key != "display" {
			continue
		}

		if isDisplayNone(keyValue[1]) {
			return true
		}
	}
	return false
}

func isDisplayNone(displayValue string) bool {
	normalizedValue := strings.ToLower(strings.TrimSpace(displayValue))
	normalizedValue = strings.ReplaceAll(normalizedValue, " ", "")
	return strings.HasPrefix(normalizedValue, "none")
}

func (node *XMLNode) getAttributeByName(name string) (string, bool) {
	for index := range node.Attrs {
		if node.Attrs[index].Name.Local == name {
			return node.Attrs[index].Value, true
		}
	}

	return "", false
}

func (node *XMLNode) getLine() (*shapes.Polyline, bool) {
	x1String, found := node.getAttributeByName("x1")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Line definition is missing x1 attribute\n")
		}
		return &shapes.Polyline{}, false
	}
	x2String, found := node.getAttributeByName("x2")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Line definition is missing x2 attribute\n")
		}
		return &shapes.Polyline{}, false
	}
	y1String, found := node.getAttributeByName("y1")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Line definition is missing y1 attribute\n")
		}
		return &shapes.Polyline{}, false
	}
	y2String, found := node.getAttributeByName("y2")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Line definition is missing y2 attribute\n")
		}
		return &shapes.Polyline{}, false
	}

	x1, err := strconv.ParseFloat(x1String, 64)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Line definition invalid x1 value\n")
		}
		return &shapes.Polyline{}, false
	}

	x2, err := strconv.ParseFloat(x2String, 64)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Line definition has an invalid x2 value\n")
		}
		return &shapes.Polyline{}, false
	}

	y1, err := strconv.ParseFloat(y1String, 64)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Line definition has an invalid y1 value\n")
		}
		return &shapes.Polyline{}, false
	}

	y2, err := strconv.ParseFloat(y2String, 64)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Line definition has an invalid y2 value\n")
		}
		return &shapes.Polyline{}, false
	}

	return shapes.NewPolyline(&[]shapes.Point{{X: x1, Y: y1}, {X: x2, Y: y2}}, nil), true
}

func (node *XMLNode) getCircle() (*shapes.Circle, bool) {
	cxString, found := node.getAttributeByName("cx")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Circle definition is missing cx attribute\n")
		}
		return &shapes.Circle{}, false
	}
	cyString, found := node.getAttributeByName("cy")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Circle definition is missing cy attribute\n")
		}
		return &shapes.Circle{}, false
	}
	rString, found := node.getAttributeByName("r")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Circle definition is missing r attribute\n")
		}
		return &shapes.Circle{}, false
	}

	x, err := strconv.ParseFloat(cxString, 64)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Circle definition has an invalid cx value\n")
		}
		return &shapes.Circle{}, false
	}

	y, err := strconv.ParseFloat(cyString, 64)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Circle definition has an invalid cy value\n")
		}
		return &shapes.Circle{}, false
	}

	r, err := strconv.ParseFloat(rString, 64)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Circle definition has an invalid r value\n")
		}
		return &shapes.Circle{}, false
	}

	return shapes.NewCircle(shapes.Point{X: x, Y: y}, r), true
}

func (node *XMLNode) getEllipse() (*shapes.Path, bool) {
	cxString, found := node.getAttributeByName("cx")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Elipse definition is missing cx attribute\n")
		}
		return &shapes.Path{}, false
	}
	cyString, found := node.getAttributeByName("cy")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Elipse definition is missing cy attribute\n")
		}
		return &shapes.Path{}, false
	}
	rxString, found := node.getAttributeByName("rx")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Elipse definition is missing rx attribute\n")
		}
		return &shapes.Path{}, false
	}
	ryString, found := node.getAttributeByName("ry")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Elipse definition is missing ry attribute\n")
		}
		return &shapes.Path{}, false
	}

	x, err := strconv.ParseFloat(cxString, 64)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Elipse definition has an invalid cx value\n")
		}
		return &shapes.Path{}, false
	}

	y, err := strconv.ParseFloat(cyString, 64)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Elipse definition has an invalid cy value\n")
		}
		return &shapes.Path{}, false
	}

	rx, err := strconv.ParseFloat(rxString, 64)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Elipse definition has an invalid rx value\n")
		}
		return &shapes.Path{}, false
	}

	ry, err := strconv.ParseFloat(ryString, 64)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Elipse definition has an invalid ry value\n")
		}
		return &shapes.Path{}, false
	}

	return shapes.NewEllipse(shapes.Point{X: x, Y: y}, rx, ry), true
}

func (node *XMLNode) getRectangle() (any, bool) {
	x, y := 0.0, 0.0
	xString, found := node.getAttributeByName("x")
	if found {
		xValue, err := strconv.ParseFloat(xString, 64)
		if err != nil {
			if utils.DebugModeEnabled() {
				log.Printf("Rectangle definition has an invalid x value\n")
			}
			return nil, false
		}
		x = xValue
	}
	yString, found := node.getAttributeByName("y")
	if found {
		yValue, err := strconv.ParseFloat(yString, 64)
		if err != nil {
			if utils.DebugModeEnabled() {
				log.Printf("Rectangle definition has an invalid y value\n")
			}
			return nil, false
		}
		y = yValue
	}

	widthString, found := node.getAttributeByName("width")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Rectangle definition is missing width attribute\n")
		}
		return nil, false
	}
	heightString, found := node.getAttributeByName("height")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Rectangle definition is missing height attribute\n")
		}
		return nil, false
	}

	rxString, found := node.getAttributeByName("rx")
	ryString := ""
	cornerDefinitionExists := false
	if found {
		ryString, found = node.getAttributeByName("ry")
		if !found {
			if utils.DebugModeEnabled() {
				log.Printf("Rectangle definition has invalid corner radius definition\n")
			}
			return nil, false
		}
		cornerDefinitionExists = true
	}

	width, err := strconv.ParseFloat(widthString, 64)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Rectangle definition has an invalid width value\n")
		}
		return nil, false
	}
	height, err := strconv.ParseFloat(heightString, 64)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Rectangle definition has an invalid height value\n")
		}
		return nil, false
	}

	rx, ry := 0.0, 0.0
	if cornerDefinitionExists {
		rxValue, err := strconv.ParseFloat(rxString, 64)
		if err != nil {
			if utils.DebugModeEnabled() {
				log.Printf("Rectangle definition has an invalid rx value\n")
			}
			return nil, false
		}
		ryValue, err := strconv.ParseFloat(ryString, 64)
		if err != nil {
			if utils.DebugModeEnabled() {
				log.Printf("Rectangle definition has an invalid ry value\n")
			}
			return nil, false
		}

		rx, ry = rxValue, ryValue
	}

	if rx == 0 && ry == 0 {
		return shapes.NewRect(x, y, width, height), true
	}

	return shapes.NewRoundedRect(x, y, width, height, rx, ry)
}

func (node *XMLNode) getPolyline() (*shapes.Polyline, bool) {
	pointsString, found := node.getAttributeByName("points")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Polyline definition is missing points attribute\n")
		}
		return &shapes.Polyline{}, false
	}
	points := parsePointsString(pointsString)
	if len(points) < 2 {
		if utils.DebugModeEnabled() {
			log.Printf("Polyline definition has no valid points definition\n")
		}
		return &shapes.Polyline{}, false
	}
	return shapes.NewPolyline(&points, nil), true
}

func (node *XMLNode) getPolygon() (*shapes.Polygon, bool) {
	pointsString, found := node.getAttributeByName("points")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Polygon definition is missing points attribute\n")
		}
		return &shapes.Polygon{}, false
	}
	points := parsePointsString(pointsString)
	if len(points) < 2 {
		if utils.DebugModeEnabled() {
			log.Printf("Polygon definition has no valid points definition\n")
		}
		return &shapes.Polygon{}, false
	}
	return shapes.NewPolygon(&points), true
}

func (node *XMLNode) getPath() (*shapes.Path, bool) {
	dString, found := node.getAttributeByName("d")
	if !found {
		if utils.DebugModeEnabled() {
			log.Printf("Path definition is missing d attribute\n")
		}
		return &shapes.Path{}, false
	}
	path, err := parsePathDataString(dString)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Printf("Path definition has invalid d attribute\n")
		}
		return &shapes.Path{}, false
	}
	return &path, true
}

func parsePointsString(pointsString string) []shapes.Point {
	pointsString = cleanString(pointsString)
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

		for i := 1; i < len(pointStrings); i += 2 {
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
