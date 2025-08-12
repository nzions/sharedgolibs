package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/nzions/sharedgolibs/pkg/util"
)

const version = "1.0.1"

// ContainerInfo represents comprehensive information about a Docker container
type ContainerInfo struct {
	Name          string            `json:"name"`
	InternalIP    string            `json:"internal_ip"`
	InternalPorts []string          `json:"internal_ports"`
	ExternalPorts []string          `json:"external_ports"`
	DNSAliases    []string          `json:"dns_aliases"`
	Version       string            `json:"version"`
	Keys          string            `json:"keys"`
	HasCurl       bool              `json:"has_curl"`
	HasWget       bool              `json:"has_wget"`
	ID            string            `json:"id"`
	Image         string            `json:"image"`
	Status        string            `json:"status"`
	Networks      map[string]string `json:"networks"`
}

func main() {
	var (
		jsonOutput  = flag.Bool("json", false, "Output in JSON format")
		help        = flag.Bool("help", false, "Show help")
		versionFlag = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *versionFlag {
		showVersion()
		return
	}

	// Print current environment manager environment
	currentEnv := util.MustGetEnv("ENVMGR_ENV", "default")
	if !*jsonOutput {
		fmt.Printf("Current envmgr environment: %s\n\n", currentEnv)
	}

	// Initialize Docker client
	dockerClient, err := initializeDockerClient()
	if err != nil {
		if *jsonOutput {
			result := map[string]interface{}{
				"envmgr_env": currentEnv,
				"error":      fmt.Sprintf("Docker not available: %v", err),
				"containers": []ContainerInfo{},
			}
			json.NewEncoder(os.Stdout).Encode(result)
		} else {
			fmt.Printf("Docker not available: %v\n", err)
		}
		return
	}
	defer dockerClient.Close()

	// Get running containers
	containers, err := getRunningContainers(dockerClient)
	if err != nil {
		if *jsonOutput {
			result := map[string]interface{}{
				"envmgr_env": currentEnv,
				"error":      fmt.Sprintf("Failed to get containers: %v", err),
				"containers": []ContainerInfo{},
			}
			json.NewEncoder(os.Stdout).Encode(result)
		} else {
			fmt.Printf("Failed to get containers: %v\n", err)
		}
		return
	}

	// Get detailed information for each container
	var containerInfos []ContainerInfo
	for _, c := range containers {
		info, err := getContainerInfo(dockerClient, c)
		if err != nil {
			if !*jsonOutput {
				fmt.Printf("Warning: Failed to get info for container %s: %v\n", c.ID[:12], err)
			}
			continue
		}
		containerInfos = append(containerInfos, info)
	}

	if *jsonOutput {
		result := map[string]interface{}{
			"envmgr_env": currentEnv,
			"containers": containerInfos,
		}
		json.NewEncoder(os.Stdout).Encode(result)
	} else {
		printContainerInfo(containerInfos)
	}
}

func showHelp() {
	fmt.Printf("envinfo version %s\n", version)
	fmt.Println("Environment and container information tool")
	fmt.Println()
	fmt.Println("Usage: envinfo [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -json           Output in JSON format")
	fmt.Println("  -version        Show version information")
	fmt.Println("  -help           Show this help message")
	fmt.Println()
	fmt.Println("This tool shows:")
	fmt.Println("  - Current envmgr environment")
	fmt.Println("  - Running Docker containers with:")
	fmt.Println("    * Container name and internal IP")
	fmt.Println("    * Internal and external ports")
	fmt.Println("    * DNS aliases assigned to container")
	fmt.Println("    * Output from --version (if supported)")
	fmt.Println("    * Output from --keys (if supported)")
	fmt.Println("    * Whether curl or wget are available")
	fmt.Println()
	fmt.Println("Source: https://github.com/nzions/sharedgolibs")
}

func showVersion() {
	fmt.Printf("envinfo version %s\n", version)
	fmt.Printf("util package version %s\n", util.Version)
}

func initializeDockerClient() (*client.Client, error) {
	// Try standard Docker client from environment first
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err == nil {
		// Test Docker connectivity
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err = dockerClient.Ping(ctx)
		if err == nil {
			return dockerClient, nil
		}
	}

	// If standard Docker isn't available, try Colima's default socket location
	homeDir, err := os.UserHomeDir()
	if err == nil {
		colimaSocketPath := fmt.Sprintf("unix://%s/.colima/default/docker.sock", homeDir)
		dockerClient, err = client.NewClientWithOpts(
			client.WithHost(colimaSocketPath),
			client.WithAPIVersionNegotiation(),
		)
		if err == nil {
			// Test Colima Docker connectivity
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, err = dockerClient.Ping(ctx)
			if err == nil {
				return dockerClient, nil
			}
		}
	}

	return nil, fmt.Errorf("docker not available")
}

func getRunningContainers(dockerClient *client.Client) ([]container.Summary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{
		All: false, // Only running containers
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list Docker containers: %w", err)
	}

	return containers, nil
}

func getContainerInfo(dockerClient *client.Client, c container.Summary) (ContainerInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get detailed container information
	inspectResult, err := dockerClient.ContainerInspect(ctx, c.ID)
	if err != nil {
		return ContainerInfo{}, fmt.Errorf("failed to inspect container: %w", err)
	}

	info := ContainerInfo{
		ID:       c.ID[:12],
		Image:    c.Image,
		Status:   c.Status,
		Networks: make(map[string]string),
	}

	// Get container name
	if len(c.Names) > 0 {
		info.Name = strings.TrimPrefix(c.Names[0], "/")
	}

	// Get network information
	for networkName, network := range inspectResult.NetworkSettings.Networks {
		info.Networks[networkName] = network.IPAddress
		if info.InternalIP == "" {
			info.InternalIP = network.IPAddress
		}

		// Collect DNS aliases
		if len(network.Aliases) > 0 {
			info.DNSAliases = append(info.DNSAliases, network.Aliases...)
		}
	}

	// Get port information using maps to avoid duplicates
	internalPortsMap := make(map[string]bool)
	externalPortsMap := make(map[string]bool)

	for _, port := range c.Ports {
		internalPort := fmt.Sprintf("%d/%s", port.PrivatePort, port.Type)
		internalPortsMap[internalPort] = true

		if port.PublicPort != 0 {
			externalPort := fmt.Sprintf("%d:%d", port.PublicPort, port.PrivatePort)
			externalPortsMap[externalPort] = true
		}
	}

	// Convert maps to slices
	for port := range internalPortsMap {
		info.InternalPorts = append(info.InternalPorts, port)
	}
	for port := range externalPortsMap {
		info.ExternalPorts = append(info.ExternalPorts, port)
	}

	// Execute commands in container to get version, keys, and check for curl/wget
	info.Version = getContainerVersionFromEntrypoint(dockerClient, c.ID, inspectResult)
	info.Keys = getContainerKeysFromEntrypoint(dockerClient, c.ID, inspectResult)

	// Check for curl and wget
	curlCheck := execInContainer(dockerClient, c.ID, []string{"sh", "-c", "command -v curl >/dev/null 2>&1 && echo 'yes' || echo 'no'"})
	info.HasCurl = strings.TrimSpace(curlCheck) == "yes"

	wgetCheck := execInContainer(dockerClient, c.ID, []string{"sh", "-c", "command -v wget >/dev/null 2>&1 && echo 'yes' || echo 'no'"})
	info.HasWget = strings.TrimSpace(wgetCheck) == "yes"

	return info, nil
}

func execInContainer(dockerClient *client.Client, containerID string, cmd []string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create exec instance
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	execResp, err := dockerClient.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return fmt.Sprintf("exec create error: %v", err)
	}

	// Start exec
	attachResp, err := dockerClient.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return fmt.Sprintf("exec attach error: %v", err)
	}
	defer attachResp.Close()

	// Read output
	output, err := io.ReadAll(attachResp.Reader)
	if err != nil {
		return fmt.Sprintf("exec read error: %v", err)
	}

	// Clean up the output by removing control characters and Docker headers
	result := string(output)

	// Remove Docker stream headers (first 8 bytes if present)
	if len(result) >= 8 && result[0] == 1 {
		result = result[8:]
	}

	// Remove any remaining control characters
	cleaned := strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, result)

	return strings.TrimSpace(cleaned)
}

func getContainerVersionFromEntrypoint(dockerClient *client.Client, containerID string, inspectResult container.InspectResponse) string {
	entrypoint := getContainerEntrypoint(inspectResult)
	if entrypoint == "" || isSystemCommand(entrypoint) {
		return "version not detected"
	}

	// First check what the entrypoint supports with --help
	helpOutput := execInContainer(dockerClient, containerID, []string{entrypoint, "--help"})

	if helpOutput == "" || strings.Contains(helpOutput, "error") {
		return "version not detected"
	}

	// Check for both single dash (-version) and double dash (--version)
	if !strings.Contains(helpOutput, "--version") && !strings.Contains(helpOutput, "-version") {
		return "version not detected"
	}

	// Try --version first, then -version if that fails
	var result string
	if strings.Contains(helpOutput, "--version") {
		result = execInContainer(dockerClient, containerID, []string{entrypoint, "--version"})
	} else {
		result = execInContainer(dockerClient, containerID, []string{entrypoint, "-version"})
	}

	if result != "" && !strings.Contains(result, "error") && !strings.Contains(result, "not found") && !strings.Contains(result, "unknown flag") && !strings.Contains(result, "invalid") {
		return result
	}

	return "version not detected"
}

func getContainerKeysFromEntrypoint(dockerClient *client.Client, containerID string, inspectResult container.InspectResponse) string {
	entrypoint := getContainerEntrypoint(inspectResult)
	if entrypoint == "" || isSystemCommand(entrypoint) {
		return "not supported"
	}

	// First check what the entrypoint supports with --help
	helpOutput := execInContainer(dockerClient, containerID, []string{entrypoint, "--help"})

	// Check for both single dash (-keys) and double dash (--keys)
	if helpOutput != "" && !strings.Contains(helpOutput, "error") && (strings.Contains(helpOutput, "--keys") || strings.Contains(helpOutput, "-keys")) {
		var result string
		if strings.Contains(helpOutput, "--keys") {
			result = execInContainer(dockerClient, containerID, []string{entrypoint, "--keys"})
		} else {
			result = execInContainer(dockerClient, containerID, []string{entrypoint, "-keys"})
		}

		if result != "" && !strings.Contains(result, "error") && !strings.Contains(result, "not found") && !strings.Contains(result, "unknown flag") && !strings.Contains(result, "invalid") {
			return result
		}
	}

	return "not supported"
}

func isSystemCommand(cmd string) bool {
	systemCommands := []string{
		// Shell commands
		"sleep", "sh", "bash", "/bin/sh", "/bin/bash", "/bin/sleep",
		"tail", "/bin/tail", "cat", "/bin/cat", "echo", "/bin/echo",
		"wait", "/bin/wait", "true", "/bin/true", "false", "/bin/false",

		// System daemons and servers (that don't typically support --version/--keys)
		"nginx", "/usr/sbin/nginx", "httpd", "/usr/sbin/httpd",
		"dockerd", "/usr/bin/dockerd", "init", "/sbin/init",
		"systemd", "/usr/lib/systemd/systemd",

		// Common container utilities
		"entrypoint.sh", "/entrypoint.sh", "docker-entrypoint.sh", "/docker-entrypoint.sh",
		"start.sh", "/start.sh", "run.sh", "/run.sh",

		// Process managers
		"supervisord", "/usr/bin/supervisord", "pm2", "/usr/bin/pm2",

		// Other utilities that don't support our flags
		"tini", "/sbin/tini", "dumb-init", "/usr/bin/dumb-init",
	}

	for _, syscmd := range systemCommands {
		if cmd == syscmd {
			return true
		}
	}
	return false
}

func getContainerEntrypoint(inspectResult container.InspectResponse) string {
	// Check for explicit entrypoint
	if len(inspectResult.Config.Entrypoint) > 0 {
		return inspectResult.Config.Entrypoint[0]
	}

	// Check for command (first element is usually the binary)
	if len(inspectResult.Config.Cmd) > 0 {
		cmd := inspectResult.Config.Cmd[0]
		// Skip shell commands
		if cmd != "sh" && cmd != "bash" && cmd != "/bin/sh" && cmd != "/bin/bash" {
			return cmd
		}
	}

	return "" // No suitable entrypoint found, will fallback to version detection
}

func printContainerInfo(containers []ContainerInfo) {
	if len(containers) == 0 {
		fmt.Println("No running containers found.")
		return
	}

	fmt.Printf("Running containers (%d):\n\n", len(containers))

	for i, container := range containers {
		fmt.Printf("Container %d: %s\n", i+1, container.Name)
		fmt.Printf("  ID: %s\n", container.ID)
		fmt.Printf("  Image: %s\n", container.Image)
		fmt.Printf("  Status: %s\n", container.Status)
		fmt.Printf("  Internal IP: %s\n", container.InternalIP)

		if len(container.InternalPorts) > 0 {
			fmt.Printf("  Internal Ports: %s\n", strings.Join(container.InternalPorts, ", "))
		}

		if len(container.ExternalPorts) > 0 {
			fmt.Printf("  External Ports: %s\n", strings.Join(container.ExternalPorts, ", "))
		}

		if len(container.DNSAliases) > 0 {
			fmt.Printf("  DNS Aliases: %s\n", strings.Join(container.DNSAliases, ", "))
		}

		if len(container.Networks) > 0 {
			fmt.Printf("  Networks:\n")
			for network, ip := range container.Networks {
				fmt.Printf("    %s: %s\n", network, ip)
			}
		}

		fmt.Printf("  Version Command: %s\n", container.Version)
		fmt.Printf("  Keys Command: %s\n", container.Keys)
		fmt.Printf("  Has curl: %t\n", container.HasCurl)
		fmt.Printf("  Has wget: %t\n", container.HasWget)

		if i < len(containers)-1 {
			fmt.Println()
		}
	}
}
