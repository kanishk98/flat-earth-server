package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type FlatEarthGraphUpdateRequest struct {
	BlockType     string
	BlockLabel    string
	BlockName     string
	AttributeName string
	NewValue      interface{}
	AttributeType string
}

func check(e error, w http.ResponseWriter) bool {
	if e != nil {
		// kanishk98: We should return more appropriate error codes.
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(e.Error()))
		return true
	}
	return false
}

func getKeyValue(key string, source map[string]interface{}) (interface{}, error) {
	value, isKeyPresent := source[key]
	if !isKeyPresent {
		return nil, fmt.Errorf("%s is a required key", key)
	}
	return value, nil
}

func generateFlatEarthGraph() []byte {
	graph, err := commands["flat-earth"].Run(args[0])
	graphBytes, err := json.Marshal(graph)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to generate JSON equivalent of FlatEarthGraph: %s", err))
		panic(err)
	}
	return graphBytes
}

func getFlatEarthGraph(w http.ResponseWriter, r *http.Request) {
	graphBytes := generateFlatEarthGraph()
	w.Write(graphBytes)
}

func getMatchingBlock(folder string, w http.ResponseWriter, blockType string, blockLabels []string) (*hclwrite.Block, string, *hclwrite.File) {
	var block *hclwrite.Block
	var fileName string
	var parsedConfig *hclwrite.File
	filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatalf("Failed to walk the %s directory", path)
			return err
		}
		if d.IsDir() {
			return nil
		} else if filepath.Ext(path) == ".tf" {
			tfBytes, err := ioutil.ReadFile(path)
			if check(err, w) {
				return err
			}
			hclConfig, diags := hclwrite.ParseConfig(tfBytes, path, hcl.Pos{Line: 0, Column: 0, Byte: 0})
			if diags != nil && diags.HasErrors() {
				log.Fatalf("Error when parsing TF config: %s", diags.Error())
				// move on
				return nil
			}
			block = hclConfig.Body().FirstMatchingBlock(blockType, blockLabels)
			if block == nil {
				return nil
			}
			fileName = path
			parsedConfig = hclConfig
			// no reason to continue looping over files, we have our block
			return fmt.Errorf("Fake error to exit loop!")
		} else {
			// it's a file we don't care about, move on
			return nil
		}
	})
	return block, fileName, parsedConfig
}

func updateFlatEarthGraph(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if check(err, w) {
		return
	}
	var updateRequest FlatEarthGraphUpdateRequest
	err = json.Unmarshal(body, &updateRequest)
	if check(err, w) {
		return
	}
	block, fileName, hclConfig := getMatchingBlock(args[0], w, updateRequest.BlockType, []string{updateRequest.BlockName, updateRequest.BlockLabel})
	block.Body().SetAttributeValue(updateRequest.AttributeName, getCtyValue(updateRequest.AttributeType, updateRequest.NewValue))
	err = ioutil.WriteFile(fileName, hclConfig.Bytes(), 0644)
	if check(err, w) {
		return
	}
	getFlatEarthGraph(w, r)
}

func getProviderSchema(w http.ResponseWriter, r *http.Request) {
	schema, err := commands["provider-schema"].GetProviderSchema(args[0])
	if check(err, w) {
		return
	}
	w.Write(schema)
}

func createNewBlock(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if check(err, w) {
		return
	}
	var newBlockRequest map[string]interface{}
	err = json.Unmarshal(body, &newBlockRequest)
	blockType, err := getKeyValue("blockType", newBlockRequest)
	if check(err, w) {
		return
	}
	blockName, err := getKeyValue("blockName", newBlockRequest)
	if check(err, w) {
		return
	}
	blockLabel, err := getKeyValue("blockLabel", newBlockRequest)
	if check(err, w) {
		return
	}
	file := hclwrite.NewFile()
	fileBody := file.Body()
	newBlock := fileBody.AppendNewBlock(blockType.(string), []string{blockName.(string), blockLabel.(string)})
	attributes, err := getKeyValue("attributes", newBlockRequest)
	if check(err, w) {
		return
	}
	for attributeKey, attributeValue := range attributes.(map[string]interface{}) {
		valueType, err := getKeyValue("type", attributeValue.(map[string]interface{}))
		if check(err, w) {
			return
		}
		valueValue, err := getKeyValue("value", attributeValue.(map[string]interface{}))
		if check(err, w) {
			return
		}
		newBlock.Body().SetAttributeValue(attributeKey, getCtyValue(valueType.(string), valueValue))
	}
	w.Write(file.Bytes())
}

func startServer(commands map[string]FlatEarthCommand, args []string) {
	var server http.Server
	server.Addr = ":8080"

	// GET requests
	http.HandleFunc("/get-flat-earth-graph", getFlatEarthGraph)
	http.HandleFunc("/get-provider-schema", getProviderSchema)

	// POST requests
	http.HandleFunc("/create-new-block", createNewBlock)
	http.HandleFunc("/update-flat-earth-graph", updateFlatEarthGraph)
	// kanishk98: if this port is in use, the app just crashes with a garbage error.
	// handle that case early on and either switch to another port (preferably)
	// or present some usable info to the user
	log.Println("Starting server 💩")
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		if err := server.Shutdown(context.Background()); err != nil {
			// error from closing listeners, or context timeout
			log.Fatalf("Failed to close connection: %v", err)
		}
		close(idleConnsClosed)
	}()
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		// error starting or closing listener
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
	<-idleConnsClosed
}
