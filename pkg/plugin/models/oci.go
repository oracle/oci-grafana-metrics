/*
** Copyright © 2023 Oracle and/or its affiliates. All rights reserved.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
 */

package models

// OCIResource represents a generic OCI resource with a name and OCID.
type OCIResource struct {
	// Name is the display name of the OCI resource.
	Name string `json:"name,omitempty"`
	// OCID is the Oracle Cloud Identifier of the OCI resource.
	OCID string `json:"ocid,omitempty"`
}

// OCIMetricNamesWithNamespace represents a namespace and its associated metric names.
type OCIMetricNamesWithNamespace struct {
	// Namespace is the OCI namespace.
	Namespace string `json:"namespace,omitempty"`
	// MetricNames is a list of metric names within the namespace.
	MetricNames []string `json:"metric_names,omitempty"`
}

// OCIMetricNamesWithResourceGroup represents a resource group and its associated metric names.
type OCIMetricNamesWithResourceGroup struct {
	// ResourceGroup is the name of the OCI resource group.
	ResourceGroup string `json:"resource_group,omitempty"`
	// MetricNames is a list of metric names within the resource group.
	MetricNames []string `json:"metric_names,omitempty"`
}

// OCIMetricDimensions represents a dimension key and its possible values.
type OCIMetricDimensions struct {
	// Key is the dimension key.
	Key string `json:"key,omitempty"`
	// Values is a list of possible values for the dimension key.
	Values []string `json:"values,omitempty"`
}

// OCIResourceTags represents a tag key and its associated values.
type OCIResourceTags struct {
	// Key is the tag key.
	Key string `json:"key,omitempty"`
	// Values is a list of possible values for the tag key.
	Values []string `json:"values,omitempty"`
}

// OCIMetricDataPoints represents a set of data points for a metric, along with associated metadata.
type OCIMetricDataPoints struct {
	// TenancyName is the name of the tenancy.
	TenancyName string
	// CompartmentName is the name of the compartment.
	CompartmentName string
	// Region is the OCI region.
	Region string
	// MetricName is the name of the metric.
	MetricName string
	// ResourceName is the name of the resource associated with the metric.
	ResourceName string
	// UniqueDataID is a unique identifier for the data.
	UniqueDataID string
	// DimensionKey is the key of the dimension used to identify the resource.
	DimensionKey string
	// DataPoints is a list of float64 values representing the metric data points.
	DataPoints []float64
	// Labels is a map of string to string representing the labels for the metric data.
	Labels map[string]string
}

// OCIResourceTagsResponse represents the response structure for OCI resource tags.
type OCIResourceTagsResponse struct {
	// ResourceID is the OCID of the resource.
	ResourceID string
	// ResourceName is the display name of the resource.
	ResourceName string
	// DefinedTags is a map of defined tag namespaces to their corresponding key-value pairs.
	DefinedTags map[string]map[string]interface{}
	// FreeFormTags is a map of free-form tag keys to their values.
	FreeFormTags map[string]string
}
