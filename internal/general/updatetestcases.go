package general

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/DavidJilg/Vecart/internal/utils"
)

func UpdateProvedSVG() {
	var configPaths []string
	configPaths = append(configPaths, "static/configs/proved/circles.json")
	configPaths = append(configPaths, "static/configs/proved/group.json")
	configPaths = append(configPaths, "static/configs/proved/lines.json")
	configPaths = append(configPaths, "static/configs/proved/polygons.json")

	Fonts = LoadFonts()

	for _, configPath := range configPaths {
		config, userConfigKeys, userShapes, err := GetConfigFromStaticAssets(configPath)
		if err != nil {
			fmt.Println(err)
			panic("Could not get config file '" + configPath + "' from static assets!")
		}
		basename := filepath.Base(configPath)
		config.OutputPath = "/static/provedSVG/" + strings.TrimSuffix(basename, filepath.Ext(basename)) + ".svg"
		config.ShapeAngleDeviationStep = 30
		config.ProcessingDpi = 10
		config.ParallelRoutines = 1
		config.OverwriteExisting = true
		config.RandomSeed = 1701

		CurrentConfigEntry = NewConfigEntry("TEST", config, userConfigKeys, userShapes)
		Config = config

		utils.WriteStringToFile(StartShapeArtGeneration(), Config.OutputPath)

		ResetStaticVariables()
	}

}

func UpdateProvedGCode() {
	var configPaths []string
	configPaths = append(configPaths, "static/configs/proved/gcode.json")
	configPaths = append(configPaths, "static/configs/proved/gcodeStampMode.json")

	Fonts = LoadFonts()

	for _, configPath := range configPaths {
		config, userConfigKeys, userShapes, err := GetConfigFromStaticAssets(configPath)
		if err != nil {
			fmt.Println(err)
			panic("Could not get config file '" + configPath + "' from static assets!")
		}
		basename := filepath.Base(configPath)
		config.OutputPath = "/static/provedGCode/" + strings.TrimSuffix(basename, filepath.Ext(basename)) + ".gcode"
		config.ParallelRoutines = 1
		config.RandomSeed = 1701
		config.OverwriteExisting = true

		CurrentConfigEntry = NewConfigEntry("TEST", config, userConfigKeys, userShapes)
		Config = config

		utils.WriteStringToFile(ConvertSVGToGCode(), Config.OutputPath)

		ResetStaticVariables()
	}

}

func ResetStaticVariables() {
	errorsOcurred = false
	occuredErrors = nil
	utils.DisableDebugMode()
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

func GetConfigFromStaticAssets(path string) (VecartConfig, []string, any, error) {
	configFile, err := StaticAssets.Open(path)
	if err != nil {
		return NewConfig(), []string{}, nil, err
	}

	content, err := utils.GetFileContentsFromStaticAssets(configFile)

	if err != nil {
		return NewConfig(), []string{}, nil, err
	}

	config := NewConfig()
	userKeys, userShapes, _, _ := config.FromJSON(content, path)

	return config, userKeys, userShapes, nil
}
