package main

import (
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path"

	"github.com/zclconf/go-cty/cty"
)

func getCtyValue(ctyType string, value interface{}) cty.Value {
	switch ctyType {
	case "string":
		return cty.StringVal(value.(string))
	case "bool":
		return cty.BoolVal(value.(bool))
	case "number":
		return cty.NumberVal(value.(*big.Float))
	case "list":
		return cty.ListVal(value.([]cty.Value))
	}
	return cty.ObjectVal(value.(map[string]cty.Value))
}

func doesFileExist(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func getFileName(folder string, isPrimary bool, blockName string, blockLabel string, newBlockRequest map[string]interface{}, w http.ResponseWriter) (string, error) {
	if !isPrimary {
		dependsOn, err := getKeyValue("dependsOn", newBlockRequest)
		if check(err, w) {
			// we let the calling method handle the error message
			return "", fmt.Errorf("")
		}
		blockLabel = dependsOn.(string)
	}
	return path.Join(folder, blockName+"_"+blockLabel+".tf"), nil
}

func writeToFile(fileName string, content []byte, w http.ResponseWriter) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if check(err, w) {
		return
	}
	_, err = f.Write(content)
	if check(err, w) {
		return
	}
	err = f.Close()
	if check(err, w) {
		return
	}
}
