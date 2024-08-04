package eureka

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type EurekaClient struct {
	Client  *http.Client
	BaseURL string
}

type Status string

const (
	StatusUp           Status = "UP"
	StatusDown         Status = "DOWN"
	StatusStarting     Status = "STARTING"
	StatusOutOfService Status = "OUT_OF_SERVICE"
	StatusUnknown      Status = "UNKNOWN"
)

type Instance struct {
	InstanceID string     `json:"instanceId,omitempty"`
	HostName   string     `json:"hostName,omitempty"`
	App        string     `json:"app,omitempty"`
	IPAddr     string     `json:"ipAddr,omitempty"`
	VipAddress string     `json:"vipAddress,omitempty"`
	SecureVip  string     `json:"secureVipAddress,omitempty"`
	Status     Status     `json:"status,omitempty"`
	Port       Port       `json:"port,omitempty"`
	SecurePort Port       `json:"securePort,omitempty"`
	DataCenter DataCenter `json:"dataCenterInfo,omitempty"`
	LeaseInfo  LeaseInfo  `json:"leaseInfo,omitempty"`
	Metadata   Metadata   `json:"metadata,omitempty"`
}

type DataCenter struct {
	Class string `json:"@class"`
	Name  string `json:"name"`
}

type LeaseInfo struct {
	RenewalIntervalInSecs int `json:"renewalIntervalInSecs"`
	DurationInSecs        int `json:"durationInSecs"`
}

type Port struct {
	Enabled bool `json:"@enabled"`
	Port    int  `json:"$"`
}

type Metadata map[string]string

func (cli *EurekaClient) RegisterApplication(instance Instance) error {
	var body = struct {
		Instance Instance `json:"instance"`
	}{
		Instance: instance,
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return err
	}

	url := cli.BaseURL + fmt.Sprintf("/eureka/v2/apps/%s", instance.App)
	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := cli.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// DeregisterApplication sends a DELETE request to the Eureka server to
// deregister the instance with the given ID.
func (cli *EurekaClient) DeregisterApplication(appID string, instanceID string) error {
	url := cli.BaseURL + fmt.Sprintf("/eureka/v2/apps/%s/%s", appID, instanceID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	resp, err := cli.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// Heartbeat sends a PUT request to the Eureka server to indicate that the
// instance with the given ID is still alive.
func (cli *EurekaClient) Heartbeat(appID string, instanceID string) error {
	url := cli.BaseURL + fmt.Sprintf("/eureka/v2/apps/%s/%s", appID, instanceID)
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := cli.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return ErrInstanceIDNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// OutOfService sends a PUT request to the Eureka server to indicate that the
// instance with the given ID is out of service.
func (cli *EurekaClient) OverrideStatus(appID string, instanceID string, status Status) error {
	url := cli.BaseURL + fmt.Sprintf("/eureka/v2/apps/%s/%s/status?value=%s", appID, instanceID, status)
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := cli.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		return ErrStatusUpdateFailed
	}

	return nil
}

// BackInService sends a DELETE request to the Eureka server to indicate that
// the instance with the given ID is back in service.
func (cli *EurekaClient) RemoveStatusOverride(appID string, instanceID string) error {
	url := cli.BaseURL + fmt.Sprintf("/eureka/v2/apps/%s/%s/status?value=UP", appID, instanceID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := cli.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		return ErrStatusUpdateFailed
	}

	return nil
}
