package utils

import (
	"bufio"
	"fmt"
	"image"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
)

func CreateFile(path string) (*os.File, error) {
	file, err := os.Create(path)
	if err != nil {
		pwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		file, err := os.Create(pwd + path)
		if err != nil {
			return nil, err
		}
		return file, nil
	}

	return file, nil
}

func WriteStringToFile(content, path string) {
	file, err := CreateFile(path)
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

func GetFileContentsFromFilePath(filePath string) (string, error) {
	content, err := GetFileContentsFromRelativeFilePath(filePath)
	if err == nil {
		return content, err
	}

	return GetFileContentsFromAbsoluteFilePath(filePath)
}

func GetFileContentsFromAbsoluteFilePath(filePath string) (string, error) {
	bytes, err := os.ReadFile(filePath) // just pass the file name
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func GetFileContentsFromRelativeFilePath(filePath string) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return GetFileContentsFromAbsoluteFilePath(pwd + filePath)
}

func GetFileContentsFromStaticAssets(file fs.File) (string, error) {
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

func GetImageFromFilePath(filePath string) (image.Image, error) {
	img, err := GetImageFromRelativeFilePath(filePath)
	if err == nil {
		return img, err
	}

	return GetImageFromAbsoluteFilePath(filePath)
}

func GetImageFromAbsoluteFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	image, _, err := image.Decode(f)
	return image, err
}

func GetImageFromRelativeFilePath(filePath string) (image.Image, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return GetImageFromAbsoluteFilePath(pwd + filePath)
}

func GetFiles(filePathParameter string) []string {
	filepath, err := os.Stat(filePathParameter)
	if err != nil {
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Cannot get current working directory")
			panic(66)
		}
		filepathTmp, err := os.Stat(pwd + filePathParameter)
		if err != nil {
			fmt.Println("Invalid directory or path")
			return nil
		}
		filepath = filepathTmp
		filePathParameter = pwd + filePathParameter
	}
	if filepath.IsDir() {
		return GetFilesInDirectory(filePathParameter)
	} else {
		content, err := GetFileContentsFromFilePath(filePathParameter)
		if err != nil {
			fmt.Println("Could not read file from '" + filePathParameter + "'")
			return nil
		}
		return []string{content}
	}
}

func GetFilesInDirectory(directoryPath string) []string {
	var files []string
	filePaths, err := os.ReadDir(directoryPath)
	if err != nil {
		fmt.Printf("Could not read files from directory '%s'\n", directoryPath)
		if DebugModeEnabled() {
			log.Println(err)
		}
		return nil
	}

	for _, file := range filePaths {
		if file.IsDir() {
			files = append(files, GetFilesInDirectory(strings.ReplaceAll(directoryPath+"/"+file.Name(), "//", "/"))...)
			continue
		}
		files = append(files, GetFiles(strings.ReplaceAll(directoryPath+"/"+file.Name(), "//", "/"))...)
	}
	return files
}
