#!/usr/bin/env bash
set -euo pipefail

# paycue o'rnatuvchi / yangilovchi skript.
#
#   Linux (to'liq): server + CLI o'rnatadi.
#     - o'rnatilmagan bo'lsa: APP_ID/APP_HASH so'raydi, .env tayyorlaydi, binarylarni
#       yuklaydi, systemd servisini sozlaydi, CLI ni o'rnatadi.
#     - o'rnatilgan bo'lsa: binarylarni oxirgi releasega yangilab, servisni qayta ishga tushiradi.
#
#   --cli-only  (yoki macOS): faqat paycue-cli ni o'rnatadi/yangilaydi (server yo'q).
#
# Foydalanish:
#   curl -fsSL https://raw.githubusercontent.com/UzStack/paycue/main/install.sh | sudo bash
#   curl -fsSL https://raw.githubusercontent.com/UzStack/paycue/main/install.sh | sudo bash -s -- --cli-only

REPO="UzStack/paycue"
INSTALL_DIR="/opt/paycue"
ENV_FILE="$INSTALL_DIR/.env"
SERVICE_FILE="/etc/systemd/system/paycue.service"
CLI_BIN="/usr/local/bin/paycue-cli"
SERVER_BIN="$INSTALL_DIR/paycue"

red()   { printf '\033[31m%s\033[0m\n' "$*"; }
green() { printf '\033[32m%s\033[0m\n' "$*"; }
info()  { printf '\033[36m%s\033[0m\n' "$*"; }

CLI_ONLY=false
for arg in "$@"; do
  case "$arg" in
    --cli-only) CLI_ONLY=true ;;
  esac
done

if [ "$(id -u)" -ne 0 ]; then
  red "Bu skript root huquqida ishlashi kerak. 'sudo' bilan ishga tushiring."
  exit 1
fi

# --- OS va arxitekturani aniqlash ---
case "$(uname -s)" in
  Linux)  OS="linux" ;;
  Darwin) OS="darwin" ;;
  *) red "Qo'llab-quvvatlanmaydigan OS: $(uname -s)"; exit 1 ;;
esac
case "$(uname -m)" in
  x86_64|amd64)   ARCH="amd64" ;;
  aarch64|arm64)  ARCH="arm64" ;;
  *) red "Qo'llab-quvvatlanmaydigan arxitektura: $(uname -m)"; exit 1 ;;
esac

# macOS da server (systemd) yo'q — faqat CLI.
if [ "$OS" = "darwin" ] && [ "$CLI_ONLY" != true ]; then
  info "macOS aniqlandi — server (systemd) qo'llab-quvvatlanmaydi, faqat CLI o'rnatiladi."
  CLI_ONLY=true
fi

DL="https://github.com/$REPO/releases/latest/download"

download_cli() {
  info "paycue-cli yuklanmoqda ($OS-$ARCH)..."
  curl -fsSL "$DL/paycue-cli-$OS-$ARCH" -o "$CLI_BIN.new"
  chmod +x "$CLI_BIN.new"
  mv "$CLI_BIN.new" "$CLI_BIN"
  green "paycue-cli o'rnatildi: $CLI_BIN"
  info "paycue-cli $("$CLI_BIN" version 2>/dev/null | awk '{print $NF}' || echo '?')"
}

download_server() {
  info "Server binary yuklanmoqda (linux-$ARCH)..."
  mkdir -p "$INSTALL_DIR"
  curl -fsSL "$DL/paycue-linux-$ARCH" -o "$SERVER_BIN.new"
  chmod +x "$SERVER_BIN.new"
  mv "$SERVER_BIN.new" "$SERVER_BIN"
  green "Server o'rnatildi: $SERVER_BIN"
  info "paycue     $("$SERVER_BIN" --version 2>/dev/null | awk '{print $NF}' || echo '?')"
}

create_service() {
  cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=paycue service
After=network.target

[Service]
User=root
Group=root
Type=simple
Restart=on-failure
RestartSec=5s
ExecStart=$SERVER_BIN
WorkingDirectory=$INSTALL_DIR

[Install]
WantedBy=multi-user.target
EOF
  systemctl daemon-reload
}

# ---------- faqat CLI ----------
if [ "$CLI_ONLY" = true ]; then
  download_cli
  echo
  green "Tayyor. paycue-cli faqat API client sifatida o'rnatildi."
  info "Boshlash:  paycue-cli            (interaktiv menu)"
  info "Masofaviy server:  paycue-cli --api https://your-server register ..."
  exit 0
fi

# ---------- Linux: UPDATE ----------
if [ -f "$ENV_FILE" ] && [ -f "$SERVER_BIN" ]; then
  info "paycue o'rnatilgan — yangilanmoqda..."
  download_server
  download_cli
  create_service
  systemctl restart paycue
  green "paycue yangilandi va qayta ishga tushirildi."
  systemctl --no-pager --full status paycue | head -n 5 || true
  exit 0
fi

# ---------- Linux: FRESH INSTALL ----------
info "paycue o'rnatilmoqda..."
echo
info "Telegram API ma'lumotlarini https://my.telegram.org dan oling."
read -rp "APP_ID: " APP_ID
read -rp "APP_HASH: " APP_HASH
read -rp "PORT [8080]: " PORT
PORT="${PORT:-8080}"

mkdir -p "$INSTALL_DIR"
cat > "$ENV_FILE" <<EOF
APP_ID=$APP_ID
APP_HASH=$APP_HASH
PORT=$PORT
DB_PATH=$INSTALL_DIR/db.sqlite3
SESSION_DIR=$INSTALL_DIR/sessions
WORKERS=10
TRANSACTION_TIMEOUT=30
DEBUG=false
EOF
chmod 600 "$ENV_FILE"
green ".env yaratildi: $ENV_FILE"

download_server
download_cli
create_service

systemctl enable --now paycue
green "paycue o'rnatildi va ishga tushdi."
echo
info "Tekshirish:   systemctl status paycue"
info "CLI:          paycue-cli            (interaktiv menu)"
info "API manzili:  http://127.0.0.1:$PORT"
