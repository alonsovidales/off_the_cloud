package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"otc/api"
	"otc/controllers"
	"runtime"
	"syscall"
	"time"
)

func main() {
	// Setting up the logging to a file inside the logs directory
	f, err := os.OpenFile(fmt.Sprintf("logs/%d.log", time.Now().UnixNano()), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// The server can be launched in any port but the default is 8282
	port := flag.Int("p", 8282, "port where the REST API will be listening")

	maxCpus := flag.Int("c", runtime.NumCPU(), "max number of CPUs to be used")
	runtime.GOMAXPROCS(*maxCpus)

	flag.Parse()

	// Launch the server and add the controllers to it
	httpApi := api.GetAPI(*port)
	httpApi.AddController("/data/", controllers.GetDataStorageController())

	go func() {
		err := httpApi.Start()
		if err != nil {
			log.Panicf("Error starting HTTP server: %v", err)
		}
	}()

	// Allows signal interruption for a graceful shutdown of the server
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	sig := <-c
	log.Printf("Signal received: %v, exiting...", sig)
	httpApi.Shutdown()
	os.Exit(0)

}
