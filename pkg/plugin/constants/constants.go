package constants

import "time"

const (
	OCI_CLI_AUTH_PROVIDER               = "oci-cli"
	OCI_INSTANCE_AUTH_PROVIDER          = "oci-instance"
	DEFAULT_PROFILE                     = "DEFAULT/"
	DEFAULT_INSTANCE_PROFILE            = "instance_profile"
	DEFAULT_RESOURCE_GROUP              = "NoResourceGroup"
	DEFAULT_RESOURCE_PLACEHOLDER        = "select resourcegroup (optional)"
	DEFAULT_RESOURCE_PLACEHOLDER_LEGACY = "select resource group"
	DEFAULT_COMPARTMENT_PLACEHOLDER     = "select compartment"
	MULTI_TENANCY_MODE_PROFILE          = "multi-profile"
	MULTI_TENANCY_MODE_POLICY           = "cross-tenancy-policy"
	YES                                 = "yes"
	NO                                  = "no"
	QUERYTYPE_TENANCIES                 = "tenancies"
	QUERYTYPE_REGIONS                   = "regions"
	QUERYTYPE_COMPARTMENTS              = "compartments"
	QUERYTYPE_NAMESPACES_WITH_METRICS   = "namespaces_with_metrics"
	QUERYTYPE_METRICS_SUMMARY           = "metrics_summary"
	CACHE_KEY_RESOURCE_TAGS             = "resourceTags"
	CACHE_KEY_RESOURCE_IDS_PER_TAG      = "resourceIDsPerTag"
	ALL_REGION                          = "all-subscribed-region"
	ALL_COMPARTMENT                     = "all-compartment"
	FETCH_FOR_NAMESPACE                 = "namespace"
	FETCH_FOR_RESOURCE_GROUP            = "resource-group"
	FETCH_FOR_DIMENSION                 = "dimension"
	FETCH_FOR_LABELDIMENSION            = "labeldimension"
	TIME_IN_MINUTES                     = 5 * time.Minute
	OCI_TARGET_COMPUTE                  = "compute"
	OCI_TARGET_VCN                      = "vcn"
	OCI_TARGET_LBAAS                    = "lbaas"
	OCI_TARGET_HEALTHCHECK              = "healthchecks"
	OCI_TARGET_DATABASE                 = "database"
	OCI_TARGET_APM                      = "apm"
	OCI_NS_APM                          = "oracle_apm_synthetics"
	OCI_NS_DB_ORACLE                    = "oracle_oci_database"
	OCI_NS_DB_EXTERNAL                  = "oracle_external_database"
	OCI_NS_DB_AUTONOMOUS                = "oci_autonomous_database"
)

var (
	OCI_NAMESPACES = map[string]string{
		"oci_computeagent":                  OCI_TARGET_COMPUTE,
		"oci_compute":                       OCI_TARGET_COMPUTE,
		"oci_compute_infrastructure_health": OCI_TARGET_COMPUTE,
		"oci_vcn":                           OCI_TARGET_VCN,
		"oci_lbaas":                         OCI_TARGET_LBAAS,
		"oci_healthchecks":                  OCI_TARGET_HEALTHCHECK,
		"oci_autonomous_database":           OCI_TARGET_DATABASE,
		"oracle_oci_database":               OCI_TARGET_DATABASE,
		"oracle_external_database":          OCI_TARGET_DATABASE,
		"oracle_apm_synthetics":             OCI_TARGET_APM,
	}
)
