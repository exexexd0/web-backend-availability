package handler

import "github.com/one-compressive/web-backend-availability/internal/app/ds"

type componentListResponse struct {
	Status string                       `json:"status"`
	Data   []ds.ComponentListSerializer `json:"data"`
}

type componentDetailResponse struct {
	Status string                       `json:"status"`
	Data   ds.ComponentDetailSerializer `json:"data"`
}

type componentFeedResponse struct {
	Status string                     `json:"status"`
	Data   ds.ComponentFeedSerializer `json:"data"`
}

type publishRequest struct {
	UptimePercent float64 `json:"uptime_percent"`
	SystemImpact  float64 `json:"system_impact"`
}

type likeRequest struct {
	Value int `json:"value"`
}

type deleteData struct {
	ID uint `json:"id"`
}

type deleteResponse struct {
	Status string     `json:"status"`
	Data   deleteData `json:"data"`
}

type likeData struct {
	ComponentID uint `json:"component_id"`
	Value       int  `json:"value"`
}

type likeResponse struct {
	Status string   `json:"status"`
	Data   likeData `json:"data"`
}
