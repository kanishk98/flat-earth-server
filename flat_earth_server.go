package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type FlatEarthGraphUpdateRequest struct {
	NodeKey       string
	AttributeName string
	NewValue      string
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func generateFlatEarthGraph() []byte {
	graph, err := commands["flat-earth"].Run(args[0])
	check(err)
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

func updateFlatEarthGraph(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	check(err)
	log.Println(string(body))
	var updateRequest FlatEarthGraphUpdateRequest
	err = json.Unmarshal(body, &updateRequest)
	check(err)
	graph, _ := commands["flat-earth"].Run(args[0])
	fileName := graph[updateRequest.NodeKey].DeclRange.Filename
	// kanishk98: Possibilities of the key not getting found here, handle that
	attrs, _ := graph[updateRequest.NodeKey].Config.JustAttributes()
	start := attrs[updateRequest.AttributeName].Expr.Range().Start
	end := attrs[updateRequest.AttributeName].Expr.Range().End
	tfBytes, err := ioutil.ReadFile(fileName)
	check(err)
	startBytes := tfBytes[:start.Byte]
	endBytes := tfBytes[end.Byte:]
	newString := string(startBytes) + updateRequest.NewValue + string(endBytes)
	err = ioutil.WriteFile(fileName, []byte(newString), 0644)
	check(err)
	getFlatEarthGraph(w, r)
}

func getProviderSchema(w http.ResponseWriter, r *http.Request) {
	schema, err := commands["provider-schema"].GetProviderSchema(args[0])
	check(err)
	w.Write(schema)
}

func startServer(commands map[string]FlatEarthCommand, args []string) {
	http.HandleFunc("/get-flat-earth-graph", getFlatEarthGraph)
	http.HandleFunc("/update-flat-earth-graph", updateFlatEarthGraph)
	http.HandleFunc("/get-provider-schema", getProviderSchema)
	// kanishk98: if this port is in use, the app just crashes with a garbage error.
	// handle that case early on and either switch to another port (preferably) 
	// or present some usable info to the user
	log.Fatal(http.ListenAndServe(":8080", nil))
}
