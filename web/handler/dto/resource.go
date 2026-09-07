package dto

type CreateResourceRequest struct {
	Name     string         `json:"name" validate:"required,max=128"`
	Provider string         `json:"provider" validate:"required,oneof=docker"`
	Config   map[string]any `json:"config"`
}

type ResourceResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Image    string `json:"image"`
	Provider string `json:"provider"`
	Status   string `json:"status"`
}

type CreateResourceResponse struct {
	Resource ResourceResponse `json:"resource"`
	TaskID   int              `json:"task_id"`
}

type CreateTaskResponse struct {
	TaskID int `json:"task_id"`
}
