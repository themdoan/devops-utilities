package cloudtrace

import (
	"context"
	"os"
	"testing"
	"time"

	cloudtracepb "cloud.google.com/go/trace/apiv1/tracepb"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestNewClientWithGCE(t *testing.T) {
	var client *Client
	var client_err error
	client, client_err = NewClientWithGCE(context.TODO())

	if client_err != nil {
		t.Logf("failed to init client: %s", client_err)
	}

	_, err := client.ListProjects(context.TODO())
	if err != nil {
		t.Logf("failed to ListProjects: %s", err)
	}

}

func TestListTraces(t *testing.T) {
	var client *Client
	var client_err error
	start := time.Now()

	projectID := os.Getenv("PROJECT_ID")
	const testConnectionTimeWindow = time.Hour * 24 * 3

	listCtx, cancel := context.WithTimeout(context.TODO(), time.Duration(time.Minute*15))
	defer func() {
		cancel()
		log.DefaultLogger.Info("Finished testConnection", "duration", time.Since(start).String())
	}()
	client, client_err = NewClientWithGCE(listCtx)

	if client_err != nil {
		t.Logf("failed to init client: %s", client_err)
	}

	it := client.tClient.ListTraces(listCtx, &cloudtracepb.ListTracesRequest{
		ProjectId: projectID,
		PageSize:  2,
		StartTime: timestamppb.New(time.Now().Add(-testConnectionTimeWindow)),
		OrderBy:   "start desc",
		View:      cloudtracepb.ListTracesRequest_COMPLETE,
		Filter:    "latency:1s",
	})

	var i int64
	entries := []*cloudtracepb.Trace{}
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.DefaultLogger.Error("error getting page", "error", err)
			break
		}
		// t.Logf("ListTraces: %v", resp)
		entries = append(entries, resp)

		i++
		if i >= 10 {
			break
		}
	}
	t.Logf("ListTraces: %v", entries)
	// alertmanager, _ := NewAlertmanager("http://127.0.0.1:9093/api/v2/alerts/", "", nil)

	// err := alertmanager.Post(context.TODO(), entries)
	// if err != nil {
	// 	t.Logf("alert post fail: %s", err)
	// }

}

func TestGetTrace2(t *testing.T) {
	var client *Client
	var client_err error
	start := time.Now()

	projectID := os.Getenv("PROJECT_ID")

	listCtx, cancel := context.WithTimeout(context.TODO(), time.Duration(time.Minute*15))
	defer func() {
		cancel()
		log.DefaultLogger.Info("Finished testConnection", "duration", time.Since(start).String())
	}()
	client, client_err = NewClientWithGCE(listCtx)

	if client_err != nil {
		t.Logf("failed to init client: %s", client_err)
	}
	req := &TraceQuery{
		ProjectID: projectID,
		TraceID:   "d44ab6d246b294c78c128c06387b5255",
	}

	trace, _ := client.GetTrace(listCtx, req)
	t.Logf("GetTrace: %s", trace)
}
