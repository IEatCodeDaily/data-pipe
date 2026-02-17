package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNewMetrics(t *testing.T) {
	// Create a new registry for testing to avoid conflicts
	reg := prometheus.NewRegistry()
	
	// Clear default registry for test
	oldRegistry := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = reg
	defer func() {
		prometheus.DefaultRegisterer = oldRegistry
	}()
	
	m := NewMetrics("test-pipeline")
	
	if m == nil {
		t.Fatal("Expected metrics to be created")
	}
	
	if m.EventsProcessed == nil {
		t.Error("EventsProcessed counter should not be nil")
	}
	
	if m.EventsErrored == nil {
		t.Error("EventsErrored counter should not be nil")
	}
	
	if m.ProcessingDuration == nil {
		t.Error("ProcessingDuration histogram should not be nil")
	}
	
	if m.PipelineStatus == nil {
		t.Error("PipelineStatus gauge should not be nil")
	}
	
	if m.SourceConnected == nil {
		t.Error("SourceConnected gauge should not be nil")
	}
	
	if m.SinkConnected == nil {
		t.Error("SinkConnected gauge should not be nil")
	}
}

func TestRecordEventProcessed(t *testing.T) {
	// Create a new registry for testing
	reg := prometheus.NewRegistry()
	oldRegistry := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = reg
	defer func() {
		prometheus.DefaultRegisterer = oldRegistry
	}()
	
	m := NewMetrics("test-pipeline")
	
	// Record some events
	m.RecordEventProcessed("test-pipeline", "insert")
	m.RecordEventProcessed("test-pipeline", "insert")
	m.RecordEventProcessed("test-pipeline", "update")
	
	// Verify the counter was incremented
	count := testutil.CollectAndCount(m.EventsProcessed)
	if count == 0 {
		t.Error("Expected events to be recorded")
	}
}

func TestRecordEventError(t *testing.T) {
	reg := prometheus.NewRegistry()
	oldRegistry := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = reg
	defer func() {
		prometheus.DefaultRegisterer = oldRegistry
	}()
	
	m := NewMetrics("test-pipeline")
	
	// Record some errors
	m.RecordEventError("test-pipeline", "source", "connection_error")
	m.RecordEventError("test-pipeline", "sink", "write_error")
	
	// Verify the counter was incremented
	count := testutil.CollectAndCount(m.EventsErrored)
	if count == 0 {
		t.Error("Expected errors to be recorded")
	}
}

func TestSetPipelineRunning(t *testing.T) {
	reg := prometheus.NewRegistry()
	oldRegistry := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = reg
	defer func() {
		prometheus.DefaultRegisterer = oldRegistry
	}()
	
	m := NewMetrics("test-pipeline")
	
	// Test setting pipeline to running
	m.SetPipelineRunning(true)
	
	// Test setting pipeline to stopped
	m.SetPipelineRunning(false)
}

func TestSetSourceConnected(t *testing.T) {
	reg := prometheus.NewRegistry()
	oldRegistry := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = reg
	defer func() {
		prometheus.DefaultRegisterer = oldRegistry
	}()
	
	m := NewMetrics("test-pipeline")
	
	// Test setting source connected
	m.SetSourceConnected(true)
	m.SetSourceConnected(false)
}

func TestSetSinkConnected(t *testing.T) {
	reg := prometheus.NewRegistry()
	oldRegistry := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = reg
	defer func() {
		prometheus.DefaultRegisterer = oldRegistry
	}()
	
	m := NewMetrics("test-pipeline")
	
	// Test setting sink connected
	m.SetSinkConnected(true)
	m.SetSinkConnected(false)
}

func TestRecordProcessingDuration(t *testing.T) {
	reg := prometheus.NewRegistry()
	oldRegistry := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = reg
	defer func() {
		prometheus.DefaultRegisterer = oldRegistry
	}()
	
	m := NewMetrics("test-pipeline")
	
	// Record some durations
	m.RecordProcessingDuration("test-pipeline", "source", 0.5)
	m.RecordProcessingDuration("test-pipeline", "sink", 0.3)
	m.RecordProcessingDuration("test-pipeline", "transform", 0.1)
	
	// Verify the histogram was updated
	count := testutil.CollectAndCount(m.ProcessingDuration)
	if count == 0 {
		t.Error("Expected durations to be recorded")
	}
}
