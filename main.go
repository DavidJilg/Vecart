package main

import (
	"bufio"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/fs"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/gift"
)

const Version = "2.0.0"

type ConfigEntry struct {
	name           string
	config         VecartConfig
	originalConfig VecartConfig
	userKeys       []string
	userShapes     any
}

func (configEntry *ConfigEntry) toShortJson() string {
	return configEntry.config.toShortJson(configEntry.userKeys, configEntry.userShapes)
}

func NewConfigEntry(name string, config VecartConfig, userKeys []string, userShapes any) ConfigEntry {
	var configEntry ConfigEntry
	configEntry.name = name
	configEntry.config = config
	configEntry.originalConfig = config.copy()
	configEntry.userKeys = userKeys
	configEntry.userShapes = userShapes
	return configEntry
}

var Log bool

var CurrentConfigEntry ConfigEntry
var Config VecartConfig

var configEntries []ConfigEntry
var RandSource *rand.Rand
var Fonts map[string]*Font

var TestMode bool

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	start := time.Now()

	batchSeedsMode, batchSize, randomConfigOrder := parseArguments()

	fmt.Printf("Vecart v%s - by David Jilg (david-jilg.com/vecart)\n", Version)
	Fonts = loadFonts()

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
			fmt.Printf("\nRunning configuration %d of %d (%s) - %s", index+1, len(configEntries), currentConfigEntry.name, time.Now().Format(time.RFC822))
		} else {
			fmt.Printf("\nRunning configuration '%s' - %s", currentConfigEntry.name, time.Now().Format(time.RFC822))
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

func parseArguments() (bool, int, bool) {
	argsWithoutProg := os.Args[1:]
	batchSeedsMode := false
	batchSize := 0
	randomConfigOrder := false

	if len(argsWithoutProg) > 0 {
		for _, argument := range argsWithoutProg {
			if argument == "--debug" || argument == "-d" {
				Log = true
				log.Println("Debug Mode Activated")
			}
		}
		for i := 0; i < len(argsWithoutProg); i++ {
			switch strings.ToLower(argsWithoutProg[i]) {
			case "--license", "-l":
				fmt.Println(getLicense())
				os.Exit(0)

			case "--updateprovedsvg", "-u":
				TestMode = true
				updateProvedSVG()
				os.Exit(0)

			case "--version", "-v":
				fmt.Printf("Vecart version %s\n", Version)
				os.Exit(0)
			case "--help", "-h":
				printHelp()
				os.Exit(0)
			case "--debug", "-d":
				{
				}
			case "--delaystart":
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
			case "--randomOrder", "-ro":
				randomConfigOrder = true

			case "--batchSeeds", "-bs":
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
			default:
				readConfig(argsWithoutProg[i])
			}
		}
	}

	return batchSeedsMode, batchSize, randomConfigOrder
}

func generateBatchSeedConfigs(batchSize int) {
	RandSource2 := rand.New(rand.NewPCG(uint64(time.Now().Unix()), uint64(time.Now().Unix())))
	var newConfigEntries []ConfigEntry
	for _, configEntry := range configEntries {
		for j := 0; j < batchSize; j++ {
			newConfig := configEntry.config.copy()

			newConfig.randomSeed = RandSource2.Int()
			basePath := newConfig.outputPath[:strings.LastIndex(newConfig.outputPath, ".")]
			newConfig.outputPath = basePath + "_" + strconv.FormatInt(int64(j), 10) + ".svg"

			newConfigEntries = append(
				newConfigEntries, NewConfigEntry(
					configEntry.name+"_"+strconv.FormatInt(int64(j), 10), newConfig, configEntry.userKeys, configEntry.userShapes))
		}
	}
	configEntries = append(configEntries, newConfigEntries...)
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
		content, err := getFileContentsFromFilePath(configPath)
		if err != nil {
			fmt.Println("Could not read config from '" + configPath + "'")
			return
		}
		var currentConfig = NewConfig()
		currentUserConfig, userShapes, otherconfigs, ok := currentConfig.fromJSON(content)
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
		if Log {
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
	config, userConfigKeys, userShapes, err := getConfigFromStaticAssets("static/configs/ellie.json")
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
		content, err := getFileContentsFromFilePath(configFilePath)
		if err != nil {
			occuredErrors = append(occuredErrors, "Could not read config from '"+configFilePath+"'")
			return false
		}

		_, _, _, success := tmpConfig.fromJSON(content)
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
	if Log {
		log.Println("\nConfig:")
		log.Printf("%s\n\n", Config.toJson())
	}

	svg := ""
	switch Config.mode {
	case Shapes:
		svg = startShapeArtGeneration()
	case SingleLine:
		svg = startSingleLineArtGeneration()
	case Mosaic:
		fmt.Printf("\nInvalid mode '%d'\n\n", Config.mode)
		return
		//svg = startMosaicArtGeneration()
	default:
		fmt.Printf("\nInvalid mode '%d'\n\n", Config.mode)
		return
	}

	if svg == "" {
		return
	}

	if fileExists(Config.outputPath) && !Config.overwriteExisting {
		basePath := Config.outputPath[:strings.LastIndex(Config.outputPath, ".")]
		index := 2
		for fileExists(Config.outputPath) {
			Config.outputPath = basePath + "_" + strconv.FormatInt(int64(index), 10) + ".svg"
			index++
		}
	}
	writeStringToFile(svg, Config.outputPath)

}

func fileExists(filePath string) bool {
	_, err := getFileContentsFromRelativeFilePath(filePath)
	return err == nil
}

func confirmContinuationWithInvalidConfig() bool {
	fmt.Println("Errors occured while parsing Config. Press 'm' or start Vecart in debug mode for more details (' \"debug\": true ' in config).")
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
			if Log {
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
			if Log {
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

	license, err := getFileContentsFromStaticAssets(licenseFile)

	if err != nil {
		fmt.Println("Can not parse file 'LICENSE' from static Vecart ressources!")
		fmt.Println(err)
		return ""
	}

	return license
}

func loadFonts() map[string]*Font {
	fonts := make(map[string]*Font)
	svgFile, err := StaticAssets.Open("static/fonts/IBM-Plex-Sans.svg")
	if err != nil {
		if Log {
			log.Println("Can not open font file 'IBM-Plex-Sans.svg' from static Vecart ressources!")
			log.Println(err)
		}
		return fonts
	}
	defer svgFile.Close()

	xmlString, err := getFileContentsFromStaticAssets(svgFile)
	if err != nil {
		if Log {
			log.Println("Can not read font file 'IBM-Plex-Sans.svg' from static Vecart ressources!")
			log.Println(err)
		}
		return fonts
	}

	font := Font{}
	font.name = "IBM-Plex-Sans"
	err = font.fromXML(xmlString)
	if err != nil {
		if Log {
			log.Println("Can not parse font file 'IBM-Plex-Sans.svg' from static Vecart ressources!")
			log.Println(err)
		}
		return fonts
	}

	fonts["IBM-Plex-Sans"] = &font

	return fonts
}

func startShapeArtGeneration() string {
	RandSource = rand.New(rand.NewPCG(uint64(Config.randomSeed), uint64(Config.randomSeed)))

	var img image.Image
	var err error
	if Config.inputPath == "" {
		ellieFile, err := StaticAssets.Open("static/ellie.png")
		if err != nil {
			fmt.Println("Can not open image 'ellie.png' from static Vecart ressources!")
			fmt.Println(err)
			return ""
		}
		defer ellieFile.Close()

		img, _, err = image.Decode(ellieFile)
		if err != nil {
			fmt.Println("Can not decode image 'ellie.png' from static Vecart ressources!")
			fmt.Println(err)
			return ""
		}
	} else {
		img, err = getImageFromFilePath(Config.inputPath)
		if err != nil {
			fmt.Printf("Can not decode image '%s'!\n", Config.inputPath)
			fmt.Println(err)
			return ""
		}
	}

	if Config.artworkWidth != 0 && Config.artworkHeight != 0 {
		imageWidthPixel := int(math.Round(mmToPixel(float64(Config.artworkWidth), Config.processingDpi)))
		imageHeightPixel := int(math.Round(mmToPixel(float64(Config.artworkHeight), Config.processingDpi)))

		imageWidth := imageWidthPixel - (imageWidthPixel % Config.quadrantWidth)
		imageHeight := imageHeightPixel - (imageHeightPixel % Config.quadrantHeight)

		img = resizeImage(img, imageWidth, imageHeight)
	} else if Config.artworkWidth != 0 || Config.artworkHeight != 0 {
		imageWidthPixel := int(math.Round(mmToPixel(float64(Config.artworkWidth), Config.processingDpi)))
		imageHeightPixel := int(math.Round(mmToPixel(float64(Config.artworkHeight), Config.processingDpi)))
		img = resizeImage(img, imageWidthPixel, imageHeightPixel)

		imageWidth := img.Bounds().Max.X - (img.Bounds().Max.X % Config.quadrantWidth)
		imageHeight := img.Bounds().Max.Y - (img.Bounds().Max.Y % Config.quadrantHeight)

		img = resizeImage(img, imageWidth, imageHeight)
	}

	greyscaleImg := image.NewGray(img.Bounds())
	draw.Draw(greyscaleImg, greyscaleImg.Bounds(), img, img.Bounds().Min, draw.Src)

	initializeForShapeGeneration(greyscaleImg, calculateNeighborRange())

	return generateShapeArt(greyscaleImg.Bounds().Max.X, greyscaleImg.Bounds().Max.Y)
}

func startSingleLineArtGeneration() string {
	RandSource = rand.New(rand.NewPCG(uint64(Config.randomSeed), uint64(Config.randomSeed)))

	var img image.Image
	var err error

	if Config.inputPath == "" {
		ellieFile, err := StaticAssets.Open("static/ellie.png")
		if err != nil {
			fmt.Println("Can not open image 'ellie.png' from static Vecart ressources!")
			fmt.Println(err)
			return ""
		}
		defer ellieFile.Close()

		img, _, err = image.Decode(ellieFile)
		if err != nil {
			fmt.Println("Can not decode image 'ellie.png' from static Vecart ressources!")
			fmt.Println(err)
			return ""
		}
	} else {
		img, err = getImageFromFilePath(Config.inputPath)
		if err != nil {
			fmt.Printf("Can not decode image '%s'!\n", Config.inputPath)
			fmt.Println(err)
			return ""
		}
	}

	if Config.artworkWidth != 0 && Config.artworkHeight != 0 {
		imageWidthPixel := int(math.Round(mmToPixel(float64(Config.artworkWidth), Config.processingDpi)))
		imageHeightPixel := int(math.Round(mmToPixel(float64(Config.artworkHeight), Config.processingDpi)))

		imageWidth := imageWidthPixel - (imageWidthPixel % Config.quadrantWidth)
		imageHeight := imageHeightPixel - (imageHeightPixel % Config.quadrantHeight)

		img = resizeImage(img, imageWidth, imageHeight)
	} else if Config.artworkWidth != 0 || Config.artworkHeight != 0 {
		imageWidthPixel := int(math.Round(mmToPixel(float64(Config.artworkWidth), Config.processingDpi)))
		imageHeightPixel := int(math.Round(mmToPixel(float64(Config.artworkHeight), Config.processingDpi)))
		img = resizeImage(img, imageWidthPixel, imageHeightPixel)

		imageWidth := img.Bounds().Max.X - (img.Bounds().Max.X % Config.quadrantWidth)
		imageHeight := img.Bounds().Max.Y - (img.Bounds().Max.Y % Config.quadrantHeight)

		img = resizeImage(img, imageWidth, imageHeight)
	}

	greyscaleImg := image.NewGray(img.Bounds())
	draw.Draw(greyscaleImg, greyscaleImg.Bounds(), img, img.Bounds().Min, draw.Src)

	initializeQuadrants(greyscaleImg, calculateNeighborRange())

	return generateSingleLineArt(greyscaleImg.Bounds().Max.X, greyscaleImg.Bounds().Max.Y)
}

func getAllShapeVariants(xOffset float64) []*Shape {
	var shapes []*Shape

	currentXOffset := 0.0

	for _, shape := range Config.shapes {
		for _, variant := range shape.Variants {
			copy := variant.transformCopy(currentXOffset, 0)
			copy.mmToPixel(Config.outputDpi)
			shapes = append(shapes, copy)
			currentXOffset += xOffset
		}
	}
	return shapes
}

func resizeImage(img image.Image, width, height int) image.Image {
	g := gift.New(gift.Resize(width, height, gift.LanczosResampling))
	dst := image.NewNRGBA(g.Bounds(img.Bounds()))
	g.Draw(dst, img)

	return dst
}

func writeStringToFile(content, path string) {
	file, err := createFile(path)
	if err != nil {
		fmt.Printf("Could not create file '%s'\n", path)
		fmt.Println(err)
		return
	}

	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	writer.WriteString(content)
}

func printHelp() {
	helpFile, err := StaticAssets.Open("static/help.txt")
	if err != nil {
		if Log {
			log.Println("Can not open help file from static Vecart ressources!")
			log.Println(err)
		}
		return
	}
	defer helpFile.Close()

	helpString, err := getFileContentsFromStaticAssets(helpFile)
	if err != nil {
		if Log {
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
}

func createFile(path string) (*os.File, error) {
	svgFile, err := os.Create(path)
	if err != nil {
		pwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		svgFile, err := os.Create(pwd + path)
		if err != nil {
			return nil, err
		}
		return svgFile, nil
	}

	return svgFile, nil
}

func getFileContentsFromFilePath(filePath string) (string, error) {
	content, err := getFileContentsFromRelativeFilePath(filePath)
	if err == nil {
		return content, err
	}

	return getFileContentsFromAbsoluteFilePath(filePath)
}

func getFileContentsFromAbsoluteFilePath(filePath string) (string, error) {
	bytes, err := os.ReadFile(filePath) // just pass the file name
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func getFileContentsFromRelativeFilePath(filePath string) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return getFileContentsFromAbsoluteFilePath(pwd + filePath)
}

func getFileContentsFromStaticAssets(file fs.File) (string, error) {
	stat, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	bs := make([]byte, stat.Size())
	_, err = file.Read(bs)
	if err != nil && err != io.EOF {
		return "", err
	}

	return string(bs), nil
}

func getImageFromFilePath(filePath string) (image.Image, error) {
	img, err := getImageFromRelativeFilePath(filePath)
	if err == nil {
		return img, err
	}

	return getImageFromAbsoluteFilePath(filePath)
}

func getImageFromAbsoluteFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	image, _, err := image.Decode(f)
	return image, err
}

func getImageFromRelativeFilePath(filePath string) (image.Image, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return getImageFromAbsoluteFilePath(pwd + filePath)
}
