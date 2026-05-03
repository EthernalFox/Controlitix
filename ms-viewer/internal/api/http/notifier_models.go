package http

import "github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"

type notifierChatResponse struct {
	ID          string  `json:"id"`
	ChatID      int64   `json:"chat_id"`
	Title       string  `json:"title"`
	Role        *string `json:"role"`
	ObjectID    *string `json:"object_id"`
	SeverityMin int16   `json:"severity_min"`
	Enabled     bool    `json:"enabled"`
}

type notifierChatsResponse struct {
	Items []notifierChatResponse `json:"items"`
	Total int                    `json:"total"`
}

func mapNotifierChatsResponse(result domain.TelegramChatListResult) notifierChatsResponse {
	items := make([]notifierChatResponse, 0, len(result.Items))
	for _, item := range result.Items {
		var objectID *string
		if item.ObjectID != nil {
			value := item.ObjectID.String()
			objectID = &value
		}
		items = append(items, notifierChatResponse{
			ID:          item.ID.String(),
			ChatID:      item.ChatID,
			Title:       item.Title,
			Role:        item.Role,
			ObjectID:    objectID,
			SeverityMin: item.SeverityMin,
			Enabled:     item.Enabled,
		})
	}

	return notifierChatsResponse{
		Items: items,
		Total: result.Total,
	}
}
