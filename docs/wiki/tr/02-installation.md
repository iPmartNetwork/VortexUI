# Kurulum

!!! success "Önerilen"
    Çalışan bir panele en hızlı ulaşmak için **tek satır kurulum betiğini** kullanın. Bağımlılıkları, veritabanı kurulumunu, HTTPS'yi ve systemd servislerini otomatik olarak yönetir.

---

## Gereksinimler

| Gereksinim | Minimum | Önerilen |
|------------|---------|----------|
| İşletim Sistemi | Ubuntu 20.04 / Debian 11 | Ubuntu 22.04+ / Debian 12 |
| RAM | 1 GB | 2 GB+ |
| Disk | 10 GB | 20 GB+ (TimescaleDB trafik verileriyle büyür) |
| CPU | 1 vCPU | 2+ vCPU |
| Go (yerel derleme için) | 1.26 | 1.26 |
| Docker (konteyner kurulumu) | 24.0+ | En son kararlı sürüm |
| Alan Adı | İsteğe bağlı | Önerilen (HTTPS + abonelikler için) |

---

## Tek Satır Kurulum

```bash
bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
```

Kurulum betiği şunları yapacaktır:

1. İşletim sisteminizi ve mimarinizi algılar
2. Bağımlılıkları kurar (PostgreSQL, Redis, Caddy)
3. VortexUI'yı indirir ve derler
4. Veritabanı göçlerini çalıştırır
5. Bir sudo yönetici hesabı oluşturur (etkileşimli istem)
6. Systemd servislerini yapılandırır
7. Caddy ile HTTPS ayarlar (alan adı verilmişse)

Tamamlandığında panele `https://alan-adiniz.com` veya `http://sunucu-ip:8080` adresinden erişebilirsiniz.

---

## Docker Compose

=== "Hızlı Başlangıç"

    ```bash
    git clone https://github.com/iPmartNetwork/VortexUI.git
    cd VortexUI/deploy
    cp ../.env.example .env
    # .env dosyasını ayarlarınıza göre düzenleyin
    docker compose up -d
    ```

=== "Üretim (Caddy HTTPS ile)"

    ```bash
    git clone https://github.com/iPmartNetwork/VortexUI.git
    cd VortexUI/deploy
    cp ../.env.example .env
    ```

    `.env` dosyasını düzenleyin:
    ```env
    VORTEX_DOMAIN=panel.example.com
    VORTEX_ADMIN_USER=admin
    VORTEX_ADMIN_PASS=guvenli-sifreniz
    VORTEX_JWT_SECRET=rastgele-32-byte-dize
    VORTEX_DB_URL=postgres://vortex:pass@db:5432/vortex?sslmode=disable
    VORTEX_REDIS_URL=redis://redis:6379/0
    ```

    Ardından:
    ```bash
    docker compose up -d
    ```

`deploy/compose.yml` şunları içerir: panel, web ön yüzü, PostgreSQL + TimescaleDB, Redis ve Caddy.

---

## Yerel Derleme

=== "Ubuntu/Debian"

    ```bash
    # Go 1.26 kurulumu
    sudo snap install go --classic
    go version  # go1.26.x göstermelidir

    # Bağımlılıkları kurun
    sudo apt update && sudo apt install -y postgresql redis-server

    # Klonlayın ve derleyin
    git clone https://github.com/iPmartNetwork/VortexUI.git
    cd VortexUI
    go build -o vortexui ./cmd/panel

    # Göçleri çalıştırın
    ./vortexui migrate

    # Yönetici oluşturun
    ./vortexui admin create --username admin --password sifreniz --sudo

    # Başlatın
    ./vortexui serve
    ```

=== "Diğer Linux"

    ```bash
    # Resmi tarball'dan Go 1.26 kurulumu
    wget https://go.dev/dl/go1.26.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.26.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin

    # Ardından Ubuntu ile aynı klonlama/derleme adımlarını izleyin
    ```

!!! warning "Go Sürümü"
    VortexUI **Go 1.26** veya üstünü gerektirir. Daha eski sürümler derleme hatası verecektir.

---

## Düğüm Ajanı Kurulumu

Düğüm ajanı uzak sunucularda çalışır ve panel ile gRPC + mTLS üzerinden iletişim kurar.

=== "Kayıt Sihirbazı (Önerilen)"

    1. Panel arayüzünde **Düğümler → Düğüm Ekle** bölümüne gidin
    2. Kayıt sihirbazı tek satırlık bir kurulum komutu oluşturur
    3. Uzak sunucunuza SSH ile bağlanıp komutu yapıştırın
    4. Ajan otomatik olarak kayıt olur, sertifika değişimini yapar ve raporlamaya başlar

=== "Manuel Kurulum"

    ```bash
    # Uzak sunucuda
    bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install-node.sh)
    ```

    Şunlar sorulacaktır:
    - Panel adresi (örn. `https://panel.example.com`)
    - Düğüm kayıt tokeni (panel arayüzünde oluşturulur)

=== "Docker Düğüm"

    ```bash
    docker run -d --name vortex-node \
      -e PANEL_ADDR=https://panel.example.com \
      -e NODE_TOKEN=kayit-tokeniniz \
      --network host \
      ghcr.io/ipmartnetwork/vortexui-node:latest
    ```

---

## Yerel Düğüm (Tek Sunucu)

Yalnızca bir sunucuya ihtiyacınız varsa **yerel düğümü** kullanın — proxy çekirdeği panel ile birlikte süreç içinde çalışır. Ayrı ajan gerekmez.

1. Kurulum sırasında yerel düğüm sorulduğunda "Evet" seçin
2. Veya sonradan: **Düğümler → Düğüm Ekle → Yerel**
3. Çekirdek seçin (Xray veya sing-box)
4. Panel çekirdek sürecini doğrudan yönetir

!!! tip
    Yerel düğüm tek sunucu kurulumları için idealdir. Çoklu sunucu dağıtımları için kayıt sihirbazı ile uzak düğümler kullanın.

---

## Ortam Değişkenleri

| Değişken | Açıklama | Varsayılan |
|----------|----------|------------|
| `VORTEX_DOMAIN` | Panel alan adı (HTTPS için) | — |
| `VORTEX_LISTEN` | API dinleme adresi | `:8080` |
| `VORTEX_DB_URL` | PostgreSQL bağlantı dizesi | `postgres://localhost/vortex` |
| `VORTEX_REDIS_URL` | Redis bağlantı dizesi | `redis://localhost:6379/0` |
| `VORTEX_JWT_SECRET` | JWT imzalama anahtarı (≥32 byte) | — (zorunlu) |
| `VORTEX_ADMIN_USER` | İlk yönetici kullanıcı adı | — |
| `VORTEX_ADMIN_PASS` | İlk yönetici şifresi | — |
| `VORTEX_TELEGRAM_TOKEN` | Telegram bot tokeni | — |
| `VORTEX_TELEGRAM_ADMIN` | Bildirimler için yönetici sohbet ID'si | — |
| `VORTEX_ZARINPAL_MERCHANT` | ZarinPal satıcı ID'si | — |
| `VORTEX_NOWPAYMENTS_KEY` | NowPayments API anahtarı | — |
| `VORTEX_NOWPAYMENTS_IPN_SECRET` | NowPayments IPN HMAC gizli anahtarı | — |
| `VORTEX_BACKUP_CRON` | Yedekleme zamanlaması (cron ifadesi) | — |
| `VORTEX_BACKUP_TELEGRAM` | Yedekleri Telegram'a gönder | `false` |
| `VORTEX_BACKUP_S3_BUCKET` | Yedeklemeler için S3 kovası | — |
| `VORTEX_METRICS_ENABLED` | Prometheus metriklerini etkinleştir | `false` |
| `VORTEX_METRICS_LISTEN` | Metrik uç noktası adresi | `:9090` |
| `VORTEX_SHARE_AUTOLIMIT` | Hesap paylaşım tespitinde otomatik limit | `false` |

---

## CLI Yönetimi

`vortexui` ikili dosyası etkileşimli bir menü sunar:

```bash
vortexui
```

```
╔══════════════════════════════════════╗
║          VortexUI Yönetimi           ║
╠══════════════════════════════════════╣
║  1) Paneli başlat                    ║
║  2) Paneli durdur                    ║
║  3) Paneli yeniden başlat            ║
║  4) Durum                            ║
║  5) Günlükler (canlı)               ║
║  6) Güncelle                         ║
║  7) Yönetici yönetimi                ║
║  8) Yedekleme                        ║
║  9) Doktor (tanılama)                ║
║  0) Çıkış                            ║
╚══════════════════════════════════════╝
```

Temel komutlar:

| Komut | Eylem |
|-------|-------|
| `vortexui update` | Son sürümü çek ve yeniden başlat |
| `vortexui admin create` | Yeni yönetici oluştur |
| `vortexui admin reset-password` | Yönetici şifresini sıfırla |
| `vortexui backup` | Anında yedekleme oluştur |
| `vortexui doctor` | Tanılama çalıştır (DB, Redis, düğümler, portlar) |
| `vortexui migrate` | Bekleyen veritabanı göçlerini çalıştır |

---

## Güncelleme

=== "Otomatik Güncelleme (Önerilen)"

    ```bash
    vortexui update
    ```

    Bu komut son sürümü çeker, yeniden derler, göçleri çalıştırır ve yeniden başlatır.

=== "Manuel Güncelleme (Panel Sunucusu)"

    ```bash
    cd /opt/VortexUI  # veya klonladığınız konum
    git pull origin master
    go build -o vortexui ./cmd/panel
    ./vortexui migrate
    sudo systemctl restart vortexui
    ```

=== "Manuel Güncelleme (Düğüm Sunucuları)"

    ```bash
    cd /opt/VortexUI-node
    git pull origin master
    go build -o vortex-node ./cmd/node
    sudo systemctl restart vortex-node
    ```

=== "Docker Güncelleme"

    ```bash
    cd /opt/VortexUI/deploy
    docker compose pull
    docker compose up -d
    ```

---

## Kurulum Sonrası Doğrulama

Kurulumdan sonra her şeyin çalıştığını doğrulayın:

1. **Panel erişilebilir** — tarayıcıda `https://alan-adiniz.com` adresini açın
2. **Giriş çalışıyor** — yönetici kimlik bilgilerinizle oturum açın
3. **Veritabanı bağlı** — Ayarlar → Sistem Bilgisi'ni kontrol edin
4. **Düğüm çevrimiçi** — yerel düğüm kullanıyorsanız Düğümler sayfasında "Çevrimiçi" göründüğünü doğrulayın
5. **Tanılama çalıştırın** — `vortexui doctor` tüm bileşenleri kontrol eder

!!! tip "Sağlık Uç Noktası"
    Panel `GET /api/health` uç noktasını sunar — bileşen durumu ile `200 OK` döndürür.
    Harici izleme için kullanın (UptimeRobot, Prometheus blackbox, vb.).
