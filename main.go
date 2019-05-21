package main

import (
	"log"

	"github.com/bitrise-io/bitrise-add-new-project/phases"
)

func main() {
	DSL, workflow, err := phases.BitriseYML(".")
	log.Printf("%+v \n %s \n %s", DSL, workflow, err)
	if err != nil {
		panic(err)
	}
	// cmd.Execute()
}
