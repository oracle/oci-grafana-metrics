package models

import "time"

// MetricsDataRequest - binds the values for metrics query
type MetricsDataRequest struct {
	TenancyOCID     string
	CompartmentOCID string
	CompartmentName string
	Region          string
	Namespace       string
	QueryText       string
	Interval        string
	ResourceGroup   string
	LegendFormat    string
	DimensionValues []string
	TagsValues      []string
	StartTime       time.Time
	EndTime         time.Time
}
