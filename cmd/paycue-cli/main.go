package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const VERSION = "2.3.0"
const defaultAPIAddr = "http://127.0.0.1:8080"

// ---- profil konfiguratsiyasi (bir nechta account) ----

type profile struct {
	API   string `json:"api"`
	Token string `json:"token"`
}

type cliConfig struct {
	Current  string             `json:"current"`
	Profiles map[string]profile `json:"profiles"`
}

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "paycue")
}

func configPath() string { return filepath.Join(configDir(), "config.json") }

func loadConfig() *cliConfig {
	cfg := &cliConfig{Profiles: map[string]profile{}}
	if b, err := os.ReadFile(configPath()); err == nil {
		_ = json.Unmarshal(b, cfg)
		if cfg.Profiles == nil {
			cfg.Profiles = map[string]profile{}
		}
	} else {
		// Eski yagona token faylini "default" profilga ko'chiramiz.
		if b, err := os.ReadFile(filepath.Join(configDir(), "token")); err == nil {
			tok := strings.TrimSpace(string(b))
			if tok != "" {
				cfg.Profiles["default"] = profile{API: defaultAPIAddr, Token: tok}
				cfg.Current = "default"
				saveConfig(cfg)
			}
		}
	}
	return cfg
}

func saveConfig(cfg *cliConfig) {
	_ = os.MkdirAll(configDir(), 0o700)
	b, _ := json.MarshalIndent(cfg, "", "  ")
	_ = os.WriteFile(configPath(), b, 0o600)
}

// ---- HTTP client ----

type client struct {
	api   string
	token string
}

func (c *client) do(method, path string, body any) (map[string]any, error) {
	var reader io.Reader
	if body != nil {
		buf, _ := json.Marshal(body)
		reader = bytes.NewReader(buf)
	}
	req, err := http.NewRequest(method, c.api+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	httpClient := &http.Client{Timeout: 6 * time.Minute} // login 2FA uchun uzun
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out map[string]any
	json.NewDecoder(resp.Body).Decode(&out)
	if resp.StatusCode >= 400 {
		return out, fmt.Errorf("API xato (%d): %v", resp.StatusCode, detailOf(out))
	}
	return out, nil
}

func detailOf(out map[string]any) string {
	if d, ok := out["data"].(map[string]any); ok {
		if s, ok := d["detail"].(string); ok {
			return s
		}
	}
	return fmt.Sprintf("%v", out)
}

func printJSON(v any) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}

// ---- profil/flag yechimi ----

type app struct {
	cfg         *cliConfig
	profileName string // tanlangan (yoki current) profil nomi
	apiFlag     string // --api (profil ustidan)
	tokenFlag   string // --token (profil ustidan)
	c           *client
}

// useProfile joriy profilni almashtirib, client'ni qayta quradi (TUI uchun).
func (a *app) useProfile(name string) {
	a.profileName = name
	prof := a.cfg.Profiles[name]
	a.c = &client{
		api:   firstNonEmpty(a.apiFlag, prof.API, os.Getenv("PAYCUE_API"), defaultAPIAddr),
		token: firstNonEmpty(a.tokenFlag, prof.Token, os.Getenv("PAYCUE_TOKEN")),
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func main() {
	var (
		apiFlag     = flag.String("api", "", "API manzili (profil/PAYCUE_API ustidan)")
		tokenFlag   = flag.String("token", "", "token (profil/PAYCUE_TOKEN ustidan)")
		profileFlag = flag.String("profile", "", "ishlatiladigan profil nomi (default: current)")
	)
	flag.Parse()
	args := flag.Args()

	cfg := loadConfig()
	profileName := firstNonEmpty(*profileFlag, cfg.Current, "default")
	prof := cfg.Profiles[profileName]

	a := &app{
		cfg:         cfg,
		profileName: profileName,
		apiFlag:     *apiFlag,
		tokenFlag:   *tokenFlag,
		c: &client{
			api:   firstNonEmpty(*apiFlag, prof.API, os.Getenv("PAYCUE_API"), defaultAPIAddr),
			token: firstNonEmpty(*tokenFlag, prof.Token, os.Getenv("PAYCUE_TOKEN")),
		},
	}

	// Argument bo'lmasa — interaktiv menu (TUI).
	if len(args) == 0 {
		runTUI(a)
		return
	}

	cmd := args[0]
	rest := args[1:]

	var err error
	switch cmd {
	case "version":
		fmt.Println("paycue-cli", VERSION)
		return
	case "profile":
		err = cmdProfile(a, rest)
	case "register":
		err = cmdRegister(a, rest)
	case "login":
		err = cmdLogin(a, rest)
	case "webhook":
		err = cmdWebhook(a.c, rest)
	case "telegram":
		err = cmdTelegram(a.c, rest)
	case "card":
		err = cmdCard(a.c, rest)
	case "transaction":
		err = cmdTransaction(a.c, rest)
	default:
		usage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Xato:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`paycue-cli — paycue API client

Argumentsiz ishga tushirilsa interaktiv menu (TUI) ochiladi.

Global flaglar:
  --api URL       API manzili (yoki PAYCUE_API, yoki profil)
  --token TOKEN   Token (yoki PAYCUE_TOKEN, yoki profil)
  --profile NAME  Ishlatiladigan profil (default: joriy profil)

Profil (bir nechta account):
  profile list                    Profillar ro'yxati
  profile current                 Joriy profil
  profile token [NAME]            Profil tokenini ko'rsatish (default: joriy)
  profile use NAME                Joriy profilni almashtirish
  profile add NAME --token T [--api URL]   Profil qo'shish/yangilash
  profile remove NAME             Profilni o'chirish

Buyruqlar:
  register --name NAME [--email E] [--phone P] --password PW [--profile NAME]
                                  Ro'yxatdan o'tish (tokenni profilga saqlaydi)
  login --login EMAIL|PHONE --password PW [--profile NAME]
                                  Parol bilan kirish (tokenni profilga saqlaydi)
  webhook [--url URL]             URL bo'lsa sozlaydi, bo'lmasa joriysini ko'rsatadi
  telegram connect [--phone +998..]  Account ulash (kod va 2FA ni interaktiv so'raydi)
  telegram list
  card add --account ID --last4 7159 [--label L]
  card list
  transaction create --card ID --amount 20000
  version`)
}

// ---- profil buyruqlari ----

func cmdProfile(a *app, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("subbuyruq kerak: list | current | token | use | add | remove")
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "list":
		if len(a.cfg.Profiles) == 0 {
			fmt.Println("Profillar yo'q. 'register' yoki 'profile add' bilan qo'shing.")
			return nil
		}
		names := make([]string, 0, len(a.cfg.Profiles))
		for n := range a.cfg.Profiles {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			marker := "  "
			if n == a.cfg.Current {
				marker = "* " // joriy profil
			}
			fmt.Printf("%s%s\t%s\n", marker, n, a.cfg.Profiles[n].API)
		}
		return nil
	case "current":
		if a.cfg.Current == "" {
			fmt.Println("Joriy profil tanlanmagan.")
			return nil
		}
		fmt.Println(a.cfg.Current)
		return nil
	case "token":
		// profile token [NAME] — profil tokenini chiqaradi (default: joriy).
		name := a.cfg.Current
		if len(rest) > 0 {
			name = rest[0]
		}
		p, ok := a.cfg.Profiles[name]
		if !ok {
			return fmt.Errorf("profil topilmadi: %s", name)
		}
		fmt.Println(p.Token)
		return nil
	case "use":
		if len(rest) == 0 {
			return fmt.Errorf("profil nomi kerak: profile use NAME")
		}
		name := rest[0]
		if _, ok := a.cfg.Profiles[name]; !ok {
			return fmt.Errorf("profil topilmadi: %s", name)
		}
		a.cfg.Current = name
		saveConfig(a.cfg)
		fmt.Println("Joriy profil:", name)
		return nil
	case "add":
		fs := flag.NewFlagSet("add", flag.ExitOnError)
		token := fs.String("token", "", "foydalanuvchi tokeni")
		api := fs.String("api", defaultAPIAddr, "API manzili")
		if len(rest) == 0 {
			return fmt.Errorf("profil nomi kerak: profile add NAME --token T")
		}
		name := rest[0]
		fs.Parse(rest[1:])
		if *token == "" {
			return fmt.Errorf("--token majburiy")
		}
		a.cfg.Profiles[name] = profile{API: *api, Token: *token}
		if a.cfg.Current == "" {
			a.cfg.Current = name
		}
		saveConfig(a.cfg)
		fmt.Println("Profil saqlandi:", name)
		return nil
	case "remove":
		if len(rest) == 0 {
			return fmt.Errorf("profil nomi kerak: profile remove NAME")
		}
		name := rest[0]
		if _, ok := a.cfg.Profiles[name]; !ok {
			return fmt.Errorf("profil topilmadi: %s", name)
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
		fmt.Println("Profil o'chirildi:", name)
		return nil
	}
	return fmt.Errorf("noma'lum subbuyruq: %s", sub)
}

// ---- buyruqlar ----

func cmdRegister(a *app, args []string) error {
	fs := flag.NewFlagSet("register", flag.ExitOnError)
	name := fs.String("name", "", "ism familiya")
	email := fs.String("email", "", "pochta")
	phone := fs.String("phone", "", "telefon")
	password := fs.String("password", "", "parol (kamida 6 belgi)")
	profName := fs.String("profile", "", "saqlanadigan profil nomi (default: default)")
	fs.Parse(args)
	out, err := a.c.do("POST", "/api/register", map[string]any{
		"name": *name, "email": *email, "phone": *phone, "password": *password,
	})
	if err != nil {
		return err
	}
	a.saveTokenFromResponse(out, *profName)
	printJSON(out["data"])
	return nil
}

func cmdLogin(a *app, args []string) error {
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	login := fs.String("login", "", "email yoki telefon")
	password := fs.String("password", "", "parol")
	profName := fs.String("profile", "", "saqlanadigan profil nomi (default: default)")
	fs.Parse(args)
	out, err := a.c.do("POST", "/api/login", map[string]any{"login": *login, "password": *password})
	if err != nil {
		return err
	}
	a.saveTokenFromResponse(out, *profName)
	printJSON(out["data"])
	return nil
}

// saveTokenFromResponse javobdagi tokenni nomli profilga saqlaydi va joriy qiladi.
func (a *app) saveTokenFromResponse(out map[string]any, profName string) {
	d, ok := out["data"].(map[string]any)
	if !ok {
		return
	}
	t, ok := d["token"].(string)
	if !ok || t == "" {
		return
	}
	pName := firstNonEmpty(profName, "default")
	a.cfg.Profiles[pName] = profile{API: a.c.api, Token: t}
	a.cfg.Current = pName
	saveConfig(a.cfg)
	fmt.Printf("Token '%s' profiliga saqlandi (joriy qilib belgilandi).\n", pName)
}

func cmdWebhook(c *client, args []string) error {
	fs := flag.NewFlagSet("webhook", flag.ExitOnError)
	url := fs.String("url", "", "webhook url (bo'sh bo'lsa joriysini ko'rsatadi)")
	fs.Parse(args)
	var out map[string]any
	var err error
	if *url == "" {
		out, err = c.do("GET", "/api/webhook", nil) // joriy webhookni ko'rish
	} else {
		out, err = c.do("POST", "/api/webhook", map[string]any{"url": *url})
	}
	if err != nil {
		return err
	}
	printJSON(out["data"])
	return nil
}

func cmdTelegram(c *client, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("subbuyruq kerak: connect | list")
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "connect":
		fs := flag.NewFlagSet("connect", flag.ExitOnError)
		phone := fs.String("phone", "", "telefon raqami")
		fs.Parse(rest)
		// Bitta interaktiv oqim: kod yuborish -> kod -> (kerak bo'lsa) 2FA.
		tgConnect(c, *phone)
		return nil
	case "list":
		out, err := c.do("GET", "/api/telegram", nil)
		if err != nil {
			return err
		}
		printJSON(out["data"])
		return nil
	}
	return fmt.Errorf("noma'lum subbuyruq: %s", sub)
}

func cmdCard(c *client, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("subbuyruq kerak: add | list")
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "add":
		fs := flag.NewFlagSet("add", flag.ExitOnError)
		account := fs.Int64("account", 0, "telegram_account_id")
		last4 := fs.String("last4", "", "oxirgi 4 raqam")
		label := fs.String("label", "", "nom")
		fs.Parse(rest)
		out, err := c.do("POST", "/api/cards", map[string]any{
			"telegram_account_id": *account, "last4": *last4, "label": *label,
		})
		if err != nil {
			return err
		}
		printJSON(out["data"])
		return nil
	case "list":
		out, err := c.do("GET", "/api/cards", nil)
		if err != nil {
			return err
		}
		printJSON(out["data"])
		return nil
	}
	return fmt.Errorf("noma'lum subbuyruq: %s", sub)
}

func cmdTransaction(c *client, args []string) error {
	if len(args) == 0 || args[0] != "create" {
		return fmt.Errorf("foydalanish: transaction create --card ID --amount N")
	}
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	card := fs.Int64("card", 0, "card_id")
	amount := fs.Int64("amount", 0, "summa")
	fs.Parse(args[1:])
	out, err := c.do("POST", "/api/transactions", map[string]any{"card_id": *card, "amount": *amount})
	if err != nil {
		return err
	}
	printJSON(out["data"])
	return nil
}
