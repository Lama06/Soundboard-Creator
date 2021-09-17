package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	inputDir  = "soundboard"
	outputDir = "website"
)

var forbiddenCharacters = strings.NewReplacer(
	"ä", "ae",
	"ö", "oe",
	"ü", "ue",
)

func removeFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	return filename[0 : len(filename)-len(ext)]
}

func copyFile(src, dst string) {
	source, err := os.Open(src)
	if err != nil {
		panic(err)
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		panic(err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		panic(err)
	}
}

type soundData struct {
	DisplayName    string
	InputFilename  string
	OutputFilename string
}

type categoryData struct {
	DisplayName    string
	InputFilename  string
	OutputFilename string
	Sounds         []soundData
}

func loadCategory(path string) categoryData {
	data := categoryData{
		DisplayName:    filepath.Base(path),
		InputFilename:  filepath.Base(path),
		OutputFilename: strings.ToLower(forbiddenCharacters.Replace(strings.ReplaceAll(filepath.Base(path), " ", "-"))),
	}

	files, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	for _, info := range files {
		if info.IsDir() {
			continue
		}

		data.Sounds = append(data.Sounds, soundData{
			DisplayName:    removeFileExtension(info.Name()),
			InputFilename:  info.Name(),
			OutputFilename: strings.ToLower(forbiddenCharacters.Replace(strings.ReplaceAll(info.Name(), " ", "-"))),
		})
	}

	return data
}

type soundboardData struct {
	DisplayName     string
	DefaultCategory string
	Categories      []categoryData
}

type soundboardConfigFile struct {
	Name            string
	DefaultCategory string
}

func loadSoundboardData() soundboardData {
	data := soundboardData{}

	files, err := os.ReadDir(inputDir)
	if err != nil {
		panic(err)
	}

	for _, info := range files {
		if !info.IsDir() {
			if info.Name() == "soundboard.json" {
				config := &soundboardConfigFile{}
				file, err := os.ReadFile(filepath.Join(inputDir, "soundboard.json"))
				if err != nil {
					panic(err)
				}
				json.Unmarshal(file, config)

				data.DisplayName = config.Name
				data.DefaultCategory = config.DefaultCategory
			}

			continue
		}

		data.Categories = append(data.Categories, loadCategory(filepath.Join(inputDir, info.Name())))
	}

	return data
}

func createSoundboard(soundboard soundboardData) {
	err := os.RemoveAll(outputDir)
	if err != nil {
		panic(err)
	}
	err = os.Mkdir(outputDir, 0700)
	if err != nil {
		panic(err)
	}

	redirectTemplateContent, err := os.ReadFile("redirect.html")
	if err != nil {
		panic(err)
	}

	soundboardTemplateContent, err := os.ReadFile("soundboard.html")
	if err != nil {
		panic(err)
	}

	for _, category := range soundboard.Categories {
		os.Mkdir(filepath.Join(outputDir, category.OutputFilename), 0700)
		for _, sound := range category.Sounds {
			copyFile(filepath.Join(inputDir, category.InputFilename, sound.InputFilename), filepath.Join(outputDir, category.OutputFilename, sound.OutputFilename))
		}

		categoryPage, err := os.Create(filepath.Join(outputDir, category.OutputFilename, "index.html"))
		if err != nil {
			panic(err)
		}

		t := template.New(category.DisplayName)
		t.Funcs(template.FuncMap{
			"currentCategory": func() categoryData {
				return category
			},
		})
		_, err = t.Parse(string(soundboardTemplateContent))
		if err != nil {
			panic(err)
		}
		t.Execute(categoryPage, soundboard)
	}

	indexPage, err := os.Create(filepath.Join(outputDir, "index.html"))
	if err != nil {
		panic(err)
	}

	t, err := template.New("Redirect").Parse(string(redirectTemplateContent))
	if err != nil {
		panic(err)
	}
	t.Execute(indexPage, "./"+strings.ToLower(forbiddenCharacters.Replace(strings.ReplaceAll(soundboard.DefaultCategory, " ", "-")))+"/index.html")
}

func main() {
	fmt.Println("Erstelle Soundboard...")
	createSoundboard(loadSoundboardData())
	fmt.Println("Fertig!")
}
