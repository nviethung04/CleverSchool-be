package services

import (
	"be-Clever School/config"
	"fmt"
	"mime"
	"net/smtp"
	"os"
)

type EmailService interface {
	SendPasswordResetEmail(email, token, username string) error
}

type emailService struct{}

func NewEmailService() EmailService {
	return &emailService{}
}

func (s *emailService) SendPasswordResetEmail(email, token, username string) error {
	// setting SMTP
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		smtpPort = "587"
	}
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	fromEmail := os.Getenv("FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = smtpUsername
	}
	fromName := os.Getenv("FROM_NAME")
	if fromName == "" {
		fromName = "Clever School System"
	}

	// URL reset password
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	resetURL := frontendURL + "/reset-password?token=" + token

	// content email
	subject := "Đặt lại mật khẩu - Clever School System"
	encodedSubject := mime.QEncoding.Encode("utf-8", subject)
	body := s.generatePasswordResetEmailBody(username, resetURL)

	// Add message
	message := fmt.Sprintf("From: %s <%s>\r\n", fromName, fromEmail)
	message += fmt.Sprintf("To: %s\r\n", email)
	message += fmt.Sprintf("Subject: %s\r\n", encodedSubject)
	message += "MIME-Version: 1.0\r\n"
	message += "Content-Type: text/html; charset=UTF-8\r\n"
	message += "\r\n"
	message += body

	// Send email
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	err := smtp.SendMail(addr, auth, fromEmail, []string{email}, []byte(message))
	if err != nil {
		config.Log.Error("Failed to send email:", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	config.Log.Info("Password reset email sent successfully to:", email)
	return nil
}

func (s *emailService) generatePasswordResetEmailBody(username, resetURL string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="vi">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Đặt lại mật khẩu</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f2f2f2;
            margin: 0;
            padding: 20px;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background-color: #ffffff;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
        }
        .header {
            background-color: #374bbf;
            color: white;
            padding: 20px;
            text-align: center;
        }
        .content {
            padding: 20px;
        }
        .button {
            display: inline-block;
            background-color: #34408d;
            color: #ffffff !important;
            padding: 12px 24px;
            text-decoration: none;
            border-radius: 5px;
            margin: 20px 0;
        }
        .button:hover {
            background-color: #4152b5;
        }
        .footer {
            background-color: #f7f7f7;
            padding: 15px 20px;
            font-size: 12px;
            color: #666;
            text-align: center;
        }
        .warning {
            background-color: #fff3cd;
            border: 1px solid #ffeaa7;
            color: #856404;
            padding: 10px;
            border-radius: 5px;
            margin: 15px 0;
        }
        @media (max-width: 600px) {
            .button {
                padding: 10px 20px;
                font-size: 16px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Đặt lại mật khẩu</h1>
        </div>

        <div class="content">
            <p>Xin chào <strong>%s</strong>,</p>

            <p>Bạn nhận được email này vì bạn đã yêu cầu đặt lại mật khẩu cho tài khoản Clever School của mình.</p>

            <p>Vui lòng nhấp vào nút bên dưới để đặt lại mật khẩu:</p>

            <div style="text-align: center;">
                <a href="%s" class="button">Đặt lại mật khẩu</a>
            </div>

            <div class="warning">
                <strong>Lưu ý:</strong>
                <ul>
                    <li>Link này sẽ hết hạn sau 2 giờ</li>
                    <li>Nếu bạn không yêu cầu đặt lại mật khẩu, vui lòng bỏ qua email này</li>
                    <li>Để bảo mật, vui lòng không chia sẻ link này với người khác</li>
                </ul>
            </div>

            <p>Nếu nút không hoạt động, bạn có thể copy và paste link sau vào trình duyệt:</p>
            <p style="word-break: break-all; color: #0066cc;">%s</p>

            <p>Trân trọng,<br>Đội ngũ Clever School System</p>
        </div>

        <div class="footer">
            <p>Email này được gửi tự động, vui lòng không trả lời email này.</p>
            <p>Nếu bạn có thắc mắc, vui lòng liên hệ với chúng tôi.</p>
        </div>
    </div>
</body>
</html>
`, username, resetURL, resetURL)
}
