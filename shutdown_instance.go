package function

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/future-architect/gcp-instance-scheduler/operator"
	"github.com/future-architect/gcp-instance-scheduler/report"

	"cloud.google.com/go/pubsub"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/net/context"
)

// Operation target label name
const TargetLabel = "state-scheduler"

// API call interval
const ShutdownInterval = 50 * time.Millisecond

type SubscribedMessage struct {
	Command string `json:"command"`
}

func ReceiveEvent(ctx context.Context, msg *pubsub.Message) error {

	projectID := os.Getenv("GCP_PROJECT")
	log.Printf("Project ID: %v", projectID)

	// decode the json message from Pub/Sub
	message, err := decode(msg.Data)
	if err != nil {
		log.Printf("Error at the fucntion 'DecodeMessage': %v", err)
	}
	log.Printf("Subscribed message(Command): %v", message.Command)

	// for multierror
	var result error

	if err := operator.SetLabelNodePoolSize(ctx, projectID, TargetLabel, ShutdownInterval); err != nil {
		result = multierror.Append(result, err)
		log.Printf("Error in setting labels on GKE cluster: %v", err)
	}

	// show cluster status
	if err := operator.ShowClusterStatus(ctx, projectID, TargetLabel); err != nil {
		result = multierror.Append(result, err)
		log.Printf("Error in stopping GKE: %v", err)
	}

	rpt, err := operator.InstanceGroupResource(ctx, projectID).FilterLabel(TargetLabel, true).ShutdownWithInterval(ctx, ShutdownInterval)
	if err != nil {
		result = multierror.Append(result, err)
		log.Printf("Some error occured in stopping gce instances: %v", err)
	}
	report.Show(rpt)

	rpt, err = operator.ComputeEngineResource(ctx, projectID).FilterLabel(TargetLabel, true).ShutdownWithInterval(ctx, ShutdownInterval)
	if err != nil {
		result = multierror.Append(result, err)
		log.Printf("Some error occured in stopping gce instances: %v", err)
	}
	report.Show(rpt)

	rpt, err = operator.SQLResource(ctx, projectID).FilterLabel(TargetLabel, true).ShutdownWithInterval(ctx, ShutdownInterval)
	if err != nil {
		result = multierror.Append(result, err)
		log.Printf("Some error occured in stopping sql instances: %v", err)
	}
	report.Show(rpt)

	log.Printf("done.")
	return result
}

func decode(payload []byte) (msgData SubscribedMessage, err error) {
	if err = json.Unmarshal(payload, &msgData); err != nil {
		log.Printf("Message[%v] ... Could not decode subscribing data: %v", payload, err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		return
	}
	return
}
