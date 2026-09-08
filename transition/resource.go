package transition

import (
	"devcloud/ent/resource"
	"fmt"
)

type ResourceEvent string

const (
	ResourceCreated      ResourceEvent = "created"
	ResourceCreateFailed ResourceEvent = "create_failed"

	ResourceStart   ResourceEvent = "start"
	ResourceStarted ResourceEvent = "started"

	ResourceStop    ResourceEvent = "stop"
	ResourceStopped ResourceEvent = "stopped"

	ResourceRestart   ResourceEvent = "restart"
	ResourceRestarted ResourceEvent = "restarted"

	ResourceRemove  ResourceEvent = "remove"
	ResourceRemoved ResourceEvent = "removed"
)

var resourceTransition = map[struct {
	from  resource.Status
	event ResourceEvent
}]resource.Status{
	// Create
	{resource.StatusCREATING, ResourceCreated}:      resource.StatusRUNNING,
	{resource.StatusCREATING, ResourceCreateFailed}: resource.StatusFAILED,
	// Restart
	{resource.StatusRUNNING, ResourceRestart}:      resource.StatusRESTARTING,
	{resource.StatusRESTARTING, ResourceRestarted}: resource.StatusRUNNING,
	// Stop
	{resource.StatusRUNNING, ResourceStop}:     resource.StatusSTOPPING,
	{resource.StatusSTOPPING, ResourceStopped}: resource.StatusSTOPPED,
	// Start
	{resource.StatusSTOPPED, ResourceStart}:    resource.StatusSTARTING,
	{resource.StatusSTARTING, ResourceStarted}: resource.StatusRUNNING,
	// remove
	{resource.StatusSTOPPED, ResourceRemove}:   resource.StatusREMOVING,
	{resource.StatusRUNNING, ResourceRemove}:   resource.StatusREMOVING,
	{resource.StatusREMOVING, ResourceRemoved}: resource.StatusREMOVED,
}

func Resource(from resource.Status, event ResourceEvent) (resource.Status, error) {
	next, ok := resourceTransition[struct {
		from  resource.Status
		event ResourceEvent
	}{from: from, event: event}]
	if !ok {
		return from, fmt.Errorf("invalid resource transition: status=%s event=%s", from, event)
	}
	return next, nil
}
