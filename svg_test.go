package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestVecart(t *testing.T) {
	var configPaths []string
	configPaths = append(configPaths, "static/configs/proved/circles.json")
	configPaths = append(configPaths, "static/configs/proved/group.json")
	configPaths = append(configPaths, "static/configs/proved/lines.json")
	configPaths = append(configPaths, "static/configs/proved/polygons.json")

	for _, configPath := range configPaths {
		config, err := getConfigFromStaticAssets(configPath)
		if err != nil {
			fmt.Println(err)
			panic("Could not get config '" + configPath + "' from static assets!")
		}
		config.shapeAngleDeviationStep = 30
		config.processingDpi = 10
		config.parallelRoutines = 1
		config.randomSeed = 1701
		config.debug = false
		basename := filepath.Base(configPath)
		config.outputPath = "/static/provedSVG/" + strings.TrimSuffix(basename, filepath.Ext(basename)) + ".svg"

		Config = config
		UserConfig = Config.toJson()
		config.outputPath = "static/provedSVG/" + strings.TrimSuffix(basename, filepath.Ext(basename)) + ".svg"

		svg := startVecart()

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
			t.Errorf("Generating SVG from '%s' failed!", configPath)
		}

		resetStaticVariables()
	}
}
