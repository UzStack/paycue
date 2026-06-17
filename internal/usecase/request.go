package usecase

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// maxWebhookAttempts bitta webhook yetkazib berish uchun urinishlar soni.
const maxWebhookAttempts = 3

// WebhookResult webhook yetkazib berishning yakuniy natijasi.
type WebhookResult struct {
	Success    bool
	Attempts   int    // jami urinishlar soni
	StatusCode int    // oxirgi HTTP javob kodi
	Err        string // oxirgi xato matni (bo'sh = muvaffaqiyat)
}

// WebhookRequest webhookni urinib ko'radi (maxWebhookAttempts gacha) va natijani
// (success, urinishlar soni, oxirgi kod/xato) qaytaradi.
func WebhookRequest(url, apiKey string, data map[string]any, log *zap.Logger) WebhookResult {
	var res WebhookResult
	for attempt := 1; attempt <= maxWebhookAttempts; attempt++ {
		res.Attempts = attempt
		code, err := postWebhook(url, apiKey, data)
		res.StatusCode = code
		if err == nil {
			res.Success = true
			res.Err = ""
			log.Info("webhook success:", zap.Int("status", code), zap.Int("attempt", attempt), zap.Any("transaction_id", data["transaction_id"]), zap.Any("amount", data["amount"]))
			return res
		}
		res.Err = err.Error()
		log.Info("webhook error:", zap.Int("status", code), zap.Int("attempt", attempt), zap.String("err", err.Error()), zap.Any("transaction_id", data["transaction_id"]), zap.Any("amount", data["amount"]))
		if attempt < maxWebhookAttempts {
			time.Sleep(5 * time.Second)
		}
	}
	return res
}

// postWebhook bitta urinish: HTTP kod va xato qaytaradi (200 + {"ok":true} muvaffaqiyat).
func postWebhook(url, apiKey string, data map[string]any) (int, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	// Client webhook haqiqiyligini shu header orqali tekshiradi.
	req.Header.Set("X-API-Key", apiKey)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	var responseData map[string]any
	decErr := json.NewDecoder(response.Body).Decode(&responseData)
	if response.StatusCode != http.StatusOK || decErr != nil || responseData["ok"] != true {
		return response.StatusCode, errors.New("webhook rad etildi: " + response.Status)
	}
	return response.StatusCode, nil
}
