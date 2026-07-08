// Copyright © 2026 Oracle and/or its affiliates. All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package plugin

import (
	"context"
	"errors"
	"testing"

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
