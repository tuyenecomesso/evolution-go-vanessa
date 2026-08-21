package whatsmeow_service

import (
	"testing"

	"github.com/patrickmn/go-cache"
	"go.mau.fi/whatsmeow/types/events"

	"github.com/EvolutionAPI/evolution-go/pkg/config"
	instance_model "github.com/EvolutionAPI/evolution-go/pkg/instance/model"
	logger_wrapper "github.com/EvolutionAPI/evolution-go/pkg/logger"
	poll_service "github.com/EvolutionAPI/evolution-go/pkg/poll/service"
)

func TestDisconnectedKeepsInstanceEligibleForStartupReconnect(t *testing.T) {
	repo := &fakeInstanceRepository{}
	cfg := &config.Config{LogDirectory: t.TempDir(), LogMaxSize: 1, LogMaxBackups: 1, LogMaxAge: 1}
	instance := &instance_model.Instance{
		Id:        "11111111-1111-1111-1111-111111111111",
		Name:      "test-instance",
		Token:     "test-token",
		Connected: true,
	}
	mycli := &MyClient{
		service:            &fakeWhatsmeowService{},
		Instance:           instance,
		userID:             instance.Id,
		token:              instance.Token,
		instanceRepository: repo,
		userInfoCache:      cache.New(cache.NoExpiration, cache.NoExpiration),
		config:             cfg,
		loggerWrapper:      logger_wrapper.NewLoggerManager(cfg),
	}
	defer mycli.loggerWrapper.GetLogger(instance.Id).Close()

	mycli.myEventHandler(&events.Disconnected{})

	if !repo.connected {
		t.Fatal("expected transient websocket disconnect to keep instance connected for ConnectOnStartup")
	}
	if repo.disconnectReason != "WhatsApp websocket disconnected; auto-reconnect active." {
		t.Fatalf("unexpected disconnect reason: %q", repo.disconnectReason)
	}
}

type fakeInstanceRepository struct {
	connected        bool
	disconnectReason string
}

func (f *fakeInstanceRepository) Create(instance instance_model.Instance) (*instance_model.Instance, error) {
	return &instance, nil
}

func (f *fakeInstanceRepository) GetInstanceByID(instanceId string) (*instance_model.Instance, error) {
	return &instance_model.Instance{Id: instanceId}, nil
}

func (f *fakeInstanceRepository) GetConnectedInstanceByID(instanceId string) (*instance_model.Instance, error) {
	return &instance_model.Instance{Id: instanceId, Connected: true}, nil
}

func (f *fakeInstanceRepository) GetInstanceByToken(token string) (*instance_model.Instance, error) {
	return &instance_model.Instance{Token: token}, nil
}

func (f *fakeInstanceRepository) GetInstanceByName(name string) (*instance_model.Instance, error) {
	return &instance_model.Instance{Name: name}, nil
}

func (f *fakeInstanceRepository) Update(instance *instance_model.Instance) error {
	return nil
}

func (f *fakeInstanceRepository) UpdateConnected(userId string, status bool, disconnectReason string) error {
	f.connected = status
	f.disconnectReason = disconnectReason
	return nil
}

func (f *fakeInstanceRepository) UpdateQrcode(userId string, qr string) error {
	return nil
}

func (f *fakeInstanceRepository) UpdateProxy(userId string, proxy string) error {
	return nil
}

func (f *fakeInstanceRepository) UpdateJid(userId string, jid string) error {
	return nil
}

func (f *fakeInstanceRepository) GetAllConnectedInstances() ([]*instance_model.Instance, error) {
	return nil, nil
}

func (f *fakeInstanceRepository) GetAllConnectedInstancesByClientName(clientName string) ([]*instance_model.Instance, error) {
	return nil, nil
}

func (f *fakeInstanceRepository) GetAll(clientName string) ([]*instance_model.Instance, error) {
	return nil, nil
}

func (f *fakeInstanceRepository) Delete(instanceId string) error {
	return nil
}

func (f *fakeInstanceRepository) GetAdvancedSettings(instanceId string) (*instance_model.AdvancedSettings, error) {
	return nil, nil
}

func (f *fakeInstanceRepository) UpdateAdvancedSettings(instanceId string, settings *instance_model.AdvancedSettings) error {
	return nil
}

type fakeWhatsmeowService struct{}

func (f *fakeWhatsmeowService) StartClient(clientData *ClientData) {}

func (f *fakeWhatsmeowService) ConnectOnStartup(clientName string) {}

func (f *fakeWhatsmeowService) StartInstance(instanceId string) error {
	return nil
}

func (f *fakeWhatsmeowService) ReconnectClient(instanceId string) error {
	return nil
}

func (f *fakeWhatsmeowService) ClearInstanceCache(instanceId string, token string) error {
	return nil
}

func (f *fakeWhatsmeowService) CallWebhook(instance *instance_model.Instance, queueName string, jsonData []byte) {
}

func (f *fakeWhatsmeowService) SendToGlobalQueues(event string, jsonData []byte, userId string) {}

func (f *fakeWhatsmeowService) ForceUpdateJid(instanceId string, number string) error {
	return nil
}

func (f *fakeWhatsmeowService) UpdateInstanceSettings(instanceId string) error {
	return nil
}

func (f *fakeWhatsmeowService) UpdateInstanceAdvancedSettings(instanceId string) error {
	return nil
}

func (f *fakeWhatsmeowService) GetPollService() poll_service.PollService {
	return nil
}
