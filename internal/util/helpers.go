package util

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"gogs.tail02d447.ts.net/rafael/service-cli/internal/config"
)

type TailscaleKeyResponse struct {
	Key string `json:"key"`
	Id  string `json:"id"`
}

func tailscaleAPIRequest(method, url, body string) ([]byte, error) {
	TAILSCALE_API_TOKEN := config.GetTailscaleApiToken()
	client := &http.Client{}
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+TAILSCALE_API_TOKEN)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received status code %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func generateComposeContent(serviceName string, image string) string {
	composeContent := fmt.Sprintf(
		`services:
  %[1]s-ts:
    image: tailscale/tailscale:latest
    container_name: %[1]s-ts
    hostname: %[1]s
    environment:
      - TS_AUTHKEY=${TS_AUTHKEY}
      - "TS_EXTRA_ARGS=--advertise-tags=tag:container --reset"
      - TS_STATE_DIR=/var/lib/tailscale
      - TS_USERSPACE=false
      - TS_SERVE_CONFIG=/config/serve.json
    volumes:
      - %[1]s-ts:/var/lib/tailscale
      - ./config:/config:ro
    devices:
      - /dev/net/tun:/dev/net/tun
    cap_add:
      - net_admin
      - sys_module
    restart: unless-stopped
    healthcheck:
      test: [ "CMD", "tailscale", "status" ]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s

  %[1]s:
    image: %[2]s
    container_name: %[1]s
    network_mode: service:%[1]s-ts
    depends_on:
      %[1]s-ts:
        condition: service_healthy
    restart: unless-stopped

volumes:
  %[1]s-ts:
    driver: local
`, serviceName, image)

	return composeContent
}

func generateComposeBareContent(serviceName string, image string) string {
	composeContent := fmt.Sprintf(
		`services:
	%[1]s:
		image: %[2]s
		container_name: %[1]s
		restart: unless-stopped
`, serviceName, image)
	return composeContent
}

func GenerateFolderService(serviceName string) (string, error) {
	SERVICE_DIR := config.GetServiceDir()
	fmt.Println("SERVICE_DIR:", SERVICE_DIR)
	serviceDir := fmt.Sprintf("%s/%s", SERVICE_DIR, serviceName)
	err := os.MkdirAll(serviceDir, os.ModePerm)
	return serviceDir, err
}

func GenerateComposeFile(serviceDir string, serviceName string, serviceImg string, bare bool) error {
	var composeContent string
	if bare {
		composeContent = generateComposeBareContent(serviceName, serviceImg)
	} else {
		composeContent = generateComposeContent(serviceName, serviceImg)
	}
	composePath := fmt.Sprintf("%s/docker-compose.yml", serviceDir)
	err := os.WriteFile(composePath, []byte(composeContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing docker-compose.yml file: %w", err)
	}
	return nil
}

func GenerateTsKey(serviceDir string, serviceName string) error {
	TAILSCALE_TAILNET := config.GetTailscaleTailnet()

	body := fmt.Sprintf(`{
	  "capabilities": {
	      "devices": {
	          "create": {
	              "reusable": false,
	              "ephemeral": false,
	              "preauthorized": true,
	              "tags": ["tag:container"]
	          }
	      }
	  },
	  "expirySeconds": 7776000,
	  "description": "Auth key for %s"
	}`, serviceName)

	res, err := tailscaleAPIRequest("POST", TAILSCALE_TAILNET, body)
	if err != nil {
		return fmt.Errorf("error making Tailscale API request: %w", err)
	}

	envPath := fmt.Sprintf("%s/.env", serviceDir)

	var keyResp TailscaleKeyResponse
	err = json.Unmarshal(res, &keyResp)
	if err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	err = os.WriteFile(envPath, []byte(fmt.Sprintf("TS_ID=%s\nTS_AUTHKEY=%s\n", keyResp.Id, keyResp.Key)), 0644)
	if err != nil {
		return fmt.Errorf("error writing .env file: %w", err)
	}

	return nil
}

func GenerateServeFile(serviceDir string, port string) error {
	serveContent := fmt.Sprintf(`{
  "TCP": {
    "443": {
      "HTTPS": true
    }
  },
  "Web": {
    "${TS_CERT_DOMAIN}:443": {
      "Handlers": {
        "/": {
          "Proxy": "http://127.0.0.1:%s"
        }
      }
    }
  }
}
`, port)

	servePath := fmt.Sprintf("%s/config/serve.json", serviceDir)
	err := os.MkdirAll(fmt.Sprintf("%s/config", serviceDir), os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}
	err = os.WriteFile(servePath, []byte(serveContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing serve.json file: %w", err)
	}

	return nil
}

func RemoveServiceFolder(serviceName string) error {
	SERVICE_DIR := config.GetServiceDir()
	serviceDir := fmt.Sprintf("%s/%s", SERVICE_DIR, serviceName)
	err := os.RemoveAll(serviceDir)
	if err != nil {
		return fmt.Errorf("error removing service directory: %w", err)
	}

	return nil
}

func RevokeTsKey(serviceName string) error {
	TAILSCALE_TAILNET := config.GetTailscaleTailnet()
	SERVICE_DIR := config.GetServiceDir()
	envPath := fmt.Sprintf("%s/%s/.env", SERVICE_DIR, serviceName)
	envFile, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("error reading .env file: %w", err)
	}

	var tsId string
	lines := strings.Split(string(envFile), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "TS_ID=") {
			tsId = strings.TrimPrefix(line, "TS_ID=")
			break
		}
	}

	if tsId == "" {
		return fmt.Errorf("TS_ID not found in .env file")
	}

	url := fmt.Sprintf("%s/%s", TAILSCALE_TAILNET, tsId)
	_, err = tailscaleAPIRequest("DELETE", url, "")
	if err != nil {
		return fmt.Errorf("error revoking Tailscale key: %w", err)
	}

	return nil
}

func GetServiceList() ([]string, error) {
	SERVICE_DIR := config.GetServiceDir()
	fmt.Println("SERVICE_DIR:", SERVICE_DIR)
	entries, err := os.ReadDir(SERVICE_DIR)
	if err != nil {
		return nil, fmt.Errorf("error reading service directory: %w", err)
	}

	var services []string
	for _, entry := range entries {
		if entry.IsDir() {
			services = append(services, entry.Name())
		}
	}

	return services, nil
}
