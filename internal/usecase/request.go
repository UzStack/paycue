package usecase

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func WebhookRequest(url string, apiKey string, data map[string]any, log *zap.Logger, retry int) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	// Client webhook haqiqiyligini shu header orqali tekshiradi.
	req.Header.Set("X-API-Key", apiKey)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	var responseData map[string]any
	err = json.NewDecoder(response.Body).Decode(&responseData)
	if response.StatusCode != http.StatusOK || err != nil || responseData["ok"] != true {
		log.Info("webhook error:", zap.Int("status", response.StatusCode), zap.Any("transaction_id", data["transaction_id"]), zap.Any("amount", data["amount"]))
		if retry < 3 {
			time.Sleep(5 * time.Second)
			log.Info("retry webhook", zap.Int("retry", retry))
			return WebhookRequest(url, apiKey, data, log, retry+1)
		}
		return errors.New("failed send webhook " + response.Status)
	}
	log.Info("webhook success:", zap.Int("status", response.StatusCode), zap.Any("transaction_id", data["transaction_id"]), zap.Any("amount", data["amount"]))
	return nil
}
