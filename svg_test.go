package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidJilg/Vecart/internal/general"
	"github.com/DavidJilg/Vecart/internal/utils"
)

func TestVecart(t *testing.T) {
	initForTests()

	var configPaths []string
	configPaths = append(configPaths, "static/configs/proved/circles.json")
	configPaths = append(configPaths, "static/configs/proved/group.json")
	configPaths = append(configPaths, "static/configs/proved/lines.json")
	configPaths = append(configPaths, "static/configs/proved/polygons.json")

	for _, configPath := range configPaths {
		config, userConfigKeys, userShapes, err := general.GetConfigFromStaticAssets(configPath)
		if err != nil {
			fmt.Println(err)
			panic("Could not get config '" + configPath + "' from static assets!")
		}
		config.ShapeAngleDeviationStep = 30
		config.ProcessingDpi = 10
		config.ParallelRoutines = 1
		config.RandomSeed = 1701
		config.OverwriteExisting = true
		basename := filepath.Base(configPath)
		config.OutputPath = "/static/provedSVG/" + strings.TrimSuffix(basename, filepath.Ext(basename)) + ".svg"
		fixedOutputPath := "static/provedSVG/" + strings.TrimSuffix(basename, filepath.Ext(basename)) + ".svg"

		general.SetConfig(config)
		general.SetConfigEntry(general.NewConfigEntry("TEST", config, userConfigKeys, userShapes))

		svg := general.StartShapeArtGeneration()

		svgFile, err := StaticAssets.Open(fixedOutputPath)
		if err != nil {
			panic("Reading '" + fixedOutputPath + "' from static assets failed!")
		}

		content, err := utils.GetFileContentsFromStaticAssets(svgFile)

		if err != nil {
			fmt.Println(err)
			panic("Could not read proved svg file '" + config.OutputPath + "' from static ressources")
		}

		if general.CleanString(content) != general.CleanString(svg) {
			utils.StringComparison(content, svg)
			t.Errorf("\nGenerating SVG from '%s' failed!", configPath)
		}

		general.ResetStaticVariables()
	}
}
