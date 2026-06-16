package domain

type Task interface {
	Paylod() any
}

// WebhookTask to'lov aniqlanganda webhook yuborish uchun ishlatiladi.
type WebhookTask struct {
	UserID  int64
	CardID  int64
	TransID string
	Amount  int64
	Action  string // confirm | cancel
}

func (w WebhookTask) Paylod() any {
	return w
}
