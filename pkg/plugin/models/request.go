/*
** Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
 */

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
