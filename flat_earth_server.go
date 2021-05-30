package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

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

func startServer(commands map[string]FlatEarthCommand, args []string) {
	http.HandleFunc("/flat-earth-graph", getFlatEarthGraph)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
