package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"github.com/tushar2708/altcsv"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	csvExportFilename := flag.String("csv-export", "i18n.csv", "filename for the CSV to export ")
	toJson := flag.Bool("toJson", false, "transforms CSV file to different JSON files")
	csvImportFilename := flag.String("csv-import", "i18n.csv", "filename of the CSV to import")
	jsonFilePrefix := flag.String("json-file-prefix", FILE_PREFIX, "")
	flag.Parse()

	if *toJson {
		files := getJsonFiles(wd)
		model := readJsonFiles(&files, *jsonFilePrefix)
		headers := make([]string, 1)
		headers[0] = "key"
		for _, f := range files {
			headers = append(headers, getLangCodeFromFilename(f, *jsonFilePrefix))
		}
		writeCsv(filepath.Join(wd, *csvExportFilename), headers, model)
	} else {
		headers, models := readCsv(csvImportFilename)

		sort.Slice(models, func(i, j int) bool {
			return models[i].key < models[j].key
		})

		log.Println(headers, models)

		writeJsons(wd, *jsonFilePrefix, headers, models)
	}
}

func writeJsons(wd string, prefix string, headers []string, models []I18nData) {
	for i, header := range headers {
		if i == 0 {
			continue
		}
		fullFilename := filepath.Join(wd, prefix+header+".json")
		writeJson(fullFilename, header, models)
	}
}

func writeJson(filename string, header string, models []I18nData) {
	jsonData := make(map[string]interface{})
	for _, model := range models {
		keyTokens := strings.Split(model.key, ".")

		var childJsonData map[string]interface{}
		if jsonData[keyTokens[0]] == nil {
			childJsonData = make(map[string]interface{})
		} else {
			err := mapstructure.Decode(jsonData[keyTokens[0]], &childJsonData)
			if err != nil {
				log.Panic(err)
			}
		}

		if len(keyTokens) > 1 {
			jsonData[keyTokens[0]] = buildJson(childJsonData, keyTokens[1:], header, model)
		} else {
			jsonData[keyTokens[0]] = model.value[header]
		}
	}

	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(jsonData)

	if err != nil {
		log.Panic(err)
	}
	err = ioutil.WriteFile(filename, buffer.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

func buildJson(jsonData map[string]interface{}, keyTokens []string, header string, model I18nData) map[string]interface{} {
	var childJsonData map[string]interface{}
	if jsonData[keyTokens[0]] == nil {
		childJsonData = make(map[string]interface{})
	} else {
		err := mapstructure.Decode(jsonData[keyTokens[0]], &childJsonData)
		if err != nil {
			log.Panic(err)
		}
	}

	if len(keyTokens) > 1 {
		jsonData[keyTokens[0]] = buildJson(childJsonData, keyTokens[1:], header, model)
	} else {
		jsonData[keyTokens[0]] = model.value[header]
	}
	return jsonData
}

func readCsv(filename *string) ([]string, []I18nData) {
	file, err := os.Open(*filename)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()

	r := altcsv.NewReader(file)
	r.Comma = ';'
	r.Quote = '"'
	content, err := r.ReadAll()
	if err != nil {
		log.Panic(err)
	}
	models := make([]I18nData, 0)
	headers := content[0]
	for i, line := range content {
		if i == 0 {
			continue
		}
		model := I18nData{key: line[0], value: map[string]string{}}
		models = append(models, model)
		for j, e := range line {
			if j == 0 {
				continue
			}

			model.value[headers[j]] = e
		}
	}

	return headers, models
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
			buildModel(&resultMaps, langCode, key.String(), value)
			return true
		})
	}
	return resultMaps
}

func getLangCodeFromFilename(fullFilename, prefix string) string {
	baseFilename := filepath.Base(fullFilename)
	return baseFilename[len(prefix) : len(baseFilename)-len(filepath.Ext(baseFilename))]
}

func buildModel(i18nData *[]I18nData, langCode, previousKey string, obj gjson.Result) {
	obj.ForEach(func(key, value gjson.Result) bool {
		newKey := previousKey + "." + key.String()
		log.Println(newKey)
		if value.IsObject() || value.IsArray() {
			buildModel(i18nData, langCode, newKey, value)
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
