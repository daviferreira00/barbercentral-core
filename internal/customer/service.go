package customer

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCustomerNotFound = errors.New("cliente não encontrado")
	ErrCPFInvalid       = errors.New("o CPF informado é inválido")
	ErrPhoneDuplicate   = errors.New("já existe um cliente cadastrado com este telefone")
)

type Service interface {
	Create(ctx context.Context, clientID string, req CreateCustomerRequest) (*Customer, error)
	Update(ctx context.Context, clientID, id string, req UpdateCustomerRequest) (*Customer, error)
	Delete(ctx context.Context, clientID, id string) error
	GetByID(ctx context.Context, clientID, id string) (*CustomerStats, error)
	List(ctx context.Context, clientID string, searchQuery string, birthMonth int, page, pageSize int) ([]CustomerStats, int, error)
	Search(ctx context.Context, clientID string, searchQuery string) ([]Customer, error)
	GetAppointmentsHistory(ctx context.Context, clientID, customerID string) ([]EnrichedAppointmentHistory, error)
	ExportCSV(ctx context.Context, clientID string) ([]byte, error)

	// Lógica de merge e criação para fluxo do portal
	GetOrCreateForPortal(ctx context.Context, clientID string, name, phone, email string) (string, error)
}

type service struct {
	repo CustomerRepository
}

func NewService(repo CustomerRepository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, clientID string, req CreateCustomerRequest) (*Customer, error) {
	// 1. Valida CPF
	if req.CPF != nil && *req.CPF != "" {
		if !ValidateCPF(*req.CPF) {
			return nil, ErrCPFInvalid
		}
	}

	// 2. Valida Telefone Duplicado
	existing, err := s.repo.GetByPhone(ctx, clientID, req.Phone)
	if err == nil && existing != nil {
		return nil, ErrPhoneDuplicate
	}

	cust := &Customer{
		ID:        uuid.New().String(),
		ClientID:  clientID,
		Name:      req.Name,
		Phone:     req.Phone,
		Email:     req.Email,
		CPF:       req.CPF,
		BirthDate: req.BirthDate,
		Notes:     req.Notes,
		CreatedAt: time.Now(),
	}

	err = s.repo.Create(ctx, cust)
	if err != nil {
		return nil, err
	}

	return cust, nil
}

func (s *service) Update(ctx context.Context, clientID, id string, req UpdateCustomerRequest) (*Customer, error) {
	// 1. Valida CPF
	if req.CPF != nil && *req.CPF != "" {
		if !ValidateCPF(*req.CPF) {
			return nil, ErrCPFInvalid
		}
	}

	// 2. Valida Telefone Duplicado para outro ID
	existing, err := s.repo.GetByPhone(ctx, clientID, req.Phone)
	if err == nil && existing != nil && existing.ID != id {
		return nil, ErrPhoneDuplicate
	}

	// 3. Busca atual
	stats, err := s.repo.GetByID(ctx, clientID, id)
	if err != nil {
		return nil, ErrCustomerNotFound
	}

	cust := &stats.Customer
	cust.Name = req.Name
	cust.Phone = req.Phone
	cust.Email = req.Email
	cust.CPF = req.CPF
	cust.BirthDate = req.BirthDate
	cust.Notes = req.Notes

	err = s.repo.Update(ctx, cust)
	if err != nil {
		return nil, err
	}

	return cust, nil
}

func (s *service) Delete(ctx context.Context, clientID, id string) error {
	return s.repo.Delete(ctx, clientID, id)
}

func (s *service) GetByID(ctx context.Context, clientID, id string) (*CustomerStats, error) {
	stats, err := s.repo.GetByID(ctx, clientID, id)
	if err != nil {
		return nil, ErrCustomerNotFound
	}
	return stats, nil
}

func (s *service) List(ctx context.Context, clientID string, searchQuery string, birthMonth int, page, pageSize int) ([]CustomerStats, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	return s.repo.List(ctx, clientID, searchQuery, birthMonth, pageSize, offset)
}

func (s *service) Search(ctx context.Context, clientID string, searchQuery string) ([]Customer, error) {
	if len(searchQuery) < 3 {
		return []Customer{}, nil
	}
	return s.repo.Search(ctx, clientID, searchQuery)
}

func (s *service) GetAppointmentsHistory(ctx context.Context, clientID, customerID string) ([]EnrichedAppointmentHistory, error) {
	return s.repo.GetAppointmentsHistory(ctx, clientID, customerID)
}

func (s *service) ExportCSV(ctx context.Context, clientID string) ([]byte, error) {
	// Carrega todos os clientes sem limite
	list, _, err := s.repo.List(ctx, clientID, "", 0, 10000, 0)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	// Adiciona o BOM UTF-8 para o Excel reconhecer os acentos
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(buf)
	w.Comma = ';' // Utilizar ponto e vírgula que é mais comum no Excel brasileiro

	// Escreve header
	header := []string{
		"Nome", "Telefone", "E-mail", "CPF", "Data de nascimento", "Total de visitas", "Última visita", "Primeira visita",
	}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	for _, c := range list {
		emailStr := ""
		if c.Email != nil {
			emailStr = *c.Email
		}
		cpfStr := ""
		if c.CPF != nil {
			cpfStr = *c.CPF
		}
		birthStr := ""
		if c.BirthDate != nil {
			birthStr = *c.BirthDate
		}
		lastStr := ""
		if c.LastVisit != nil {
			lastStr = *c.LastVisit
		}
		firstStr := ""
		if c.FirstVisit != nil {
			firstStr = *c.FirstVisit
		}

		row := []string{
			c.Name,
			c.Phone,
			emailStr,
			cpfStr,
			birthStr,
			strconv.Itoa(c.TotalVisits),
			lastStr,
			firstStr,
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}

	w.Flush()
	return buf.Bytes(), nil
}

func (s *service) GetOrCreateForPortal(ctx context.Context, clientID string, name, phone, email string) (string, error) {
	// 1. Tenta buscar por telefone
	existing, err := s.repo.GetByPhone(ctx, clientID, phone)
	if err != nil {
		return "", err
	}

	if existing != nil {
		return existing.ID, nil
	}

	// 2. Não existe, cria um novo
	cust := &Customer{
		ID:        uuid.New().String(),
		ClientID:  clientID,
		Name:      name,
		Phone:     phone,
		Email:     &email,
		CPF:       nil,
		BirthDate: nil,
		Notes:     nil,
		CreatedAt: time.Now(),
	}

	err = s.repo.Create(ctx, cust)
	if err != nil {
		return "", err
	}

	return cust.ID, nil
}

// Helper para validação de CPF (Algoritmo Oficial)
func ValidateCPF(cpf string) bool {
	// Remove caracteres não-numéricos
	re := regexp.MustCompile(`[^\d]`)
	cpf = re.ReplaceAllString(cpf, "")

	if len(cpf) != 11 {
		return false
	}

	// Elimina CPFs conhecidos com todos os números iguais
	var allEqual = true
	for i := 1; i < 11; i++ {
		if cpf[i] != cpf[0] {
			allEqual = false
			break
		}
	}
	if allEqual {
		return false
	}

	// Validação do primeiro dígito
	var sum int
	for i := 0; i < 9; i++ {
		val, _ := strconv.Atoi(string(cpf[i]))
		sum += val * (10 - i)
	}
	rest := sum % 11
	d1 := 0
	if rest >= 2 {
		d1 = 11 - rest
	}
	valD1, _ := strconv.Atoi(string(cpf[9]))
	if valD1 != d1 {
		return false
	}

	// Validação do segundo dígito
	sum = 0
	for i := 0; i < 10; i++ {
		val, _ := strconv.Atoi(string(cpf[i]))
		sum += val * (11 - i)
	}
	rest = sum % 11
	d2 := 0
	if rest >= 2 {
		d2 = 11 - rest
	}
	valD2, _ := strconv.Atoi(string(cpf[10]))
	return valD2 == d2
}
