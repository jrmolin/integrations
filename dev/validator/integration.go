package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Integration is an interface that defines the operations that can be performed on a single integration.
type Integration interface {
	Load() error
	Check() error
	Fix() error
	Logs() error
	Validate() []string
	SummarizeLogs() error
	Summarize(bool) (IntegrationStats, error)
	GenerateSampleLogs() error
}

// localIntegration is a concrete implementation of the Integration interface for a local integration package.
type localIntegration struct {
	path string

	manifest *Manifest
}

// Manifest represents the structure of the manifest.yml file.
type Manifest struct {
	Name        string       `yaml:"name"`
	Version     string       `yaml:"version"`
	Title       string       `yaml:"title"`
	Description string       `yaml:"description"`
	Type        string       `yaml:"type"`
	License     string       `yaml:"license"`
	Maintainers []string     `yaml:"maintainers"`
	Categories  []string     `yaml:"categories"`
	Icons       []Icon       `yaml:"icons"`
	Owner       Owner        `yaml:"owner"`
	Dependencies Dependencies `yaml:"dependencies"`

	DataStreams []DataStream
	Assets      Assets
}

// Icon represents an icon for the integration.
type Icon struct {
	Path string `yaml:"path"`
	Type string `yaml:"type"`
}

// Owner represents the owner of the integration.
type Owner struct {
	Name   string `yaml:"name"`
	Github string `yaml:"github"`
}

// DataStream represents a data stream in the integration.
type DataStream struct {
	path        string
	Name        string
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`

	LogFiles    []string
}

// Assets represents the assets associated with a data stream.
type Assets struct {
	Dashboards      []Asset
	SavedSearches   []Asset
	Visualizations  []Asset
	Pipelines       []Asset
}

// Asset represents a single asset, like a dashboard or saved search.
type Asset struct {
	ID   string
	Path string
}

// Dependencies represents the dependencies of the integration.
type Dependencies struct {
	Kibana         string            `yaml:"kibana"`
	Elasticsearch  string            `yaml:"elasticsearch"`
	Integrations   map[string]string `yaml:"integrations"`
}

// NewIntegration creates a new Integration instance for the given path.
func NewIntegration(path string) (Integration, error) {
	integ := localIntegration{path: path}
	if err := integ.Load() ; err != nil {
		return nil, err
	}
	return &integ, nil
}

func NewDataStream(path string) (*DataStream, error) {
	data := DataStream{path: path, Name: filepath.Base(path)}
	if err := data.Load() ; err != nil {
		return nil, err
	}
	return &data, nil
}

func (d *DataStream) Load() error {

	manifest := filepath.Join(d.path, "manifest.yml")

	data, err := os.ReadFile(manifest)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, d); err != nil {
		return err
	}

	logPath := filepath.Join(d.path, "_dev/test/pipeline")
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		entries, err := os.ReadDir(logPath)
		if err != nil {
			return err
		}
		for _, e := range entries {
			fname := e.Name()
			if isLogFile(fname) {
				d.LogFiles = append(d.LogFiles, fname)
			}
		}
	}

	systemLogPath := filepath.Join(d.path, "_dev/test/system")
	if _, err := os.Stat(systemLogPath); !os.IsNotExist(err) {
		entries, err := os.ReadDir(systemLogPath)
		if err != nil {
			return err
		}
		for _, e := range entries {
			fname := e.Name()
			if isLogFile(fname) {
				d.LogFiles = append(d.LogFiles, fname)
			}
		}
	}

	return nil
}

// Check performs all checks on the integration.
func (i *localIntegration) Check() error {
	errors := i.Validate()
	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, ", "))
	}
	return nil
}

// Fix attempts to fix any issues with the integration.
func (i *localIntegration) Load() error {
	if i.manifest != nil {
		return nil
	}

	manifestPath := filepath.Join(i.path, "manifest.yml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return err
	}

	// list the data_stream directory
	dataStreamsPath := filepath.Join(i.path, "data_stream")
	if _, err := os.Stat(dataStreamsPath); !os.IsNotExist(err) {
		dataStreamEntries, err := os.ReadDir(dataStreamsPath)
		if err != nil {
			return err
		}
		for _, dsEntry := range dataStreamEntries {
			if !dsEntry.IsDir() {
				continue
			}

			// for each directory, create a new DataStream object and append it
			dataStreamDir := filepath.Join(dataStreamsPath, dsEntry.Name())
			dataStream, err := NewDataStream(dataStreamDir)

			if err != nil {
				return err
			}

			manifest.DataStreams = append(manifest.DataStreams, *dataStream)
		}
	}

	type att struct {
		// name = attributes.title
		Title string `json:"title"`
	}
	type dashboard struct {
		Att att `json:"attributes"`
		// id = id (should also be the name of the file)
		Id  string `json:"id"`
	}

	// add dashboards
	dashboardsPath := filepath.Join(i.path, "kibana/dashboard")
	if _, err := os.Stat(dashboardsPath); !os.IsNotExist(err) {
		entries, err := os.ReadDir(dashboardsPath)
		if err != nil {
			return err
		}
		for _, e := range entries {
			// read the file
			// unmarshal the json to an object
			dash := filepath.Join(dashboardsPath, e.Name())

			data, err := os.ReadFile(dash)
			if err != nil {
				return err
			}

			var dash_object dashboard
			if err := yaml.Unmarshal(data, &dash_object); err != nil {
				return err
			}

			// create an asset
			asset := Asset{ID: dash_object.Id, Path: dash}

			manifest.Assets.Dashboards = append(
				manifest.Assets.Dashboards, asset)
		}
	}

	i.manifest = &manifest
	return nil
}

// Fix attempts to fix any issues with the integration.
func (i *localIntegration) Fix() error {
	var manifest *Manifest = i.manifest
	if manifest.Owner.Name == "" {
		manifest.Owner.Name = "Elastic"
	}

	if manifest.Owner.Github == "" {
		manifest.Owner.Github = "elastic"
	}

	data, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}

	return os.WriteFile(i.path, data, 0644)
}

// Logs returns the logs for the integration.
func (i *localIntegration) Logs() error {
	fmt.Printf("Getting logs for integration at %s\n", i.path)
	// TODO: Implement logs logic.
	return nil
}

// Validate validates the integration's manifest.
func (i *localIntegration) Validate() []string {
	var errors []string

	var manifest *Manifest = i.manifest
	if manifest.Name == "" {
		errors = append(errors, "field 'name' is required")
	}
	if manifest.Version == "" {
		errors = append(errors, "field 'version' is required")
	}
	if manifest.Title == "" {
		errors = append(errors, "field 'title' is required")
	}
	if manifest.Description == "" {
		errors = append(errors, "field 'description' is required")
	}
	if manifest.Type != "integration" {
		errors = append(errors, "field 'type' must be 'integration'")
	}

	if manifest.Owner.Name == "" {
		errors = append(errors, "field 'owner.name' is required")
	}
	if manifest.Owner.Github == "" {
		errors = append(errors, "field 'owner.github' is required")
	}

	for i, ds := range manifest.DataStreams {
		if ds.Name == "" {
			errors = append(errors, fmt.Sprintf("data_stream %d: field 'name' is required", i))
		}
		if ds.Title == "" {
			errors = append(errors, fmt.Sprintf("data_stream %d: field 'title' is required", i))
		}
	}

	return errors
}

func (i *localIntegration) Summarize(verbose bool) (IntegrationStats, error) {
	stats := IntegrationStats{
		Name: filepath.Base(i.path),
	}

	manifest := i.manifest

	stats.DataStreams = len(i.manifest.DataStreams)
	stats.Dashboards = len(i.manifest.Assets.Dashboards)

	if verbose {
		fmt.Printf("Integration: %s\n", stats.Name)
		fmt.Printf("  Data Streams: %d\n", stats.DataStreams)
	}

	for _, ds := range manifest.DataStreams {
		if verbose {
			fmt.Printf("    - %s\n", ds.Name)
		}
		stats.SampleLogs += len(ds.LogFiles)
	}


	if verbose {
		fmt.Printf("  Dashboards: %d\n", stats.Dashboards)
		for _, dash := range manifest.Assets.Dashboards {
			fmt.Printf("    - %s\n", dash.ID)
		}
		fmt.Printf("  Sample Logs: %d\n", stats.SampleLogs)
	}
	return stats, nil
}

func (i *localIntegration) GenerateSampleLogs() error {
	fmt.Printf("Generating sample logs for integration at %s\n", i.path)
	// TODO: Implement generate sample logs logic.
	return nil
}

func (i *localIntegration) SummarizeLogs() error {
	dataStreamsPath := filepath.Join(filepath.Dir(i.path), "data_stream")
	if _, err := os.Stat(dataStreamsPath); os.IsNotExist(err) {
		return nil // No data_streams directory, nothing to summarize.
	}

	dataStreamEntries, err := os.ReadDir(dataStreamsPath)
	if err != nil {
		return fmt.Errorf("could not read data_stream directory: %w", err)
	}

	fmt.Printf("  Log Summary:\n")

	for _, dsEntry := range dataStreamEntries {
		if !dsEntry.IsDir() {
			continue
		}
		dataStreamName := dsEntry.Name()
		logPaths := []string{
			filepath.Join(dataStreamsPath, dataStreamName, "_dev/test/pipeline"),
			filepath.Join(dataStreamsPath, dataStreamName, "_dev/test/system"),
		}

		totalLogs := 0
	
		uniqueMessages := make(map[string]int)

		for _, logPath := range logPaths {
			if _, err := os.Stat(logPath); os.IsNotExist(err) {
				continue
			}

			logEntries, err := os.ReadDir(logPath)
			if err != nil {
				return fmt.Errorf("could not read log directory %s: %w", logPath, err)
			}

			for _, logEntry := range logEntries {
				if !logEntry.IsDir() && strings.HasSuffix(logEntry.Name(), ".log") {
					totalLogs++
					logFilePath := filepath.Join(logPath, logEntry.Name())
					data, err := os.ReadFile(logFilePath)
					if err != nil {
						fmt.Printf("    WARN: could not read log file %s: %v\n", logFilePath, err)
						continue
					}

					re := regexp.MustCompile(`%[A-Z0-9_]+-.-(\d+)`)
					matches := re.FindAllStringSubmatch(string(data), -1)
					for _, match := range matches {
						uniqueMessages[match[1]]++
					}
				}
			}
		}

		if totalLogs > 0 {
		
uniqueCount := 0
			for _, count := range uniqueMessages {
				if count == 1 {
				
uniqueCount++
				}
			}
			fmt.Printf("    Data Stream: %s\n", dataStreamName)
			fmt.Printf("      Total log files: %d\n", totalLogs)
			fmt.Printf("      Unique message IDs: %d\n", len(uniqueMessages))
			fmt.Printf("      Messages that appear once: %d\n", uniqueCount)
		}
	}

	return nil
}

func isLogFile(name string) bool {
	if strings.Contains(name, "-config.yml") {
		return false
	}
	if strings.Contains(name, "-expected.json") {
		return false
	}
	return true
}
