package main

import (
	"embed"
	"flag"
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

type AppFile struct {
	Pattern       string `yaml:"pattern"`
	MinSizeInMb   int    `yaml:"minSizeMb"`
	Delete        bool   `yaml:"delete"`
	OlderThanDays int    `yaml:"olderThanDays"`
}

type AppInput struct {
	Filenames []AppFile   `yaml:"filenames"`
	Folders   []AppFolder `yaml:"folders"`
}

func main() {
	appInput := parseUserInpit()

	filesToDelete := []string{}

	for idxFolder, folder := range appInput.Folders {
		log.Printf("Searching folder %d: [%s]", idxFolder+1, folder.Name)

		if _, err := os.Stat(folder.Name); os.IsNotExist(err) {
			log.Printf("Error: directory [%s] doesn't exists.", folder.Name)
			continue
		}

		for _, filenameInfo := range appInput.Filenames {
			log.Printf("Compiling regex: [%s]", filenameInfo.Pattern)
			newRegexpt, err := regexp.Compile(filenameInfo.Pattern)

			if err != nil {
				log.Panicf("ERRROR: [%s]\n", err.Error())
			}

			filepath.WalkDir(folder.Name, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					log.Printf("ERROR: [%s]", err.Error())
				}

				info, err := d.Info()

				if err != nil {
					log.Printf("Error: coult not read info of the file [%s], error: [%s]", path, err.Error())
				}

				matched := !d.IsDir() &&
					newRegexpt.MatchString(d.Name()) &&
					info.ModTime().Before(time.Now().AddDate(0, 0, -filenameInfo.OlderThanDays)) &&
					(info.Size()/1000) >= int64(filenameInfo.MinSizeInMb)

				if matched {
					if filenameInfo.Delete {
						log.Printf("File [%s] marked to be deleted.", path)
						filesToDelete = append(filesToDelete, path)
					} else {
						log.Printf("Found file [%s]", path)
					}
				}

				return nil
			})
		}
	}
	log.Printf("Files to delete: [%d]", len(filesToDelete))
}

func parseUserInpit() AppInput {
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
