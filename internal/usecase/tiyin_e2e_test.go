package usecase

import (
	"database/sql"
	"encoding/json"
	"os"
	"testing"

	"github.com/UzStack/paycue/internal/domain"
	"github.com/UzStack/paycue/internal/repository"
	_ "github.com/mattn/go-sqlite3"
)

func tdb(t *testing.T) *sql.DB {
	f, _ := os.CreateTemp("", "paycue-*.sqlite3")
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	db, err := sql.Open("sqlite3", f.Name())
	if err != nil {
		t.Fatal(err)
	}
	repository.InitTables(db)
	// карта uchun minimal qatorlar
	db.Exec("INSERT INTO telegram_accounts(id,user_id,phone) VALUES(1,1,'+998901112233')")
	db.Exec("INSERT INTO cards(id,telegram_account_id,last4,number) VALUES(1,1,'7159','8600111122227159')")
	return db
}

func TestTiyinMarshal(t *testing.T) {
	cases := map[domain.Tiyin]string{100000: "1000", 100001: "1000.01", 99999: "999.99", 0: "0"}
	for in, want := range cases {
		b, _ := json.Marshal(in)
		if string(b) != want {
			t.Errorf("Tiyin(%d) -> %s, kutilgan %s", in, b, want)
		}
	}
}

func TestCreateUpDown(t *testing.T) {
	db := tdb(t)
	// 1-chi: 1000 so'm -> 100000 tiyin (1000.00)
	a1, _, err := CreateTransactionForCard(db, 1, 1000, 30, false)
	if err != nil || a1 != 100000 {
		t.Fatalf("up #1: got %d err %v", a1, err)
	}
	// 2-chi (band, yuqoriga): 100001 (1000.01)
	a2, _, _ := CreateTransactionForCard(db, 1, 1000, 30, false)
	if a2 != 100001 {
		t.Fatalf("up #2: got %d, kutilgan 100001", a2)
	}
	// pastga: band 100000 dan keyin 99999 (999.99)
	a3, _, _ := CreateTransactionForCard(db, 1, 1000, 30, true)
	if a3 != 99999 {
		t.Fatalf("down: got %d, kutilgan 99999", a3)
	}
}

func TestAmountMigration(t *testing.T) {
	f, _ := os.CreateTemp("", "paycue-mig-*.sqlite3")
	f.Close()
	defer os.Remove(f.Name())
	db, _ := sql.Open("sqlite3", f.Name())
	repository.InitTables(db) // birinchi marta meta=1 qo'yiladi, ma'lumot yo'q

	// Eski (so'mdagi) yozuvni qo'lda kiritamiz va meta'ni "migratsiya qilinmagan"ga qaytaramiz
	db.Exec("INSERT INTO cards(id,telegram_account_id,last4) VALUES(9,1,'0000')")
	db.Exec("INSERT INTO transactions(card_id,amount,transaction_id) VALUES(9,20001,'old-uuid')")
	db.Exec("DELETE FROM meta WHERE key='amount_in_tiyin'")

	repository.InitTables(db) // migratsiya: 20001 so'm -> 2000100 tiyin
	var amt int64
	db.QueryRow("SELECT amount FROM transactions WHERE transaction_id='old-uuid'").Scan(&amt)
	if amt != 2000100 {
		t.Fatalf("migratsiya: got %d, kutilgan 2000100", amt)
	}
	// ikkinchi marta ishga tushganda yana ×100 qilmasligi kerak
	repository.InitTables(db)
	db.QueryRow("SELECT amount FROM transactions WHERE transaction_id='old-uuid'").Scan(&amt)
	if amt != 2000100 {
		t.Fatalf("migratsiya takrorlandi: got %d", amt)
	}
}
