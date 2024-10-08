package main

//TODO
//Other Art Types
//Generate GCode

import (
	"bufio"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/fs"
	"math"
	"math/rand/v2"
	"os"
	"strings"
	"time"

	"github.com/disintegration/gift"
)

const Version = "1.0"

var Config VecartConfig
var UserConfig string
var RandSource *rand.Rand
var Fonts map[string]*Font

func main() {
	start := time.Now()
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) > 0 {
		switch strings.ToLower(argsWithoutProg[0]) {
		case "license", "--license", "-license", "-l":
			fmt.Println(getLicense())
			return
		case "updateprovedsvg":
			updateProvedSVG()
			return
		case "help", "-help", "h", "-h":
			printUsage()
		}
	}

	fmt.Printf("Vecart v%s - by David Jilg (david-jilg.com/vecart)\n\n", Version)

	Fonts = loadFonts()

	Config = NewConfig()

	if len(argsWithoutProg) > 0 {
		content, err := getFileContentsFromFilePath(argsWithoutProg[0])
		if err != nil {
			fmt.Println("Could not read config from '" + argsWithoutProg[0] + "'")
			return
		}
		ok := Config.fromJSON(content)
		if !ok {

			fmt.Println("Errors occured while parsing Config. Start Vecart in debug mode for more details (' \"debug\": true ' in config).")
			for {
				option, ok := askForOption("\nContinue despite the errors?", []string{"yes", "no", "more info"})

				if !ok || option == "no" {
					return
				}

				if option == "more info" {
					for _, errorString := range errors {
						fmt.Println(errorString)
					}
					continue
				}

				break
			}

		}
		UserConfig = content
	} else {
		printUsage()
		fmt.Println()
		fmt.Println("No configuration provided. Continuing with example configuration!")
		config, err := getConfigFromStaticAssets("static/configs/ellie.json")
		if err != nil {
			fmt.Println(err)
			panic("Could not get example config file from static assets!")
		}
		Config = config
	}

	if Config.debug {
		fmt.Println("\nConfig:")
		fmt.Printf("%s\n\n", Config.toJson())
	}

	svg := startVecart()
	if svg != "" {
		writeStringToFile(svg, Config.outputPath)
		duration := time.Since(start)
		fmt.Printf("\n\nVecart finished in %s\n\n", duration.Round(time.Second))
	}
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
			if Config.debug {
				fmt.Printf("Error ocurred while reading user input: '%e'\n", err)
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
			if Config.debug {
				fmt.Printf("Error ocurred while reading user input: '%e'\n", err)
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
		if Config.debug {
			fmt.Println("Can not open font file 'IBM-Plex-Sans.svg' from static Vecart ressources!")
			fmt.Println(err)
		}
		return fonts
	}
	defer svgFile.Close()

	xmlString, err := getFileContentsFromStaticAssets(svgFile)
	if err != nil {
		if Config.debug {
			fmt.Println("Can not read font file 'IBM-Plex-Sans.svg' from static Vecart ressources!")
			fmt.Println(err)
		}
		return fonts
	}

	font := Font{}
	font.name = "IBM-Plex-Sans"
	err = font.fromXML(xmlString)
	if err != nil {
		if Config.debug {
			fmt.Println("Can not parse font file 'IBM-Plex-Sans.svg' from static Vecart ressources!")
			fmt.Println(err)
		}
		return fonts
	}

	fonts["IBM-Plex-Sans"] = &font

	return fonts
}

func startVecart() string {
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
			fmt.Println("Can not decode image 'ellie.jpg' from static Vecart ressources!")
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

	initialize(greyscaleImg, calculateNeighborRange())

	return generateVectorArt(greyscaleImg.Bounds().Max.X, greyscaleImg.Bounds().Max.Y)
}

func getAllShapeVariants(xOffset float64) []*Shape{
	var shapes []*Shape

	currentXOffset := 0.0

	for _, shape := range Config.shapes{
		for _, variant := range shape.Variants{
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

func printUsage() {
	fmt.Println("Usage")
	fmt.Println("  Vecart [pathToJSONConfig]")
	fmt.Println("  Vecart --help")
	fmt.Println("  Vecart --license")
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
