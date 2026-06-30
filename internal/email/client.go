package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

type SendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

type Client struct {
	apiKey    string
	fromEmail string
}

func NewClient(apiKey, fromEmail string) *Client {
	return &Client{
		apiKey:    apiKey,
		fromEmail: fromEmail,
	}
}

func (c *Client) FormatFrom(name string) string {
	if name == "" {
		name = "BarberCentral"
	}
	if c.fromEmail == "" {
		return fmt.Sprintf("%s <no-reply@barbercentral.com.br>", name)
	}
	return fmt.Sprintf("%s <%s>", name, c.fromEmail)
}

func (c *Client) Send(to, subject, html string) error {
	// Se a API Key não estiver preenchida (ambiente dev), printamos no log e pulamos chamada de rede
	if c.apiKey == "" || c.apiKey == "re_dev_key" {
		log.Info().
			Str("to", to).
			Str("subject", subject).
			Msg("[DEV EMAIL] Envio de e-mail simulado (sem RESEND_API_KEY)")
		return nil
	}

	reqBody := SendRequest{
		From:    c.FormatFrom("BarberCentral"),
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *Client) SendWelcomeEmail(to, name, resetLink string) error {
	subject := "Bem-vindo ao BarberCentral — defina sua senha"
	html := `
	<div style="font-family: Arial, sans-serif; background-color: #f8fafc; padding: 40px 0; color: #1e293b; margin: 0;">
		<div style="max-width: 500px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1); border: 1px solid #e2e8f0; overflow: hidden;">
			<div style="background-color: #1e1b4b; padding: 24px; text-align: center;">
				<h1 style="color: #ffffff; margin: 0; font-size: 24px; font-weight: bold; letter-spacing: -0.025em;">BarberCentral</h1>
			</div>
			<div style="padding: 32px 24px;">
				<h2 style="margin-top: 0; font-size: 20px; font-weight: 700; color: #1e1b4b;">Seja muito bem-vindo!</h2>
				<p style="font-size: 15px; line-height: 1.6; color: #475569;">Olá ` + name + `,</p>
				<p style="font-size: 15px; line-height: 1.6; color: #475569;">Sua conta de acesso ao painel do BarberCentral foi criada com sucesso pelo administrador. Clique no botão abaixo para definir sua senha de acesso inicial:</p>
				<div style="text-align: center; margin: 32px 0 24px;">
					<a href="` + resetLink + `" style="background-color: #6366f1; color: #ffffff; padding: 12px 24px; font-size: 15px; font-weight: 600; text-decoration: none; border-radius: 8px; display: inline-block; box-shadow: 0 4px 6px -1px rgba(99,102,241,0.3);">Definir minha Senha</a>
				</div>
				<p style="font-size: 13px; line-height: 1.5; color: #94a3b8;">Por motivos de segurança, este link de definição expira em 24 horas.</p>
			</div>
			<div style="background-color: #f1f5f9; padding: 16px; text-align: center; border-top: 1px solid #e2e8f0; font-size: 12px; color: #64748b;">
				&copy; 2026 BarberCentral. Todos os direitos reservados.
			</div>
		</div>
	</div>`
	return c.Send(to, subject, html)
}

func (c *Client) SendResetPasswordEmail(to, name, resetLink string) error {
	subject := "Redefinir sua senha — BarberCentral"
	html := `
	<div style="font-family: Arial, sans-serif; background-color: #f8fafc; padding: 40px 0; color: #1e293b; margin: 0;">
		<div style="max-width: 500px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1); border: 1px solid #e2e8f0; overflow: hidden;">
			<div style="background-color: #1e1b4b; padding: 24px; text-align: center;">
				<h1 style="color: #ffffff; margin: 0; font-size: 24px; font-weight: bold; letter-spacing: -0.025em;">BarberCentral</h1>
			</div>
			<div style="padding: 32px 24px;">
				<h2 style="margin-top: 0; font-size: 20px; font-weight: 700; color: #1e1b4b;">Recuperação de Senha</h2>
				<p style="font-size: 15px; line-height: 1.6; color: #475569;">Olá ` + name + `,</p>
				<p style="font-size: 15px; line-height: 1.6; color: #475569;">Recebemos uma solicitação para redefinir a senha da sua conta no BarberCentral. Clique no botão abaixo para escolher uma nova senha:</p>
				<div style="text-align: center; margin: 32px 0 24px;">
					<a href="` + resetLink + `" style="background-color: #6366f1; color: #ffffff; padding: 12px 24px; font-size: 15px; font-weight: 600; text-decoration: none; border-radius: 8px; display: inline-block; box-shadow: 0 4px 6px -1px rgba(99,102,241,0.3);">Redefinir Senha</a>
				</div>
				<p style="font-size: 13px; line-height: 1.5; color: #94a3b8;">Por motivos de segurança, este link expira em 1 hora. Se você não solicitou a redefinição de senha, nenhuma ação é necessária.</p>
			</div>
			<div style="background-color: #f1f5f9; padding: 16px; text-align: center; border-top: 1px solid #e2e8f0; font-size: 12px; color: #64748b;">
				&copy; 2026 BarberCentral. Todos os direitos reservados.
			</div>
		</div>
	</div>`
	return c.Send(to, subject, html)
}

func (c *Client) SendMagicLinkEmail(to, name, verifyLink string) error {
	subject := "Seu link de acesso — BarberCentral"
	html := `
	<div style="font-family: Arial, sans-serif; background-color: #f8fafc; padding: 40px 0; color: #1e293b; margin: 0;">
		<div style="max-width: 500px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1); border: 1px solid #e2e8f0; overflow: hidden;">
			<div style="background-color: #1e1b4b; padding: 24px; text-align: center;">
				<h1 style="color: #ffffff; margin: 0; font-size: 24px; font-weight: bold; letter-spacing: -0.025em;">BarberCentral</h1>
			</div>
			<div style="padding: 32px 24px;">
				<h2 style="margin-top: 0; font-size: 20px; font-weight: 700; color: #1e1b4b;">Acesso Rápido</h2>
				<p style="font-size: 15px; line-height: 1.6; color: #475569;">Olá ` + name + `,</p>
				<p style="font-size: 15px; line-height: 1.6; color: #475569;">Clique no botão abaixo para acessar sua conta diretamente no BarberCentral sem precisar de senha:</p>
				<div style="text-align: center; margin: 32px 0 24px;">
					<a href="` + verifyLink + `" style="background-color: #6366f1; color: #ffffff; padding: 12px 24px; font-size: 15px; font-weight: 600; text-decoration: none; border-radius: 8px; display: inline-block; box-shadow: 0 4px 6px -1px rgba(99,102,241,0.3);">Entrar no Painel</a>
				</div>
				<p style="font-size: 13px; line-height: 1.5; color: #94a3b8;">Este link de acesso rápido expira em 15 minutos. Se você não solicitou este acesso, apenas ignore este e-mail.</p>
			</div>
			<div style="background-color: #f1f5f9; padding: 16px; text-align: center; border-top: 1px solid #e2e8f0; font-size: 12px; color: #64748b;">
				&copy; 2026 BarberCentral. Todos os direitos reservados.
			</div>
		</div>
	</div>`
	return c.Send(to, subject, html)
}
