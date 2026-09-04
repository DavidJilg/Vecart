package utils

import "strconv"

const GCodePrecision int = 6

// Set Feedrate
func F(f int) string {
	return "F" + strconv.Itoa(f) //+ " ;Set Feedrate"
}

// Straight line on x axis at max speed
func G00X(x float64) string {
	return "G00 X" + FormatFloat(x, GCodePrecision)
}

// Straight line on y axis at max speed
func G00Y(y float64) string {
	return "G00 Y" + FormatFloat(y, GCodePrecision)
}

// Straight line on z axis at max speed
func G00Z(z float64) string {
	return "G00 Z" + FormatFloat(z, GCodePrecision)
}

// Straight line on x/y axis at max speed
func G00XY(x float64, y float64) string {
	return "G00 X" + FormatFloat(x, GCodePrecision) +
		" Y" + FormatFloat(y, GCodePrecision)
}

// Straight line on x/y/z at max speed
func G00XYZ(x float64, y float64, z float64) string {
	return "G00 X" + FormatFloat(x, GCodePrecision) +
		" Y" + FormatFloat(y, GCodePrecision) +
		" Z" + FormatFloat(z, GCodePrecision)
}

// Straight line on x axis at F rate
func G01X(x float64, feedrate int) string {
	return "G01 X" + FormatFloat(x, GCodePrecision) + " F" + strconv.Itoa(feedrate)
}

// Straight line on y axis at F rate
func G01Y(y float64, feedrate int) string {
	return "G01 Y" + FormatFloat(y, GCodePrecision) + " F" + strconv.Itoa(feedrate)
}

// Straight line on z axis at F rate
func G01Z(z float64, feedrate int) string {
	return "G01 Z" + FormatFloat(z, GCodePrecision) + " F" + strconv.Itoa(feedrate)
}

// Straight line on x/y axis at F rate
func G01XY(x float64, y float64, feedrate int) string {
	return "G01 X" + FormatFloat(x, GCodePrecision) +
		" Y" + FormatFloat(y, GCodePrecision) + " F" + strconv.Itoa(feedrate)
}

// Straight line on x/y/z axis at F rate
func G01XYZ(x float64, y float64, z float64, feedrate int) string {
	return "G01 X" + FormatFloat(x, GCodePrecision) +
		" Y" + FormatFloat(y, GCodePrecision) +
		" Z" + FormatFloat(z, GCodePrecision) + " F" + strconv.Itoa(feedrate)
}

// Clockwise arc
func G02(x float64, y float64, i float64, j float64, feedrate int) string {
	return "G02 X" + FormatFloat(x, GCodePrecision) +
		" Y" + FormatFloat(y, GCodePrecision) +
		" I" + FormatFloat(i, GCodePrecision) +
		" J" + FormatFloat(j, GCodePrecision) + " F" + strconv.Itoa(feedrate)
}

// Anti-Clockwise arc
func G03(x float64, y float64, i float64, j float64, feedrate int) string {
	return "G03 X" + FormatFloat(x, GCodePrecision) +
		" Y" + FormatFloat(y, GCodePrecision) +
		" I" + FormatFloat(i, GCodePrecision) +
		" J" + FormatFloat(j, GCodePrecision) + " F" + strconv.Itoa(feedrate)
}

// Set units to millimeters
func G21() string {
	return "G21" //"G21 ;Set units to millimeters"
}

// Go home
func G28() string {
	return "G28" //"G28 ;Go Home"
}

// Go home after moving z axis
func G28Z(z float64) string {
	return "G028 Z" + FormatFloat(z, GCodePrecision)
}

// Go home after moving x/y axis
func G28XY(x float64, y float64) string {
	return "G028 X" + FormatFloat(x, GCodePrecision) +
		" Y" + FormatFloat(y, GCodePrecision)
}

// Go home through intermediary point
func G28XYZ(x float64, y float64, z float64) string {
	return "G028 X" + FormatFloat(x, GCodePrecision) +
		" Y" + FormatFloat(y, GCodePrecision) +
		" Z" + FormatFloat(z, GCodePrecision)
}

// Enable Absolute positioning
func G90() string {
	return "G90" //"G90 ;Absolute Positioning"
}

// Pause
func M0() string {
	return "M0" //"M0 ;Pause"
}
