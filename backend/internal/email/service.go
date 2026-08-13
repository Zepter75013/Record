package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
)

type EmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUser     string
	smtpPassword string
	fromEmail    string
	fromName     string
}

func NewEmailService() *EmailService {
	return &EmailService{
		smtpHost:     os.Getenv("SMTP_HOST"),
		smtpPort:     os.Getenv("SMTP_PORT"),
		smtpUser:     os.Getenv("SMTP_USER"),
		smtpPassword: os.Getenv("SMTP_PASSWORD"),
		fromEmail:    os.Getenv("SMTP_FROM"),
		fromName:     os.Getenv("SMTP_FROM_NAME"),
	}
}

func (s *EmailService) SendPasswordResetEmail(toEmail, resetToken string) error {
	frontendURL := os.Getenv("FRONTEND_URL")
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, resetToken)

	subject := "Réinitialisation de votre mot de passe"

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #4b3d7a 0%, #d87d3a 100%); color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 8px 8px; }
        .button { display: inline-block; padding: 12px 30px; background: #d87d3a; color: white; text-decoration: none; border-radius: 8px; font-weight: bold; margin: 20px 0; }
        .footer { text-align: center; margin-top: 20px; font-size: 0.9em; color: #666; }
        .warning { background: #fff3cd; border-left: 4px solid #ffc107; padding: 12px; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔒 Réinitialisation de mot de passe</h1>
        </div>
        <div class="content">
            <p>Bonjour,</p>
            <p>Vous avez demandé à réinitialiser votre mot de passe pour <strong>Vinyl Manager</strong>.</p>
            <p>Cliquez sur le bouton ci-dessous pour créer un nouveau mot de passe :</p>
            
            <div style="text-align: center;">
                <a href="{{.ResetLink}}" class="button">Réinitialiser mon mot de passe</a>
            </div>

            <div class="warning">
                ⚠️ <strong>Important :</strong> Ce lien est valable pendant <strong>1 heure</strong> uniquement.
            </div>

            <p>Si vous n'êtes pas à l'origine de cette demande, vous pouvez ignorer cet email en toute sécurité.</p>
            
            <p style="margin-top: 30px; font-size: 0.9em; color: #666;">
                Si le bouton ne fonctionne pas, copiez-collez ce lien dans votre navigateur :<br>
                <code style="background: #e9ecef; padding: 5px 10px; border-radius: 4px; display: inline-block; margin-top: 8px;">{{.ResetLink}}</code>
            </p>
        </div>
        <div class="footer">
            <p>© 2026 Vinyl Manager - Gestionnaire de collection de disques</p>
        </div>
    </div>
</body>
</html>
`

	var body bytes.Buffer
	t := template.Must(template.New("email").Parse(tmpl))
	err := t.Execute(&body, map[string]string{
		"ResetLink": resetLink,
	})
	if err != nil {
		return fmt.Errorf("erreur template email: %w", err)
	}

	message := fmt.Sprintf("From: %s <%s>\r\n", s.fromName, s.fromEmail)
	message += fmt.Sprintf("To: %s\r\n", toEmail)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "MIME-Version: 1.0\r\n"
	message += "Content-Type: text/html; charset=UTF-8\r\n"
	message += "\r\n"
	message += body.String()

	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPassword, s.smtpHost)
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)

	err = smtp.SendMail(addr, auth, s.fromEmail, []string{toEmail}, []byte(message))
	if err != nil {
		return fmt.Errorf("erreur envoi email: %w", err)
	}

	return nil
}
