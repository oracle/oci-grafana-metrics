package models

// QueryModel ...
type QueryModel struct {
	QueryText         string   `json:"queryText"`
	TenancyName       string   `json:"tenancyName"`
	TenancyOCID       string   `json:"tenancyOCID"`
	TenancyLegacy     string   `json:"tenancy"`
	CompartmentName   string   `json:"compartmentName"`
	CompartmentOCID   string   `json:"compartmentOCID"`
	CompartmentLegacy string   `json:"compartment"`
	Region            string   `json:"region"`
	Namespace         string   `json:"namespace"`
	Metric            string   `json:"metric"`
	Interval          string   `json:"interval"`
	Statistic         string   `json:"statistic"`
	LegendFormat      string   `json:"legendFormat"`
	ResourceGroup     string   `json:"resourceGroup,omitempty"`
	DimensionValues   []string `json:"dimensionValues,omitempty"`
	TagsValues        []string `json:"tagsValues,omitempty"`
}
