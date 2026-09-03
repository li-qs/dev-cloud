package dto

type CreateResourceRequest struct {
	Name     string         `json:"name" validate:"required,max=128"`
	Provider string         `json:"provider" validate:"required,oneof=postgres redis nginx"`
	Config   map[string]any `json:"config"`
}

type ResourceResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Provider string `json:"provider"`
	Status   string `json:"status"`
}

type CreateResourceResponse struct {
	Resource ResourceResponse `json:"resource"`
	TaskID   int              `json:"task_id"`
}
