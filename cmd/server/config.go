package main

import (
	"breast-implant-warranty-system/core/singleton"
)

type Config struct {
	PORT        string `env:"PORT" envDefault:"8080"`
	ENVIRONMENT string `env:"ENVIRONMENT" envDefault:"development"`

	// JWT 設定
	JWT_SECRET         string `env:"JWT_SECRET" envSecretPath:"JWT_SECRET_SECRET_PATH" validate:"required"`
	JWT_REFRESH_SECRET string `env:"JWT_REFRESH_SECRET" envSecretPath:"JWT_REFRESH_SECRET_SECRET_PATH" validate:"required"`
	JWT_EXPIRE_HOURS   int    `env:"JWT_EXPIRE_HOURS" envDefault:"24"`

	// 加密設定
	ENCRYPTION_KEY string `env:"ENCRYPTION_KEY" envSecretPath:"ENCRYPTION_KEY_SECRET_PATH" validate:"required"`

	// Mailgun 設定
	MAILGUN_DOMAIN     string `env:"MAILGUN_DOMAIN" envDefault:"mail.gmmed.com.tw"`
	MAILGUN_API_KEY    string `env:"MAILGUN_API_KEY" envSecretPath:"MAILGUN_API_KEY_SECRET_PATH" validate:"required"`
	MAILGUN_FROM_EMAIL string `env:"MAILGUN_FROM_EMAIL" envDefault:"noreply@gmmed.com.tw"`

	// 公司設定
	COMPANY_NAME               string `env:"COMPANY_NAME" envDefault:"偉鉅股份有限公司"`
	COMPANY_EMAIL              string `env:"COMPANY_EMAIL" envDefault:"info@mail.gmmed.com.tw"`
	COMPANY_NOTIFICATION_EMAIL string `env:"COMPANY_NOTIFICATION_EMAIL" envDefault:"item.search@gmail.com"`

	// 伺服器設定
	SERVER_HOST string `env:"SERVER_HOST" envDefault:"localhost"`
	DEBUG       bool   `env:"DEBUG" envDefault:"true"`
	LOG_LEVEL   string `env:"LOG_LEVEL" envDefault:"info"`
	APP_URL     string `env:"APP_URL" envDefault:"http://localhost:5173"`

	// Email 範本設定
	EMAIL_TEMPLATE_SUBJECT     string `env:"EMAIL_TEMPLATE_SUBJECT" envDefault:"{patient_surname} 您的植入物保固已完成登錄"`
	EMAIL_TEMPLATE_SENDER_NAME string `env:"EMAIL_TEMPLATE_SENDER_NAME" envDefault:"偉鉅股份客服部"`

	// 安全設定
	CORS_ALLOWED_ORIGINS string `env:"CORS_ALLOWED_ORIGINS"`

	singleton.PostgresDBConfig
}

func (cfg *Config) PostgresDBName() string {
	return cfg.POSTGRES_DB_NAME
}
func (cfg *Config) PostgresReadDBHost() string {
	return cfg.POSTGRES_WRITE_DB_HOST
}
func (cfg *Config) PostgresReadDBPort() string {
	return cfg.POSTGRES_WRITE_DB_PORT
}
func (cfg *Config) PostgresReadDBUser() string {
	return cfg.POSTGRES_WRITE_DB_USER
}
func (cfg *Config) PostgresReadDBPassword() string {
	return cfg.POSTGRES_WRITE_DB_PASSWORD
}
func (cfg *Config) PostgresWriteDBHost() string {
	return cfg.POSTGRES_WRITE_DB_HOST
}
func (cfg *Config) PostgresWriteDBPort() string {
	return cfg.POSTGRES_WRITE_DB_PORT
}
func (cfg *Config) PostgresWriteDBUser() string {
	return cfg.POSTGRES_WRITE_DB_USER
}
func (cfg *Config) PostgresWriteDBPassword() string {
	return cfg.POSTGRES_WRITE_DB_PASSWORD
}
func (cfg *Config) EncryptionKey() string {
	return cfg.ENCRYPTION_KEY
}
func (cfg *Config) MailgunDomain() string {
	return cfg.MAILGUN_DOMAIN
}
func (cfg *Config) MailgunAPIKey() string {
	return cfg.MAILGUN_API_KEY
}
func (cfg *Config) MailgunFromEmail() string {
	return cfg.MAILGUN_FROM_EMAIL
}
func (cfg *Config) EmailTemplateSubject() string {
	return cfg.EMAIL_TEMPLATE_SUBJECT
}
func (cfg *Config) EmailTemplateSenderName() string {
	return cfg.EMAIL_TEMPLATE_SENDER_NAME
}
func (cfg *Config) CompanyName() string {
	return cfg.COMPANY_NAME
}
func (cfg *Config) CompanyEmail() string {
	return cfg.COMPANY_EMAIL
}
func (cfg *Config) CompanyNotificationEmail() string {
	return cfg.COMPANY_NOTIFICATION_EMAIL
}
func (cfg *Config) CORSAllowedOrigins() string {
	return cfg.CORS_ALLOWED_ORIGINS
}
func (cfg *Config) JWTSecret() string {
	return cfg.JWT_SECRET
}
func (cfg *Config) JWTRefreshSecret() string {
	return cfg.JWT_REFRESH_SECRET
}
func (cfg *Config) JWTExpireHours() int {
	return cfg.JWT_EXPIRE_HOURS
}
func (cfg *Config) AppURL() string {
	return cfg.APP_URL
}
