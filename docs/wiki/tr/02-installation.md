# 2. Kurulum

!!! important
    Kurulumdan önce **80 ve 443 portlarını** (HTTPS) ve alan adı DNS'ini hazırlayın.

---

## Ön Koşullar

| Öğe | Docker (önerilen) | Native |
|------|:--------------------:|:------:|
| İşletim sistemi | Linux (Ubuntu 22.04+) | Linux |
| RAM | 2 GB minimum | 2 GB+ |
| CPU | 1 vCPU | 1+ |
| Disk | 10 GB | 10 GB |
| Docker + Compose v2 | ✅ | Yalnızca DB/Redis |
| Go 1.26 | — | ✅ (otomatik kurulur) |
| Portlar | 80, 443 (+ inbound'lar) | aynı |

---

## Yöntem 1: Tek Satır Kurulum (Önerilen)

```bash
bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
```

Kurulum betiği **etkileşimlidir** ve iki ana soru sorar:

### 1. Kurulum yöntemi

| Seçenek | Açıklama |
|--------|-------------|
| **Docker Compose** *(önerilen)* | Tam yığın konteynerlerde: web · panel · node · PostgreSQL · Redis |
| **Native (systemd)** | Go ikili dosyası servis olarak; DB/Redis Docker'da; SPA Caddy ile |

### 2. Panel erişimi

| Seçenek | Açıklama |
|--------|-------------|
| **Alan adı + HTTPS** | Caddy Let's Encrypt sertifikası alır — 80 ve 443 portları açık olmalı |
| **IP + HTTP** | Özel port (ör. 8080) |

### Etkileşimsiz kurulum (betik/CI)

```bash
VORTEXUI_METHOD=docker \
VORTEXUI_NONINTERACTIVE=1 \
VORTEXUI_ADMIN_USER=admin \
VORTEXUI_ADMIN_PASS='your-strong-password' \
bash install.sh
```

### Kurulum çıktısı

- Kurulum yolu: `/opt/vortexui` (`VORTEXUI_DIR` ile geçersiz kılınabilir)
- `/usr/local/bin` içinde `vortexui` komutu
- Ortam dosyası: `deploy/.env` (JWT, DB şifresi, alan adı)
- mTLS sertifikaları: `deploy/certs/`
- Panel URL'si + ilk admin kimlik bilgileri terminalde yazdırılır

---

## Yöntem 2: Manuel Docker Compose

```bash
git clone https://github.com/iPmartNetwork/VortexUI && cd VortexUI

# Generate secrets
echo "JWT_SECRET=$(openssl rand -hex 32)" >> deploy/.env
echo "DB_PASSWORD=$(openssl rand -hex 16)" >> deploy/.env
echo "SITE_ADDRESS=panel.example.com" >> deploy/.env
echo "ACME_EMAIL=admin@example.com" >> deploy/.env

make certs
docker compose --env-file deploy/.env -f deploy/compose.yml up -d --build

# Create admin
docker compose -f deploy/compose.yml exec panel \
  /usr/local/bin/panel admin create --username admin --password 'change-me' --sudo
```

### Yığın servisleri

| Servis | Rol |
|---------|------|
| `db` | PostgreSQL 16 + TimescaleDB |
| `redis` | Redis 7 |
| `panel` | API + local node (host network) |
| `web` | Caddy + SPA (HTTPS) |

---

## Yöntem 3: Native Kurulum (Geliştirme/Gelişmiş)

```bash
git clone https://github.com/iPmartNetwork/VortexUI && cd VortexUI

docker compose up -d          # PostgreSQL + Redis
cp .env.example .env
# Fill VORTEX_JWT_SECRET with: openssl rand -hex 32

make build
make certs
make run-panel

# Another terminal — create admin
./bin/panel admin create --username admin --password 'your-password' --sudo
```

Frontend (geliştirme):

```bash
cd web && npm install && npm run dev
```

---

## Node Agent Kurulumu (Çok Sunuculu)

Çok node'lu bir filo için, her ayrı sunucuda:

```bash
VORTEX_NODE_LISTEN=:50051 \
VORTEX_CORE=xray \
VORTEX_CORE_BIN=/usr/local/bin/xray \
VORTEX_TLS_CERT=node.crt \
VORTEX_TLS_KEY=node.key \
VORTEX_TLS_CA=ca.crt \
./bin/node
```

Ardından panelde: **Nodes → Add Node** — adresi ve mTLS sertifikasını kaydedin.

---

## Local Node

Ayrı bir agent olmadan tek sunuculu kurulum için:

```env
VORTEX_LOCAL_NODE=true
VORTEX_LOCAL_NODE_NAME=local
VORTEX_LOCAL_NODE_HOST=your-public-ip-or-domain
VORTEX_CORE=xray
VORTEX_CORE_BIN=/usr/local/bin/xray
```

Docker Compose'ta bu varsayılan olarak etkindir.

---

## Önemli Ortam Değişkenleri

| Değişken | Varsayılan | Açıklama |
|----------|---------|-------------|
| `VORTEX_HTTP_ADDR` | `:8080` | Panel HTTP adresi |
| `VORTEX_DATABASE_URL` | — | **Gerekli** — PostgreSQL |
| `VORTEX_JWT_SECRET` | — | **Gerekli** — minimum 32 bayt |
| `VORTEX_REDIS_URL` | `redis://localhost:6379/0` | Redis |
| `VORTEX_LOCAL_NODE` | `false` | In-process node |
| `VORTEX_SHARE_AUTOLIMIT` | `false` | Hesap paylaşımında otomatik sınırlama |
| `VORTEX_WEBHOOK_URL` | — | Bildirim webhook'u |
| `VORTEX_TELEGRAM_TOKEN` | — | Telegram bot token'ı |
| `VORTEX_CF_API_TOKEN` | — | Cloudflare DNS otomasyonu |

Tam liste: [`.env.example`](../../../.env.example)

---

## Kurulum Sonrası Sağlık Kontrolü

```bash
vortexui status
curl -s http://127.0.0.1:8080/api/health
```

Beklenen yanıt: `{"status":"ok"}`

---

## Güncelleme

```bash
vortexui update
# or
cd /opt/vortexui && git pull && docker compose -f deploy/compose.yml up -d --build
```

Kurulum betiğini yeniden çalıştırmak **güvenlidir** — sırlar ve DB verileri korunur.
