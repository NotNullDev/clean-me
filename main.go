package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"sigs.k8s.io/yaml"
)

//go:embed config-sample.yaml
var exampleConfig embed.FS

type AppFolder struct {
	Name string `yaml:"name"`
}

// TODO: update yaml
type AppFile struct {
	Pattern       string `yaml:"pattern"`
	MinSizeInMb   int    `yaml:"minSizeMb"`
	Action        string `yaml:"action"`
	ActionTo      string `yaml:"actionTo"`
	PreservePath  bool   `yaml:"preservePath"`
	OlderThanDays int    `yaml:"olderThanDays"`
}

type AppInput struct {
	Filenames          []AppFile   `yaml:"filenames"`
	Folders            []AppFolder `yaml:"folders"`
	PromptBeforeAction bool        `yaml:"promptBeforeActions"` // kind of -WhatIf from ps1
}

type InternalAppFile struct {
	file AppFile
	path string
}

func main() {
	appInput := parseUserInput()

	filesToProcess := []InternalAppFile{}

	for idxFolder, folder := range appInput.Folders {
		log.Printf("Searching folder %d: [%s]", idxFolder+1, folder.Name)

		if _, err := os.Stat(folder.Name); os.IsNotExist(err) {
			log.Printf("Error: directory [%s] doesn't exists.", folder.Name)
			continue
		}

		for _, filenameInfo := range appInput.Filenames {
			log.Printf("Compiling regex: [%s]", filenameInfo.Pattern)
			newRegexp, err := regexp.Compile(filenameInfo.Pattern)

			if err != nil {
				log.Panicf("ERROR: [%s]\n", err.Error())
			}

			filepath.WalkDir(folder.Name, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					log.Printf("ERROR: [%s]", err.Error())
				}

				info, err := d.Info()

				if err != nil {
					log.Printf("Error: could not read info of the file [%s], error: [%s]", path, err.Error())
				}

				matched := !d.IsDir() &&
					newRegexp.MatchString(d.Name()) &&
					info.ModTime().Before(time.Now().AddDate(0, 0, -filenameInfo.OlderThanDays)) &&
					(info.Size()/1000) >= int64(filenameInfo.MinSizeInMb)

				if matched {
					switch filenameInfo.Action {
					case "delete":
						log.Printf("File [%s] marked to be deleted.\n", path)
						filesToProcess = append(filesToProcess, InternalAppFile{
							file: filenameInfo,
							path: path,
						})
					case "print":
						log.Printf("Found file [%s]\n", path)
					case "move":
						log.Printf("File [%s] marked to be moved.\n", path)
						filesToProcess = append(filesToProcess, InternalAppFile{
							file: filenameInfo,
							path: path,
						})
					case "copy":
						log.Printf("File [%s] marked to be copied.\n", path)
						filesToProcess = append(filesToProcess, InternalAppFile{
							file: filenameInfo,
							path: path,
						})
					case "none":

					default:
						log.Printf("ERROR: action not supported: [%s]", filenameInfo.Action)
					}
				}

				return nil
			})
		}
	}

	processResult(filesToProcess)
}

func processResult(filesToProcess []InternalAppFile) {
	// todo: list everything and? [prompt user (to be )]

	for _, internalFile := range filesToProcess {
		switch internalFile.file.Action {
		case "delete":
			err := deleteFile(internalFile)

			if err != nil {
				log.Printf("ERROR: [delete] [%s]", err.Error())
			}
		case "copy":
			err := copyFile(internalFile)

			if err != nil {
				log.Printf("ERROR: [copy] [%s]", err.Error())
			}

		case "move":
			err := moveFile(internalFile)

			if err != nil {
				log.Printf("ERROR: [move] [%s]", err.Error())
			}
		}
	}
}

func deleteFile(fileToDelete InternalAppFile) (err error) {
	err = os.Remove(fileToDelete.path)
	if err != nil {
		err = fmt.Errorf("can't delete file [%s], info: [%s]", fileToDelete.path, err.Error())
	}
	return err
}

func copyFile(fileToCopy InternalAppFile) (err error) {
	fContent, err := os.ReadFile(fileToCopy.path)

	if err != nil {
		err = fmt.Errorf("can't read file [%s], info: [%s]", fileToCopy.path, err.Error())
		return
	}

	err = os.WriteFile(fileToCopy.file.ActionTo, fContent, 0777)

	if err != nil {
		err = fmt.Errorf("can't copy file [%s], info: [%s]", fileToCopy.path, err.Error())
		return
	}
	return nil
}

func moveFile(fileToMove InternalAppFile) (err error) {
	err = copyFile(fileToMove)
	if err != nil {
		return
	}
	err = deleteFile(fileToMove)
	return err
}

func parseUserInput() AppInput {
	configFileName := flag.String("config", "", "Path to the configuration file. ")
	gen := flag.Bool("gen", false, "If set to true, new example config file called [config.xml] will be generated in the current directory.")
	flag.Parse()

	if *gen {
		if _, err := os.Stat("config.yaml"); err == nil {
			log.Printf("Error: file [config.yaml] is present in the current directory. Please remove it in order to generate example config.")
			os.Exit(1)
		}

		exampleConfigFileContent, _ := exampleConfig.ReadFile("config-sample.yaml")

		os.WriteFile("config.yaml", exampleConfigFileContent, 0666)

		println("File [config.yaml] has been generated")
		os.Exit(1)
	}

	if *configFileName == "" {
		flag.Usage()
		os.Exit(1)
	}

	configFile, err := os.ReadFile(*configFileName)

	if err != nil {
		log.Printf("Could not read file [%s]", *configFileName)
		os.Exit(1)
	}

	parsedInput := AppInput{}

	yaml.Unmarshal(configFile, &parsedInput)

	log.Printf("%v\n", parsedInput)

	return parsedInput
}
