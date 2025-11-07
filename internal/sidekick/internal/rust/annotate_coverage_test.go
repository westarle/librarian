// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rust

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/googleapis/librarian/internal/sidekick/internal/config"
	"github.com/googleapis/librarian/internal/sidekick/internal/parser"
)

func TestAnnotationCoverage(t *testing.T) {
	// Find the testdata directory
	testdataDir, err := filepath.Abs("../../testdata/googleapis-full")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		t.Skipf("testdata directory not found at %s", testdataDir)
	}

	// Find all service config YAML files
	var serviceConfigs []string
	err = filepath.Walk(testdataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".yaml") {
			// Simple heuristic to identify service configs
			if strings.Contains(path, "google/cloud/") || strings.Contains(path, "google/iam/") || strings.Contains(path, "google/monitoring/") || strings.Contains(path, "google/logging/") || strings.Contains(path, "google/storage/") || strings.Contains(path, "google/pubsub/") || strings.Contains(path, "google/bigtable/") || strings.Contains(path, "google/bigquery/") || strings.Contains(path, "google/spanner/") || strings.Contains(path, "google/datastore/") || strings.Contains(path, "google/firestore/") || strings.Contains(path, "google/cloud/compute/") {
				serviceConfigs = append(serviceConfigs, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(serviceConfigs) == 0 {
		t.Skip("no service config files found in testdata")
	}

	totalMethods := 0
	methodsWithAnnotation := 0
	methodsWithMultipleAnnotations := 0
	methodsWithNoAnnotationButFallback := 0
	fallbackTypes := make(map[string]int)
	unmatchedMethods := 0

	type veneerStats struct {
		total     int
		annotated int
		fallback  int
	}
	veneerCoverage := make(map[string]*veneerStats)
	veneerPrefixes := map[string]string{
		"google/storage/":   "storage",
		"google/pubsub/":    "pubsub",
		"google/bigtable/":  "bigtable",
		"google/spanner/":   "spanner",
		"google/cloud/bigquery/":  "bigquery",
		"google/datastore/": "datastore",
		"google/firestore/": "firestore",
		"google/cloud/compute/": "compute",
	}

	codec, err := newCodec("protobuf", map[string]string{})
	if err != nil {
		t.Fatal(err)
	}

	for i, sc := range serviceConfigs {
		if i%50 == 0 {
			t.Logf("Processing %d/%d: %s", i+1, len(serviceConfigs), sc)
		}
		relPath, _ := filepath.Rel(testdataDir, sc)
		dir := filepath.Dir(relPath)

		var currentVeneer string
		for prefix, veneer := range veneerPrefixes {
			if strings.Contains(relPath, prefix) {
				currentVeneer = veneer
				if _, ok := veneerCoverage[currentVeneer]; !ok {
					veneerCoverage[currentVeneer] = &veneerStats{}
				}
				break
			}
		}

		cfg := &config.Config{
			General: config.GeneralConfig{
					SpecificationFormat: "protobuf",
					ServiceConfig:       relPath,
					SpecificationSource: dir,
			},
			Source: map[string]string{
				"googleapis-root": testdataDir,
			},
		}

		model, err := parser.CreateModel(cfg)
		if err != nil {
			// t.Logf("Failed to create model for %s: %v", sc, err)
			continue
		}

		// We don't strictly need full annotation, but it helps populate PathInfo
		annotateModel(model, codec)

		for _, service := range model.Services {
			for _, method := range service.Methods {
				totalMethods++
				if currentVeneer != "" {
					veneerCoverage[currentVeneer].total++
				}

				annotatedFields := 0
				if method.InputType != nil {
					for _, field := range method.InputType.Fields {
						if field.IsResourceReference {
							annotatedFields++
						}
					}
				}

				if annotatedFields > 0 {
					methodsWithAnnotation++
					if currentVeneer != "" {
						veneerCoverage[currentVeneer].annotated++
					}
					if annotatedFields > 1 {
						methodsWithMultipleAnnotations++
					}
				} else {
					// Fallback heuristic check
					fallbackMatch := false
					if method.PathInfo != nil && method.PathInfo.Codec != nil {
						ann, ok := method.PathInfo.Codec.(*pathInfoAnnotation)
						if ok {
							// UniqueParameters only contains fields that are present in the HTTP path.
							// This acts as a strong filter, reducing false positives for name-based matching.
							for _, p := range ann.UniqueParameters {
								if p.FieldName == "name" {
									fallbackMatch = true
									fallbackTypes["name"]++
									break
								} else if p.FieldName == "parent" {
									fallbackMatch = true
									fallbackTypes["parent"]++
									break
								} else if p.FieldName == "resource" {
									fallbackMatch = true
									fallbackTypes["resource"]++
									break
								} else if strings.HasSuffix(p.FieldName, ".name") {
									fallbackMatch = true
									fallbackTypes["*.name"]++
									break
								}
							}
						}
					}

					if fallbackMatch {
						methodsWithNoAnnotationButFallback++
						if currentVeneer != "" {
							veneerCoverage[currentVeneer].fallback++
						}
					} else {
						unmatchedMethods++
					}
				}
			}
		}
	}

	t.Logf("Total methods: %d", totalMethods)
	t.Logf("Methods with annotation: %d (%.2f%%)", methodsWithAnnotation, float64(methodsWithAnnotation)/float64(totalMethods)*100)
	t.Logf("  - Multiple annotations: %d", methodsWithMultipleAnnotations)
	t.Logf("Methods with NO annotation but matched fallback: %d (%.2f%%)", methodsWithNoAnnotationButFallback, float64(methodsWithNoAnnotationButFallback)/float64(totalMethods)*100)
	t.Logf("  - Fallback breakdown: %v", fallbackTypes)
	t.Logf("Unmatched methods: %d (%.2f%%)", unmatchedMethods, float64(unmatchedMethods)/float64(totalMethods)*100)

	totalCovered := methodsWithAnnotation + methodsWithNoAnnotationButFallback
	t.Logf("Total covered methods: %d (%.2f%%)", totalCovered, float64(totalCovered)/float64(totalMethods)*100)

	t.Logf("\n=== Veneer Coverage Analysis ===")
	for veneer, stats := range veneerCoverage {
		covered := stats.annotated + stats.fallback
		percentage := 0.0
		if stats.total > 0 {
			percentage = float64(covered) / float64(stats.total) * 100
		}
		t.Logf("%s: Total=%d, Covered=%d (%.2f%%) [Annotated=%d, Fallback=%d]", veneer, stats.total, covered, percentage, stats.annotated, stats.fallback)
	}
}