package main

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/manifoldco/promptui"
)

// selectMenu strelka (↑/↓) bilan tanlanadigan menu. Tanlangan indeks qaytadi.
// Ctrl-C yoki ESC bo'lsa ok=false.
func selectMenu(label string, items []string) (int, bool) {
	s := promptui.Select{
		Label: label,
		Items: items,
		Size:  len(items),
	}
	i, _, err := s.Run()
	if err != nil {
		return 0, false
	}
	return i, true
}

func ask(label string) string {
	v, err := (&promptui.Prompt{Label: label}).Run()
	if err != nil {
		return ""
	}
	return v
}

func askRequired(label string) (string, bool) {
	p := promptui.Prompt{
		Label: label,
		Validate: func(s string) error {
			if s == "" {
				return errors.New("bo'sh bo'lmasin")
			}
			return nil
		},
	}
	v, err := p.Run()
	if err != nil {
		return "", false
	}
	return v, true
}

func askPassword(label string) string {
	v, err := (&promptui.Prompt{Label: label, Mask: '*'}).Run()
	if err != nil {
		return ""
	}
	return v
}

func show(out map[string]any, err error) {
	if err != nil {
		fmt.Println("❌ Xato:", err)
		return
	}
	fmt.Println("✅ Muvaffaqiyatli:")
	printJSON(out["data"])
}

// runTUI strelka bilan boshqariladigan interaktiv menu.
func runTUI(a *app) {
	for {
		tok := "yo'q"
		if a.c.token != "" {
			tok = "bor"
		}
		label := fmt.Sprintf("paycue-cli  (profil: %s | token: %s)", a.profileName, tok)
		idx, ok := selectMenu(label, []string{
			"Ro'yxatdan o'tish (register)",
			"Kirish (login)",
			"Profillar",
			"Webhook",
			"Telegram accountlar",
			"Cartalar",
			"Tranzaksiyalar",
			"Chiqish",
		})
		if !ok {
			return
		}
		switch idx {
		case 0:
			tuiRegister(a)
		case 1:
			tuiLogin(a)
		case 2:
			tuiProfiles(a)
		case 3:
			tuiWebhook(a)
		case 4:
			tuiTelegram(a)
		case 5:
			tuiCards(a)
		case 6:
			tuiTransaction(a)
		case 7:
			fmt.Println("Xayr!")
			return
		}
	}
}

func tuiRegister(a *app) {
	name, ok := askRequired("Ism familiya")
	if !ok {
		return
	}
	email := ask("Email (ixtiyoriy)")
	phone := ask("Telefon (ixtiyoriy)")
	password := askPassword("Parol (ixtiyoriy, bo'sh=parolsiz; berilsa ≥6 belgi)")
	profName := ask("Saqlanadigan profil nomi (bo'sh=default)")
	if profName == "" {
		profName = "default"
	}
	out, err := a.c.do("POST", "/api/register", map[string]any{
		"name": name, "email": email, "phone": phone, "password": password,
	})
	if err == nil {
		a.saveTokenFromResponse(out, profName)
		a.useProfile(a.cfg.Current)
	}
	show(out, err)
}

func tuiLogin(a *app) {
	login, ok := askRequired("Email yoki telefon")
	if !ok {
		return
	}
	password := askPassword("Parol")
	profName := ask("Saqlanadigan profil nomi (bo'sh=default)")
	if profName == "" {
		profName = "default"
	}
	out, err := a.c.do("POST", "/api/login", map[string]any{"login": login, "password": password})
	if err == nil {
		a.saveTokenFromResponse(out, profName)
		a.useProfile(a.cfg.Current)
	}
	show(out, err)
}

func tuiProfiles(a *app) {
	for {
		names := make([]string, 0, len(a.cfg.Profiles))
		for n := range a.cfg.Profiles {
			names = append(names, n)
		}
		sort.Strings(names)
		fmt.Println("\n--- Profillar (joriy: " + a.cfg.Current + ") ---")
		for _, n := range names {
			marker := "  "
			if n == a.cfg.Current {
				marker = "* "
			}
			fmt.Printf("%s%s\t%s\n", marker, n, a.cfg.Profiles[n].API)
		}
		idx, ok := selectMenu("Profillar", []string{
			"Profilga o'tish", "Tokenni ko'rsatish", "Profil qo'shish", "Profil o'chirish", "Orqaga",
		})
		if !ok {
			return
		}
		switch idx {
		case 0:
			if len(names) == 0 {
				fmt.Println("Profillar yo'q.")
				break
			}
			i, ok := selectMenu("Qaysi profilga o'tamiz?", names)
			if !ok {
				break
			}
			a.cfg.Current = names[i]
			saveConfig(a.cfg)
			a.useProfile(names[i])
			fmt.Println("Joriy profil:", names[i])
		case 1:
			// Tokenni ko'rsatish (nusxalash uchun).
			if len(names) == 0 {
				break
			}
			i, ok := selectMenu("Qaysi profil tokeni?", names)
			if !ok {
				break
			}
			fmt.Println("\nToken (" + names[i] + "):")
			fmt.Println(a.cfg.Profiles[names[i]].Token)
		case 2:
			name, ok := askRequired("Profil nomi")
			if !ok {
				break
			}
			token, ok := askRequired("Token")
			if !ok {
				break
			}
			api := ask("API manzili (bo'sh=" + defaultAPIAddr + ")")
			if api == "" {
				api = defaultAPIAddr
			}
			a.cfg.Profiles[name] = profile{API: api, Token: token}
			if a.cfg.Current == "" {
				a.cfg.Current = name
			}
			saveConfig(a.cfg)
			fmt.Println("Saqlandi:", name)
		case 3:
			if len(names) == 0 {
				break
			}
			i, ok := selectMenu("Qaysi profilni o'chiramiz?", names)
			if !ok {
				break
			}
			name := names[i]
			delete(a.cfg.Profiles, name)
			if a.cfg.Current == name {
				a.cfg.Current = ""
				for n := range a.cfg.Profiles {
					a.cfg.Current = n
					break
				}
			}
			saveConfig(a.cfg)
			fmt.Println("O'chirildi:", name)
		case 4:
			return
		}
	}
}

func tuiWebhook(a *app) {
	idx, ok := selectMenu("Webhook", []string{"Joriy webhookni ko'rish", "Webhook sozlash", "Orqaga"})
	if !ok {
		return
	}
	switch idx {
	case 0:
		out, err := a.c.do("GET", "/api/webhook", nil)
		show(out, err)
	case 1:
		url, ok := askRequired("Webhook URL")
		if !ok {
			return
		}
		out, err := a.c.do("POST", "/api/webhook", map[string]any{"url": url})
		show(out, err)
	}
}

func tuiTelegram(a *app) {
	for {
		idx, ok := selectMenu("Telegram accountlar", []string{
			"Ro'yxatni ko'rish", "Account ulash", "Account o'chirish", "Orqaga",
		})
		if !ok {
			return
		}
		switch idx {
		case 0:
			out, err := a.c.do("GET", "/api/telegram", nil)
			show(out, err)
		case 1:
			tgConnect(a.c, "")
		case 2:
			id, ok := askRequired("O'chiriladigan telegram_account_id")
			if !ok {
				break
			}
			out, err := a.c.do("DELETE", "/api/telegram/"+id, nil)
			show(out, err)
		case 3:
			return
		}
	}
}

// tgConnect bitta oqimda Telegram account ulaydi:
// telefon -> kod yuborish -> kodni so'rash -> tasdiqlash -> (kerak bo'lsa) 2FA parol.
func tgConnect(c *client, phone string) {
	if phone == "" {
		var ok bool
		phone, ok = askRequired("Telefon raqami (+998...)")
		if !ok {
			return
		}
	}

	out, err := c.do("POST", "/api/telegram/send-code", map[string]any{"phone": phone})
	if err != nil {
		show(out, err)
		return
	}
	accountID := int64Of(out, "telegram_account_id")
	fmt.Println("📩 Tasdiqlash kodi Telegramga yuborildi.")

	code, ok := askRequired("Telegramdagi kodni kiriting")
	if !ok {
		return
	}

	out, err = c.do("POST", "/api/telegram/verify", map[string]any{
		"telegram_account_id": accountID, "code": code, "password": "",
	})
	if err != nil {
		show(out, err)
		return
	}

	// 2FA kerak bo'lsa parol so'rab, qaytadan tasdiqlaymiz.
	if d, ok := out["data"].(map[string]any); ok {
		if need, _ := d["need_password"].(bool); need {
			fmt.Println("🔐 2FA yoqilgan — parol kerak.")
			pass := askPassword("2FA parolingiz")
			out, err = c.do("POST", "/api/telegram/verify", map[string]any{
				"telegram_account_id": accountID, "code": code, "password": pass,
			})
			if err != nil {
				show(out, err)
				return
			}
		}
	}
	show(out, err)
}

// int64Of javob data'sidan raqamli maydonni oladi (JSON raqami float64 bo'ladi).
func int64Of(out map[string]any, key string) int64 {
	if d, ok := out["data"].(map[string]any); ok {
		if v, ok := d[key].(float64); ok {
			return int64(v)
		}
	}
	return 0
}

func tuiCards(a *app) {
	for {
		idx, ok := selectMenu("Cartalar", []string{"Ro'yxatni ko'rish", "Carta qo'shish", "Carta o'chirish", "Orqaga"})
		if !ok {
			return
		}
		switch idx {
		case 0:
			out, err := a.c.do("GET", "/api/cards", nil)
			show(out, err)
		case 1:
			id, _ := strconv.ParseInt(ask("telegram_account_id"), 10, 64)
			number, ok := askRequired("Carta raqami (to'liq)")
			if !ok {
				break
			}
			owner := ask("Carta egasining ismi")
			out, err := a.c.do("POST", "/api/cards", map[string]any{
				"telegram_account_id": id, "number": number, "owner_name": owner,
			})
			show(out, err)
		case 2:
			id, ok := askRequired("O'chiriladigan card id")
			if !ok {
				break
			}
			out, err := a.c.do("DELETE", "/api/cards/"+id, nil)
			show(out, err)
		case 3:
			return
		}
	}
}

func tuiTransaction(a *app) {
	for {
		idx, ok := selectMenu("Tranzaksiyalar", []string{"Yaratish", "Ro'yxatni ko'rish", "O'chirish", "Orqaga"})
		if !ok {
			return
		}
		switch idx {
		case 0:
			card, _ := strconv.ParseInt(ask("card_id (bo'sh = avtomatik eng kam yuklangan)"), 10, 64)
			amountStr, ok := askRequired("Summa (amount)")
			if !ok {
				break
			}
			amount, _ := strconv.ParseInt(amountStr, 10, 64)
			out, err := a.c.do("POST", "/api/transactions", map[string]any{"card_id": card, "amount": amount})
			show(out, err)
		case 1:
			out, err := a.c.do("GET", "/api/transactions", nil)
			if err != nil {
				show(out, err)
				break
			}
			printTransactions(out["data"])
		case 2:
			id, ok := askRequired("O'chiriladigan transaction id")
			if !ok {
				break
			}
			out, err := a.c.do("DELETE", "/api/transactions/"+id, nil)
			show(out, err)
		case 3:
			return
		}
	}
}
