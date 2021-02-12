package main

import (
	"github.com/tidwall/gjson"
	"github.com/tushar2708/altcsv"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type I18nData struct {
	key   string
	value map[string]string
}

const FILE_PREFIX = "locale-"

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

	files := getJsonFiles(filepath.Join(wd, ".."))
	log.Println(files)
	log.Println("---")
	log.Println("---")
	log.Println("---")

	model := readJsonFiles(&files, FILE_PREFIX)
	log.Println(model)

	headers := make([]string, 1)
	headers[0] = "key"
	for _, f := range files {
		headers = append(headers, getLangCodeFromFilename(f, FILE_PREFIX))
	}

	writeCsv(filepath.Join(wd, "..", "i18n.csv"), headers, model)
}

func getJsonFiles(path string) (resultFilenames []string) {
	infoList, err := ioutil.ReadDir(path)
	if err != nil {
		log.Panic(err)
	}

	for _, info := range infoList {
		if filepath.Ext(info.Name()) == ".json" {
			fullFilename := filepath.Join(path, info.Name())
			log.Println("found file: ", fullFilename)
			resultFilenames = append(resultFilenames, fullFilename)
		}
	}

	return resultFilenames
}

func readJsonFiles(files *[]string, prefix string) (resultMaps []I18nData) {
	for _, fullFilename := range *files {
		if filepath.Ext(fullFilename) != ".json" {
			continue
		}

		fileContent, err := ioutil.ReadFile(fullFilename)
		if err != nil {
			log.Panic("readJsonFiles error on ReadFile", fullFilename)
		}

		langCode := getLangCodeFromFilename(fullFilename, prefix)

		parsedResult := gjson.ParseBytes(fileContent)
		parsedResult.ForEach(func(key, value gjson.Result) bool {
			log.Println(key)
			build(&resultMaps, langCode, key.String(), value)
			return true
		})
	}
	return resultMaps
}

func getLangCodeFromFilename(fullFilename, prefix string) string {
	baseFilename := filepath.Base(fullFilename)
	return baseFilename[len(prefix) : len(baseFilename)-len(filepath.Ext(baseFilename))]
}

func build(i18nData *[]I18nData, langCode, previousKey string, obj gjson.Result) {
	obj.ForEach(func(key, value gjson.Result) bool {
		newKey := previousKey + "." + key.String()
		log.Println(newKey)
		if value.IsObject() || value.IsArray() {
			build(i18nData, langCode, newKey, value)
		} else {
			index := findIndex(newKey, i18nData)
			if index >= 0 {
				// update
				(*i18nData)[index].value[langCode] = value.String()
			} else {
				// add
				m := make(map[string]string, 1)
				m[langCode] = value.String()
				*i18nData = append(*i18nData, I18nData{key: newKey, value: m})
			}
		}
		return true
	})
}

func findIndex(key string, i18nData *[]I18nData) int {
	for i, data := range *i18nData {
		if data.key == key {
			return i
		}
	}
	return -1
}

func writeCsv(exportFile string, headers []string, i18nData []I18nData) {
	file, err := os.Create(exportFile)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()
	w := altcsv.NewWriter(file)
	w.Comma = ';'
	w.AllQuotes = true
	w.Quote = '"'
	w.Write(headers)
	for _, d := range i18nData {
		line := make([]string, 1)
		line[0] = d.key
		for i, langCode := range headers {
			if i == 0 {
				continue
			}
			line = append(line, d.value[langCode])
		}
		w.Write(line)
	}
	w.Flush()
}
