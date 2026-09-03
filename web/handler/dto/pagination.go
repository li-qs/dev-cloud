package dto

type PaginationRequest struct {
	Page     int `query:"page" validate:"min=1"`
	PageSize int `query:"page_size" validate:"min=1,max=100"`
}

func (p *PaginationRequest) SetDefaults() {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.PageSize == 0 {
		p.PageSize = 20
	}
}
