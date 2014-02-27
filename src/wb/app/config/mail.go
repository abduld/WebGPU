package config

var (
	MailgunKey string
)

func InitMailConfig() {
	MailgunKey, _ = NestedRevelConfig.String("mailgun.key")
}
