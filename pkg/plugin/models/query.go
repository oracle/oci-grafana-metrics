/*
** Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
 */

package models

// QueryModel ...
type QueryModel struct {
	QueryText       string   `json:"queryText"`
	TenancyName     string   `json:"tenancyName"`
	TenancyOCID     string   `json:"tenancy"`
	CompartmentName string   `json:"compartmentName"`
	CompartmentOCID string   `json:"compartment"`
	Region          string   `json:"region"`
	Namespace       string   `json:"namespace"`
	Metric          string   `json:"metric"`
	RawQuery        bool     `json:"rawQuery"`
	Interval        string   `json:"interval"`
	Statistic       string   `json:"statistic"`
	LegendFormat    string   `json:"legendFormat"`
	ResourceGroup   string   `json:"resourcegroup,omitempty"`
	DimensionValues []string `json:"dimensionValues,omitempty"`
	TagsValues      []string `json:"tagsValues,omitempty"`
}
