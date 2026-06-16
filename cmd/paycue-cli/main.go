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
	"strings"
	"time"
)

const VERSION = "2.0.0"

func defaultAPI() string {
	if v := os.Getenv("PAYCUE_API"); v != "" {
		return v
	}
	return "http://127.0.0.1:8080"
}

func tokenPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "paycue", "token")
}

func loadToken() string {
	if v := os.Getenv("PAYCUE_TOKEN"); v != "" {
		return v
	}
	b, err := os.ReadFile(tokenPath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func saveToken(token string) {
	p := tokenPath()
	_ = os.MkdirAll(filepath.Dir(p), 0o700)
	_ = os.WriteFile(p, []byte(token), 0o600)
}

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

func main() {
	var (
		api   = flag.String("api", defaultAPI(), "API manzili")
		token = flag.String("token", loadToken(), "foydalanuvchi tokeni")
	)
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	c := &client{api: *api, token: *token}
	cmd := args[0]
	rest := args[1:]

	var err error
	switch cmd {
	case "version":
		fmt.Println("paycue-cli", VERSION)
		return
	case "register":
		err = cmdRegister(c, rest)
	case "webhook":
		err = cmdWebhook(c, rest)
	case "telegram":
		err = cmdTelegram(c, rest)
	case "card":
		err = cmdCard(c, rest)
	case "transaction":
		err = cmdTransaction(c, rest)
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

Global flaglar:
  --api URL      API manzili (yoki PAYCUE_API env, default http://127.0.0.1:8080)
  --token TOKEN  Token (yoki PAYCUE_TOKEN env, yoki ~/.config/paycue/token)

Buyruqlar:
  register --name NAME [--email E] [--phone P]   Ro'yxatdan o'tish (token qaytaradi va saqlaydi)
  webhook --url URL                              Webhook URL sozlash (secret qaytaradi)
  telegram send-code --phone +998..              Telegram account ulashni boshlash (kod yuboradi)
  telegram verify --account ID --code 12345 [--password 2FA]
  telegram list                                  Telegram accountlar ro'yxati
  card add --account ID --last4 7159 [--label L] Carta qo'shish
  card list                                      Cartalar ro'yxati
  transaction create --card ID --amount 20000    Transaction yaratish
  version`)
}

func cmdRegister(c *client, args []string) error {
	fs := flag.NewFlagSet("register", flag.ExitOnError)
	name := fs.String("name", "", "ism familiya")
	email := fs.String("email", "", "pochta")
	phone := fs.String("phone", "", "telefon")
	fs.Parse(args)
	out, err := c.do("POST", "/api/register", map[string]any{"name": *name, "email": *email, "phone": *phone})
	if err != nil {
		return err
	}
	if d, ok := out["data"].(map[string]any); ok {
		if t, ok := d["token"].(string); ok {
			saveToken(t)
			fmt.Println("Token saqlandi:", tokenPath())
		}
	}
	printJSON(out["data"])
	return nil
}

func cmdWebhook(c *client, args []string) error {
	fs := flag.NewFlagSet("webhook", flag.ExitOnError)
	url := fs.String("url", "", "webhook url")
	fs.Parse(args)
	out, err := c.do("POST", "/api/webhook", map[string]any{"url": *url})
	if err != nil {
		return err
	}
	printJSON(out["data"])
	return nil
}

func cmdTelegram(c *client, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("subbuyruq kerak: send-code | verify | list")
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "send-code":
		fs := flag.NewFlagSet("send-code", flag.ExitOnError)
		phone := fs.String("phone", "", "telefon raqami")
		fs.Parse(rest)
		out, err := c.do("POST", "/api/telegram/send-code", map[string]any{"phone": *phone})
		if err != nil {
			return err
		}
		printJSON(out["data"])
		return nil
	case "verify":
		fs := flag.NewFlagSet("verify", flag.ExitOnError)
		account := fs.Int64("account", 0, "telegram_account_id")
		code := fs.String("code", "", "tasdiqlash kodi")
		password := fs.String("password", "", "2FA paroli (kerak bo'lsa)")
		fs.Parse(rest)
		out, err := c.do("POST", "/api/telegram/verify", map[string]any{
			"telegram_account_id": *account, "code": *code, "password": *password,
		})
		if err != nil {
			return err
		}
		printJSON(out["data"])
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
