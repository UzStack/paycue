package usecase

import (
	"database/sql"
	"errors"

	"github.com/UzStack/paycue/internal/repository"
)

// maxIncrement bitta carta uchun bo'sh summa qidirishda urinishlar chegarasi (tiyinda).
// 100000 tiyin = 1000 so'm — har carta uchun shu oraliqda noyob summa qidiriladi.
const maxIncrement = 100000

// CreateTransactionForCard berilgan carta uchun hozir band bo'lmagan summani topadi
// va transaction yaratadi. amountSom — so'mda kiritilgan asosiy summa; ichkarida
// tiyinga (×100) o'tkaziladi va noyoblik tiyin qadamida qidiriladi (1000.00, 1000.01 ...).
// countDown=true bo'lsa summa pastga kamayadi (1000.00, 999.99, ...). Qaytadigan
// qiymat — tiyinda. Increment har carta bo'yicha alohida.
func CreateTransactionForCard(db *sql.DB, cardID, amountSom int64, timeoutMins int, countDown bool) (int64, string, error) {
	base := amountSom * 100 // tiyin
	a := base
	for i := 0; i < maxIncrement; i++ {
		if a <= 0 {
			break // pastga hisoblashda 0 dan o'tib ketdik
		}
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
		if countDown {
			a--
		} else {
			a++
		}
	}
	return 0, "", errors.New("bo'sh summa topilmadi (limit oshib ketdi)")
}
