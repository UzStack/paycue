package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

var stdin = bufio.NewReader(os.Stdin)

func prompt(label string) string {
	fmt.Printf("%s: ", label)
	line, _ := stdin.ReadString('\n')
	return strings.TrimSpace(line)
}

func promptDefault(label, def string) string {
	v := prompt(fmt.Sprintf("%s [%s]", label, def))
	if v == "" {
		return def
	}
	return v
}

func pause() {
	fmt.Print("\nDavom etish uchun Enter...")
	stdin.ReadString('\n')
}

func header(a *app) {
	fmt.Println("\n==================== paycue-cli ====================")
	tok := "yo'q"
	if a.c.token != "" {
		tok = "bor"
	}
	fmt.Printf("Profil: %s | API: %s | token: %s\n", a.profileName, a.c.api, tok)
	fmt.Println("----------------------------------------------------")
}

// runTUI interaktiv menyu.
func runTUI(a *app) {
	for {
		header(a)
		fmt.Println(`1) Ro'yxatdan o'tish (register)
2) Kirish (login)
3) Profillar
4) Webhook sozlash
5) Telegram accountlar
6) Cartalar
7) Transaction yaratish
0) Chiqish`)
		switch prompt("Tanlang") {
		case "1":
			tuiRegister(a)
		case "2":
			tuiLogin(a)
		case "3":
			tuiProfiles(a)
		case "4":
			tuiWebhook(a)
		case "5":
			tuiTelegram(a)
		case "6":
			tuiCards(a)
		case "7":
			tuiTransaction(a)
		case "0", "q", "exit":
			fmt.Println("Xayr!")
			return
		default:
			fmt.Println("Noma'lum tanlov.")
		}
	}
}

func show(out map[string]any, err error) {
	if err != nil {
		fmt.Println("❌ Xato:", err)
		return
	}
	fmt.Println("✅ Muvaffaqiyatli:")
	printJSON(out["data"])
}

func tuiRegister(a *app) {
	name := prompt("Ism familiya")
	email := prompt("Email (ixtiyoriy)")
	phone := prompt("Telefon (ixtiyoriy)")
	password := prompt("Parol (kamida 6 belgi)")
	profName := promptDefault("Saqlanadigan profil nomi", "default")
	out, err := a.c.do("POST", "/api/register", map[string]any{
		"name": name, "email": email, "phone": phone, "password": password,
	})
	if err == nil {
		a.saveTokenFromResponse(out, profName)
		a.useProfile(a.cfg.Current)
	}
	show(out, err)
	pause()
}

func tuiLogin(a *app) {
	login := prompt("Email yoki telefon")
	password := prompt("Parol")
	profName := promptDefault("Saqlanadigan profil nomi", "default")
	out, err := a.c.do("POST", "/api/login", map[string]any{"login": login, "password": password})
	if err == nil {
		a.saveTokenFromResponse(out, profName)
		a.useProfile(a.cfg.Current)
	}
	show(out, err)
	pause()
}

func tuiProfiles(a *app) {
	for {
		fmt.Println("\n--- Profillar ---")
		names := make([]string, 0, len(a.cfg.Profiles))
		for n := range a.cfg.Profiles {
			names = append(names, n)
		}
		sort.Strings(names)
		if len(names) == 0 {
			fmt.Println("(profillar yo'q)")
		}
		for _, n := range names {
			marker := "  "
			if n == a.cfg.Current {
				marker = "* "
			}
			fmt.Printf("%s%s\t%s\n", marker, n, a.cfg.Profiles[n].API)
		}
		fmt.Println("\n1) Profilga o'tish   2) Profil qo'shish   3) Profil o'chirish   0) Orqaga")
		switch prompt("Tanlang") {
		case "1":
			name := prompt("Profil nomi")
			if _, ok := a.cfg.Profiles[name]; !ok {
				fmt.Println("Profil topilmadi.")
				break
			}
			a.cfg.Current = name
			saveConfig(a.cfg)
			a.useProfile(name)
			fmt.Println("Joriy profil:", name)
		case "2":
			name := prompt("Profil nomi")
			token := prompt("Token")
			api := promptDefault("API manzili", defaultAPIAddr)
			if name == "" || token == "" {
				fmt.Println("nom va token majburiy.")
				break
			}
			a.cfg.Profiles[name] = profile{API: api, Token: token}
			if a.cfg.Current == "" {
				a.cfg.Current = name
			}
			saveConfig(a.cfg)
			fmt.Println("Saqlandi:", name)
		case "3":
			name := prompt("O'chiriladigan profil nomi")
			if _, ok := a.cfg.Profiles[name]; !ok {
				fmt.Println("Profil topilmadi.")
				break
			}
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
		case "0", "":
			return
		}
	}
}

func tuiWebhook(a *app) {
	url := prompt("Webhook URL")
	out, err := a.c.do("POST", "/api/webhook", map[string]any{"url": url})
	show(out, err)
	pause()
}

func tuiTelegram(a *app) {
	for {
		fmt.Println("\n--- Telegram accountlar ---")
		fmt.Println("1) Account ulash (kod yuborish)   2) Kodni tasdiqlash   3) Ro'yxat   0) Orqaga")
		switch prompt("Tanlang") {
		case "1":
			phone := prompt("Telefon raqami (+998...)")
			out, err := a.c.do("POST", "/api/telegram/send-code", map[string]any{"phone": phone})
			show(out, err)
			if err == nil {
				fmt.Println("\nKod Telegramga yuborildi. '2' bilan tasdiqlang.")
			}
			pause()
		case "2":
			id, _ := strconv.ParseInt(prompt("telegram_account_id"), 10, 64)
			code := prompt("Tasdiqlash kodi")
			password := prompt("2FA paroli (bo'lmasa bo'sh qoldiring)")
			out, err := a.c.do("POST", "/api/telegram/verify", map[string]any{
				"telegram_account_id": id, "code": code, "password": password,
			})
			show(out, err)
			pause()
		case "3":
			out, err := a.c.do("GET", "/api/telegram", nil)
			show(out, err)
			pause()
		case "0", "":
			return
		}
	}
}

func tuiCards(a *app) {
	for {
		fmt.Println("\n--- Cartalar ---")
		fmt.Println("1) Carta qo'shish   2) Ro'yxat   0) Orqaga")
		switch prompt("Tanlang") {
		case "1":
			id, _ := strconv.ParseInt(prompt("telegram_account_id"), 10, 64)
			last4 := prompt("Oxirgi 4 raqam")
			label := prompt("Nom (ixtiyoriy)")
			out, err := a.c.do("POST", "/api/cards", map[string]any{
				"telegram_account_id": id, "last4": last4, "label": label,
			})
			show(out, err)
			pause()
		case "2":
			out, err := a.c.do("GET", "/api/cards", nil)
			show(out, err)
			pause()
		case "0", "":
			return
		}
	}
}

func tuiTransaction(a *app) {
	card, _ := strconv.ParseInt(prompt("card_id"), 10, 64)
	amount, _ := strconv.ParseInt(prompt("Summa (amount)"), 10, 64)
	out, err := a.c.do("POST", "/api/transactions", map[string]any{"card_id": card, "amount": amount})
	show(out, err)
	pause()
}
