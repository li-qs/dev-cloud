// Package apperr 定义跨层共享的业务错误，避免 repo 反向依赖 web 层。
package apperr

import "errors"

var (
	// ErrNoPendingTask 表示当前没有可领取的任务，属于正常空闲状态。
	ErrNoPendingTask = errors.New("no pending task")

	// ErrInvalidState 表示资源当前状态不允许该操作。
	ErrInvalidState = errors.New("invalid resource state")

	// ErrConflict 表示资源状态在读取后被并发修改，操作需要重试。
	ErrConflict = errors.New("resource state conflict")
)
