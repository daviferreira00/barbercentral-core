package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"barbercentral-core/internal/planlimit"
)

var ErrEmailBelongsToAdmin = errors.New("email_already_exists: este e-mail já pertence a um administrador da plataforma")

type AdminService interface {
	ListClients(ctx context.Context) ([]Client, error)
	GetClientByID(ctx context.Context, id string) (*Client, error)
	CreateClient(ctx context.Context, req CreateClientRequest) (*Client, error)
	UpdateClient(ctx context.Context, id string, req UpdateClientRequest) (*Client, error)
	BlockClient(ctx context.Context, id string) error
	UnblockClient(ctx context.Context, id string) error
	BlockClientWithReason(ctx context.Context, id, reason, performedBy string) error
	UnblockClientWithReason(ctx context.Context, id, reason, performedBy string) error
	UpdateClientPlan(ctx context.Context, id, planID string) error
	ListClientUsers(ctx context.Context, clientID string) ([]ClientUser, error)
	CreateClientUser(ctx context.Context, clientID string, req CreateClientUserRequest) (*ClientUser, error)
	UpdateClientUser(ctx context.Context, userID string, req UpdateClientUserRequest) error
	DeleteClientUser(ctx context.Context, userID string) error

	ListPlans(ctx context.Context) ([]planlimit.Plan, error)
	CreatePlan(ctx context.Context, p *planlimit.Plan) (*planlimit.Plan, error)
	UpdatePlan(ctx context.Context, id string, p *planlimit.Plan) (*planlimit.Plan, error)

	ListAllUsers(ctx context.Context) ([]AdminUserResponse, error)
	CreateUser(ctx context.Context, req CreateUserRequest) (interface{}, error)
	UpdateUser(ctx context.Context, userType string, id string, req UpdateUserRequest) error
	DeleteUser(ctx context.Context, userType string, id string) error
}

type adminService struct {
	repo AdminRepository
}

func NewAdminService(repo AdminRepository) AdminService {
	return &adminService{repo: repo}
}

func (s *adminService) ListClients(ctx context.Context) ([]Client, error) {
	return s.repo.ListClients(ctx)
}

func (s *adminService) GetClientByID(ctx context.Context, id string) (*Client, error) {
	return s.repo.GetClientByID(ctx, id)
}

func (s *adminService) CreateClient(ctx context.Context, req CreateClientRequest) (*Client, error) {
	c := &Client{
		ID:           uuid.New().String(),
		PlanID:       req.PlanID,
		Name:         req.Name,
		Slug:         req.Slug,
		CustomDomain: req.CustomDomain,
		Status:       "active",
		CreatedAt:    time.Now(),
	}

	err := s.repo.CreateClient(ctx, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s *adminService) UpdateClient(ctx context.Context, id string, req UpdateClientRequest) (*Client, error) {
	c, err := s.repo.GetClientByID(ctx, id)
	if err != nil {
		return nil, err
	}

	c.PlanID = req.PlanID
	c.Name = req.Name
	c.Slug = req.Slug
	c.CustomDomain = req.CustomDomain

	err = s.repo.UpdateClient(ctx, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s *adminService) BlockClient(ctx context.Context, id string) error {
	return s.repo.UpdateClientStatus(ctx, id, "blocked")
}

func (s *adminService) UnblockClient(ctx context.Context, id string) error {
	return s.repo.UpdateClientStatus(ctx, id, "active")
}

func (s *adminService) BlockClientWithReason(ctx context.Context, id, reason, performedBy string) error {
	err := s.repo.UpdateClientStatus(ctx, id, "blocked")
	if err != nil {
		return err
	}
	logID := uuid.New().String()
	return s.repo.CreateBlockLog(ctx, logID, id, "block", reason, performedBy)
}

func (s *adminService) UnblockClientWithReason(ctx context.Context, id, reason, performedBy string) error {
	err := s.repo.UpdateClientStatus(ctx, id, "active")
	if err != nil {
		return err
	}
	logID := uuid.New().String()
	return s.repo.CreateBlockLog(ctx, logID, id, "unblock", reason, performedBy)
}

func (s *adminService) UpdateClientPlan(ctx context.Context, id, planID string) error {
	return s.repo.UpdateClientPlan(ctx, id, planID)
}

func (s *adminService) ListClientUsers(ctx context.Context, clientID string) ([]ClientUser, error) {
	return s.repo.ListClientUsers(ctx, clientID)
}

// CreateClientUser vincula um usuário a uma barbearia. Se já existir uma
// identidade (user_account) com o e-mail informado, apenas cria o vínculo
// novo (client_user_link) — não duplica o usuário. Só cria identidade nova
// quando o e-mail é inédito na plataforma.
func (s *adminService) CreateClientUser(ctx context.Context, clientID string, req CreateClientUserRequest) (*ClientUser, error) {
	allowed, curr, max, err := planlimit.CheckUsersLimit(ctx, s.repo.GetDB(), clientID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("plan_limit_exceeded:max_users:%d:%d", curr, max)
	}

	isAdminEmail, err := s.repo.EmailExistsAsAdmin(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if isAdminEmail {
		return nil, ErrEmailBelongsToAdmin
	}

	existing, err := s.repo.FindUserAccountByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return s.repo.LinkExistingUser(ctx, clientID, existing.ID, req.Role)
	}

	// Senha padrão local para testes/fallback
	pwd := req.Password
	if pwd == "" {
		pwd = "senha_padrao_local"
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return s.repo.CreateUserWithLink(ctx, clientID, req, string(hashBytes))
}

// UpdateClientUser/DeleteClientUser operam por link_id (id de client_user_link),
// não mais por user_id — um mesmo usuário pode ter vários vínculos.
func (s *adminService) UpdateClientUser(ctx context.Context, linkID string, req UpdateClientUserRequest) error {
	return s.repo.UpdateClientUserLink(ctx, linkID, req)
}

func (s *adminService) DeleteClientUser(ctx context.Context, linkID string) error {
	return s.repo.DeleteClientUserLink(ctx, linkID)
}

func (s *adminService) ListPlans(ctx context.Context) ([]planlimit.Plan, error) {
	return s.repo.ListPlans(ctx)
}

func (s *adminService) CreatePlan(ctx context.Context, p *planlimit.Plan) (*planlimit.Plan, error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	err := s.repo.CreatePlan(ctx, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *adminService) UpdatePlan(ctx context.Context, id string, p *planlimit.Plan) (*planlimit.Plan, error) {
	p.ID = id
	err := s.repo.UpdatePlan(ctx, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *adminService) ListAllUsers(ctx context.Context) ([]AdminUserResponse, error) {
	return s.repo.ListAllUsers(ctx)
}

// CreateUser (endpoint genérico /admin/users) — para "client", segue a mesma
// regra de achar-ou-criar por e-mail do CreateClientUser.
func (s *adminService) CreateUser(ctx context.Context, req CreateUserRequest) (interface{}, error) {
	if req.Type == "admin" {
		exists, err := s.repo.CheckEmailExists(ctx, req.Email, "")
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("email_already_exists: O e-mail já está em uso na plataforma")
		}

		pwd := req.Password
		if pwd == "" {
			pwd = "senha_padrao_local"
		}
		hashBytes, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}

		id := uuid.New().String()
		err = s.repo.CreateAdminUser(ctx, id, req.Name, req.Email, string(hashBytes))
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"id":    id,
			"type":  "admin",
			"name":  req.Name,
			"email": req.Email,
		}, nil
	} else if req.Type == "client" {
		if req.ClientID == "" {
			return nil, fmt.Errorf("client_id é obrigatório para usuários de barbearia")
		}

		isAdminEmail, err := s.repo.EmailExistsAsAdmin(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if isAdminEmail {
			return nil, ErrEmailBelongsToAdmin
		}

		role := req.Role
		if role == "" {
			role = "owner"
		}

		existing, err := s.repo.FindUserAccountByEmail(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return s.repo.LinkExistingUser(ctx, req.ClientID, existing.ID, role)
		}

		allowed, curr, max, err := planlimit.CheckUsersLimit(ctx, s.repo.GetDB(), req.ClientID)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, fmt.Errorf("plan_limit_exceeded:max_users:%d:%d", curr, max)
		}

		pwd := req.Password
		if pwd == "" {
			pwd = "senha_padrao_local"
		}
		hashBytes, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}

		return s.repo.CreateUserWithLink(ctx, req.ClientID, CreateClientUserRequest{
			Name: req.Name, Email: req.Email, Role: role, Password: req.Password,
		}, string(hashBytes))
	}

	return nil, fmt.Errorf("tipo de usuário inválido: %s", req.Type)
}

// UpdateUser: para "client", id é o link_id (não mais o id do usuário) —
// name/email/password afetam a identidade global; role/status, só o vínculo.
func (s *adminService) UpdateUser(ctx context.Context, userType string, id string, req UpdateUserRequest) error {
	var passwordHash string
	if req.Password != "" {
		hashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		passwordHash = string(hashBytes)
	}

	if userType == "admin" {
		exists, err := s.repo.CheckEmailExists(ctx, req.Email, id)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("email_already_exists: O e-mail já está em uso na plataforma")
		}
		return s.repo.UpdateAdminUser(ctx, id, req.Name, req.Email, passwordHash)
	} else if userType == "client" {
		role := req.Role
		if role == "" {
			role = "owner"
		}
		status := req.Status
		if status == "" {
			status = "active"
		}

		userID, err := s.repo.GetUserIDByLinkID(ctx, id)
		if err != nil {
			return err
		}
		exists, err := s.repo.CheckEmailExists(ctx, req.Email, userID)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("email_already_exists: O e-mail já está em uso na plataforma")
		}

		return s.repo.UpdateClientUserLink(ctx, id, UpdateClientUserRequest{
			Name:         req.Name,
			Email:        req.Email,
			Role:         role,
			Status:       status,
			PasswordHash: passwordHash,
		})
	}

	return fmt.Errorf("tipo de usuário inválido: %s", userType)
}

func (s *adminService) DeleteUser(ctx context.Context, userType string, id string) error {
	if userType == "admin" {
		return s.repo.DeleteAdminUser(ctx, id)
	} else if userType == "client" {
		return s.repo.DeleteClientUserLink(ctx, id)
	}
	return fmt.Errorf("tipo de usuário inválido: %s", userType)
}
