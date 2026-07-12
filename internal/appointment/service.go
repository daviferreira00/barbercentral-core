package appointment

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"sort"
	"time"

	"github.com/google/uuid"

	"barbercentral-core/internal/clientconfig"
	"barbercentral-core/internal/customer"
	"barbercentral-core/internal/email"
	"barbercentral-core/internal/loyalty"
	"barbercentral-core/internal/professional"
	svc "barbercentral-core/internal/service"
	"barbercentral-core/internal/stock"
	"barbercentral-core/internal/chat"
)

var (
	ErrSlotNotAvailable = errors.New("o horário selecionado já não está mais disponível")
	ErrCancelPolicy     = errors.New("o cancelamento não é permitido por estar fora da antecedência mínima exigida")
)

type Service interface {
	ListBlockedSlots(ctx context.Context, clientID, professionalID, startDate, endDate string) ([]BlockedSlot, error)
	CreateBlockedSlot(ctx context.Context, clientID, userID string, req CreateBlockedSlotRequest) (*BlockedSlot, error)
	DeleteBlockedSlot(ctx context.Context, clientID, id string) error
	List(ctx context.Context, clientID, professionalID, startDate, endDate string) ([]EnrichedAppointment, error)
	GetByID(ctx context.Context, clientID, id string) (*EnrichedAppointment, error)
	UpdateStatus(ctx context.Context, clientID, id, status, userID, notes string) error

	GetAvailability(ctx context.Context, slug string, professionalID string, serviceIDs []string, date string) ([]TimeSlot, error)
	CreatePublic(ctx context.Context, slug string, req CreatePublicAppointmentRequest) (*EnrichedAppointment, error)
	GetByCancelToken(ctx context.Context, token string) (*EnrichedAppointment, error)
	CancelByToken(ctx context.Context, token string) error
	SendUpcomingReminders(ctx context.Context) (int, error)

	SendWhatsAppNotification(ctx context.Context, slug, phone, message string) error

	// FASE-05 / FASE-06 adicionais
	Create(ctx context.Context, clientID, userID string, req CreateAppointmentRequest) (*EnrichedAppointment, error)
	Update(ctx context.Context, clientID, id, userID string, req UpdateAppointmentRequest) (*EnrichedAppointment, error)
	Cancel(ctx context.Context, clientID, id, userID, reason string) error
	GetStatusLogs(ctx context.Context, clientID, appointmentID string) ([]AppointmentStatusLog, error)
	GetAvailabilityInternal(ctx context.Context, clientID, professionalID string, serviceIDs []string, date string) ([]TimeSlot, error)
}

type service struct {
	repo        AppointmentRepository
	configRepo  clientconfig.ConfigRepository
	profRepo    professional.ProfessionalRepository
	svcRepo     svc.ServiceRepository
	emailClient *email.Client
	custService    customer.Service
	stockService   stock.Service
	loyaltyService loyalty.Service
	chatService    chat.Service
}

func NewService(
	repo AppointmentRepository,
	configRepo clientconfig.ConfigRepository,
	profRepo professional.ProfessionalRepository,
	svcRepo svc.ServiceRepository,
	emailClient *email.Client,
	custService customer.Service,
	stockService stock.Service,
	loyaltyService loyalty.Service,
	chatService chat.Service,
) Service {
	return &service{
		repo:           repo,
		configRepo:     configRepo,
		profRepo:       profRepo,
		svcRepo:        svcRepo,
		emailClient:    emailClient,
		custService:    custService,
		stockService:   stockService,
		loyaltyService: loyaltyService,
		chatService:    chatService,
	}
}

func (s *service) SendWhatsAppNotification(ctx context.Context, slug, phone, message string) error {
	cfg, err := s.configRepo.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}
	_, err = s.chatService.SendMessage(ctx, cfg.ClientID, chat.SendMessageRequest{
		ContactNumber: phone,
		Content:       message,
	})
	return err
}

func (s *service) ListBlockedSlots(ctx context.Context, clientID, professionalID, startDate, endDate string) ([]BlockedSlot, error) {
	return s.repo.ListBlockedSlots(ctx, clientID, professionalID, startDate, endDate)
}

func (s *service) CreateBlockedSlot(ctx context.Context, clientID, userID string, req CreateBlockedSlotRequest) (*BlockedSlot, error) {
	slot := &BlockedSlot{
		ID:             uuid.New().String(),
		ClientID:       clientID,
		ProfessionalID: req.ProfessionalID,
		Date:           req.Date,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		Reason:         req.Reason,
		CreatedBy:      userID,
		CreatedAt:      time.Now(),
	}

	err := s.repo.CreateBlockedSlot(ctx, slot)
	if err != nil {
		return nil, err
	}

	return slot, nil
}

func (s *service) DeleteBlockedSlot(ctx context.Context, clientID, id string) error {
	return s.repo.DeleteBlockedSlot(ctx, clientID, id)
}

func (s *service) List(ctx context.Context, clientID, professionalID, startDate, endDate string) ([]EnrichedAppointment, error) {
	return s.repo.List(ctx, clientID, professionalID, startDate, endDate)
}

func (s *service) GetByID(ctx context.Context, clientID, id string) (*EnrichedAppointment, error) {
	return s.repo.GetByID(ctx, clientID, id)
}

func (s *service) UpdateStatus(ctx context.Context, clientID, id, status, userID, notes string) error {
	app, err := s.repo.GetByID(ctx, clientID, id)
	if err != nil {
		return err
	}

	oldStatus := app.Status
	if oldStatus == status {
		return nil
	}

	err = s.repo.UpdateStatus(ctx, clientID, id, status)
	if err != nil {
		return err
	}

	// Registra log
	statusLog := &AppointmentStatusLog{
		ID:            uuid.New().String(),
		AppointmentID: id,
		FromStatus:    &oldStatus,
		ToStatus:      status,
		ChangedBy:     userID,
		Notes:         func() *string { if notes != "" { return &notes }; return nil }(),
		CreatedAt:     time.Now(),
	}
	_ = s.repo.CreateStatusLog(ctx, statusLog)

	// Notifica cliente se confirmado pelo barbeiro
	if status == "confirmed" && app.CustomerPhone != nil && *app.CustomerPhone != "" {
		dateFormatted := ""
		tDate, err := time.Parse("2006-01-02", app.Date)
		if err == nil {
			dateFormatted = tDate.Format("02/01/2006")
		} else {
			dateFormatted = app.Date
		}
		timeFormatted := app.StartTime[:5]
		servicesStr := ""
		for i, srv := range app.Services {
			if i > 0 {
				servicesStr += ", "
			}
			servicesStr += srv.ServiceName
		}
		
		msg := fmt.Sprintf("Olá, %s! Seu agendamento para o dia %s às %s foi confirmado pelo barbeiro.\n\nProfissional: %s\nServiços: %s",
			*app.CustomerName, dateFormatted, timeFormatted, app.ProfessionalName, servicesStr)
			
		_, _ = s.chatService.SendMessage(ctx, clientID, chat.SendMessageRequest{
			ContactNumber: *app.CustomerPhone,
			Content:       msg,
		})
	}

	// Notifica cliente se cancelado
	if status == "cancelled" && s.emailClient != nil && app.CustomerEmail != nil && *app.CustomerEmail != "" {
		subject := "Agendamento Cancelado"
		body := fmt.Sprintf("<p>Olá %s, seu agendamento para o dia %s às %s foi cancelado.</p>",
			*app.CustomerName, app.Date, app.StartTime)
		if notes != "" {
			body += fmt.Sprintf("<p><strong>Motivo:</strong> %s</p>", notes)
		}
		_ = s.emailClient.Send(*app.CustomerEmail, subject, body)
	}

	// Baixa automática de estoque se concluído
	if status == "completed" && s.stockService != nil {
		var serviceIDs []string
		for _, sVal := range app.Services {
			serviceIDs = append(serviceIDs, sVal.ServiceID)
		}
		_ = s.stockService.TriggerAutomaticDrop(ctx, clientID, id, serviceIDs)
	}

	// Acúmulo automático de fidelidade se concluído e pago
	if status == "completed" && s.loyaltyService != nil && app.CustomerID != nil && *app.CustomerID != "" {
		payStatus, payAmount, errPay := s.repo.GetPaymentStatus(ctx, clientID, id)
		if errPay == nil && payStatus == "paid" {
			_ = s.loyaltyService.TriggerAutomaticEarn(ctx, clientID, *app.CustomerID, id, payAmount)
		}
	}

	return nil
}

func (s *service) GetAvailability(ctx context.Context, slug string, professionalID string, serviceIDs []string, date string) ([]TimeSlot, error) {
	cfg, err := s.configRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	totalDuration := 0
	for _, sID := range serviceIDs {
		var svcDuration int
		if professionalID != "" {
			linked, err := s.profRepo.GetLinkedServices(ctx, cfg.ClientID, professionalID)
			if err == nil {
				for _, link := range linked {
					if link.ServiceID == sID && link.CustomDuration != nil {
						svcDuration = *link.CustomDuration
					}
				}
			}
		}

		if svcDuration == 0 {
			orig, err := s.svcRepo.GetByID(ctx, cfg.ClientID, sID)
			if err == nil {
				svcDuration = orig.DurationMinutes
			} else {
				svcDuration = 30
			}
		}
		totalDuration += svcDuration
	}

	if totalDuration == 0 {
		return nil, errors.New("duração total dos serviços é zero ou serviços inválidos")
	}

	tDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, errors.New("formato de data inválido, deve ser YYYY-MM-DD")
	}
	weekday := int(tDate.Weekday())

	profs, err := s.profRepo.List(ctx, cfg.ClientID, "active")
	profNameByID := make(map[string]string)
	if err == nil {
		for _, p := range profs {
			profNameByID[p.ID] = p.Name
		}
	}

	var targetProfs []string
	if professionalID != "" {
		targetProfs = append(targetProfs, professionalID)
	} else {
		for _, p := range profs {
			targetProfs = append(targetProfs, p.ID)
		}
	}

	uniqueSlotsMap := make(map[string]TimeSlot)

	for _, profID := range targetProfs {
		schedules, err := s.profRepo.GetSchedule(ctx, cfg.ClientID, profID)
		if err != nil {
			continue
		}

		var sched *professional.ProfessionalSchedule
		for i := range schedules {
			if schedules[i].Weekday == weekday && schedules[i].Enabled == 1 {
				sched = &schedules[i]
				break
			}
		}

		if sched == nil {
			continue
		}

		apps, err := s.repo.List(ctx, cfg.ClientID, profID, date, date)
		if err != nil {
			apps = []EnrichedAppointment{}
		}

		blocks, err := s.repo.ListBlockedSlots(ctx, cfg.ClientID, profID, date, date)
		if err != nil {
			blocks = []BlockedSlot{}
		}

		startT, _ := time.Parse("15:04:05", sched.StartTime)
		endT, _ := time.Parse("15:04:05", sched.EndTime)

		currT := startT
		for currT.Add(time.Duration(totalDuration) * time.Minute).Before(endT) || currT.Add(time.Duration(totalDuration)*time.Minute).Equal(endT) {
			slotEnd := currT.Add(time.Duration(totalDuration) * time.Minute)

			collision := false

			if cfg.BlockLunchEnabled == 1 {
				lunchStart, err1 := time.Parse("15:04:05", cfg.BlockLunchStart)
				lunchEnd, err2 := time.Parse("15:04:05", cfg.BlockLunchEnd)
				if err1 == nil && err2 == nil {
					if currT.Before(lunchEnd) && slotEnd.After(lunchStart) {
						collision = true
					}
				}
			}

			if !collision {
				for _, b := range blocks {
					bStart, _ := time.Parse("15:04:05", b.StartTime)
					bEnd, _ := time.Parse("15:04:05", b.EndTime)
					if currT.Before(bEnd) && slotEnd.After(bStart) {
						collision = true
						break
					}
				}
			}

			if !collision {
				for _, app := range apps {
					if app.Status == "cancelled" {
						continue
					}
					appStart, _ := time.Parse("15:04:05", app.StartTime)
					appEnd, _ := time.Parse("15:04:05", app.EndTime)
					interval := time.Duration(cfg.IntervalBetweenMinutes) * time.Minute

					if currT.Before(appEnd.Add(interval)) && slotEnd.Add(interval).After(appStart) {
						collision = true
						break
					}
				}
			}

			if !collision {
				loc, _ := time.LoadLocation(cfg.Timezone)
				if loc == nil {
					loc = time.Local
				}

				now := time.Now().In(loc)
				slotTime, _ := time.ParseInLocation("2006-01-02 15:04:05", date+" "+currT.Format("15:04:05"), loc)

				minAdvance := now.Add(time.Duration(cfg.MinAdvanceHours) * time.Hour)
				maxAdvance := now.Add(time.Duration(cfg.MaxAdvanceDays) * 24 * time.Hour)

				if slotTime.Before(minAdvance) || slotTime.After(maxAdvance) {
					collision = true
				}
			}

			if !collision {
				slotShowStart := currT.Format("15:04")
				slotShowEnd := slotEnd.Format("15:04")
				uniqueSlotsMap[slotShowStart] = TimeSlot{
					StartTime:        slotShowStart,
					EndTime:          slotShowEnd,
					ProfessionalID:   profID,
					ProfessionalName: profNameByID[profID],
				}
			}

			currT = currT.Add(30 * time.Minute)
		}
	}

	var sortedKeys []string
	for k := range uniqueSlotsMap {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	var result []TimeSlot
	for _, k := range sortedKeys {
		result = append(result, uniqueSlotsMap[k])
	}

	return result, nil
}

func (s *service) CreatePublic(ctx context.Context, slug string, req CreatePublicAppointmentRequest) (*EnrichedAppointment, error) {
	cfg, err := s.configRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if req.CustomerEmail != "" {
		if _, err := mail.ParseAddress(req.CustomerEmail); err != nil {
			return nil, errors.New("endereço de e-mail inválido")
		}
	}

	tDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("formato de data inválido, deve ser YYYY-MM-DD")
	}

	totalDuration := 0
	totalPrice := 0.0
	var appServices []AppointmentService

	for _, sID := range req.ServiceIDs {
		var svcDuration int
		var svcPrice float64

		linked, err := s.profRepo.GetLinkedServices(ctx, cfg.ClientID, req.ProfessionalID)
		if err == nil {
			for _, link := range linked {
				if link.ServiceID == sID {
					if link.CustomDuration != nil {
						svcDuration = *link.CustomDuration
					}
					if link.CustomPrice != nil {
						svcPrice = *link.CustomPrice
					}
				}
			}
		}

		if svcDuration == 0 || svcPrice == 0.0 {
			orig, err := s.svcRepo.GetByID(ctx, cfg.ClientID, sID)
			if err != nil {
				return nil, fmt.Errorf("serviço com ID %s não encontrado", sID)
			}
			if svcDuration == 0 {
				svcDuration = orig.DurationMinutes
			}
			if svcPrice == 0.0 {
				svcPrice = orig.Price
			}
		}

		totalDuration += svcDuration
		totalPrice += svcPrice

		appServices = append(appServices, AppointmentService{
			ID:              uuid.New().String(),
			ServiceID:       sID,
			Price:           svcPrice,
			DurationMinutes: svcDuration,
		})
	}

	startTimeParsed, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		startTimeParsed, err = time.Parse("15:04:05", req.StartTime)
		if err != nil {
			return nil, errors.New("formato de hora inicial inválido")
		}
	}

	endTimeParsed := startTimeParsed.Add(time.Duration(totalDuration) * time.Minute)
	startTimeStr := startTimeParsed.Format("15:04:00")
	endTimeStr := endTimeParsed.Format("15:04:00")

	availability, err := s.GetAvailability(ctx, slug, req.ProfessionalID, req.ServiceIDs, req.Date)
	if err != nil {
		return nil, err
	}

	slotAvailable := false
	reqTimeFormatted := startTimeParsed.Format("15:04")
	for _, slot := range availability {
		if slot.StartTime == reqTimeFormatted {
			slotAvailable = true
			break
		}
	}

	if !slotAvailable {
		return nil, ErrSlotNotAvailable
	}

	// 4.5. MERGE INTELIGENTE DE CLIENTES (FASE-06)
	var customerIDPtr *string
	customerID, err := s.custService.GetOrCreateForPortal(ctx, cfg.ClientID, req.CustomerName, req.CustomerPhone, req.CustomerEmail)
	if err == nil && customerID != "" {
		customerIDPtr = &customerID
	}

	appID := uuid.New().String()
	cancelToken := uuid.New().String()

	for i := range appServices {
		appServices[i].AppointmentID = appID
	}

	var customerEmailPtr *string
	if req.CustomerEmail != "" {
		customerEmailPtr = &req.CustomerEmail
	}

	app := &Appointment{
		ID:             appID,
		ClientID:       cfg.ClientID,
		ProfessionalID: req.ProfessionalID,
		CustomerID:     customerIDPtr,
		Date:           req.Date,
		StartTime:      startTimeStr,
		EndTime:        endTimeStr,
		Status:         "pending",
		Notes:          req.Notes,
		CancelToken:    &cancelToken,
		CustomerName:   &req.CustomerName,
		CustomerPhone:  &req.CustomerPhone,
		CustomerEmail:  customerEmailPtr,
		ReminderSent:   0,
		Source:         "online",
		CreatedAt:      time.Now(),
	}

	err = s.repo.Create(ctx, app, appServices)
	if err != nil {
		return nil, err
	}

	enriched, err := s.repo.GetByID(ctx, cfg.ClientID, appID)
	if err != nil {
		return enriched, nil
	}

	if s.emailClient != nil && req.CustomerEmail != "" {
		subject := fmt.Sprintf("Agendamento Confirmado - %s", cfg.ClientName)
		body := fmt.Sprintf(`
			<div style="font-family: sans-serif; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #eee; border-radius: 12px;">
				<h2 style="color: %s;">Seu agendamento foi confirmado!</h2>
				<p>Olá <strong>%s</strong>, tudo bem?</p>
				<p>Seu horário na barbearia <strong>%s</strong> está garantido.</p>
				<hr style="border: 0; border-top: 1px solid #eee; margin: 20px 0;" />
				<p><strong>Profissional:</strong> %s</p>
				<p><strong>Data:</strong> %s às %s</p>
				<p><strong>Preço Estimado:</strong> R$ %.2f</p>
				%s
				<hr style="border: 0; border-top: 1px solid #eee; margin: 20px 0;" />
				<p style="font-size: 12px; color: #666;">
					Precisa cancelar? Você pode fazer isso até %d horas antes do horário marcado clicando no link abaixo:
				</p>
				<p style="margin-top: 15px;">
					<a href="http://localhost:3000/agendamento/cancelar/%s" style="background-color: #dc2626; color: white; padding: 10px 18px; text-decoration: none; border-radius: 6px; font-weight: bold; font-size: 13px; display: inline-block;">
						Cancelar Agendamento
					</a>
				</p>
			</div>
		`, cfg.ColorPrimary, req.CustomerName, cfg.ClientName, enriched.ProfessionalName,
			tDate.Format("02/01/2006"), req.StartTime, totalPrice,
			func() string {
				if req.Notes != nil && *req.Notes != "" {
					return fmt.Sprintf("<p><strong>Observações:</strong> <em>%s</em></p>", *req.Notes)
				}
				return ""
			}(),
			cfg.CancellationPolicyHours, cancelToken)

		_ = s.emailClient.Send(req.CustomerEmail, subject, body)
	}

	// Envia WhatsApp automático de criação/recebimento
	if enriched.CustomerPhone != nil && *enriched.CustomerPhone != "" {
		dateFormatted := ""
		tDate, err := time.Parse("2006-01-02", enriched.Date)
		if err == nil {
			dateFormatted = tDate.Format("02/01/2006")
		} else {
			dateFormatted = enriched.Date
		}
		timeFormatted := enriched.StartTime[:5]
		servicesStr := ""
		for i, srv := range enriched.Services {
			if i > 0 {
				servicesStr += ", "
			}
			servicesStr += srv.ServiceName
		}
		
		msg := fmt.Sprintf("Olá, %s! Seu agendamento para o dia %s às %s foi realizado com sucesso.\n\nProfissional: %s\nServiços: %s",
			*enriched.CustomerName, dateFormatted, timeFormatted, enriched.ProfessionalName, servicesStr)
			
		_, _ = s.chatService.SendMessage(ctx, cfg.ClientID, chat.SendMessageRequest{
			ContactNumber: *enriched.CustomerPhone,
			Content:       msg,
		})
	}

	return enriched, nil
}

func (s *service) GetByCancelToken(ctx context.Context, token string) (*EnrichedAppointment, error) {
	return s.repo.GetByCancelToken(ctx, token)
}

func (s *service) CancelByToken(ctx context.Context, token string) error {
	app, err := s.repo.GetByCancelToken(ctx, token)
	if err != nil {
		return err
	}

	if app.Status == "cancelled" {
		return nil
	}

	cfg, err := s.configRepo.GetByClientID(ctx, app.ClientID)
	if err != nil {
		return err
	}

	loc, _ := time.LoadLocation(cfg.Timezone)
	if loc == nil {
		loc = time.Local
	}

	now := time.Now().In(loc)
	appTime, err := time.ParseInLocation("2006-01-02 15:04:05", app.Date+" "+app.StartTime, loc)
	if err == nil {
		limitTime := appTime.Add(-time.Duration(cfg.CancellationPolicyHours) * time.Hour)
		if now.After(limitTime) {
			return ErrCancelPolicy
		}
	}

	err = s.UpdateStatus(ctx, app.ClientID, app.ID, "cancelled", "public_cancel_token", "Cancelado pelo link público")
	return err
}

func (s *service) SendUpcomingReminders(ctx context.Context) (int, error) {
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	if loc == nil {
		loc = time.Local
	}

	tomorrow := time.Now().In(loc).Add(24 * time.Hour).Format("2006-01-02")
	list, err := s.repo.GetUpcomingAppointmentsForReminder(ctx, tomorrow)
	if err != nil {
		return 0, err
	}

	sentCount := 0
	for _, app := range list {
		if app.CustomerEmail == nil || *app.CustomerEmail == "" {
			continue
		}

		clientName, err := s.configRepo.GetClientName(ctx, app.ClientID)
		if err != nil {
			clientName = "Barbearia"
		}

		colorPrimary := "#1a1a1a"
		cfg, err := s.configRepo.GetByClientID(ctx, app.ClientID)
		if err == nil {
			colorPrimary = cfg.ColorPrimary
		}

		subject := fmt.Sprintf("Lembrete de Agendamento - %s", clientName)
		cancelToken := ""
		if app.CancelToken != nil {
			cancelToken = *app.CancelToken
		}

		body := fmt.Sprintf(`
			<div style="font-family: sans-serif; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #eee; border-radius: 12px;">
				<h2 style="color: %s;">Lembrete: Seu horário é amanhã!</h2>
				<p>Olá %s,</p>
				<p>Este é um lembrete automático do seu agendamento na barbearia <strong>%s</strong>.</p>
				<hr style="border: 0; border-top: 1px solid #eee; margin: 20px 0;" />
				<p><strong>Profissional:</strong> %s</p>
				<p><strong>Horário:</strong> Amanhã (%s) às %s</p>
				<hr style="border: 0; border-top: 1px solid #eee; margin: 20px 0;" />
				<p style="font-size: 11px; color: #999;">
					Se precisar cancelar ou remarcar, utilize o link de cancelamento abaixo:
				</p>
				<p>
					<a href="http://localhost:3000/agendamento/cancelar/%s" style="color: #dc2626; font-weight: bold; font-size: 12px;">
						Cancelar Agendamento
					</a>
				</p>
			</div>
		`, colorPrimary, *app.CustomerName, clientName, app.ProfessionalName,
			app.Date, app.StartTime, cancelToken)

		err = s.emailClient.Send(*app.CustomerEmail, subject, body)
		if err == nil {
			_ = s.repo.MarkReminderSent(ctx, app.ID)
			sentCount++
		}
	}

	return sentCount, nil
}

func (s *service) Create(ctx context.Context, clientID, userID string, req CreateAppointmentRequest) (*EnrichedAppointment, error) {
	cfg, err := s.configRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	totalDuration := 0
	totalPrice := 0.0
	var appServices []AppointmentService

	for _, sID := range req.ServiceIDs {
		var svcDuration int
		var svcPrice float64

		linked, err := s.profRepo.GetLinkedServices(ctx, clientID, req.ProfessionalID)
		if err == nil {
			for _, link := range linked {
				if link.ServiceID == sID {
					if link.CustomDuration != nil {
						svcDuration = *link.CustomDuration
					}
					if link.CustomPrice != nil {
						svcPrice = *link.CustomPrice
					}
				}
			}
		}

		if svcDuration == 0 || svcPrice == 0.0 {
			orig, err := s.svcRepo.GetByID(ctx, clientID, sID)
			if err != nil {
				return nil, fmt.Errorf("serviço com ID %s não encontrado", sID)
			}
			if svcDuration == 0 {
				svcDuration = orig.DurationMinutes
			}
			if svcPrice == 0.0 {
				svcPrice = orig.Price
			}
		}

		totalDuration += svcDuration
		totalPrice += svcPrice

		appServices = append(appServices, AppointmentService{
			ID:              uuid.New().String(),
			ServiceID:       sID,
			Price:           svcPrice,
			DurationMinutes: svcDuration,
		})
	}

	startTimeParsed, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		startTimeParsed, err = time.Parse("15:04:05", req.StartTime)
		if err != nil {
			return nil, errors.New("formato de hora inicial inválido")
		}
	}
	endTimeParsed := startTimeParsed.Add(time.Duration(totalDuration) * time.Minute)
	startTimeStr := startTimeParsed.Format("15:04:00")
	endTimeStr := endTimeParsed.Format("15:04:00")

	// Validação de Conflito de Horário
	apps, err := s.repo.List(ctx, clientID, req.ProfessionalID, req.Date, req.Date)
	if err == nil {
		for _, app := range apps {
			if app.Status == "cancelled" {
				continue
			}
			appStart, _ := time.Parse("15:04:05", app.StartTime)
			appEnd, _ := time.Parse("15:04:05", app.EndTime)
			interval := time.Duration(cfg.IntervalBetweenMinutes) * time.Minute

			if startTimeParsed.Before(appEnd.Add(interval)) && endTimeParsed.Add(interval).After(appStart) {
				return nil, ErrSlotNotAvailable
			}
		}
	}

	// Validação com Bloqueio de Almoço global
	if cfg.BlockLunchEnabled == 1 {
		lunchStart, err1 := time.Parse("15:04:05", cfg.BlockLunchStart)
		lunchEnd, err2 := time.Parse("15:04:05", cfg.BlockLunchEnd)
		if err1 == nil && err2 == nil {
			if startTimeParsed.Before(lunchEnd) && endTimeParsed.After(lunchStart) {
				return nil, ErrSlotNotAvailable
			}
		}
	}

	// Validação com Bloqueios
	blocks, err := s.repo.ListBlockedSlots(ctx, clientID, req.ProfessionalID, req.Date, req.Date)
	if err == nil {
		for _, b := range blocks {
			bStart, _ := time.Parse("15:04:05", b.StartTime)
			bEnd, _ := time.Parse("15:04:05", b.EndTime)
			if startTimeParsed.Before(bEnd) && endTimeParsed.After(bStart) {
				return nil, ErrSlotNotAvailable
			}
		}
	}

	// Merge por telefone no painel
	var customerID *string = req.CustomerID
	if customerID == nil && req.CustomerName != nil && *req.CustomerName != "" && req.CustomerPhone != nil && *req.CustomerPhone != "" {
		emailVal := ""
		if req.CustomerEmail != nil {
			emailVal = *req.CustomerEmail
		}
		cID, err := s.custService.GetOrCreateForPortal(ctx, clientID, *req.CustomerName, *req.CustomerPhone, emailVal)
		if err == nil {
			customerID = &cID
		}
	}

	appID := uuid.New().String()
	cancelToken := uuid.New().String()

	for i := range appServices {
		appServices[i].AppointmentID = appID
	}

	app := &Appointment{
		ID:             appID,
		ClientID:       clientID,
		ProfessionalID: req.ProfessionalID,
		CustomerID:     customerID,
		Date:           req.Date,
		StartTime:      startTimeStr,
		EndTime:        endTimeStr,
		Status:         "confirmed", // direto como confirmado no painel
		Notes:          req.Notes,
		CancelToken:    &cancelToken,
		CustomerName:   req.CustomerName,
		CustomerPhone:  req.CustomerPhone,
		CustomerEmail:  req.CustomerEmail,
		ReminderSent:   0,
		Source:         "panel",
		CreatedAt:      time.Now(),
	}

	err = s.repo.Create(ctx, app, appServices)
	if err != nil {
		return nil, err
	}

	// Cria status log
	logNotes := "Criado pelo painel administrativo"
	statusLog := &AppointmentStatusLog{
		ID:            uuid.New().String(),
		AppointmentID: appID,
		FromStatus:    nil,
		ToStatus:      "confirmed",
		ChangedBy:     userID,
		Notes:         &logNotes,
		CreatedAt:     time.Now(),
	}
	_ = s.repo.CreateStatusLog(ctx, statusLog)

	enriched, err := s.repo.GetByID(ctx, clientID, appID)
	if err != nil {
		return enriched, nil
	}

	// Envia email de confirmação
	if s.emailClient != nil && req.CustomerEmail != nil && *req.CustomerEmail != "" {
		clientName, err := s.configRepo.GetClientName(ctx, clientID)
		if err != nil {
			clientName = "Barbearia"
		}
		subject := fmt.Sprintf("Agendamento Confirmado - %s", clientName)
		body := fmt.Sprintf("<p>Olá %s, seu agendamento para o dia %s às %s foi agendado com sucesso pelo painel.</p>",
			*req.CustomerName, req.Date, req.StartTime)
		_ = s.emailClient.Send(*req.CustomerEmail, subject, body)
	}

	return enriched, nil
}

func (s *service) Update(ctx context.Context, clientID, id, userID string, req UpdateAppointmentRequest) (*EnrichedAppointment, error) {
	app, err := s.repo.GetByID(ctx, clientID, id)
	if err != nil {
		return nil, err
	}

	cfg, err := s.configRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	totalDuration := 0
	for _, sSvc := range app.Services {
		totalDuration += sSvc.Duration
	}

	startTimeParsed, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		startTimeParsed, err = time.Parse("15:04:05", req.StartTime)
		if err != nil {
			return nil, errors.New("formato de hora inicial inválido")
		}
	}
	endTimeParsed := startTimeParsed.Add(time.Duration(totalDuration) * time.Minute)
	startTimeStr := startTimeParsed.Format("15:04:00")
	endTimeStr := endTimeParsed.Format("15:04:00")

	// Conflito
	apps, err := s.repo.List(ctx, clientID, req.ProfessionalID, req.Date, req.Date)
	if err == nil {
		for _, other := range apps {
			if other.ID == id || other.Status == "cancelled" {
				continue
			}
			appStart, _ := time.Parse("15:04:05", other.StartTime)
			appEnd, _ := time.Parse("15:04:05", other.EndTime)
			interval := time.Duration(cfg.IntervalBetweenMinutes) * time.Minute

			if startTimeParsed.Before(appEnd.Add(interval)) && endTimeParsed.Add(interval).After(appStart) {
				return nil, ErrSlotNotAvailable
			}
		}
	}

	// Validação com Bloqueio de Almoço global
	if cfg.BlockLunchEnabled == 1 {
		lunchStart, err1 := time.Parse("15:04:05", cfg.BlockLunchStart)
		lunchEnd, err2 := time.Parse("15:04:05", cfg.BlockLunchEnd)
		if err1 == nil && err2 == nil {
			if startTimeParsed.Before(lunchEnd) && endTimeParsed.After(lunchStart) {
				return nil, ErrSlotNotAvailable
			}
		}
	}

	// Bloqueios
	blocks, err := s.repo.ListBlockedSlots(ctx, clientID, req.ProfessionalID, req.Date, req.Date)
	if err == nil {
		for _, b := range blocks {
			bStart, _ := time.Parse("15:04:05", b.StartTime)
			bEnd, _ := time.Parse("15:04:05", b.EndTime)
			if startTimeParsed.Before(bEnd) && endTimeParsed.After(bStart) {
				return nil, ErrSlotNotAvailable
			}
		}
	}

	oldDate := app.Date
	oldTime := app.StartTime

	app.ProfessionalID = req.ProfessionalID
	app.Date = req.Date
	app.StartTime = startTimeStr
	app.EndTime = endTimeStr

	var appServices []AppointmentService
	for _, sSvc := range app.Services {
		appServices = append(appServices, AppointmentService{
			ID:              uuid.New().String(),
			AppointmentID:   app.ID,
			ServiceID:       sSvc.ServiceID,
			Price:           sSvc.Price,
			DurationMinutes: sSvc.Duration,
		})
	}

	err = s.repo.Update(ctx, &app.Appointment, appServices)
	if err != nil {
		return nil, err
	}

	// Log
	logNotes := fmt.Sprintf("Reagendado: de %s %s para %s %s", oldDate, oldTime[:5], req.Date, req.StartTime[:5])
	statusLog := &AppointmentStatusLog{
		ID:            uuid.New().String(),
		AppointmentID: app.ID,
		FromStatus:    &app.Status,
		ToStatus:      app.Status,
		ChangedBy:     userID,
		Notes:         &logNotes,
		CreatedAt:     time.Now(),
	}
	_ = s.repo.CreateStatusLog(ctx, statusLog)

	enriched, err := s.repo.GetByID(ctx, clientID, app.ID)
	if err != nil {
		return enriched, nil
	}

	if s.emailClient != nil && app.CustomerEmail != nil && *app.CustomerEmail != "" {
		subject := "Horário de Agendamento Alterado"
		body := fmt.Sprintf("<p>Olá %s, seu agendamento foi remarcado.</p><p><strong>Novo Horário:</strong> %s às %s</p>",
			*app.CustomerName, req.Date, req.StartTime)
		_ = s.emailClient.Send(*app.CustomerEmail, subject, body)
	}

	return enriched, nil
}

func (s *service) Cancel(ctx context.Context, clientID, id, userID, reason string) error {
	return s.UpdateStatus(ctx, clientID, id, "cancelled", userID, reason)
}

func (s *service) GetStatusLogs(ctx context.Context, clientID, appointmentID string) ([]AppointmentStatusLog, error) {
	return s.repo.GetStatusLogs(ctx, clientID, appointmentID)
}

func (s *service) GetAvailabilityInternal(ctx context.Context, clientID, professionalID string, serviceIDs []string, date string) ([]TimeSlot, error) {
	slug, err := s.configRepo.GetClientSlug(ctx, clientID)
	if err != nil {
		return nil, err
	}
	return s.GetAvailability(ctx, slug, professionalID, serviceIDs, date)
}
