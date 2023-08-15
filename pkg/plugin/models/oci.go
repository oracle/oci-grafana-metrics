/*
** Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
 */

package models

type OCIResource struct {
	Name string `json:"name,omitempty"`
	OCID string `json:"ocid,omitempty"`
}

type OCIMetricNamesWithNamespace struct {
	Namespace   string   `json:"namespace,omitempty"`
	MetricNames []string `json:"metric_names,omitempty"`
}

type OCIMetricNamesWithResourceGroup struct {
	ResourceGroup string   `json:"resource_group,omitempty"`
	MetricNames   []string `json:"metric_names,omitempty"`
}

type OCIMetricDimensions struct {
	Key    string   `json:"key,omitempty"`
	Values []string `json:"values,omitempty"`
}

type OCIResourceTags struct {
	Key    string   `json:"key,omitempty"`
	Values []string `json:"values,omitempty"`
}

type OCIMetricDataPoints struct {
	TenancyName     string
	CompartmentName string
	Region          string
	MetricName      string
	ResourceName    string
	UniqueDataID    string
	DataPoints      []float64
	Labels          map[string]string
}

type OCIResourceTagsResponse struct {
	ResourceID   string
	ResourceName string
	DefinedTags  map[string]map[string]interface{}
	FreeFormTags map[string]string
}
