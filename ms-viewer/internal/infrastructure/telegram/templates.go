package telegram

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type templatePayload struct {
	EventType        string
	Severity         string
	ObjectName       string
	DeviceName       string
	TagName          string
	State            string
	Value            string
	Timestamp        string
	EscalationNotice string
}

type Renderer struct{}

func NewRenderer() *Renderer {
	return &Renderer{}
}

func (renderer *Renderer) Render(
	event domain.AlarmEvent,
	routeContext domain.AlarmRouteContext,
	escalated bool,
	escalationDelay time.Duration,
) (string, error) {
	return RenderAlarmMessage(event, routeContext, escalated, escalationDelay)
}

var (
	raisedTemplate = template.Must(template.New("raised").Parse(strings.TrimSpace(`
🚨 *Тревога*: {{.Severity}}
Объект: {{.ObjectName}}
Устройство: {{.DeviceName}}
Тег: *{{.TagName}}*
Значение: {{.Value}}
Состояние: {{.State}}
Время: {{.Timestamp}}
`)))

	clearedTemplate = template.Must(template.New("cleared").Parse(strings.TrimSpace(`
✅ *Тревога снята*
Объект: {{.ObjectName}}
Устройство: {{.DeviceName}}
Тег: *{{.TagName}}*
Состояние: {{.State}}
Время: {{.Timestamp}}
`)))

	ackedTemplate = template.Must(template.New("acked").Parse(strings.TrimSpace(`
👁 *Тревога квитирована*
Объект: {{.ObjectName}}
Устройство: {{.DeviceName}}
Тег: *{{.TagName}}*
Состояние: {{.State}}
Время: {{.Timestamp}}
`)))

	escalatedTemplate = template.Must(template.New("escalated").Parse(strings.TrimSpace(`
⚠️ *Эскалация*: {{.Severity}}
{{.EscalationNotice}}
Объект: {{.ObjectName}}
Устройство: {{.DeviceName}}
Тег: *{{.TagName}}*
Состояние: {{.State}}
Время: {{.Timestamp}}
`)))
)

func RenderAlarmMessage(
	event domain.AlarmEvent,
	routeContext domain.AlarmRouteContext,
	escalated bool,
	escalationDelay time.Duration,
) (string, error) {
	payload := templatePayload{
		EventType:  event.EventType.String(),
		Severity:   telegramEscape(resolveSeverity(event.StateTo)),
		ObjectName: telegramEscape(nonEmpty(routeContext.ObjectName, "n/a")),
		DeviceName: telegramEscape(nonEmpty(routeContext.DeviceName, "n/a")),
		TagName:    telegramEscape(nonEmpty(routeContext.TagName, "n/a")),
		State:      telegramEscape(event.StateTo.String()),
		Value: telegramEscape(formatValue(
			event.Value,
			routeContext.UnitSymbol,
		)),
		Timestamp: telegramEscape(event.TS.UTC().Format(time.RFC3339)),
	}

	if escalated {
		minutes := int(escalationDelay / time.Minute)
		if minutes <= 0 {
			minutes = 1
		}
		payload.EscalationNotice = telegramEscape(
			fmt.Sprintf("Не квитировано %d минут", minutes),
		)
	}

	selectedTemplate := raisedTemplate
	switch {
	case escalated:
		selectedTemplate = escalatedTemplate
	case event.EventType == domain.AlarmEventCleared:
		selectedTemplate = clearedTemplate
	case event.EventType == domain.AlarmEventAcked:
		selectedTemplate = ackedTemplate
	}

	buffer := bytes.NewBuffer(nil)
	if executeError := selectedTemplate.Execute(buffer, payload); executeError != nil {
		return "", fmt.Errorf("render telegram template: %w", executeError)
	}

	return buffer.String(), nil
}

func TelegramEscape(source string) string {
	return telegramEscape(source)
}

func telegramEscape(source string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(source)
}

func resolveSeverity(state domain.AlarmState) string {
	switch state {
	case domain.AlarmStateHiHi, domain.AlarmStateLoLo:
		return "критично"
	case domain.AlarmStateHi, domain.AlarmStateLo:
		return "высокий"
	case domain.AlarmStateCommLoss, domain.AlarmStateOffline, domain.AlarmStateBad, domain.AlarmStateUncertain:
		return "предупреждение"
	default:
		return "инфо"
	}
}

func formatValue(value *float64, unitSymbol string) string {
	if value == nil {
		return "n/a"
	}
	if strings.TrimSpace(unitSymbol) == "" {
		return fmt.Sprintf("%.3f", *value)
	}
	return fmt.Sprintf("%.3f %s", *value, unitSymbol)
}

func nonEmpty(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}
