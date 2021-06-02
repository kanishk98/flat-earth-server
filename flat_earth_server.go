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
	NewValue      []byte
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func generateFlatEarthGraph() []byte {
	graph, err := commands["flat-earth"].Run(args[0])

	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to generate FlatEarthGraph: %s", err))
		panic(err)
	}

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
	if err != nil {
		panic(err)
	}
	log.Println(string(body))
	var updateRequest FlatEarthGraphUpdateRequest
	err = json.Unmarshal(body, &updateRequest)
	if err != nil {
		panic(err)
	}
	graph, _ := commands["flat-earth"].Run(args[0])
	fileName := graph[updateRequest.NodeKey].DeclRange.Filename
	// kanishk98: Possibilities of the key not getting found here, handle that
	attrs, _ := graph[updateRequest.NodeKey].Config.JustAttributes()
	start := attrs[updateRequest.AttributeName].Expr.Range().Start
	end := attrs[updateRequest.AttributeName].Expr.Range().End
	tfBytes, err := ioutil.ReadFile(fileName)
	check(err)
	newBytes := append(tfBytes[start.Byte+1:], updateRequest.NewValue...)
	newBytes = append(newBytes, tfBytes[end.Byte+1:]...)
	err = ioutil.WriteFile("./stuff.tf", newBytes, 0644)
	check(err)
	getFlatEarthGraph(w, r)
}

func startServer(commands map[string]FlatEarthCommand, args []string) {
	http.HandleFunc("/get-flat-earth-graph", getFlatEarthGraph)
	http.HandleFunc("/update-flat-earth-graph", updateFlatEarthGraph)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
