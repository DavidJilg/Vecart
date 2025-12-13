package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

func updateProvedSVG() {
	var configPaths []string
	configPaths = append(configPaths, "static/configs/proved/circles.json")
	configPaths = append(configPaths, "static/configs/proved/group.json")
	configPaths = append(configPaths, "static/configs/proved/lines.json")
	configPaths = append(configPaths, "static/configs/proved/polygons.json")

	Fonts = loadFonts()

	for _, configPath := range configPaths {
		config, userConfigKeys, userShapes, err := getConfigFromStaticAssets(configPath)
		if err != nil {
			fmt.Println(err)
			panic("Could not get config file '" + configPath + "' from static assets!")
		}
		basename := filepath.Base(configPath)
		config.outputPath = "/static/provedSVG/" + strings.TrimSuffix(basename, filepath.Ext(basename)) + ".svg"
		config.shapeAngleDeviationStep = 30
		config.processingDpi = 10
		config.parallelRoutines = 1
		config.overwriteExisting = true
		config.randomSeed = 1701

		CurrentConfigEntry = NewConfigEntry("TEST", config, userConfigKeys, userShapes)
		Config = config

		writeStringToFile(startShapeArtGeneration(), Config.outputPath)

		resetStaticVariables()
	}

}

func resetStaticVariables() {
	errorsOcurred = false
	occuredErrors = nil
	Log = false
	Config = NewConfig()
	CurrentConfigEntry = ConfigEntry{}
	configEntries = []ConfigEntry{}
	RandSource = nil
	quadrants = nil
	ShapeCount = 0
	ShapeCountMutex = &sync.Mutex{}
	currentSpinnerFrame = 0

	stopSpinnerBool = false
	stopSpinnerMutex = sync.Mutex{}
	lastAjustedDarkness = 0.0
	lastAjustedDarknessChange = 0
	finishQuadrantsMutex = sync.Mutex{}
	finishQuadrantsStop = false
}

func getConfigFromStaticAssets(path string) (VecartConfig, []string, any, error) {
	configFile, err := StaticAssets.Open(path)
	if err != nil {
		return NewConfig(), []string{}, nil, err
	}

	content, err := getFileContentsFromStaticAssets(configFile)

	if err != nil {
		return NewConfig(), []string{}, nil, err
	}

	config := NewConfig()
	userKeys, userShapes, _, _ := config.fromJSON(content)

	return config, userKeys, userShapes, nil
}
