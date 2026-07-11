package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	ListInstances(ctx context.Context) ([]WhatsAppInstance, error)
	CreateInstance(ctx context.Context, req CreateInstanceRequest) (*WhatsAppInstance, error)
	ConnectInstance(ctx context.Context, name string) (any, error)
	StateInstance(ctx context.Context, name string) (any, error)
	LogoutInstance(ctx context.Context, name string) (any, error)
	DeleteInstance(ctx context.Context, name string) (any, error)
	LinkInstance(ctx context.Context, req LinkInstanceRequest) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func getEvoCredentials() (string, string) {
	url := os.Getenv("EVOLUTION_API_URL")
	if url == "" {
		url = "https://wapi.hub1.com.br"
	}
	key := os.Getenv("EVOLUTION_API_KEY")
	if key == "" {
		key = "d9a294a36e9ce21bc82dee90764a8b9e07cba939bc236e79"
	}
	return url, key
}

func (s *service) ListInstances(ctx context.Context) ([]WhatsAppInstance, error) {
	dbInstances, err := s.repo.ListInstances(ctx)
	if err != nil {
		return nil, err
	}

	evoURL, evoKey := getEvoCredentials()
	req, err := http.NewRequestWithContext(ctx, "GET", evoURL+"/instance/fetchInstances", nil)
	if err != nil {
		return dbInstances, nil
	}
	req.Header.Set("apikey", evoKey)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return dbInstances, nil
	}
	defer resp.Body.Close()

	var evoList []EvoInstance
	if resp.StatusCode == http.StatusOK {
		_ = json.NewDecoder(resp.Body).Decode(&evoList)
	}

	dbMap := make(map[string]*WhatsAppInstance)
	for i := range dbInstances {
		dbMap[dbInstances[i].InstanceName] = &dbInstances[i]
	}

	var result []WhatsAppInstance

	for _, evo := range evoList {
		if dbInst, exists := dbMap[evo.Name]; exists {
			dbInst.ConnectionStatus = evo.ConnectionStatus
			dbInst.OwnerJid = evo.OwnerJid
			dbInst.ProfilePicUrl = evo.ProfilePicUrl
			dbInst.Number = evo.Number
			result = append(result, *dbInst)
			delete(dbMap, evo.Name)
		}
	}

	for _, dbInst := range dbMap {
		dbInst.ConnectionStatus = "close"
		result = append(result, *dbInst)
	}

	return result, nil
}

func (s *service) CreateInstance(ctx context.Context, req CreateInstanceRequest) (*WhatsAppInstance, error) {
	evoURL, evoKey := getEvoCredentials()

	existing, _ := s.repo.GetInstanceByName(ctx, req.InstanceName)
	if existing != nil {
		return nil, fmt.Errorf("já existe uma instância com o nome %s", req.InstanceName)
	}

	bodyData := map[string]any{
		"instanceName": req.InstanceName,
		"qrcode":       true,
		"integration":  "WHATSAPP-BAILEYS",
	}
	jsonBody, _ := json.Marshal(bodyData)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", evoURL+"/instance/create", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", evoKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erro na Evolution API (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var evoResult struct {
		Instance struct {
			InstanceID string `json:"instanceId"`
		} `json:"instance"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&evoResult)

	inst := &WhatsAppInstance{
		ID:             evoResult.Instance.InstanceID,
		InstanceName:   req.InstanceName,
		ClientID:       req.ClientID,
		ProfessionalID: req.ProfessionalID,
	}
	if inst.ID == "" {
		inst.ID = uuid.New().String()
	}

	err = s.repo.CreateInstance(ctx, inst)
	if err != nil {
		return nil, err
	}

	return inst, nil
}

func (s *service) ConnectInstance(ctx context.Context, name string) (any, error) {
	evoURL, evoKey := getEvoCredentials()
	req, err := http.NewRequestWithContext(ctx, "GET", evoURL+"/instance/connect/"+name, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", evoKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result any
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func (s *service) StateInstance(ctx context.Context, name string) (any, error) {
	evoURL, evoKey := getEvoCredentials()
	req, err := http.NewRequestWithContext(ctx, "GET", evoURL+"/instance/connectionState/"+name, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", evoKey)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result any
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func (s *service) LogoutInstance(ctx context.Context, name string) (any, error) {
	evoURL, evoKey := getEvoCredentials()
	req, err := http.NewRequestWithContext(ctx, "DELETE", evoURL+"/instance/logout/"+name, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", evoKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result any
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func (s *service) DeleteInstance(ctx context.Context, name string) (any, error) {
	_ = s.repo.DeleteInstance(ctx, name)

	evoURL, evoKey := getEvoCredentials()
	req, err := http.NewRequestWithContext(ctx, "DELETE", evoURL+"/instance/delete/"+name, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", evoKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result any
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func (s *service) LinkInstance(ctx context.Context, req LinkInstanceRequest) error {
	inst, err := s.repo.GetInstanceByName(ctx, req.InstanceName)
	if err != nil {
		return err
	}
	if inst == nil {
		newInst := &WhatsAppInstance{
			InstanceName:   req.InstanceName,
			ClientID:       req.ClientID,
			ProfessionalID: req.ProfessionalID,
		}
		return s.repo.CreateInstance(ctx, newInst)
	}

	return s.repo.UpdateInstanceLink(ctx, req.InstanceName, req.ClientID, req.ProfessionalID)
}
