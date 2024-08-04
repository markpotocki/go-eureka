package eureka

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
		assert.Equal(t, "/eureka/v2/apps/test-app", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var body struct {
			Instance Instance `json:"instance"`
		}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&body); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, instance, body.Instance)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	cli := &EurekaClient{
		Client:  srv.Client(),
		BaseURL: srv.URL,
	}

	assert.NoError(t, cli.RegisterApplication(instance))
}

func TestEurekaClient_DeregisterApplication(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/eureka/v2/apps/test-app/test-instance", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cli := &EurekaClient{
		Client:  srv.Client(),
		BaseURL: srv.URL,
	}

	appID := "test-app"
	instanceID := "test-instance"

	assert.NoError(t, cli.DeregisterApplication(appID, instanceID))
}

func TestEurekaClient_Heartbeat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/eureka/v2/apps/test-app/test-instance", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cli := &EurekaClient{
		Client:  srv.Client(),
		BaseURL: srv.URL,
	}

	appID := "test-app"
	instanceID := "test-instance"

	assert.NoError(t, cli.Heartbeat(appID, instanceID))
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

	assert.EqualError(t, cli.Heartbeat(appID, instanceID), ErrInstanceIDNotFound.Error())
}

func TestEurekaClient_OverrideStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/eureka/v2/apps/test-app/test-instance/status", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "OUT_OF_SERVICE", r.URL.Query().Get("value"))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cli := &EurekaClient{
		Client:  srv.Client(),
		BaseURL: srv.URL,
	}

	appID := "test-app"
	instanceID := "test-instance"

	assert.NoError(t, cli.OverrideStatus(appID, instanceID, StatusOutOfService))
}

func TestEurekaClient_OverrideStatusFailed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cli := &EurekaClient{
		Client:  srv.Client(),
		BaseURL: srv.URL,
	}

	assert.EqualError(t, cli.OverrideStatus("test-app", "test-instance", StatusOutOfService), ErrStatusUpdateFailed.Error())
}

func TestEurekaClient_RemoveStatusOverride(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/eureka/v2/apps/test-app/test-instance/status", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "UP", r.URL.Query().Get("value"))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cli := &EurekaClient{
		Client:  srv.Client(),
		BaseURL: srv.URL,
	}

	appID := "test-app"
	instanceID := "test-instance"

	assert.NoError(t, cli.RemoveStatusOverride(appID, instanceID))
}

func TestEurekaClient_RemoveStatusOverrideFailed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cli := &EurekaClient{
		Client:  srv.Client(),
		BaseURL: srv.URL,
	}

	assert.EqualError(t, cli.RemoveStatusOverride("test-app", "test-instance"), ErrStatusUpdateFailed.Error())
}
