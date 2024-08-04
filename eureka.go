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

type Instance struct {
	InstanceID string `json:"instanceId"`
	HostName   string `json:"hostName"`
	App        string `json:"app"`
	IPAddr     string `json:"ipAddr"`
	VipAddress string `json:"vipAddress"`
	SecureVip  string `json:"secureVipAddress"`
	Status     string `json:"status"`
	Port       Port   `json:"port"`
}

type Port struct {
	Enabled bool `json:"@enabled"`
	Port    int  `json:"$"`
}

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
func (cli *EurekaClient) OutOfService(appID string, instanceID string) error {
	url := cli.BaseURL + fmt.Sprintf("/eureka/v2/apps/%s/%s/status?value=OUT_OF_SERVICE", appID, instanceID)
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
func (cli *EurekaClient) BackInService(appID string, instanceID string) error {
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
