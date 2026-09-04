package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidJilg/Vecart/internal/general"
	"github.com/DavidJilg/Vecart/internal/utils"
)

func TestGCode(t *testing.T) {
	initForTests()

	var configPaths []string
	configPaths = append(configPaths, "static/configs/proved/gcode.json")

	for _, configPath := range configPaths {
		config, userConfigKeys, userShapes, err := general.GetConfigFromStaticAssets(configPath)
		if err != nil {
			fmt.Println(err)
			panic("Could not get config '" + configPath + "' from static assets!")
		}
		config.ParallelRoutines = 1
		config.RandomSeed = 1701
		config.OverwriteExisting = true

		basename := filepath.Base(configPath)
		config.OutputPath = "static/provedGCode/" + strings.TrimSuffix(basename, filepath.Ext(basename)) + ".gcode"

		general.SetConfig(config)
		general.SetConfigEntry(general.NewConfigEntry("TEST", config, userConfigKeys, userShapes))

		gcodeString := general.ConvertSVGToGCode()

		gcodeFile, err := StaticAssets.Open(config.OutputPath)
		if err != nil {
			panic("Reading '" + config.OutputPath + "' from static assets failed!")
		}

		content, err := utils.GetFileContentsFromStaticAssets(gcodeFile)

		if err != nil {
			fmt.Println(err)
			panic("Could not read proved Gcode file '" + config.OutputPath + "' from static ressources")
		}

		if general.CleanString(content) != general.CleanString(gcodeString) {
			utils.StringComparison(content, gcodeString)
			t.Errorf("\nGenerating GCode from '%s' failed!", configPath)
		}

		general.ResetStaticVariables()
	}
}
