package utils

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

var flagsInitialized = false

func MMToPixel(millimeters float64, dpi float64) float64 {
	return (millimeters * dpi) / 25.4
}

func PixelToMM(pixels float64, dpi float64) float64 {
	return (pixels * 25.4) / dpi
}

func InitFlags() {
	if flagsInitialized {
		return
	}
	flag.Bool("debugMode", false, "Running in Debug mode")
	flag.Bool("testMode", false, "Running in Test mode")
	flag.String("Version", "0.0.0", "Version of the application")
	flagsInitialized = true
}

func DebugModeEnabled() bool {
	return flag.Lookup("debugMode").Value.(flag.Getter).Get().(bool)
}

func EnableDebugMode() {
	flag.Set("debugMode", "true")
}

func DisableDebugMode() {
	flag.Set("debugMode", "false")
}

func TestModeEnabled() bool {
	return flag.Lookup("testMode").Value.(flag.Getter).Get().(bool)
}

func EnableTestMode() {
	flag.Set("testMode", "true")
}

func DisableTestMode() {
	flag.Set("testMode", "false")
}

func GetVersion() string {
	return flag.Lookup("Version").Value.(flag.Getter).Get().(string)
}

func SetVersion(version string) {
	flag.Set("Version", version)
}

func FormatFloat(value float64, precision int) string {
	return strconv.FormatFloat(value, 'f', precision, 64)
}

func GCodeSlicesEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func StringComparison(a, b string) {
	fmt.Print("\n\n")
	var linesA = strings.Split(a, "\n")
	var linesB = strings.Split(b, "\n")

	if len(linesA) != len(linesB) {
		fmt.Printf("Different number of lines: A=%d, B=%d\n", len(linesA), len(linesB))
	}

	for i := 0; i < len(linesA) && i < len(linesB); i++ {
		if linesA[i] != linesB[i] {
			fmt.Printf("Line %d differs:\nA: %s\nB: %s\n\n", i+1, linesA[i], linesB[i])
		}
	}
}
