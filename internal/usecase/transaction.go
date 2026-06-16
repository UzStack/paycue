package usecase

import (
	"database/sql"
	"errors"

	"github.com/UzStack/paycue/internal/repository"
)

// maxIncrement bitta carta uchun bo'sh summa qidirishda urinishlar chegarasi.
const maxIncrement = 100000

// CreateTransactionForCard berilgan carta uchun hozir band bo'lmagan eng kichik
// summani topadi va transaction yaratadi. Increment har carta bo'yicha alohida.
func CreateTransactionForCard(db *sql.DB, cardID, amount int64, timeoutMins int) (int64, string, error) {
	a := amount
	for a < amount+maxIncrement {
		free, err := repository.CheckTransaction(db, cardID, a, timeoutMins)
		if err != nil {
			return 0, "", err
		}
		if free {
			transID, err := repository.CreateTransaction(db, cardID, a)
			if err != nil {
				return 0, "", err
			}
			return a, transID, nil
		}
		a++
	}
	return 0, "", errors.New("bo'sh summa topilmadi (limit oshib ketdi)")
}
