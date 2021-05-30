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

func getFlatEarthGraph(w http.ResponseWriter, r *http.Request) {
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
	// carry out update magic here, then re-generate graph and return it
	getFlatEarthGraph(w, r)
}

func startServer(commands map[string]FlatEarthCommand, args []string) {
	http.HandleFunc("/get-flat-earth-graph", getFlatEarthGraph)
	http.HandleFunc("/update-flat-earth-graph", updateFlatEarthGraph)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
