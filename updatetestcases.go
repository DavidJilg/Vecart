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
		config, err := getConfigFromStaticAssets(configPath)
		if err != nil {
			fmt.Println(err)
			panic("Could not get config file '" + configPath + "' from static assets!")
		}
		basename := filepath.Base(configPath)
		config.outputPath = "/static/provedSVG/" + strings.TrimSuffix(basename, filepath.Ext(basename)) + ".svg"
		config.shapeAngleDeviationStep = 30
		config.processingDpi = 10
		config.parallelRoutines = 1
		config.randomSeed = 1701
		config.debug = false

		Config = config
		UserConfig = Config.toJson()

		writeStringToFile(startVecart(), Config.outputPath)

		resetStaticVariables()
	}

}

func resetStaticVariables() {
	errorsOcurred = false
	errors = nil
	debug = false
	Config = NewConfig()
	UserConfig = ""
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

func getConfigFromStaticAssets(path string) (VecartConfig, error) {
	configFile, err := StaticAssets.Open(path)
	if err != nil {
		return NewConfig(), err
	}

	content, err := getFileContentsFromStaticAssets(configFile)

	if err != nil {
		return NewConfig(), err
	}

	config := NewConfig()
	config.fromJSON(content)

	return config, nil
}
