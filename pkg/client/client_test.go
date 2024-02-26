package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestAPIClient_PostEvent(t *testing.T) {
	tests := []struct {
		name    string
		event   Event
		wantErr bool
	}{
		{
			name: "post events with different context types",
			event: Event{
				Source: "test",

				Context: []ContextElement{
					{
						Type: "test",
						Data: "test",
					},
					{
						Type: "test slice data",
						Data: []string{"test", "whatever", "can", "be", "here"},
					},
					{
						Type: "test map data",
						Data: map[string]string{"test": "test", "whatever": "whatever"},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(http.MethodPost, "http://mockedcore.localhost/api/v1/events",
				func(req *http.Request) (*http.Response, error) {

					var eventSource Event
					if err := json.NewDecoder(req.Body).Decode(&eventSource); err != nil {
						t.Errorf("error decoding request body: %v", err)
					}

					authorizationHeader := req.Header.Get("Authorization")
					if len(authorizationHeader) == 0 {
						t.Errorf("expected Authorization header to be set")
					}

					if eventSource.Source != tt.event.Source {
						t.Errorf("expected event source to be %s, got %s", tt.event.Source, eventSource.Source)
					}

					if len(eventSource.Context) != len(tt.event.Context) {
						t.Errorf("expected event context to have %d elements, got %d", len(tt.event.Context), len(eventSource.Context))
					}

					responder, _ := httpmock.NewJsonResponder(http.StatusAccepted, "")
					return responder(req)
				})
			c := NewAPIClient("someJWTtoken")
			c.WithBaseURL("http://mockedcore.localhost/api/v1")
			c.WithHTTPClient(http.DefaultClient)
			ctx := context.TODO()
			if err := c.PostEvent(ctx, tt.event); (err != nil) != tt.wantErr {
				t.Errorf("APIClient.PostEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
