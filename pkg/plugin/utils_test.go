package plugin

import (
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

// Regression tests for https://github.com/oracle/oci-grafana-metrics/issues/370
func TestGenerateCustomMetricLabel(t *testing.T) {
	o := &OCIDatasource{logger: log.DefaultLogger}

	dims := map[string]string{
		"resourceId":   "ocid1.dataflowrun.oc1.iad.aaaa",
		"resourceName": "cluster-a",
		"executorId":   "3",
		"region":       "us-ashburn-1",
	}
	pluginLabels := map[string]string{
		"tenancy":   "DEFAULT",
		"unique_id": "ocid1.vnic.oc1.iad.bbbb",
		"region":    "plugin-region",
	}

	tests := []struct {
		name     string
		format   string
		dims     map[string]string
		labels   map[string]string
		want     string
		resolved bool
	}{
		{"multiple dimensions", "Cluster: {{resourceName}} - Executor: {{executorId}}", dims, nil, "Cluster: cluster-a - Executor: 3", true},
		{"metric placeholder", "{{metric}} on {{resourceName}}", dims, nil, "CpuUtilization on cluster-a", true},
		{"whitespace in placeholder", "{{ resourceName }}", dims, nil, "cluster-a", true},
		{"trailing whitespace only", "{{resourceName }}", dims, nil, "cluster-a", true},
		{"plugin label unique_id", "{{ unique_id }}", map[string]string{}, pluginLabels, "ocid1.vnic.oc1.iad.bbbb", true},
		{"plugin label tenancy", "{{tenancy}}/{{resourceName}}", dims, pluginLabels, "DEFAULT/cluster-a", true},
		{"dimension wins over plugin label", "{{region}}", dims, pluginLabels, "us-ashburn-1", true},
		{"unknown placeholder is reported", "{{resourceName}} {{nope}}", dims, pluginLabels, "cluster-a {{nope}}", false},
		{"missing dimension in namespace", "{{resourceDisplayName }}", map[string]string{"resourceId": "x"}, pluginLabels, "{{resourceDisplayName }}", false},
		{"nil maps", "{{metric}}", nil, nil, "CpuUtilization", true},
		{"regex metacharacters in value", "{{resourceName}}", map[string]string{"resourceName": "a$1.b"}, nil, "a$1.b", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := o.generateCustomMetricLabel(tt.format, "CpuUtilization", tt.dims, tt.labels)
			if got != tt.want || ok != tt.resolved {
				t.Errorf("got (%q, %v), want (%q, %v)", got, ok, tt.want, tt.resolved)
			}
		})
	}
}
