package main

import (
	"mi-ca/internal/controller"
	"mi-ca/internal/env"
	"mi-ca/internal/kubernetes"
	"mi-ca/internal/logging"
	"os"
	"sync"
)

func main() {
	if !env.PopulateEnvironment() {
		os.Exit(1)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	logging.Init()

	kubernetes.Connect()

	go controller.Connect()
	wg.Wait()
}
