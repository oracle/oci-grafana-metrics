package models

import "time"

// MetricsDataRequest - binds the values for metrics query
type MetricsDataRequest struct {
	TenancyOCID     string
	CompartmentOCID string
	Region          string
	Namespace       string
	QueryText       string
	Interval        string
	ResourceGroup   string
	StartTime       time.Time
	EndTime         time.Time
}
