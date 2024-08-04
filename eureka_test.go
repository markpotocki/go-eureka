package eureka

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEurekaClient_RegisterApplication(t *testing.T) {
	instance := Instance{
		InstanceID: "test-instance",
		HostName:   "localhost",
		App:        "test-app",
		IPAddr:     "127.0.0.1",
		VipAddress: "test-app",
		Status:     "UP",
		Port: Port{
			Enabled: true,
			Port:    8080,
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/eureka/v2/apps/test-app" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		var body struct {
			Instance Instance `json:"instance"`
		}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Instance != instance {
			t.Errorf("unexpected body: %+v", body)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	cli := &EurekaClient{
		Client:  srv.Client(),
		BaseURL: srv.URL,
	}

	if err := cli.RegisterApplication(instance); err != nil {
		t.Fatal(err)
	}
}

func TestEurekaClient_DeregisterApplication(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/eureka/v2/apps/test-app/test-instance" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cli := &EurekaClient{
		Client:  srv.Client(),
		BaseURL: srv.URL,
	}

	appID := "test-app"
	instanceID := "test-instance"

	if err := cli.DeregisterApplication(appID, instanceID); err != nil {
		t.Fatal(err)
	}
}

func TestEurekaClient_Heartbeat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/eureka/v2/apps/test-app/test-instance" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cli := &EurekaClient{
		Client:  srv.Client(),
		BaseURL: srv.URL,
	}

	appID := "test-app"
	instanceID := "test-instance"

	if err := cli.Heartbeat(appID, instanceID); err != nil {
		t.Fatal(err)
	}
}

func TestEurekaClient_HeartbeatNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	cli := &EurekaClient{
		Client:  srv.Client(),
		BaseURL: srv.URL,
	}

	appID := "test-app"
	instanceID := "test-instance"

	err := cli.Heartbeat(appID, instanceID)
	if err != ErrInstanceIDNotFound {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEurekaClient_OutOfService(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/eureka/v2/apps/test-app/test-instance/status" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("value") != "OUT_OF_SERVICE" {
			t.Errorf("unexpected query: %s", r.URL.Query().Get("value"))
		}
		if r.Method != http.MethodPut {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cli := &EurekaClient{
		Client:  srv.Client(),
		BaseURL: srv.URL,
	}

	appID := "test-app"
	instanceID := "test-instance"

	if err := cli.OutOfService(appID, instanceID); err != nil {
		t.Fatal(err)
	}
}

func TestEurekaClient_OutOfServiceUpdateFailed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cli := &EurekaClient{
		Client:  srv.Client(),
		BaseURL: srv.URL,
	}

	err := cli.OutOfService("test-app", "test-instance")
	if err != ErrStatusUpdateFailed {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEurekaClient_BackInService(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/eureka/v2/apps/test-app/test-instance/status" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("value") != "UP" {
			t.Errorf("unexpected query: %s", r.URL.Query().Get("value"))
		}
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cli := &EurekaClient{
		Client:  srv.Client(),
		BaseURL: srv.URL,
	}

	appID := "test-app"
	instanceID := "test-instance"

	if err := cli.BackInService(appID, instanceID); err != nil {
		t.Fatal(err)
	}
}

func TestEurekaClient_BackInServiceUpdateFailed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cli := &EurekaClient{
		Client:  srv.Client(),
		BaseURL: srv.URL,
	}

	err := cli.BackInService("test-app", "test-instance")
	if err != ErrStatusUpdateFailed {
		t.Fatalf("unexpected error: %v", err)
	}
}
