package exporter_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/internal/exporter"
	"github.com/thomaspoignant/go-feature-flag/internal/model"
	"github.com/thomaspoignant/go-feature-flag/testutils"
)

func TestDataExporterScheduler_flushWithTime(t *testing.T) {
	mockExporter := testutils.MockExporter{Bulk: true}
	dc := exporter.NewDataExporterScheduler(
		context.Background(), 10*time.Millisecond, 1000, &mockExporter, log.New(os.Stdout, "", 0))
	go dc.StartDaemon()
	defer dc.Close()

	inputEvents := []exporter.FeatureEvent{
		exporter.NewFeatureEvent(ffuser.NewAnonymousUser("ABCD"), "random-key",
			&model.FlagData{Percentage: testconvert.Float64(100)}, "YO", model.VariationDefault, false),
	}

	for _, event := range inputEvents {
		dc.AddEvent(event)
	}

	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, inputEvents, mockExporter.GetExportedEvents())
}

func TestDataExporterScheduler_flushWithNumberOfEvents(t *testing.T) {
	mockExporter := testutils.MockExporter{Bulk: true}
	dc := exporter.NewDataExporterScheduler(
		context.Background(), 10*time.Minute, 100, &mockExporter, log.New(os.Stdout, "", 0))
	go dc.StartDaemon()
	defer dc.Close()

	var inputEvents []exporter.FeatureEvent
	for i := 0; i <= 100; i++ {
		inputEvents = append(inputEvents, exporter.NewFeatureEvent(ffuser.NewAnonymousUser("ABCD"),
			"random-key", &model.FlagData{Percentage: testconvert.Float64(100)},
			"YO", model.VariationDefault, false))
	}
	for _, event := range inputEvents {
		dc.AddEvent(event)
	}
	assert.Equal(t, inputEvents[:100], mockExporter.GetExportedEvents())
}

func TestDataExporterScheduler_defaultFlush(t *testing.T) {
	mockExporter := testutils.MockExporter{Bulk: true}
	dc := exporter.NewDataExporterScheduler(
		context.Background(), 0, 0, &mockExporter, log.New(os.Stdout, "", 0))
	go dc.StartDaemon()
	defer dc.Close()

	var inputEvents []exporter.FeatureEvent
	for i := 0; i <= 100000; i++ {
		inputEvents = append(inputEvents, exporter.NewFeatureEvent(ffuser.NewAnonymousUser("ABCD"),
			"random-key", &model.FlagData{Percentage: testconvert.Float64(100)},
			"YO", model.VariationDefault, false))
	}
	for _, event := range inputEvents {
		dc.AddEvent(event)
	}
	assert.Equal(t, inputEvents[:100000], mockExporter.GetExportedEvents())
}

func TestDataExporterScheduler_exporterReturnError(t *testing.T) {
	mockExporter := testutils.MockExporter{Err: errors.New("random err"), ExpectedNumberErr: 1, Bulk: true}

	file, _ := ioutil.TempFile("", "log")
	defer file.Close()
	defer os.Remove(file.Name())
	logger := log.New(file, "", 0)

	dc := exporter.NewDataExporterScheduler(
		context.Background(), 0, 100, &mockExporter, logger)
	go dc.StartDaemon()
	defer dc.Close()

	var inputEvents []exporter.FeatureEvent
	for i := 0; i <= 200; i++ {
		inputEvents = append(inputEvents, exporter.NewFeatureEvent(ffuser.NewAnonymousUser("ABCD"),
			"random-key", &model.FlagData{Percentage: testconvert.Float64(100)},
			"YO", model.VariationDefault, false))
	}
	for _, event := range inputEvents {
		dc.AddEvent(event)
	}
	assert.Equal(t, inputEvents[:201], mockExporter.GetExportedEvents())

	// read log
	logs, _ := ioutil.ReadFile(file.Name())
	assert.Regexp(t, "\\["+testutils.RFC3339Regex+"\\] error while exporting data: random err\\n", string(logs))
}

func TestDataExporterScheduler_nonBulkExporter(t *testing.T) {
	mockExporter := testutils.MockExporter{Bulk: false}
	dc := exporter.NewDataExporterScheduler(
		context.Background(), 0, 0, &mockExporter, log.New(os.Stdout, "", 0))
	defer dc.Close()

	var inputEvents []exporter.FeatureEvent
	for i := 0; i < 100; i++ {
		inputEvents = append(inputEvents, exporter.NewFeatureEvent(ffuser.NewAnonymousUser("ABCD"),
			"random-key", &model.FlagData{Percentage: testconvert.Float64(100)},
			"YO", model.VariationDefault, false))
	}
	for _, event := range inputEvents {
		dc.AddEvent(event)
	}
	assert.Equal(t, inputEvents[:100], mockExporter.GetExportedEvents())
}
