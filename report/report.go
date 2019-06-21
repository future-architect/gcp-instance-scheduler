package report

import (
	"log"

	"github.com/future-architect/gcp-instance-scheduler/model"
)

func Show(report *model.ShutdownReport) {
	log.Println("!REPORT!")
	log.Println("[Shutdown Resource]")

	for i, resource := range report.DoneResources {
		log.Printf(">> Resouce(%v): %v\n", i+1, resource)
	}

	log.Println("[Already Shutdown Resource]")
	for i, resource := range report.AlreadyShutdownResources {
		log.Printf(">> Resouce(%v): %v\n", i+1, resource)
	}
}
