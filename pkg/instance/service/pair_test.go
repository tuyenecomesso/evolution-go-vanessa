package instance_service

import (
	"testing"

	"go.mau.fi/whatsmeow"

	"github.com/EvolutionAPI/evolution-go/pkg/config"
	instance_model "github.com/EvolutionAPI/evolution-go/pkg/instance/model"
	logger_wrapper "github.com/EvolutionAPI/evolution-go/pkg/logger"
)

// Regression test for the bug where Pair() logged an error from PairPhone but
// returned nil as the error anyway, so callers (and the HTTP handler) always
// saw a 200 OK with an empty pairing code instead of the real failure.
func TestPair_MissingClient_ReturnsError(t *testing.T) {
	cfg := &config.Config{LogDirectory: t.TempDir()}
	svc := instances{
		clientPointer: map[string]*whatsmeow.Client{},
		loggerWrapper: logger_wrapper.NewLoggerManager(cfg),
	}

	instance := &instance_model.Instance{Id: "missing-instance"}
	result, err := svc.Pair(&PairStruct{Phone: "244900000000"}, instance)

	if err == nil {
		t.Fatal("expected an error when no client is registered for the instance, got nil")
	}
	if result != nil {
		t.Fatalf("expected nil result on error, got %+v", result)
	}
}

func TestPair_NilClientPointerEntry_ReturnsError(t *testing.T) {
	cfg := &config.Config{LogDirectory: t.TempDir()}
	svc := instances{
		clientPointer: map[string]*whatsmeow.Client{"nil-client-instance": nil},
		loggerWrapper: logger_wrapper.NewLoggerManager(cfg),
	}

	instance := &instance_model.Instance{Id: "nil-client-instance"}
	result, err := svc.Pair(&PairStruct{Phone: "244900000000"}, instance)

	if err == nil {
		t.Fatal("expected an error when the registered client is nil, got nil")
	}
	if result != nil {
		t.Fatalf("expected nil result on error, got %+v", result)
	}
}
