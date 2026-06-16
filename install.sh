#!/usr/bin/env bash
set -euo pipefail

# paycue o'rnatuvchi / yangilovchi skript.
#   - o'rnatilmagan bo'lsa: APP_ID/APP_HASH so'raydi, .env tayyorlaydi, binarylarni
#     yuklaydi, systemd servisini sozlaydi va CLI ni o'rnatadi.
#   - o'rnatilgan bo'lsa: oxirgi releasedan binarylarni yangilab, servisni qayta ishga tushiradi.
#
# Foydalanish:  curl -fsSL https://raw.githubusercontent.com/UzStack/paycue/main/install.sh | sudo bash

REPO="UzStack/paycue"
INSTALL_DIR="/opt/paycue"
ENV_FILE="$INSTALL_DIR/.env"
SERVICE_FILE="/etc/systemd/system/paycue.service"
CLI_BIN="/usr/local/bin/paycue-cli"
SERVER_BIN="$INSTALL_DIR/paycue"

red()   { printf '\033[31m%s\033[0m\n' "$*"; }
green() { printf '\033[32m%s\033[0m\n' "$*"; }
info()  { printf '\033[36m%s\033[0m\n' "$*"; }

if [ "$(id -u)" -ne 0 ]; then
  red "Bu skript root huquqida ishlashi kerak. 'sudo' bilan ishga tushiring."
  exit 1
fi

# --- arxitekturani aniqlash ---
case "$(uname -m)" in
  x86_64|amd64)   ARCH="amd64" ;;
  aarch64|arm64)  ARCH="arm64" ;;
  *) red "Qo'llab-quvvatlanmaydigan arxitektura: $(uname -m)"; exit 1 ;;
esac

DL="https://github.com/$REPO/releases/latest/download"

download_binaries() {
  info "Binarylar yuklanmoqda (linux-$ARCH)..."
  mkdir -p "$INSTALL_DIR"
  curl -fsSL "$DL/paycue-linux-$ARCH"     -o "$SERVER_BIN.new"
  curl -fsSL "$DL/paycue-cli-linux-$ARCH" -o "$CLI_BIN.new"
  chmod +x "$SERVER_BIN.new" "$CLI_BIN.new"
  mv "$SERVER_BIN.new" "$SERVER_BIN"
  mv "$CLI_BIN.new" "$CLI_BIN"
  green "Binarylar yangilandi."
  # O'rnatilgan versiyalarni ko'rsatamiz (oxirgi maydon — versiya raqami).
  info "paycue     $("$SERVER_BIN" --version 2>/dev/null | awk '{print $NF}' || echo '?')"
  info "paycue-cli $("$CLI_BIN" version 2>/dev/null | awk '{print $NF}' || echo '?')"
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

if [ -f "$ENV_FILE" ] && [ -f "$SERVER_BIN" ]; then
  # ---------- UPDATE ----------
  info "paycue o'rnatilgan — yangilanmoqda..."
  download_binaries
  create_service
  systemctl restart paycue
  green "paycue yangilandi va qayta ishga tushirildi."
  systemctl --no-pager --full status paycue | head -n 5 || true
  exit 0
fi

# ---------- FRESH INSTALL ----------
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

download_binaries
create_service

systemctl enable --now paycue
green "paycue o'rnatildi va ishga tushdi."
echo
info "Tekshirish:   systemctl status paycue"
info "CLI:          paycue-cli register --name 'Ism' --email pochta@example.com"
info "API manzili:  http://127.0.0.1:$PORT"
