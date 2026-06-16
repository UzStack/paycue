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
			"Transaction yaratish",
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
	password := askPassword("Parol (kamida 6 belgi)")
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
			"Profilga o'tish", "Profil qo'shish", "Profil o'chirish", "Orqaga",
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
		case 2:
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
		case 3:
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
			"Ro'yxatni ko'rish", "Account ulash (kod yuborish)", "Kodni tasdiqlash", "Orqaga",
		})
		if !ok {
			return
		}
		switch idx {
		case 0:
			out, err := a.c.do("GET", "/api/telegram", nil)
			show(out, err)
		case 1:
			phone, ok := askRequired("Telefon raqami (+998...)")
			if !ok {
				break
			}
			out, err := a.c.do("POST", "/api/telegram/send-code", map[string]any{"phone": phone})
			show(out, err)
			if err == nil {
				fmt.Println("Kod Telegramga yuborildi. 'Kodni tasdiqlash' bilan davom eting.")
			}
		case 2:
			id, _ := strconv.ParseInt(ask("telegram_account_id"), 10, 64)
			code, ok := askRequired("Tasdiqlash kodi")
			if !ok {
				break
			}
			password := askPassword("2FA paroli (bo'lmasa bo'sh qoldiring)")
			out, err := a.c.do("POST", "/api/telegram/verify", map[string]any{
				"telegram_account_id": id, "code": code, "password": password,
			})
			show(out, err)
		case 3:
			return
		}
	}
}

func tuiCards(a *app) {
	for {
		idx, ok := selectMenu("Cartalar", []string{"Ro'yxatni ko'rish", "Carta qo'shish", "Orqaga"})
		if !ok {
			return
		}
		switch idx {
		case 0:
			out, err := a.c.do("GET", "/api/cards", nil)
			show(out, err)
		case 1:
			id, _ := strconv.ParseInt(ask("telegram_account_id"), 10, 64)
			last4, ok := askRequired("Oxirgi 4 raqam")
			if !ok {
				break
			}
			label := ask("Nom (ixtiyoriy)")
			out, err := a.c.do("POST", "/api/cards", map[string]any{
				"telegram_account_id": id, "last4": last4, "label": label,
			})
			show(out, err)
		case 2:
			return
		}
	}
}

func tuiTransaction(a *app) {
	card, _ := strconv.ParseInt(ask("card_id"), 10, 64)
	amountStr, ok := askRequired("Summa (amount)")
	if !ok {
		return
	}
	amount, _ := strconv.ParseInt(amountStr, 10, 64)
	out, err := a.c.do("POST", "/api/transactions", map[string]any{"card_id": card, "amount": amount})
	show(out, err)
}
