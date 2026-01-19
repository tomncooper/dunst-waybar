//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testBinaryName      = "dunst-waybar"
	containerImage      = "archlinux:latest"
	testImageName       = "dunst-waybar-test"
	testImageTag        = "latest"
	dbusSessionAddress  = "unix:path=/tmp/dbus-session"
	displayNumber       = ":99"
)

var (
	// customImageBuilt tracks if we've already built our custom image
	customImageBuilt = false
)

// WaybarOutput represents the expected JSON output format
type WaybarOutput struct {
	Text    string `json:"text"`
	Tooltip string `json:"tooltip"`
	Class   string `json:"class"`
	Alt     string `json:"alt"`
}

// TestBinaryBasics tests basic binary functionality
func TestBinaryBasics(t *testing.T) {
	ctx := context.Background()

	// Build binary for testing
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Create container
	container := createTestContainer(t, ctx, binaryPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}()

	t.Run("VersionFlag", func(t *testing.T) {
		testVersionFlag(t, ctx, container)
	})

	t.Run("ErrorStateWithoutDbus", func(t *testing.T) {
		testErrorStateWithoutDbus(t, ctx, container)
	})
}

// TestBinaryWithDbus tests binary with D-Bus running
func TestBinaryWithDbus(t *testing.T) {
	ctx := context.Background()

	// Build binary for testing
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Create container with D-Bus
	container := createTestContainerWithDbus(t, ctx, binaryPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}()

	t.Run("ErrorStateWithoutDunst", func(t *testing.T) {
		testErrorStateWithoutDunst(t, ctx, container)
	})

	t.Run("OutputFormat", func(t *testing.T) {
		testOutputFormat(t, ctx, container)
	})
}

// TestBinaryWithDunst tests binary with Dunst actually running
func TestBinaryWithDunst(t *testing.T) {
	ctx := context.Background()

	// Build binary for testing
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Create container with D-Bus and Dunst
	container := createTestContainerWithDunst(t, ctx, binaryPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}()

	t.Run("UnpausedState", func(t *testing.T) {
		testUnpausedState(t, ctx, container)
	})

	t.Run("PausedState", func(t *testing.T) {
		testPausedState(t, ctx, container)
	})

	t.Run("StateTransition", func(t *testing.T) {
		testStateTransition(t, ctx, container)
	})
}

// buildBinary builds the dunst-waybar binary for testing
func buildBinary(t *testing.T) string {
	t.Helper()

	// Get project root
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	// Create temporary directory for binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, testBinaryName)

	// Build binary
	t.Logf("Building binary at %s", binaryPath)
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/dunst-waybar")
	cmd.Dir = projectRoot
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Build failed: %v\nOutput: %s", err, string(output))
	}

	// Verify binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatalf("Binary was not created at %s", binaryPath)
	}

	t.Logf("Binary built successfully at %s", binaryPath)
	return binaryPath
}

// createTestContainer creates a basic test container
func createTestContainer(t *testing.T, ctx context.Context, binaryPath string) testcontainers.Container {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image: containerImage,
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      binaryPath,
				ContainerFilePath: "/usr/local/bin/" + testBinaryName,
				FileMode:          0755,
			},
		},
		Cmd:        []string{"sleep", "infinity"},
		WaitingFor: wait.ForLog(""),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	return container
}

// createTestContainerWithDbus creates a container with D-Bus installed
func createTestContainerWithDbus(t *testing.T, ctx context.Context, binaryPath string) testcontainers.Container {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image: containerImage,
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      binaryPath,
				ContainerFilePath: "/usr/local/bin/" + testBinaryName,
				FileMode:          0755,
			},
		},
		Cmd:        []string{"sleep", "infinity"},
		WaitingFor: wait.ForLog(""),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// Install and start D-Bus inside the running container
	t.Log("Installing D-Bus in container...")
	_, _, err = container.Exec(ctx, []string{
		"/bin/bash", "-c",
		"pacman -Sy --noconfirm dbus && mkdir -p /run/dbus && dbus-daemon --system --fork",
	})
	if err != nil {
		t.Fatalf("Failed to install and start D-Bus: %v", err)
	}

	// Wait a bit for D-Bus to fully start
	time.Sleep(2 * time.Second)
	t.Log("D-Bus started successfully")

	return container
}

// testVersionFlag tests the --version flag
func testVersionFlag(t *testing.T, ctx context.Context, container testcontainers.Container) {
	t.Helper()

	code, output, err := container.Exec(ctx, []string{testBinaryName, "--version"})
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	outputStr, err := readExecOutput(output)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	if code != 0 {
		t.Errorf("Expected exit code 0, got %d", code)
	}

	if !strings.Contains(outputStr, "dunst-waybar") {
		t.Errorf("Expected version output to contain 'dunst-waybar', got: %s", outputStr)
	}

	t.Logf("Version output: %s", outputStr)
}

// testErrorStateWithoutDbus tests error state when D-Bus is not available
func testErrorStateWithoutDbus(t *testing.T, ctx context.Context, container testcontainers.Container) {
	t.Helper()

	// Run binary with timeout (it should exit immediately with error)
	code, output, err := container.Exec(ctx, []string{
		"/bin/bash", "-c",
		fmt.Sprintf("timeout 2 %s || true", testBinaryName),
	})
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	outputStr, err := readExecOutput(output)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	// Should output error state JSON before exiting
	if !strings.Contains(outputStr, `"class":"error"`) {
		t.Errorf("Expected error state in output, got: %s", outputStr)
	}

	// Exit code may vary (timeout or error from binary)
	t.Logf("Output: %s (exit code: %d)", outputStr, code)
}

// testErrorStateWithoutDunst tests error state when Dunst is not running
func testErrorStateWithoutDunst(t *testing.T, ctx context.Context, container testcontainers.Container) {
	t.Helper()

	// Run binary with timeout
	_, output, err := container.Exec(ctx, []string{
		"/bin/bash", "-c",
		fmt.Sprintf("timeout 2 %s || true", testBinaryName),
	})
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	outputStr, err := readExecOutput(output)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	// Should output error state since Dunst is not running
	if !strings.Contains(outputStr, `"class":"error"`) {
		t.Errorf("Expected error state in output, got: %s", outputStr)
	}

	t.Logf("Output: %s", outputStr)
}

// testOutputFormat tests that the output is valid JSON in correct format
func testOutputFormat(t *testing.T, ctx context.Context, container testcontainers.Container) {
	t.Helper()

	// Run binary with timeout
	_, output, err := container.Exec(ctx, []string{
		"/bin/bash", "-c",
		fmt.Sprintf("timeout 2 %s || true", testBinaryName),
	})
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	outputStr, err := readExecOutput(output)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	// Extract first JSON line
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")
	if len(lines) == 0 {
		t.Fatal("No output received")
	}

	// Find first line that contains JSON by looking for opening brace
	var jsonLine string
	for _, line := range lines {
		// Find the first { and extract from there
		idx := strings.Index(line, "{")
		if idx >= 0 {
			jsonLine = line[idx:]
			break
		}
	}

	if jsonLine == "" {
		t.Fatalf("No JSON output found in: %q", outputStr)
	}

	t.Logf("JSON output: %s", jsonLine)

	// Parse JSON
	var waybarOutput WaybarOutput
	if err := json.Unmarshal([]byte(jsonLine), &waybarOutput); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, jsonLine)
	}

	// Verify required fields are present
	if waybarOutput.Text == "" {
		t.Error("text field is empty")
	}
	if waybarOutput.Tooltip == "" {
		t.Error("tooltip field is empty")
	}
	if waybarOutput.Class == "" {
		t.Error("class field is empty")
	}
	if waybarOutput.Alt == "" {
		t.Error("alt field is empty")
	}

	// Since Dunst is not running, should be error state
	if waybarOutput.Class != "error" {
		t.Errorf("Expected class 'error', got '%s'", waybarOutput.Class)
	}

	t.Logf("Parsed output: %+v", waybarOutput)
}

// readExecOutput reads the output from container exec
func readExecOutput(reader io.Reader) (string, error) {
	buf := new(strings.Builder)
	_, err := io.Copy(buf, reader)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// execWithEnv executes a command in the container with D-Bus environment set
func execWithEnv(ctx context.Context, container testcontainers.Container, command string) (int, io.Reader, error) {
	return container.Exec(ctx, []string{
		"/bin/bash", "-c",
		fmt.Sprintf("export DBUS_SESSION_BUS_ADDRESS=%s && %s", dbusSessionAddress, command),
	})
}

// getCustomImageRequest returns a ContainerRequest that builds or uses the custom test image
func getCustomImageRequest(t *testing.T, binaryPath string) testcontainers.ContainerRequest {
	t.Helper()

	// Build custom image from Dockerfile (only once, then cached)
	return testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    ".",
			Dockerfile: "Dockerfile",
			Repo:       testImageName,
			Tag:        testImageTag,
			KeepImage:  true, // Keep image for reuse across tests
			PrintBuildLog: false,
		},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      binaryPath,
				ContainerFilePath: "/usr/local/bin/" + testBinaryName,
				FileMode:          0755,
			},
		},
		Cmd:        []string{"sleep", "infinity"},
		WaitingFor: wait.ForLog(""),
	}
}

// createTestContainerWithDunst creates a container with D-Bus and Dunst pre-installed
func createTestContainerWithDunst(t *testing.T, ctx context.Context, binaryPath string) testcontainers.Container {
	t.Helper()

	req := getCustomImageRequest(t, binaryPath)

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// Packages are already installed via Dockerfile, just start services
	
	// Start Xvfb (virtual display)
	t.Log("Starting Xvfb...")
	_, _, err = container.Exec(ctx, []string{
		"/bin/bash", "-c",
		fmt.Sprintf("Xvfb %s -screen 0 1024x768x24 &", displayNumber),
	})
	if err != nil {
		t.Fatalf("Failed to start Xvfb: %v", err)
	}

	time.Sleep(2 * time.Second)

	// Start D-Bus session bus
	t.Log("Starting D-Bus session...")
	_, _, err = container.Exec(ctx, []string{
		"/bin/bash", "-c",
		fmt.Sprintf("dbus-daemon --session --fork --address=%s", dbusSessionAddress),
	})
	if err != nil {
		t.Fatalf("Failed to start D-Bus: %v", err)
	}

	time.Sleep(2 * time.Second)

	// Start Dunst with virtual display
	t.Log("Starting Dunst...")
	_, _, err = container.Exec(ctx, []string{
		"/bin/bash", "-c",
		fmt.Sprintf("export DISPLAY=%s && export DBUS_SESSION_BUS_ADDRESS=%s && dunst &", displayNumber, dbusSessionAddress),
	})
	if err != nil {
		t.Fatalf("Failed to start Dunst: %v", err)
	}

	// Wait for Dunst to start
	time.Sleep(3 * time.Second)

	// Verify Dunst is running by checking process
	code, output, err := container.Exec(ctx, []string{
		"/bin/bash", "-c",
		"pgrep -x dunst",
	})
	if err != nil {
		t.Fatalf("Failed to check Dunst process: %v", err)
	}
	if code != 0 {
		outputStr, _ := readExecOutput(output)
		t.Fatalf("Dunst process is not running (exit code: %d, output: %s)", code, outputStr)
	}

	// Verify Dunst is accessible via D-Bus
	code, output, err = execWithEnv(ctx, container, "dunstctl is-paused")
	if err != nil {
		t.Fatalf("Failed to check Dunst status: %v", err)
	}
	if code != 0 {
		outputStr, _ := readExecOutput(output)
		t.Logf("Warning: dunstctl check failed (exit code: %d, output: %s)", code, outputStr)
		// Continue anyway, sometimes dunstctl takes a moment
	}

	t.Log("Dunst started successfully")
	return container
}

// testUnpausedState tests that binary correctly reports unpaused state
func testUnpausedState(t *testing.T, ctx context.Context, container testcontainers.Container) {
	t.Helper()

	// Ensure Dunst is unpaused
	_, _, err := execWithEnv(ctx, container, "dunstctl set-paused false")
	if err != nil {
		t.Fatalf("Failed to unpause Dunst: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Run binary with timeout
	_, output, err := execWithEnv(ctx, container, fmt.Sprintf("timeout 2 %s || true", testBinaryName))
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	outputStr, err := readExecOutput(output)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	// Extract and parse JSON
	jsonLine := extractJSON(t, outputStr)
	var waybarOutput WaybarOutput
	if err := json.Unmarshal([]byte(jsonLine), &waybarOutput); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, jsonLine)
	}

	// Verify unpaused state
	if waybarOutput.Class != "unpaused" {
		t.Errorf("Expected class 'unpaused', got '%s'", waybarOutput.Class)
	}
	if waybarOutput.Alt != "unpaused" {
		t.Errorf("Expected alt 'unpaused', got '%s'", waybarOutput.Alt)
	}

	t.Logf("✓ Unpaused state: %+v", waybarOutput)
}

// testPausedState tests that binary correctly reports paused state
func testPausedState(t *testing.T, ctx context.Context, container testcontainers.Container) {
	t.Helper()

	// Pause Dunst
	_, _, err := execWithEnv(ctx, container, "dunstctl set-paused true")
	if err != nil {
		t.Fatalf("Failed to pause Dunst: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Run binary with timeout
	_, output, err := execWithEnv(ctx, container, fmt.Sprintf("timeout 2 %s || true", testBinaryName))
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	outputStr, err := readExecOutput(output)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	// Extract and parse JSON
	jsonLine := extractJSON(t, outputStr)
	var waybarOutput WaybarOutput
	if err := json.Unmarshal([]byte(jsonLine), &waybarOutput); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, jsonLine)
	}

	// Verify paused state
	if waybarOutput.Class != "paused" {
		t.Errorf("Expected class 'paused', got '%s'", waybarOutput.Class)
	}
	if waybarOutput.Alt != "paused" {
		t.Errorf("Expected alt 'paused', got '%s'", waybarOutput.Alt)
	}

	t.Logf("✓ Paused state: %+v", waybarOutput)

	// Unpause for next test
	execWithEnv(ctx, container, "dunstctl set-paused false")
}

// testStateTransition tests that binary correctly handles state transitions
func testStateTransition(t *testing.T, ctx context.Context, container testcontainers.Container) {
	t.Helper()

	// Start with unpaused state
	_, _, err := execWithEnv(ctx, container, "dunstctl set-paused false")
	if err != nil {
		t.Fatalf("Failed to unpause Dunst: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Test initial unpaused state
	_, output, err := execWithEnv(ctx, container, fmt.Sprintf("timeout 1 %s || true", testBinaryName))
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	outputStr1, _ := readExecOutput(output)
	jsonLine1 := extractJSON(t, outputStr1)
	var state1 WaybarOutput
	if err := json.Unmarshal([]byte(jsonLine1), &state1); err != nil {
		t.Fatalf("Failed to parse initial state: %v", err)
	}

	if state1.Class != "unpaused" {
		t.Errorf("Initial state should be unpaused, got '%s'", state1.Class)
	}
	t.Logf("✓ Initial state: unpaused")

	// Pause Dunst
	_, _, err = execWithEnv(ctx, container, "dunstctl set-paused true")
	if err != nil {
		t.Fatalf("Failed to pause Dunst: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Test paused state
	_, output, err = execWithEnv(ctx, container, fmt.Sprintf("timeout 1 %s || true", testBinaryName))
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	outputStr2, _ := readExecOutput(output)
	jsonLine2 := extractJSON(t, outputStr2)
	var state2 WaybarOutput
	if err := json.Unmarshal([]byte(jsonLine2), &state2); err != nil {
		t.Fatalf("Failed to parse paused state: %v", err)
	}

	if state2.Class != "paused" {
		t.Errorf("After pause, state should be paused, got '%s'", state2.Class)
	}
	t.Logf("✓ Transitioned to: paused")

	// Unpause Dunst
	_, _, err = execWithEnv(ctx, container, "dunstctl set-paused false")
	if err != nil {
		t.Fatalf("Failed to unpause Dunst: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Test back to unpaused
	_, output, err = execWithEnv(ctx, container, fmt.Sprintf("timeout 1 %s || true", testBinaryName))
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	outputStr3, _ := readExecOutput(output)
	jsonLine3 := extractJSON(t, outputStr3)
	var state3 WaybarOutput
	if err := json.Unmarshal([]byte(jsonLine3), &state3); err != nil {
		t.Fatalf("Failed to parse unpaused state: %v", err)
	}

	if state3.Class != "unpaused" {
		t.Errorf("After unpause, state should be unpaused, got '%s'", state3.Class)
	}
	t.Logf("✓ Transitioned back to: unpaused")

	t.Logf("✓ State transition test passed: unpaused → paused → unpaused")
}

// extractJSON extracts the first JSON object from output
func extractJSON(t *testing.T, output string) string {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		idx := strings.Index(line, "{")
		if idx >= 0 {
			return line[idx:]
		}
	}
	t.Fatalf("No JSON found in output: %q", output)
	return ""
}
