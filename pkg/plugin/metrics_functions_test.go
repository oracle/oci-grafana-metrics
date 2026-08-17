// Copyright © 2026 Oracle and/or its affiliates. All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package plugin

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

type testServiceError struct {
	status int
	code   string
}

func (e testServiceError) Error() string          { return e.code }
func (e testServiceError) GetHTTPStatusCode() int { return e.status }
func (e testServiceError) GetMessage() string     { return e.code }
func (e testServiceError) GetCode() string        { return e.code }
func (e testServiceError) GetOpcRequestID() string {
	return "test-request"
}

func compartment(id string) identity.Compartment {
	return identity.Compartment{Id: common.String(id)}
}

func TestListCompartmentsSinglePageOmitsPage(t *testing.T) {
	calls := 0
	got, err := listCompartments(context.Background(), identity.ListCompartmentsRequest{}, func(
		_ context.Context,
		request identity.ListCompartmentsRequest,
	) (identity.ListCompartmentsResponse, error) {
		calls++
		if request.Page != nil {
			t.Fatalf("initial Page = %q, want nil", *request.Page)
		}
		return identity.ListCompartmentsResponse{Items: []identity.Compartment{compartment("one")}}, nil
	})

	if err != nil {
		t.Fatalf("listCompartments returned error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
	if len(got) != 1 || *got[0].Id != "one" {
		t.Fatalf("compartments = %v, want [one]", got)
	}
}

func TestListCompartmentsUsesNextPageToken(t *testing.T) {
	calls := 0
	got, err := listCompartments(context.Background(), identity.ListCompartmentsRequest{}, func(
		_ context.Context,
		request identity.ListCompartmentsRequest,
	) (identity.ListCompartmentsResponse, error) {
		calls++
		switch calls {
		case 1:
			if request.Page != nil {
				t.Fatalf("initial Page = %q, want nil", *request.Page)
			}
			return identity.ListCompartmentsResponse{
				Items:       []identity.Compartment{compartment("one")},
				OpcNextPage: common.String("next-page"),
			}, nil
		case 2:
			if request.Page == nil || *request.Page != "next-page" {
				t.Fatalf("second Page = %v, want next-page", request.Page)
			}
			return identity.ListCompartmentsResponse{Items: []identity.Compartment{compartment("two")}}, nil
		default:
			t.Fatalf("unexpected call %d", calls)
			return identity.ListCompartmentsResponse{}, nil
		}
	})

	if err != nil {
		t.Fatalf("listCompartments returned error: %v", err)
	}
	if len(got) != 2 || *got[0].Id != "one" || *got[1].Id != "two" {
		t.Fatalf("compartments = %v, want [one two]", got)
	}
}

func TestListCompartmentsStopsOnEmptyNextPage(t *testing.T) {
	calls := 0
	got, err := listCompartments(context.Background(), identity.ListCompartmentsRequest{}, func(
		_ context.Context,
		_ identity.ListCompartmentsRequest,
	) (identity.ListCompartmentsResponse, error) {
		calls++
		return identity.ListCompartmentsResponse{
			Items:       []identity.Compartment{compartment("one")},
			OpcNextPage: common.String(""),
		}, nil
	})

	if err != nil {
		t.Fatalf("listCompartments returned error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
	if len(got) != 1 || *got[0].Id != "one" {
		t.Fatalf("compartments = %v, want [one]", got)
	}
}

func TestListCompartmentsRestartsOnceForInvalidToken(t *testing.T) {
	calls := 0
	got, err := listCompartments(context.Background(), identity.ListCompartmentsRequest{}, func(
		_ context.Context,
		request identity.ListCompartmentsRequest,
	) (identity.ListCompartmentsResponse, error) {
		calls++
		switch calls {
		case 1:
			return identity.ListCompartmentsResponse{
				Items:       []identity.Compartment{compartment("discarded")},
				OpcNextPage: common.String("stale-token"),
			}, nil
		case 2:
			if request.Page == nil || *request.Page != "stale-token" {
				t.Fatalf("second Page = %v, want stale-token", request.Page)
			}
			return identity.ListCompartmentsResponse{}, testServiceError{status: 400, code: "InvalidPaginationToken"}
		case 3:
			if request.Page != nil {
				t.Fatalf("restart Page = %q, want nil", *request.Page)
			}
			return identity.ListCompartmentsResponse{
				Items:       []identity.Compartment{compartment("one")},
				OpcNextPage: common.String("fresh-token"),
			}, nil
		case 4:
			if request.Page == nil || *request.Page != "fresh-token" {
				t.Fatalf("fourth Page = %v, want fresh-token", request.Page)
			}
			return identity.ListCompartmentsResponse{Items: []identity.Compartment{compartment("two")}}, nil
		default:
			t.Fatalf("unexpected call %d", calls)
			return identity.ListCompartmentsResponse{}, nil
		}
	})

	if err != nil {
		t.Fatalf("listCompartments returned error: %v", err)
	}
	if len(got) != 2 || *got[0].Id != "one" || *got[1].Id != "two" {
		t.Fatalf("compartments = %v, want [one two]", got)
	}
}

func TestListCompartmentsDoesNotRestartTwice(t *testing.T) {
	calls := 0
	got, err := listCompartments(context.Background(), identity.ListCompartmentsRequest{}, func(
		_ context.Context,
		request identity.ListCompartmentsRequest,
	) (identity.ListCompartmentsResponse, error) {
		calls++
		if calls == 1 || calls == 3 {
			return identity.ListCompartmentsResponse{
				Items:       []identity.Compartment{compartment("partial")},
				OpcNextPage: common.String("token"),
			}, nil
		}
		if request.Page == nil || *request.Page != "token" {
			t.Fatalf("Page = %v, want token", request.Page)
		}
		return identity.ListCompartmentsResponse{}, testServiceError{status: 400, code: "InvalidParameter"}
	})

	if err == nil {
		t.Fatal("listCompartments returned nil error")
	}
	if got != nil {
		t.Fatalf("compartments = %v, want nil", got)
	}
	if calls != 4 {
		t.Fatalf("calls = %d, want 4", calls)
	}
}

func TestListCompartmentsDoesNotRetryInitialInvalidParameter(t *testing.T) {
	calls := 0
	got, err := listCompartments(context.Background(), identity.ListCompartmentsRequest{}, func(
		_ context.Context,
		request identity.ListCompartmentsRequest,
	) (identity.ListCompartmentsResponse, error) {
		calls++
		if request.Page != nil {
			t.Fatalf("initial Page = %q, want nil", *request.Page)
		}
		return identity.ListCompartmentsResponse{}, testServiceError{status: 400, code: "InvalidParameter"}
	})

	if err == nil {
		t.Fatal("listCompartments returned nil error")
	}
	if got != nil {
		t.Fatalf("compartments = %v, want nil", got)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestListCompartmentsReturnsOtherErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "network", err: errors.New("network failure")},
		{name: "service", err: testServiceError{status: 500, code: "InternalError"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			got, err := listCompartments(context.Background(), identity.ListCompartmentsRequest{}, func(
				_ context.Context,
				_ identity.ListCompartmentsRequest,
			) (identity.ListCompartmentsResponse, error) {
				calls++
				return identity.ListCompartmentsResponse{}, test.err
			})

			if err == nil {
				t.Fatal("listCompartments returned nil error")
			}
			if got != nil {
				t.Fatalf("compartments = %v, want nil", got)
			}
			if calls != 1 {
				t.Fatalf("calls = %d, want 1", calls)
			}
		})
	}
}

// sdkTimeAt returns a common.SDKTime n minutes after a fixed base time.
func sdkTimeAt(base time.Time, minutes int) common.SDKTime {
	return common.SDKTime{Time: base.Add(time.Duration(minutes) * time.Minute)}
}

// assertAlignedSeries checks that aligned[idx] contains exactly the expected
// nullable values: the series' own value where present, nil where OCI did not
// report a datapoint for it.
func assertAlignedSeries(t *testing.T, aligned [][]*float64, idx int, label string, expected []*float64) {
	t.Helper()
	got := aligned[idx]
	if len(got) != len(expected) {
		t.Fatalf("series %s: length = %d, want %d", label, len(got), len(expected))
	}
	for j := range expected {
		if expected[j] == nil {
			if got[j] != nil {
				t.Fatalf("series %s: datapoint %d = %v, want nil (no datapoint of its own at this timestamp)", label, j, *got[j])
			}
			continue
		}
		if got[j] == nil {
			t.Fatalf("series %s: datapoint %d = nil, want %v", label, j, *expected[j])
		}
		if *got[j] != *expected[j] {
			t.Fatalf("series %s: datapoint %d = %v, want %v", label, j, *got[j], *expected[j])
		}
	}
}

func float64Ptr(v float64) *float64 { return &v }

// TestAlignDataPointsToUnionTimestamps reproduces the production data-corruption
// failure from oracle/oci-grafana-metrics#285/#323: when series report
// datapoints at different timestamps (e.g. OCI object storage only emits
// datapoints for buckets with traffic), each series must keep its own values
// and get nulls — never another series' values — at timestamps it did not
// report.
func TestAlignDataPointsToUnionTimestamps(t *testing.T) {
	base := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)

	// t1..t6
	union := []common.SDKTime{
		sdkTimeAt(base, 1), sdkTimeAt(base, 2), sdkTimeAt(base, 3),
		sdkTimeAt(base, 4), sdkTimeAt(base, 5), sdkTimeAt(base, 6),
	}

	// Series A is dense (t1..t6), series B is sparse (only t2, t5),
	// series C is sparse (only t3).
	dataValuesPerSeries := map[int]map[common.SDKTime]float64{
		0: {
			union[0]: 1, union[1]: 2, union[2]: 3,
			union[3]: 4, union[4]: 5, union[5]: 6,
		},
		1: {
			union[1]: 20, union[4]: 50,
		},
		2: {
			union[2]: 30,
		},
	}

	aligned := alignDataPointsToUnionTimestamps(3, dataValuesPerSeries, union)

	if len(aligned) != 3 {
		t.Fatalf("aligned series count = %d, want 3", len(aligned))
	}

	// Series A keeps exactly its own six values.
	assertAlignedSeries(t, aligned, 0, "A", []*float64{
		float64Ptr(1), float64Ptr(2), float64Ptr(3),
		float64Ptr(4), float64Ptr(5), float64Ptr(6),
	})

	// Series B has values only at t2 and t5, nil everywhere else. Under the
	// old positional realignment B would have been padded with A's values;
	// these nil assertions catch that regression.
	assertAlignedSeries(t, aligned, 1, "B", []*float64{
		nil, float64Ptr(20), nil, nil, float64Ptr(50), nil,
	})

	// Series C has a value only at t3.
	assertAlignedSeries(t, aligned, 2, "C", []*float64{
		nil, nil, float64Ptr(30), nil, nil, nil,
	})

	// No value ever crosses series: every populated cell equals the value the
	// series itself reported, not a neighbor's (e.g. B at t2 must be 20, not
	// A's 2).
	for i, want := range dataValuesPerSeries {
		for ts, v := range want {
			var j int
			for j = range union {
				if union[j] == ts {
					break
				}
			}
			if got := aligned[i][j]; got == nil || *got != v {
				t.Fatalf("series %d at %v = %v, want its own value %v", i, ts.Time, got, v)
			}
		}
	}
}

func TestAlignDataPointsToUnionTimestampsEdgeCases(t *testing.T) {
	base := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	union := []common.SDKTime{sdkTimeAt(base, 1), sdkTimeAt(base, 2)}

	// Single series with all timestamps present is unchanged.
	single := alignDataPointsToUnionTimestamps(1, map[int]map[common.SDKTime]float64{
		0: {union[0]: 7, union[1]: 8},
	}, union)
	assertAlignedSeries(t, single, 0, "single", []*float64{float64Ptr(7), float64Ptr(8)})

	// Zero resources yields no series.
	if got := alignDataPointsToUnionTimestamps(0, map[int]map[common.SDKTime]float64{}, union); len(got) != 0 {
		t.Fatalf("zero-resources alignment returned %d series, want 0", len(got))
	}
}
