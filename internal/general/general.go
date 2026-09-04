package general

import (
	"bufio"
	"embed"
	"fmt"
	"image"
	"log"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DavidJilg/Vecart/internal/shapes"

	"github.com/disintegration/gift"

	"github.com/DavidJilg/Vecart/internal/utils"
)

var configEntries []ConfigEntry
var Fonts map[string]*Font
var StaticAssets embed.FS
var License embed.FS
var Config VecartConfig
var CurrentConfigEntry ConfigEntry
var RandSource *rand.Rand

type ConfigEntry struct {
	name           string
	config         VecartConfig
	originalConfig VecartConfig
	userKeys       []string
	userShapes     any
}

func (configEntry *ConfigEntry) toShortJson() string {
	return configEntry.config.ToShortJson(configEntry.userKeys, configEntry.userShapes)
}

func NewConfigEntry(name string, config VecartConfig, userKeys []string, userShapes any) ConfigEntry {
	var configEntry ConfigEntry
	configEntry.name = name
	configEntry.config = config
	configEntry.originalConfig = config.Copy()
	configEntry.userKeys = userKeys
	configEntry.userShapes = userShapes
	return configEntry
}

func InitStaticAssets(StaticAssetsInput embed.FS, LicenseInput embed.FS) {
	StaticAssets = StaticAssetsInput
	License = LicenseInput
}

func SetConfig(config VecartConfig) {
	Config = config
}

func SetConfigEntry(configEntry ConfigEntry) {
	CurrentConfigEntry = configEntry
}

func SetFonts(fonts map[string]*Font) {
	Fonts = fonts
}

func Main() {
	start := time.Now()
	batchSeedsMode, batchSize, randomConfigOrder := parseArguments()
	fmt.Printf("Vecart v%s - by David Jilg (david-jilg.com/vecart)\n", utils.GetVersion())

	Fonts = LoadFonts()

	if len(configEntries) == 0 {
		getExampleConfig()
	}

	if batchSeedsMode && batchSize > 1 {
		generateBatchSeedConfigs(batchSize)
	}

	if randomConfigOrder {
		randSource := rand.New(rand.NewPCG(uint64(time.Now().Unix()), uint64(time.Now().Unix())))
		randSource.Shuffle(len(configEntries), func(i, j int) {
			configEntries[i], configEntries[j] = configEntries[j], configEntries[i]
		})
	}

	for index, currentConfigEntry := range configEntries {
		iterationStart := time.Now()
		if len(configEntries) > 1 {
			fmt.Printf("\nRunning configuration %d of %d (%s) - %s\n", index+1, len(configEntries), currentConfigEntry.name, time.Now().Format(time.RFC822))
		} else {
			fmt.Printf("\nRunning configuration '%s' - %s\n", currentConfigEntry.name, time.Now().Format(time.RFC822))
		}

		runIteration(currentConfigEntry)

		if len(configEntries) > 1 {
			duration := time.Since(iterationStart)
			fmt.Printf("\nFinished iteration %d of %d in %s  - %s\n\n", index+1, len(configEntries), duration.Round(time.Second), time.Now().Format(time.RFC822))
		}
	}

	duration := time.Since(start)
	fmt.Printf("\n\nVecart finished in %s - %s\n\n", duration.Round(time.Second), time.Now().Format(time.RFC822))
}

func generateBatchSeedConfigs(batchSize int) {
	RandSource2 := rand.New(rand.NewPCG(uint64(time.Now().Unix()), uint64(time.Now().Unix())))
	var newConfigEntries []ConfigEntry
	for _, configEntry := range configEntries {
		for j := 0; j < batchSize; j++ {
			newConfig := configEntry.config.Copy()

			newConfig.RandomSeed = RandSource2.Int()
			basePath := newConfig.OutputPath[:strings.LastIndex(newConfig.OutputPath, ".")]
			if newConfig.Mode == GCode {
				newConfig.OutputPath = basePath + "_" + strconv.FormatInt(int64(j), 10) + ".gcode"
			} else {
				newConfig.OutputPath = basePath + "_" + strconv.FormatInt(int64(j), 10) + ".svg"
			}
			newConfigEntries = append(
				newConfigEntries, NewConfigEntry(
					configEntry.name+"_"+strconv.FormatInt(int64(j), 10), newConfig, configEntry.userKeys, configEntry.userShapes))
		}
	}
	configEntries = append(configEntries, newConfigEntries...)
}

func parseArguments() (bool, int, bool) {
	argsWithoutProg := os.Args[1:]
	batchSeedsMode := false
	batchSize := 0
	randomConfigOrder := false

	if len(argsWithoutProg) > 0 {
		for _, argument := range argsWithoutProg {
			if argument == "--debug" || argument == "-d" {
				utils.EnableDebugMode()
				log.Println("Debug Mode Activated")
			}
		}
		for i := 0; i < len(argsWithoutProg); i++ {
			switch strings.ToLower(argsWithoutProg[i]) {
			case "--license", "-l":
				fmt.Println(getLicense())
				os.Exit(0)

			case "--updateProvedFiles", "--updateprovedfiles", "-u":
				utils.EnableTestMode()
				UpdateProvedSVG()
				UpdateProvedGCode()
				os.Exit(0)

			case "--version", "-v":
				fmt.Printf("Vecart version %s\n", utils.GetVersion())
				os.Exit(0)
			case "--help", "-h":
				printHelp()
				os.Exit(0)
			case "--delaystart", "--delayStart", "-ds":
				{
					if i >= len(argsWithoutProg)-1 {
						fmt.Println("Invalid number of arguments for delay start!")
						printUsage()
						os.Exit(66)
					}
					i++

					delayStart, err := strconv.Atoi(argsWithoutProg[i])
					if err != nil {
						fmt.Println("Invalid delay start second value'" + argsWithoutProg[i] + "'")
						os.Exit(66)
					}

					log.Printf("Delaying start for %d seconds\n", delayStart)
					time.Sleep(time.Duration(delayStart) * time.Second)
					log.Println("Starting")
				}
			case "--randomorder", "--randomOrder", "-ro":
				randomConfigOrder = true

			case "--batchseeds", "--batchSeeds", "-bs":
				if i >= len(argsWithoutProg)-1 {
					fmt.Println("Invalid number of arguments for batchSeeds!")
					printUsage()
					os.Exit(66)
				}
				i++

				batchSizeTmp, err := strconv.Atoi(argsWithoutProg[i])
				if err != nil {
					fmt.Println("Invalid batch size'" + argsWithoutProg[i] + "'")
					os.Exit(66)
				}
				batchSize = batchSizeTmp
				batchSeedsMode = true
			case "--debug", "--Debug", "-d":
				// Already handled above
			default:
				readConfig(argsWithoutProg[i])
			}
		}
	}

	return batchSeedsMode, batchSize, randomConfigOrder
}

func readConfig(configPath string) {
	filepath, err := os.Stat(configPath)
	if err != nil {
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Cannot get current working directory")
			panic(66)
		}
		filepathTmp, err := os.Stat(pwd + configPath)
		if err != nil {
			fmt.Println("Invalid config directory or path")
			return
		}
		filepath = filepathTmp
		configPath = pwd + configPath
	}
	if filepath.IsDir() {
		readDirectory(configPath)
	} else {
		content, err := utils.GetFileContentsFromFilePath(configPath)
		if err != nil {
			fmt.Println("Could not read config from '" + configPath + "'")
			return
		}
		var currentConfig = NewConfig()
		currentUserConfig, userShapes, otherconfigs, ok := currentConfig.FromJSON(content, configPath)
		if !ok {
			if !confirmContinuationWithInvalidConfig() {
				return
			}
		}
		firstConfigName := configPath
		indexOffset := 0
		if len(otherconfigs) > 0 {
			firstConfigName = configPath + "_gs_0"
			indexOffset = 1
		}
		configEntries = append(configEntries, NewConfigEntry(firstConfigName, currentConfig, currentUserConfig, userShapes))
		for i, config := range otherconfigs {
			configEntries = append(configEntries, NewConfigEntry(configPath+"_gs_"+strconv.FormatInt(int64(i+indexOffset), 10), *config, currentUserConfig, userShapes))
		}
	}
}

func readDirectory(directoryPath string) {
	configFiles, err := os.ReadDir(directoryPath)
	if err != nil {
		fmt.Printf("Could not read configs from directory '%s'\n", directoryPath)
		if utils.DebugModeEnabled() {
			log.Println(err)
		}
		return
	}

	for _, configFile := range configFiles {
		if !strings.HasSuffix(configFile.Name(), ".json") {
			continue
		}
		if configFile.IsDir() {
			readDirectory(strings.ReplaceAll(directoryPath+"/"+configFile.Name(), "//", "/"))
			continue
		}
		readConfig(strings.ReplaceAll(directoryPath+"/"+configFile.Name(), "//", "/"))
	}
}

func getExampleConfig() {
	fmt.Println("No configuration provided. Continuing with example configuration!")
	config, userConfigKeys, userShapes, err := GetConfigFromStaticAssets("static/configs/ellie.json")
	if err != nil {
		fmt.Println(err)
		panic("Could not get example config file from static assets!")
	}
	configEntries = append(configEntries, NewConfigEntry("ExampleConfig", config, userConfigKeys, userShapes))
}

func checkBatchModeConfigs(configFilePaths []string) bool {
	ok := true
	for _, configFilePath := range configFilePaths {
		tmpConfig := NewConfig()
		content, err := utils.GetFileContentsFromFilePath(configFilePath)
		if err != nil {
			occuredErrors = append(occuredErrors, "Could not read config from '"+configFilePath+"'")
			return false
		}

		_, _, _, success := tmpConfig.FromJSON(content, configFilePath)
		if !success {
			occuredErrors = append(occuredErrors, "Finished parsing config with errors from '"+configFilePath+"'")
			ok = false
		}
	}

	return ok
}

func runIteration(configEntry ConfigEntry) {
	Config = configEntry.config
	CurrentConfigEntry = configEntry
	if utils.DebugModeEnabled() {
		log.Println("\nConfig:")
		log.Printf("%s\n\n", Config.ToJson())
	}

	fileContent := ""
	switch Config.Mode {
	case ShapeArt:
		fileContent = StartShapeArtGeneration()
	case SingleLine:
		fmt.Printf("\nInvalid mode '%d'\n\n", Config.Mode)
		return
		//fileContent = StartSingleLineArtGeneration()
	case Mosaic:
		fmt.Printf("\nInvalid mode '%d'\n\n", Config.Mode)
		return
		//fileContent = startMosaicArtGeneration()
	case GCode:
		fileContent = ConvertSVGToGCode()
	default:
		fmt.Printf("\nInvalid mode '%d'\n\n", Config.Mode)
		return
	}

	if fileContent == "" {
		return
	}

	if fileExists(Config.OutputPath) && !Config.OverwriteExisting {
		basePath := Config.OutputPath[:strings.LastIndex(Config.OutputPath, ".")]
		index := 2
		for fileExists(Config.OutputPath) {
			Config.OutputPath = basePath + "_" + strconv.FormatInt(int64(index), 10) + ".svg"
			index++
		}
	}
	utils.WriteStringToFile(fileContent, Config.OutputPath)

}

func fileExists(filePath string) bool {
	_, err := utils.GetFileContentsFromRelativeFilePath(filePath)
	return err == nil
}

func confirmContinuationWithInvalidConfig() bool {
	fmt.Println("Errors occured while parsing Config. Press 'm' or start Vecart in debug mode for more details (' \"--debug\" flag).")
	for {
		option, ok := askForOption("\nContinue despite the errors?", []string{"yes", "no", "more info"})

		if !ok || option == "no" {
			os.Exit(1)
		}

		if option == "more info" {
			for _, errorString := range occuredErrors {
				fmt.Println(errorString)
			}
			continue
		}

		break
	}

	return true
}

func askForOption(question string, options []string) (string, bool) {
	reader := bufio.NewReader(os.Stdin)
	iterations := 0
	var optionsFirstChar []string
	var optionsSuffix []string
	for _, option := range options {
		suffix := ""
		for index, char := range option {
			if index == 0 {
				optionsFirstChar = append(optionsFirstChar, string(rune(char)))
				continue
			}
			suffix += string(rune(char))
		}
		optionsSuffix = append(optionsSuffix, suffix)
	}
	for {
		fmt.Printf("%s [", question)
		for index := range options {
			fmt.Printf("(%s)%s", optionsFirstChar[index], optionsSuffix[index])
			if index != len(options)-1 {
				fmt.Print(" | ")
			}
		}
		fmt.Print("]: ")

		response, err := reader.ReadString('\n')
		if err != nil {
			if utils.DebugModeEnabled() {
				log.Printf("Error ocurred while reading user input: '%e'\n", err)
			}
		}

		response = strings.ToLower(strings.TrimSpace(response))

		for index, option := range options {
			if response == strings.ToLower(strings.TrimSpace(option)) || response == strings.ToLower(strings.TrimSpace(optionsFirstChar[index])) {
				return option, true
			}
		}

		iterations++
		if iterations >= 3 {
			fmt.Println("Received Invalid input three times. Aborting!")
			return "", false
		}
	}
}

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)
	iterations := 0

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			if utils.DebugModeEnabled() {
				log.Printf("Error ocurred while reading user input: '%e'\n", err)
			}
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
		iterations++
		if iterations >= 3 {
			fmt.Println("Received Invalid input three times. Aborting!")
			return false
		}
	}
}

func getLicense() string {
	licenseFile, err := License.Open("LICENSE")
	if err != nil {
		fmt.Println("Can not open file 'LICENSE' from static Vecart ressources!")
		fmt.Println(err)
		return ""
	}
	defer licenseFile.Close()

	license, err := utils.GetFileContentsFromStaticAssets(licenseFile)

	if err != nil {
		fmt.Println("Can not parse file 'LICENSE' from static Vecart ressources!")
		fmt.Println(err)
		return ""
	}

	return license
}

func getAllShapeVariants(xOffset float64) []*shapes.Shape {
	var shapeList []*shapes.Shape

	currentXOffset := 0.0

	for _, shape := range Config.ShapeList {
		for _, variant := range shape.Variants {
			copy := variant.TransformCopy(currentXOffset, 0)
			copy.MMToPixel(Config.OutputDpi)
			shapeList = append(shapeList, copy)
			currentXOffset += xOffset
		}
	}
	return shapeList
}

func ResizeImage(img image.Image, width, height int) image.Image {
	g := gift.New(gift.Resize(width, height, gift.LanczosResampling))
	dst := image.NewNRGBA(g.Bounds(img.Bounds()))
	g.Draw(dst, img)

	return dst
}

func printHelp() {
	helpFile, err := StaticAssets.Open("static/help.txt")
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Println("Can not open help file from static Vecart ressources!")
			log.Println(err)
		}
		return
	}
	defer helpFile.Close()

	helpString, err := utils.GetFileContentsFromStaticAssets(helpFile)
	if err != nil {
		if utils.DebugModeEnabled() {
			log.Println("Can not read help text from static Vecart ressources!")
			log.Println(err)
		}
		return
	}
	fmt.Println()
	fmt.Println(helpString)
	fmt.Println()
	printUsage()
	fmt.Println()
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  Vecart")
	fmt.Println("  Vecart [options] [pathToConfig/pathToConfigDirectory] ...")

	fmt.Println("  Options:")
	fmt.Println("      --version /-v                | Show the Vecart version")
	fmt.Println("      --help / -h                  | Show help message")
	fmt.Println("      --license / -l               | Show license information")
	fmt.Println("      --debug / -d                 | Enable debug mode with more verbose logging")
	fmt.Println("      --batchseed / -bs [nrOfRuns] | Generate multiple variants of all configs with different random seeds")
	fmt.Println("      --randomOrder / -ro          | Randomize the order of multiple configs")
	fmt.Println("      --delayStart / -ds           | Wait before starting.")
	fmt.Println("      --updateProvedFiles / -us      | Update the proved SVG and GCode files.")
}
