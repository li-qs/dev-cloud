package transition

import (
	"devcloud/ent/resource"
	"fmt"
)

// ResourceEvent 资源状态机的事件。
//
// 状态机采用「两段式」约定：
//   - 进入事件（create/start/stop/restart/remove）：由 API 入队时触发，把资源推进到中间态；
//   - 完成事件（*_success / *_failed）：由 worker 执行结束后触发，把资源推进到终态或回滚。
type ResourceEvent string

// 资源状态迁移事件。无后缀为「进入中间态」事件，success/failed 为「执行结果」事件。
const (
	ResourceCreate        ResourceEvent = "create"
	ResourceCreateSuccess ResourceEvent = "create_success"
	ResourceCreateFailed  ResourceEvent = "create_failed"

	ResourceStart        ResourceEvent = "start"
	ResourceStartSuccess ResourceEvent = "start_success"
	ResourceStartFailed  ResourceEvent = "start_failed"

	ResourceRestart        ResourceEvent = "restart"
	ResourceRestartSuccess ResourceEvent = "restart_success"
	ResourceRestartFailed  ResourceEvent = "restart_failed"

	ResourceStop        ResourceEvent = "stop"
	ResourceStopSuccess ResourceEvent = "stop_success"
	ResourceStopFailed  ResourceEvent = "stop_failed"

	ResourceRemove        ResourceEvent = "remove"
	ResourceRemoveSuccess ResourceEvent = "remove_success"
	ResourceRemoveFailed  ResourceEvent = "remove_failed"
)

// resourceTransitions 资源状态迁移表：key 为 (当前状态, 事件)，value 为目标状态。
// 未出现在表中的组合即非法迁移。"" 作为 create 的「无前置状态」哨兵。
var resourceTransitions = map[struct {
	current resource.Status
	event   ResourceEvent
}]resource.Status{
	// Create
	{"", ResourceCreate}:                             resource.StatusCREATING,
	{resource.StatusCREATING, ResourceCreateSuccess}: resource.StatusSTOPPED,
	{resource.StatusCREATING, ResourceCreateFailed}:  resource.StatusFAILED,
	// Start
	{resource.StatusSTOPPED, ResourceStart}:         resource.StatusSTARTING,
	{resource.StatusSTARTING, ResourceStartSuccess}: resource.StatusRUNNING,
	{resource.StatusSTARTING, ResourceStartFailed}:  resource.StatusSTOPPED,
	// Restart
	{resource.StatusRUNNING, ResourceRestart}:           resource.StatusRESTARTING,
	{resource.StatusRESTARTING, ResourceRestartSuccess}: resource.StatusRUNNING,
	{resource.StatusRESTARTING, ResourceRestartFailed}:  resource.StatusRUNNING,
	// Stop
	{resource.StatusRUNNING, ResourceStop}:         resource.StatusSTOPPING,
	{resource.StatusSTOPPING, ResourceStopSuccess}: resource.StatusSTOPPED,
	{resource.StatusSTOPPING, ResourceStopFailed}:  resource.StatusRUNNING,
	// Remove
	{resource.StatusSTOPPED, ResourceRemove}:         resource.StatusREMOVING,
	{resource.StatusFAILED, ResourceRemove}:          resource.StatusREMOVING,
	{resource.StatusREMOVING, ResourceRemoveSuccess}: resource.StatusREMOVED,
	{resource.StatusREMOVING, ResourceRemoveFailed}:  resource.StatusSTOPPED,
}

// NextResourceStatus 根据当前状态与事件计算下一个状态；
// 迁移非法（表中无该组合）时返回错误，调用方不应写库。
func NextResourceStatus(current resource.Status, event ResourceEvent) (resource.Status, error) {
	next, ok := resourceTransitions[struct {
		current resource.Status
		event   ResourceEvent
	}{
		current: current,
		event:   event,
	}]
	if !ok {
		return current, fmt.Errorf("invalid resource transition: status=%s event=%s", current, event)
	}
	return next, nil
}
