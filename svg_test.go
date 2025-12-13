package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestVecart(t *testing.T) {
	TestMode = true

	var configPaths []string
	configPaths = append(configPaths, "static/configs/proved/circles.json")
	configPaths = append(configPaths, "static/configs/proved/group.json")
	configPaths = append(configPaths, "static/configs/proved/lines.json")
	configPaths = append(configPaths, "static/configs/proved/polygons.json")

	for _, configPath := range configPaths {
		config, userConfigKeys, userShapes, err := getConfigFromStaticAssets(configPath)
		if err != nil {
			fmt.Println(err)
			panic("Could not get config '" + configPath + "' from static assets!")
		}
		config.shapeAngleDeviationStep = 30
		config.processingDpi = 10
		config.parallelRoutines = 1
		config.randomSeed = 1701
		config.overwriteExisting = true
		basename := filepath.Base(configPath)
		config.outputPath = "/static/provedSVG/" + strings.TrimSuffix(basename, filepath.Ext(basename)) + ".svg"

		CurrentConfigEntry = NewConfigEntry("TEST", config, userConfigKeys, userShapes)
		Config = config
		config.outputPath = "static/provedSVG/" + strings.TrimSuffix(basename, filepath.Ext(basename)) + ".svg"

		svg := startShapeArtGeneration()

		svgFile, err := StaticAssets.Open(config.outputPath)
		if err != nil {
			panic("Reading '" + config.outputPath + "' from static assets failed!")
		}

		content, err := getFileContentsFromStaticAssets(svgFile)

		if err != nil {
			fmt.Println(err)
			panic("Could not read proved svg file '" + config.outputPath + "' from static ressources")
		}

		if cleanString(content) != cleanString(svg) {
			stringComparison(content, svg)
			t.Errorf("\nGenerating SVG from '%s' failed!", configPath)
		}

		resetStaticVariables()
	}
}

func stringComparison(a, b string) {
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
