package layers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

const originalSuricataYamlPath string = "/root/obs-integ/suricata.yaml"
const tempSuricataYamlPath string = "/root/obs-integ/temp-suricata.yaml"

// Define a template to append the af-packet section dynamically
const afPacketTemplate = `
af-packet:
{{range .Interfaces}}
  - interface: {{.}}
    cluster-type: cluster_flow
{{end}}
`

// copySuricataConfig copies the original Suricata config to a temporary file
func copySuricataConfig() error {
	srcFile, err := os.Open(originalSuricataYamlPath)
	if err != nil {
		slog.Error("Failed to open original Suricata config",
			"error", err,
			"path", originalSuricataYamlPath)
		return fmt.Errorf("error opening original suricata.yaml: %v", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(tempSuricataYamlPath)
	if err != nil {
		slog.Error("Failed to create temporary Suricata config",
			"error", err,
			"path", tempSuricataYamlPath)
		return fmt.Errorf("error creating temp-suricata.yaml: %v", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		slog.Error("Failed to copy Suricata config",
			"error", err,
			"source", originalSuricataYamlPath,
			"destination", tempSuricataYamlPath)
		return fmt.Errorf("error copying suricata.yaml to temp-suricata.yaml: %v", err)
	}

	slog.Info("Successfully copied Suricata config",
		"source", originalSuricataYamlPath,
		"destination", tempSuricataYamlPath)
	return nil
}

// updateSuricataConfig updates Suricata's config with dynamically added af-packet section
func updateSuricataConfig(interfaces []string) error {
	content, err := os.ReadFile(tempSuricataYamlPath)
	if err != nil {
		slog.Error("Failed to read temporary Suricata config",
			"error", err,
			"path", tempSuricataYamlPath)
		return fmt.Errorf("error reading temp-suricata.yaml: %v", err)
	}

	// Prepare af-packet section using a template
	tmpl, err := template.New("afPacket").Parse(afPacketTemplate)
	if err != nil {
		slog.Error("Failed to parse af-packet template", "error", err)
		return fmt.Errorf("error parsing af-packet template: %v", err)
	}

	// Prepare the data for the template
	data := struct {
		Interfaces []string
	}{
		Interfaces: interfaces,
	}

	// Create a temporary buffer to hold the generated af-packet section
	var afPacketContent strings.Builder
	if err := tmpl.Execute(&afPacketContent, data); err != nil {
		slog.Error("Failed to execute af-packet template", "error", err)
		return fmt.Errorf("error executing af-packet template: %v", err)
	}

	// Dynamically append the generated af-packet section to the Suricata config
	updatedContent := strings.ReplaceAll(string(content), "# af-packet-section-placeholder", afPacketContent.String())

	err = os.WriteFile(tempSuricataYamlPath, []byte(updatedContent), 0644)
	if err != nil {
		slog.Error("Failed to write updated Suricata config",
			"error", err,
			"path", tempSuricataYamlPath)
		return fmt.Errorf("error writing to temp-suricata.yaml: %v", err)
	}

	slog.Info("Updated Suricata config with network interfaces",
		"interfaces", interfaces,
		"path", tempSuricataYamlPath)
	return nil
}

// ReceiverFunc processes network interfaces and configures Suricata
func ReceiverFunc(ctx context.Context, ch *Channels, interfaces []string) error {
	// If the list is empty, fetch all possible non-loopback interfaces
	if len(interfaces) == 0 {
		availableInterfaces, err := getAllNonLoopbackInterfaces()
		if err != nil {
			slog.Error("Failed to find available network interfaces", "error", err)
			return fmt.Errorf("failed to find suitable network interfaces: %v", err)
		}
		interfaces = availableInterfaces
		slog.Info("Discovered non-loopback interfaces", "interfaces", interfaces)
	}

	slog.Info("Using network interfaces", "interfaces", interfaces)

	if err := copySuricataConfig(); err != nil {
		slog.Error("Failed to copy Suricata config", "error", err)
		return fmt.Errorf("failed to copy suricata config: %v", err)
	}

	// Update Suricata config with dynamically generated af-packet section
	if err := updateSuricataConfig(interfaces); err != nil {
		slog.Error("Failed to update Suricata config", "error", err)
		return fmt.Errorf("failed to update suricata config: %v", err)
	}

	// Prepare command with multiple `-i` options
	cmdArgs := []string{"-oL", "suricata", "-c", tempSuricataYamlPath}
	for _, iface := range interfaces {
		cmdArgs = append(cmdArgs, "-i", iface)
	}

	cmd := exec.Command("stdbuf", cmdArgs...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		slog.Error("Failed to create stdout pipe", "error", err)
		return fmt.Errorf("error creating StdoutPipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		slog.Error("Failed to start Suricata", "error", err)
		return fmt.Errorf("error starting command: %v", err)
	}

	defer func() {
		if err := cmd.Process.Kill(); err != nil {
			slog.Error("Failed to kill Suricata process", "error", err)
		}
	}()

	slog.Info("Started Suricata process",
		"interfaces", interfaces,
		"config_path", tempSuricataYamlPath)

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			slog.Info("Context cancelled, stopping receiver")
			return ctx.Err()
		default:
			data := scanner.Bytes()
			event := &SuricataHTTPEvent{}
			if err := json.Unmarshal(data, event); err == nil {
				ch.LogsChan <- event
			}
		}
	}

	if err := scanner.Err(); err != nil {
		slog.Error("Scanner error", "error", err)
		return fmt.Errorf("scanner error: %v", err)
	}

	return nil
}

// Get all non-loopback interfaces
func getAllNonLoopbackInterfaces() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		slog.Error("Failed to get network interfaces", "error", err)
		return nil, err
	}

	var nonLoopbackInterfaces []string
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagUp != 0 {
			nonLoopbackInterfaces = append(nonLoopbackInterfaces, iface.Name)
		}
	}

	if len(nonLoopbackInterfaces) == 0 {
		return nil, fmt.Errorf("no suitable network interface found")
	}

	return nonLoopbackInterfaces, nil
}
