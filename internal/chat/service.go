package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"barbercentral-core/internal/whatsapp"
)

type Service interface {
	ListChats(ctx context.Context, clientID string) ([]Chat, error)
	ListMessages(ctx context.Context, clientID, chatID string) ([]Message, error)
	SendMessage(ctx context.Context, clientID string, req SendMessageRequest) (*Message, error)
	ProcessWebhook(ctx context.Context, payload WebhookPayload) error
	SendButtonsMessage(ctx context.Context, clientID, number, title, text, footer string) error
}

type service struct {
	repo   Repository
	waRepo whatsapp.Repository
}

func NewService(repo Repository, waRepo whatsapp.Repository) Service {
	return &service{
		repo:   repo,
		waRepo: waRepo,
	}
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

func cleanNumber(number string) string {
	var clean []rune
	for _, r := range number {
		if r >= '0' && r <= '9' {
			clean = append(clean, r)
		}
	}
	s := string(clean)
	if len(s) == 10 || len(s) == 11 {
		s = "55" + s
	}
	return s
}

func (s *service) ListChats(ctx context.Context, clientID string) ([]Chat, error) {
	return s.repo.ListChats(ctx, clientID)
}

func (s *service) ListMessages(ctx context.Context, clientID, chatID string) ([]Message, error) {
	// Primeiro valida se o chat pertence ao clientID
	chat, err := s.repo.GetChatByID(ctx, clientID, chatID)
	if err != nil {
		return nil, err
	}

	// Zera o contador de não lidas ao abrir a conversa
	_ = s.repo.ResetUnreadCount(ctx, clientID, chatID)

	return s.repo.ListMessages(ctx, chat.ID)
}

func (s *service) SendMessage(ctx context.Context, clientID string, req SendMessageRequest) (*Message, error) {
	if req.ContactNumber == "" || req.Content == "" {
		return nil, fmt.Errorf("número de contato e conteúdo são obrigatórios")
	}

	// 1. Achar o canal de WhatsApp ativo da barbearia
	instances, err := s.waRepo.ListInstancesByClientID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	var activeInst *whatsapp.WhatsAppInstance
	for i := range instances {
		activeInst = &instances[i]
		break // usamos a primeira cadastrada
	}

	if activeInst == nil {
		return nil, fmt.Errorf("sua barbearia não possui nenhum canal de WhatsApp cadastrado ou ativo")
	}

	// 2. Chamar Evolution API para disparar a mensagem
	evoURL, evoKey := getEvoCredentials()
	payload := map[string]interface{}{
		"number": cleanNumber(req.ContactNumber),
		"options": map[string]interface{}{
			"delay":       1200,
			"presence":    "composing",
			"linkPreview": false,
		},
		"textMessage": map[string]interface{}{
			"text": req.Content,
		},
	}

	bodyBytes, _ := json.Marshal(payload)
	postURL := fmt.Sprintf("%s/message/sendText/%s", evoURL, activeInst.InstanceName)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", postURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("apikey", evoKey)
	httpReq.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("falha ao enviar mensagem via Evolution API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Evolution API retornou status de erro: %d", resp.StatusCode)
	}

	// 3. Registrar conversa e mensagem no banco local
	contactName := req.ContactNumber
	chat, err := s.repo.GetOrCreateChat(ctx, clientID, cleanNumber(req.ContactNumber), contactName)
	if err != nil {
		return nil, err
	}

	msg := &Message{
		ID:        uuid.New().String(),
		ChatID:    chat.ID,
		MessageID: "out_" + uuid.New().String()[:12],
		Direction: DirectionOutbound,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	err = s.repo.SaveMessage(ctx, msg)
	if err != nil {
		return nil, err
	}

	err = s.repo.UpdateChatLastMessage(ctx, chat.ID, req.Content, false)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *service) ProcessWebhook(ctx context.Context, payload WebhookPayload) error {
	// 1. Identificar se a instância pertence a algum cliente
	inst, err := s.waRepo.GetInstanceByName(ctx, payload.Instance)
	if err != nil {
		return err
	}
	if inst == nil || inst.ClientID == nil {
		return nil // Ignora webhook de instâncias não pertencentes ao sistema
	}
	clientID := *inst.ClientID

	remoteJid := payload.Data.Key.RemoteJid
	if remoteJid == "" || strings.Contains(remoteJid, "@g.us") {
		return nil // Ignorar mensagens de grupo
	}
	contactNumber := cleanNumber(remoteJid)

	// 2. Tratar resposta aos botões de confirmação/cancelamento
	selectedButtonID := payload.Data.Message.ButtonsResponseMessage.SelectedButtonId
	if selectedButtonID != "" {
		return s.processButtonResponse(ctx, clientID, contactNumber, selectedButtonID)
	}

	// Tratar mensagem de texto convencional
	textMsg := payload.Data.Message.Conversation
	if textMsg == "" {
		textMsg = payload.Data.Message.ExtendedTextMessage.Text
	}

	if textMsg == "" {
		return nil // Ignorar se não houver conteúdo legível de texto
	}

	direction := DirectionInbound
	if payload.Data.Key.FromMe {
		direction = DirectionOutbound
	}

	contactName := payload.Data.PushName
	if contactName == "" {
		contactName = contactNumber
	}

	chat, err := s.repo.GetOrCreateChat(ctx, clientID, contactNumber, contactName)
	if err != nil {
		return err
	}

	msg := &Message{
		ID:        uuid.New().String(),
		ChatID:    chat.ID,
		MessageID: payload.Data.Key.ID,
		Direction: direction,
		Content:   textMsg,
		CreatedAt: time.Now(),
	}

	err = s.repo.SaveMessage(ctx, msg)
	if err != nil {
		return err
	}

	incrementUnread := (direction == DirectionInbound)
	return s.repo.UpdateChatLastMessage(ctx, chat.ID, textMsg, incrementUnread)
}

func (s *service) processButtonResponse(ctx context.Context, clientID, contactNumber, buttonID string) error {
	// Limpar o número para buscar correspondência no banco de dados do agendamento (geralmente sem código de país se DDD local)
	cleanPhone := contactNumber
	if len(cleanPhone) > 10 && cleanPhone[:2] == "55" {
		cleanPhone = cleanPhone[2:]
	}

	// 1. Buscar o ID do agendamento pendente mais recente deste cliente
	appID, err := s.repo.GetLatestPendingAppointment(ctx, clientID, cleanPhone)
	if err != nil {
		return fmt.Errorf("nenhum agendamento pendente encontrado para o contato %s: %w", cleanPhone, err)
	}

	// 2. Mapear status com base no clique do botão
	status := "confirmed"
	if buttonID == "cancel_appointment" {
		status = "cancelled"
	}

	// 3. Atualizar status no banco
	err = s.repo.UpdateAppointmentStatus(ctx, clientID, appID, status)
	if err != nil {
		return fmt.Errorf("erro ao atualizar status do agendamento %s para %s: %w", appID, status, err)
	}

	return nil
}

func (s *service) SendButtonsMessage(ctx context.Context, clientID, number, title, text, footer string) error {
	// 1. Achar o canal de WhatsApp ativo da barbearia
	instances, err := s.waRepo.ListInstancesByClientID(ctx, clientID)
	if err != nil {
		return err
	}

	var activeInst *whatsapp.WhatsAppInstance
	for i := range instances {
		activeInst = &instances[i]
		break // usamos a primeira cadastrada
	}

	if activeInst == nil {
		return fmt.Errorf("barbearia não possui canal de WhatsApp cadastrado ou ativo")
	}

	// 2. Chamar Evolution API para enviar botões
	evoURL, evoKey := getEvoCredentials()
	payload := map[string]interface{}{
		"number":      cleanNumber(number),
		"title":       title,
		"description": text,
		"footer":      footer,
		"buttons": []map[string]interface{}{
			{
				"id":          "confirm_appointment",
				"displayText": "Confirmar",
			},
			{
				"id":          "cancel_appointment",
				"displayText": "Cancelar",
			},
		},
	}

	bodyBytes, _ := json.Marshal(payload)
	postURL := fmt.Sprintf("%s/message/sendButtons/%s", evoURL, activeInst.InstanceName)
	req, err := http.NewRequestWithContext(ctx, "POST", postURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("apikey", evoKey)
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("falha ao enviar mensagem de botão: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Evolution API retornou status %d", resp.StatusCode)
	}

	// Registrar a mensagem enviada no chat local
	chat, err := s.repo.GetOrCreateChat(ctx, clientID, cleanNumber(number), number)
	if err == nil {
		msg := &Message{
			ID:        uuid.New().String(),
			ChatID:    chat.ID,
			MessageID: "out_btn_" + uuid.New().String()[:10],
			Direction: DirectionOutbound,
			Content:   fmt.Sprintf("[Botões de Lembrete: %s]", text),
			CreatedAt: time.Now(),
		}
		_ = s.repo.SaveMessage(ctx, msg)
		_ = s.repo.UpdateChatLastMessage(ctx, chat.ID, msg.Content, false)
	}

	return nil
}
