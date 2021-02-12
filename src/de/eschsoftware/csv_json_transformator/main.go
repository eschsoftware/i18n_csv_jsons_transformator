package main

import (
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type I18nData struct {
	key string
	value map[string]string
}

var wd string

func init() {
	path, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}
	wd = path
}

func main() {
	log.Println(wd)

	files := getJsonAndCsvFiles(filepath.Join(wd, ".."))
	log.Println(files)
	log.Println("---")
	log.Println("---")
	log.Println("---")

	log.Println(readJsonFiles(&files))

}

func getJsonAndCsvFiles(path string) (resultFilenames []string) {
	infoList, err := ioutil.ReadDir(path)
	if err != nil {
		log.Panic(err)
	}

	for _, info := range infoList {
		if filepath.Ext(info.Name()) == ".json" || filepath.Ext(info.Name()) == ".csv" {
			fullFilename := filepath.Join(path, info.Name())
			log.Println("found file: ", fullFilename)
			resultFilenames = append(resultFilenames, fullFilename)
		}
	}

	return resultFilenames
}

func readJsonFiles(files *[]string) (resultMaps []I18nData) {
	for _, fullFilename := range *files {
		if filepath.Ext(fullFilename) != ".json" {
			continue
		}

		fileContent, err := ioutil.ReadFile(fullFilename)
		if err != nil {
			log.Panic("readJsonFiles error on ReadFile", fullFilename)
		}
		//var objmap map[string]json.RawMessage
		//json.Unmarshal(fileContent, &objmap)

		parsedResult := gjson.ParseBytes(fileContent)
		parsedResult.ForEach(func(key, value gjson.Result) bool {

		})
	}
	return resultMaps
}

func build(i18nData *I18nData, langCode, previousKey string, obj gjson.Result) {
		obj.ForEach(func(key, value gjson.Result) bool {
			newKey := previousKey + "." + key.String()
			if value.IsObject() || value.IsArray() {
				build(i18nData, newKey, value)
			} else {

			}
		})

}
