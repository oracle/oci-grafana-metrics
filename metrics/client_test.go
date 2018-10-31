package metrics

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/oracle/oci-go-sdk/common"
)

var compartment = "ocid1.compartment.oc1..aaaaaaaa6xkpezjs764l2bhm7dxmd2cliykmuomd26tuwttbl7xifuliywqq"

func TestSummarizeMetrics(t *testing.T) {
	configProvider := common.DefaultConfigProvider()
	c, err := NewTelemetryClientWithConfigurationProvider(configProvider)
	if err != nil {
		log.Printf("error with client")
		t.Error(err)
	}
	c.SetRegion("us-ashburn-1")
	end := time.Now().UTC()
	start := end.Add(time.Hour * -1).UTC()
	start = start.Truncate(time.Millisecond)
	end = end.Truncate(time.Millisecond)

	body := SummarizeMetricsDataDetails{
		Query:      common.String("DiskIopsWritten[5m].sum()"),
		Namespace:  common.String("oci_computeagent"),
		StartTime:  &common.SDKTime{start},
		EndTime:    &common.SDKTime{end},
		Resolution: common.String("5m"),
	}

	request := SummarizeMetricsDataRequest{
		CompartmentId:               common.String(compartment),
		SummarizeMetricsDataDetails: body,
	}

	res, err := c.SummarizeMetricsData(context.TODO(), request)
	fmt.Println(spew.Sdump(res))
	if err != nil {
		log.Printf("error with request")
		t.Error(err)
	}

}

func TestListMetrics(t *testing.T) {
	configProvider := common.DefaultConfigProvider()
	c, err := NewTelemetryClientWithConfigurationProvider(configProvider)
	if err != nil {
		log.Printf("error with client")
		t.Error(err)
	}
	c.SetRegion("us-ashburn-1")

	reqDetails := ListMetricsDetails{
		Namespace: common.String("oci_lbaas"),
	}
	listMetrics := ListMetricsRequest{
		CompartmentId:      common.String(compartment),
		ListMetricsDetails: reqDetails,
	}
	resp, e := c.ListMetrics(context.TODO(), listMetrics)
	fmt.Println(spew.Sdump(resp))
	if e != nil {
		t.Error(e)
	}
}
