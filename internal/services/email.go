package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/sirupsen/logrus"

	"breast-implant-warranty-system/internal/models"
)

type MailgunConfigItf interface {
	MailgunDomain() string
	MailgunAPIKey() string
	MailgunFromEmail() string
	EmailTemplateSenderName() string
	EmailTemplateSubject() string
	CompanyNotificationEmail() string
	CompanyName() string
	CompanyEmail() string
}

// EmailService 信件服務
type EmailService struct {
	mg     *mailgun.MailgunImpl
	cfg    MailgunConfigItf
	logger *logrus.Logger
}

// NewEmailService 建立新的信件服務
func NewEmailService(cfg MailgunConfigItf) *EmailService {
	mg := mailgun.NewMailgun(cfg.MailgunDomain(), cfg.MailgunAPIKey())

	return &EmailService{
		mg:     mg,
		cfg:    cfg,
		logger: logrus.New(),
	}
}

// WarrantyConfirmationData 保固確認信件數據
type WarrantyConfirmationData struct {
	PatientName        string
	PatientSurname     string
	ProductBrand       string
	ProductModel       string
	ProductSize        string
	ProductType        string
	SerialNumber       string
	SecondSerialNumber string
	SurgeryDate        string
	WarrantyEndDate    string
	HospitalName       string
	DoctorName         string
	CompanyName        string
	CompanyEmail       string
	RegistrationID     string
	IsLifetimeWarranty bool
}

// SendWarrantyConfirmation 發送保固確認信件
func (s *EmailService) SendWarrantyConfirmation(warranty *models.WarrantyRegistration) error {
	// 檢查 registration 資料
	if warranty == nil {
		return fmt.Errorf("warranty registration is nil")
	}
	if !warranty.PatientEmail.Valid || warranty.PatientEmail.String == "" {
		return fmt.Errorf("patient email is nil")
	}
	if !warranty.PatientName.Valid || warranty.PatientName.String == "" {
		return fmt.Errorf("patient name is nil")
	}
	if !warranty.SurgeryDate.Valid {
		return fmt.Errorf("surgery date is nil")
	}
	if !warranty.WarrantyEndDate.Valid {
		return fmt.Errorf("warranty end date is nil")
	}
	if !warranty.ProductID.Valid {
		return fmt.Errorf("product ID is nil")
	}
	if !warranty.ProductSerialNumber.Valid || warranty.ProductSerialNumber.String == "" {
		return fmt.Errorf("product serial number is nil")
	}

	// 準備信件數據
	data := s.prepareWarrantyData(warranty)

	// 生成信件內容
	subject := s.generateSubject(data)
	htmlBody, err := s.generateHTMLBody(data)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate email HTML body")
		return fmt.Errorf("failed to generate email content: %w", err)
	}

	textBody := s.generateTextBody(data)

	// 建立信件訊息
	message := mailgun.NewMessage(
		fmt.Sprintf("%s <%s>", s.cfg.EmailTemplateSenderName(), s.cfg.MailgunFromEmail()),
		subject,
		textBody,
		warranty.PatientEmail.String,
	)

	// 設置HTML內容
	message.SetHTML(htmlBody)

	// 添加標籤用於追蹤
	message.AddTag("warranty-confirmation")
	message.AddTag("automated")

	// 設置自定義變數
	message.AddVariable("warranty_id", warranty.ID)
	message.AddVariable("patient_name", warranty.PatientName)

	// 發送信件
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, id, err := s.mg.Send(ctx, message)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"warranty_id":   warranty.ID,
			"patient_email": warranty.PatientEmail,
		}).Error("Failed to send warranty confirmation email")
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"warranty_id":   warranty.ID,
		"patient_email": warranty.PatientEmail,
		"message_id":    id,
		"response":      resp,
	}).Info("Warranty confirmation email sent successfully")

	return nil
}

// SendNotificationToCompany 發送通知信件給公司
func (s *EmailService) SendNotificationToCompany(warranty *models.WarrantyRegistration) error {
	if s.cfg.CompanyNotificationEmail() == "" {
		s.logger.Warn("Company notification email not configured, skipping notification")
		return nil
	}

	data := s.prepareWarrantyData(warranty)

	subject := fmt.Sprintf("新保固登記通知 - %s", data.PatientName)
	htmlBody, err := s.generateCompanyNotificationHTML(data)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate company notification HTML")
		return fmt.Errorf("failed to generate notification content: %w", err)
	}

	textBody := s.generateCompanyNotificationText(data)

	message := mailgun.NewMessage(
		fmt.Sprintf("%s <%s>", s.cfg.EmailTemplateSenderName(), s.cfg.MailgunFromEmail()),
		subject,
		textBody,
		s.cfg.CompanyNotificationEmail(),
	)

	message.SetHTML(htmlBody)
	message.AddTag("company-notification")
	message.AddTag("automated")
	message.AddVariable("warranty_id", warranty.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, id, err := s.mg.Send(ctx, message)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"warranty_id":   warranty.ID,
			"company_email": s.cfg.CompanyNotificationEmail,
		}).Error("Failed to send company notification email")
		return fmt.Errorf("failed to send company notification: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"warranty_id":   warranty.ID,
		"company_email": s.cfg.CompanyNotificationEmail,
		"message_id":    id,
		"response":      resp,
	}).Info("Company notification email sent successfully")

	return nil
}

// prepareWarrantyData 準備信件數據
func (s *EmailService) prepareWarrantyData(warranty *models.WarrantyRegistration) *WarrantyConfirmationData {
	// 提取姓氏（假設姓氏是名字的第一個字）
	surname := ""
	patientName := warranty.PatientName.String
	if len(patientName) > 0 {
		runes := []rune(patientName)
		surname = string(runes[0])
	}

	// 格式化日期
	surgeryDate := warranty.SurgeryDate.Time.Format("2006年01月02日")
	warrantyEndDate := warranty.WarrantyEndDate.Time.Format("2006年01月02日")

	// 檢查是否為終身保固
	isLifetimeWarranty := false
	if warranty.NullableProduct.WarrantyYears.Valid && warranty.NullableProduct.WarrantyYears.Int64 == 0 {
		isLifetimeWarranty = true
		warrantyEndDate = "終身保固"
	}

	// 處理第二個序號
	secondSerialNumber := ""
	if warranty.ProductSerialNumber2.Valid && warranty.ProductSerialNumber2.String != "" {
		secondSerialNumber = warranty.ProductSerialNumber2.String
	}

	data := &WarrantyConfirmationData{
		PatientName:        patientName,
		PatientSurname:     surname,
		SerialNumber:       warranty.ProductSerialNumber.String,
		SecondSerialNumber: secondSerialNumber,
		SurgeryDate:        surgeryDate,
		WarrantyEndDate:    warrantyEndDate,
		HospitalName:       warranty.HospitalName.String,
		DoctorName:         warranty.DoctorName.String,
		CompanyName:        s.cfg.CompanyName(),
		CompanyEmail:       s.cfg.CompanyEmail(),
		RegistrationID:     warranty.ID,
		IsLifetimeWarranty: isLifetimeWarranty,
	}

	// 填充產品資訊
	if warranty.NullableProduct.ModelNumber.Valid {
		data.ProductBrand = warranty.NullableProduct.Brand.String
		data.ProductModel = warranty.NullableProduct.ModelNumber.String
		if warranty.NullableProduct.Size.Valid {
			data.ProductSize = warranty.NullableProduct.Size.String
		}
		data.ProductType = warranty.NullableProduct.Type.String
	}

	return data
}

// generateSubject 生成信件主題
func (s *EmailService) generateSubject(data *WarrantyConfirmationData) string {
	subject := s.cfg.EmailTemplateSubject()
	subject = strings.ReplaceAll(subject, "{patient_surname}", data.PatientSurname)
	subject = strings.ReplaceAll(subject, "{patient_name}", data.PatientName)
	subject = strings.ReplaceAll(subject, "{company_name}", data.CompanyName)
	return subject
}

// generateTextBody 生成純文字信件內容
func (s *EmailService) generateTextBody(data *WarrantyConfirmationData) string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("親愛的 %s 女士您好，\n\n", data.PatientName))
	buffer.WriteString("您的植入物保固登記已完成，以下是您的保固資訊：\n\n")
	buffer.WriteString("=== 保固資訊 ===\n")
	buffer.WriteString(fmt.Sprintf("登記編號：%s\n", data.RegistrationID))
	buffer.WriteString(fmt.Sprintf("產品品牌：%s\n", data.ProductBrand))
	buffer.WriteString(fmt.Sprintf("產品型號：%s\n", data.ProductModel))
	buffer.WriteString(fmt.Sprintf("產品尺寸：%s\n", data.ProductSize))
	buffer.WriteString(fmt.Sprintf("產品類型：%s\n", data.ProductType))
	buffer.WriteString(fmt.Sprintf("序號：%s\n", data.SerialNumber))

	if data.SecondSerialNumber != "" {
		buffer.WriteString(fmt.Sprintf("第二序號：%s\n", data.SecondSerialNumber))
	}

	buffer.WriteString(fmt.Sprintf("手術日期：%s\n", data.SurgeryDate))
	buffer.WriteString(fmt.Sprintf("保固期限：%s\n", data.WarrantyEndDate))
	buffer.WriteString(fmt.Sprintf("手術醫院：%s\n", data.HospitalName))
	buffer.WriteString(fmt.Sprintf("主治醫師：%s\n", data.DoctorName))

	buffer.WriteString("\n請妥善保存此信件作為保固憑證。\n")
	buffer.WriteString("如有任何術後問題，請聯繫您的手術醫院。\n")
	buffer.WriteString(fmt.Sprintf("\n此致\n%s\n", data.CompanyName))

	return buffer.String()
}

// generateHTMLBody 生成HTML信件內容
func (s *EmailService) generateHTMLBody(data *WarrantyConfirmationData) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html lang="zh-TW">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>保固確認通知</title>
    <style>
        body { font-family: 'Microsoft JhengHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 20px; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 0 10px rgba(0,0,0,0.1); }
        .header { text-align: center; border-bottom: 3px solid #007bff; padding-bottom: 20px; margin-bottom: 30px; }
        .header h1 { color: #007bff; margin: 0; font-size: 24px; }
        .content { margin-bottom: 30px; }
        .info-table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        .info-table th, .info-table td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        .info-table th { background-color: #f8f9fa; font-weight: bold; width: 30%; }
        .highlight { background-color: #e7f3ff; padding: 15px; border-left: 4px solid #007bff; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; color: #666; font-size: 14px; }
        .warranty-badge { display: inline-block; background: #28a745; color: white; padding: 5px 10px; border-radius: 15px; font-size: 12px; font-weight: bold; }
        .lifetime-badge { background: #ffc107; color: #000; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🛡️ 植入物保固確認通知</h1>
            <p>{{.CompanyName}}</p>
        </div>

        <div class="content">
            <p>親愛的 <strong>{{.PatientName}}</strong>，</p>
            <p>您的植入物保固登記已完成！以下是您的保固資訊：</p>

            <div class="highlight">
                <strong>登記編號：{{.RegistrationID}}</strong>
                {{if .IsLifetimeWarranty}}
                <span class="warranty-badge lifetime-badge">終身保固</span>
                {{else}}
                <span class="warranty-badge">有限保固</span>
                {{end}}
            </div>

            <table class="info-table">
                <tr><th>產品品牌</th><td>{{.ProductBrand}}</td></tr>
                <tr><th>產品型號</th><td>{{.ProductModel}}</td></tr>
                <tr><th>產品尺寸</th><td>{{.ProductSize}}</td></tr>
                <tr><th>產品類型</th><td>{{.ProductType}}</td></tr>
                <tr><th>序號</th><td>{{.SerialNumber}}</td></tr>
                {{if .SecondSerialNumber}}
                <tr><th>第二序號</th><td>{{.SecondSerialNumber}}</td></tr>
                {{end}}
                <tr><th>手術日期</th><td>{{.SurgeryDate}}</td></tr>
                <tr><th>保固期限</th><td><strong>{{.WarrantyEndDate}}</strong></td></tr>
                <tr><th>手術醫院</th><td>{{.HospitalName}}</td></tr>
                <tr><th>主治醫師</th><td>{{.DoctorName}}</td></tr>
            </table>

            <div class="highlight">
                <strong>📋 重要提醒：</strong><br>
                • 請妥善保存此信件作為保固憑證<br>
                • 如需保固服務，請提供登記編號<br>
                • 如有任何術後問題，請聯繫您的手術醫院
            </div>
        </div>

        <div class="footer">
            <p><strong>{{.CompanyName}}</strong></p>
            <p><small>此為系統自動發送的信件，請勿直接回覆</small></p>
        </div>
    </div>
</body>
</html>`

	t, err := template.New("warranty_confirmation").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse email template: %w", err)
	}

	var buffer bytes.Buffer
	err = t.Execute(&buffer, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute email template: %w", err)
	}

	return buffer.String(), nil
}

// generateCompanyNotificationHTML 生成公司通知HTML
func (s *EmailService) generateCompanyNotificationHTML(data *WarrantyConfirmationData) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html lang="zh-TW">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>新保固登記通知</title>
    <style>
        body { font-family: 'Microsoft JhengHei', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 20px; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 0 10px rgba(0,0,0,0.1); }
        .header { text-align: center; border-bottom: 3px solid #dc3545; padding-bottom: 20px; margin-bottom: 30px; }
        .header h1 { color: #dc3545; margin: 0; font-size: 24px; }
        .info-table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        .info-table th, .info-table td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        .info-table th { background-color: #f8f9fa; font-weight: bold; width: 30%; }
        .alert { background-color: #fff3cd; border: 1px solid #ffeaa7; padding: 15px; border-radius: 5px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔔 新保固登記通知</h1>
            <p>{{.CompanyName}} - 內部通知</p>
        </div>

        <div class="alert">
            <strong>新的保固登記已完成，請注意處理相關事務。</strong>
        </div>

        <table class="info-table">
            <tr><th>登記編號</th><td>{{.RegistrationID}}</td></tr>
            <tr><th>患者姓名</th><td>{{.PatientName}}</td></tr>
            <tr><th>產品品牌</th><td>{{.ProductBrand}}</td></tr>
            <tr><th>產品型號</th><td>{{.ProductModel}}</td></tr>
            <tr><th>序號</th><td>{{.SerialNumber}}</td></tr>
            {{if .SecondSerialNumber}}
            <tr><th>第二序號</th><td>{{.SecondSerialNumber}}</td></tr>
            {{end}}
            <tr><th>手術日期</th><td>{{.SurgeryDate}}</td></tr>
            <tr><th>保固期限</th><td>{{.WarrantyEndDate}}</td></tr>
            <tr><th>手術醫院</th><td>{{.HospitalName}}</td></tr>
            <tr><th>主治醫師</th><td>{{.DoctorName}}</td></tr>
        </table>
    </div>
</body>
</html>`

	t, err := template.New("company_notification").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse company notification template: %w", err)
	}

	var buffer bytes.Buffer
	err = t.Execute(&buffer, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute company notification template: %w", err)
	}

	return buffer.String(), nil
}

// generateCompanyNotificationText 生成公司通知純文字內容
func (s *EmailService) generateCompanyNotificationText(data *WarrantyConfirmationData) string {
	var buffer bytes.Buffer

	buffer.WriteString("新保固登記通知\n")
	buffer.WriteString("==================\n\n")
	buffer.WriteString("新的保固登記已完成，請注意處理相關事務。\n\n")
	buffer.WriteString("登記資訊：\n")
	buffer.WriteString(fmt.Sprintf("登記編號：%s\n", data.RegistrationID))
	buffer.WriteString(fmt.Sprintf("患者姓名：%s\n", data.PatientName))
	buffer.WriteString(fmt.Sprintf("產品品牌：%s\n", data.ProductBrand))
	buffer.WriteString(fmt.Sprintf("產品型號：%s\n", data.ProductModel))
	buffer.WriteString(fmt.Sprintf("序號：%s\n", data.SerialNumber))

	if data.SecondSerialNumber != "" {
		buffer.WriteString(fmt.Sprintf("第二序號：%s\n", data.SecondSerialNumber))
	}

	buffer.WriteString(fmt.Sprintf("手術日期：%s\n", data.SurgeryDate))
	buffer.WriteString(fmt.Sprintf("保固期限：%s\n", data.WarrantyEndDate))
	buffer.WriteString(fmt.Sprintf("手術醫院：%s\n", data.HospitalName))
	buffer.WriteString(fmt.Sprintf("主治醫師：%s\n", data.DoctorName))

	return buffer.String()
}
