package operator

import (
	"log"
	"time"

	"github.com/future-architect/gcp-instance-scheduler/model"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/net/context"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

type SQLListCall struct {
	InstanceList *sqladmin.InstancesListCall
	ProjectID    string
	Error        error
}

type SQLShutdownCall struct {
	TargetList *sqladmin.InstancesListResponse
	ProjectID  string
	Error      error
}

func SQLResource(ctx context.Context, projectID string) *SQLListCall {
	// reporting error list
	var res error

	// create SQL service
	sqlService, err := sqladmin.NewService(ctx)
	if err != nil {
		res = multierror.Append(res, err)
	}

	// get target SQL instance list
	list := sqladmin.NewInstancesService(sqlService).List(projectID)

	return &SQLListCall{
		InstanceList: list,
		ProjectID:    projectID,
		Error:        res,
	}
}

func (r *SQLListCall) FilterLabel(targetLabel string, flag bool) *SQLShutdownCall {
	// reporting error list
	var res = r.Error

	list, err := r.InstanceList.Filter("userLabels." + targetLabel + "=true").Do()
	if err != nil {
		res = multierror.Append(res, err)
	}

	return &SQLShutdownCall{
		TargetList: list,
		ProjectID:  r.ProjectID,
		Error:      res,
	}
}

func (r *SQLShutdownCall) ShutdownWithInterval(ctx context.Context, interval time.Duration) (*model.ShutdownReport, error) {
	var res = r.Error
	var doneRes []string
	var alreadyRes []string

	// create SQL service
	sqlService, err := sqladmin.NewService(ctx)
	if err != nil {
		res = multierror.Append(res, err)
	}

	for _, instance := range r.TargetList.Items {
		// do not change replica instance's activation policy
		if instance.InstanceType == "READ_REPLICA_INSTANCE" {
			continue
		}

		// do not change instance's activation policy which is already "NEVER"
		if instance.Settings.ActivationPolicy == "NEVER" {
			alreadyRes = append(alreadyRes, instance.Name)
			continue
		}

		// stop the target instance
		instance.Settings.ActivationPolicy = "NEVER"

		// apply the settings
		_, err := sqladmin.NewInstancesService(sqlService).Patch(r.ProjectID, instance.Name, instance).Do()
		if err != nil {
			res = multierror.Append(res, err)
		}
		alreadyRes = append(alreadyRes, instance.Name)
		time.Sleep(interval)
	}
	log.Printf("Success in stopping SQL instances: Done.")

	return &model.ShutdownReport{
		DoneResources:            doneRes,
		AlreadyShutdownResources: alreadyRes,
	}, res
}
