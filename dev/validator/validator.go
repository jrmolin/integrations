package validator

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
)

// IntegrationStats holds the summary statistics for a single integration.
type IntegrationStats struct {
	Name              string
	DataStreams       int
	SampleLogs        int
	Dashboards        int
	UniqueMessageIDs  int
	SingleMessageIDs  int
}


func getPackages() ([]string, error) {
	packagesDir := "packages"
	entries, err := os.ReadDir(packagesDir)
	if err != nil {
		return nil, fmt.Errorf("could not read packages directory: %w", err)
	}

	var packages []string
	for _, entry := range entries {
		if entry.IsDir() {
			pkgPath := filepath.Join(packagesDir, entry.Name())
			manifestPath := filepath.Join(pkgPath, "manifest.yml")
			if _, err := os.Stat(manifestPath); err == nil {
				packages = append(packages, pkgPath)
			}
		}
	}
	return packages, nil
}

func Check(fix bool) error {
	all_packages, err := getPackages()
	if err != nil {
		return err
	}

	for _, packagePath := range all_packages {
		fmt.Printf("Processing integration at %s\n", filepath.Dir(packagePath))
		integration, err := NewIntegration(packagePath)
		if err != nil {
			fmt.Printf("  ERROR: failed to create integration: %v\n", err)
			continue
		}

		if fix {
			if err := integration.Fix(); err != nil {
				fmt.Printf("  ERROR: failed to fix integration: %v\n", err)
			} else {
				fmt.Println("  SUCCESS: Integration fixed.")
			}
		} else {
			if err := integration.Check(); err != nil {
				fmt.Printf("  ERROR: validation failed: %v\n", err)
			} else {
				fmt.Println("  SUCCESS: Integration is valid.")
			}
		}
	}
	return nil
}

func printStatisticalSummary(stats []IntegrationStats) {
	fmt.Println("\n--- Global Summary ---")

	totalDataStreams := 0
	totalSampleLogs := 0
	totalDashboards := 0

	for _, s := range stats {
		totalDataStreams += s.DataStreams
		totalSampleLogs += s.SampleLogs
		totalDashboards += s.Dashboards
	}

	avgDataStreams := float64(totalDataStreams) / float64(len(stats))
	avgSampleLogs := float64(totalSampleLogs) / float64(len(stats))
	avgDashboards := float64(totalDashboards) / float64(len(stats))

	var dsVariance, slVariance, dbVariance float64
	for _, s := range stats {
		dsVariance += math.Pow(float64(s.DataStreams)-avgDataStreams, 2)
		slVariance += math.Pow(float64(s.SampleLogs)-avgSampleLogs, 2)
		dbVariance += math.Pow(float64(s.Dashboards)-avgDashboards, 2)
	}

	dsStdDev := math.Sqrt(dsVariance / float64(len(stats)))
	slStdDev := math.Sqrt(slVariance / float64(len(stats)))
	dbStdDev := math.Sqrt(dbVariance / float64(len(stats)))

	fmt.Printf("Total Integrations: %d\n", len(stats))
	fmt.Printf("Data Streams:\n")
	fmt.Printf("  Average: %.2f\n", avgDataStreams)
	fmt.Printf("  Std Dev: %.2f\n", dsStdDev)
	fmt.Printf("Sample Logs:\n")
	fmt.Printf("  Average: %.2f\n", avgSampleLogs)
	fmt.Printf("  Std Dev: %.2f\n", slStdDev)
	fmt.Printf("Dashboards:\n")
	fmt.Printf("  Average: %.2f\n", avgDashboards)
	fmt.Printf("  Std Dev: %.2f\n", dbStdDev)
}

func Summarize(pkg string, verbose bool) error {
	if pkg != "" {
		packagePath := filepath.Join("packages", pkg)
		if _, err := os.Stat(packagePath); os.IsNotExist(err) {
			return fmt.Errorf("package %s not found", pkg)
		}
		integration, err := NewIntegration(packagePath)
		if err != nil {
			return fmt.Errorf("failed to create integration: %w", err)
		}
		var stats []IntegrationStats
		s, err := integration.Summarize(verbose)
		stats = append(stats, s)
		printStatisticalSummary(stats)
		return err
	}

	all_packages, err := getPackages()
	if err != nil {
		return err
	}

	var stats []IntegrationStats
	for _, packagePath := range all_packages {
		integration, err := NewIntegration(packagePath)
		if err != nil {
			fmt.Printf("  ERROR: failed to create integration: %v\n", err)
			continue
		}
		s, err := integration.Summarize(verbose)
		if err != nil {
			fmt.Printf("  ERROR: failed to summarize integration: %v\n", err)
			continue
		}
		stats = append(stats, s)
	}

	printStatisticalSummary(stats)

	return nil
}

func GenerateSampleLogs(pkg string) error {
	fmt.Printf("Generating sample logs for package: %s\n", pkg)
	// TODO: Implement generate sample logs logic.
	return nil
}
