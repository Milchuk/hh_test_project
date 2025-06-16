package models

type Goods struct {
	ID          int     `json:"id"`
	ProjectID   int     `json:"projectId"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Priority    int     `json:"priority"`
	Removed     bool    `json:"removed"`
	CreatedAt   string  `json:"createdAt"`
}

type GoodsCreate struct {
	Name string `json:"name" validate:"required,notblank,max=100"`
}

type GoodsUpdate struct {
	Name        *string `json:"name" validate:"notblank,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
	Priority    int     `json:"priority" validate:"notblank"`
	Removed     bool    `json:"removed" validate:"notblank"`
}
