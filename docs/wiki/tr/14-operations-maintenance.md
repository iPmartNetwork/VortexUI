# Operasyonlar ve Bakım

---

## Caddy ile HTTPS

VortexUI, otomatik HTTPS için web sunucusu olarak Caddy kullanır:

- **Otomatik Let's Encrypt** — sıfır yapılandırma ile sertifika düzenleme ve yenileme
- **HTTP → HTTPS yönlendirme** — tüm trafik HTTPS'ye zorlanır
- **SPA yönlendirme** — React ön yüzü doğru şekilde sunulur
- **Ters proxy** — API istekleri Go backend'e yönlendirilir
- **DoH uç noktası** — DoH etkinse `/dns-query` sunulur

### Caddyfile Yapısı

```
{$VORTEX_DOMAIN} {
    encode gzip
    handle /api/* {
        reverse_proxy localhost:8080
    }
    handle /sub/* {
        reverse_proxy localhost:8080
    }
    handle /dns-query {
        reverse_proxy localhost:8053
    }
    handle {
        root * /opt/vortexui/web/dist
        try_files {path} /index.html
        file_server
    }
}
```

!!! tip
    Özel TLS ayarlarına ihtiyacınız varsa (örn. istemci sertifikası, belirli şifre paketleri), `deploy/Caddyfile` dosyasını düzenleyin.

---

## Prometheus Metrikleri

**Etkinleştirme:** `VORTEX_METRICS_ENABLED=true` ve `VORTEX_METRICS_LISTEN=:9090` ayarlayın.

### Mevcut Metrikler

| Metrik | Tür | Açıklama |
|--------|-----|----------|
| `vortex_users_total` | Gauge | Duruma göre toplam kullanıcı sayısı |
| `vortex_traffic_bytes` | Counter | Trafik baytları (yükleme/indirme) |
| `vortex_connections_active` | Gauge | Mevcut aktif bağlantılar |
| `vortex_nodes_status` | Gauge | Düğüm durumu (1=çevrimiçi, 0=çevrimdışı) |
| `vortex_node_cpu_percent` | Gauge | Düğüm CPU kullanımı |
| `vortex_node_memory_percent` | Gauge | Düğüm bellek kullanımı |
| `vortex_orders_total` | Counter | Duruma göre siparişler |
| `vortex_api_requests_total` | Counter | Uç nokta/duruma göre API istekleri |
| `vortex_api_latency_seconds` | Histogram | API yanıt gecikmesi |

### Grafana Gösterge Paneli

`deploy/grafana/vortexui-dashboard.json` dosyasından hazır gösterge panelini içe aktarın:

1. Grafana'yı açın → Dashboards → Import
2. JSON dosyasını yükleyin veya içeriğini yapıştırın
3. Prometheus veri kaynağınızı seçin
4. Gösterge paneli gösterir: trafik trendleri, düğüm sağlığı, kullanıcı etkinliği, API performansı

---

## Ölçekleme ve Yüksek Erişilebilirlik

### Yatay Ölçekleme

Büyük dağıtımlar için:

| Bileşen | Ölçekleme Stratejisi |
|---------|----------------------|
| Panel API | Yük dengeleyici arkasında birden fazla örnek çalıştır |
| Veritabanı | Okuma replikaları ile PostgreSQL |
| Redis | Redis Cluster veya Sentinel |
| Düğümler | Gerektiği kadar ekle (otomatik yönetilir) |
| Federasyon | Bölgesel dağıtım için birden fazla paneli bağla |

### Küme Modu

Birden fazla panel örneği çalıştırın:

1. Tüm örnekler aynı PostgreSQL ve Redis'i paylaşır
2. Önde bir yük dengeleyici kullanın (Caddy, nginx, HAProxy)
3. SSE bağlantıları yapışkandır (oturum benzeşimi önerilir)
4. Arka plan işleri Redis tabanlı dağıtık kilitler kullanır

### Veritabanı Hususları

| Sorun | Çözüm |
|-------|-------|
| Bağlantı havuzlama | PostgreSQL önünde pgBouncer |
| Okuma ölçekleme | Analitik sorguları için okuma replikaları |
| Yazma verimi | Trafik verileri için TimescaleDB hypertable'ları |
| Yedekleme | Zaman noktası kurtarma için pg_dump + WAL arşivleme |

---

## Veritabanı Bakımı

### TimescaleDB İpuçları

Trafik verileri verimli zaman serisi sorguları için TimescaleDB hypertable'larında saklanır:

```sql
-- Chunk boyutlarını kontrol et
SELECT show_chunks('traffic_data');

-- Tutma politikası ayarla (90 gün sakla)
SELECT add_retention_policy('traffic_data', INTERVAL '90 days');

-- Eski chunk'ları sıkıştır
SELECT add_compression_policy('traffic_data', INTERVAL '7 days');
```

### Rutin Bakım

```bash
# Vacuum ve analiz
psql -U vortex -d vortex -c "VACUUM ANALYZE;"

# Veritabanı boyutunu kontrol et
psql -U vortex -d vortex -c "SELECT pg_size_pretty(pg_database_size('vortex'));"

# En büyük tabloları listele
psql -U vortex -d vortex -c "
  SELECT relname, pg_size_pretty(pg_relation_size(oid))
  FROM pg_class
  ORDER BY pg_relation_size(oid) DESC
  LIMIT 10;
"
```

### Göçler

Göçler panel başlangıcında otomatik çalışır. Manuel çalıştırmak için:

```bash
vortexui migrate
```

Göç durumunu kontrol edin:

```bash
vortexui migrate status
```

---

## Günlük Yönetimi

### Panel Günlükleri

```bash
# Canlı günlükler
journalctl -u vortexui -f

# Son 100 satır
journalctl -u vortexui -n 100

# Seviyeye göre filtrele
journalctl -u vortexui | grep ERROR
```

### Düğüm Ajanı Günlükleri

```bash
journalctl -u vortex-node -f
```

### Günlük Seviyeleri

Ortam değişkeni ile yapılandırın:

```
VORTEX_LOG_LEVEL=info   # debug, info, warn, error
VORTEX_LOG_FORMAT=json  # json veya text
```

### Günlük Rotasyonu

Systemd journal varsayılan olarak rotasyonu yönetir. Limitleri yapılandırın:

```ini
# /etc/systemd/journald.conf
SystemMaxUse=500M
MaxRetentionSec=30d
```

---

## Performans Ayarlama

### Sistem Limitleri

```bash
# Dosya tanımlayıcı limitini artır
echo "* soft nofile 65535" >> /etc/security/limits.conf
echo "* hard nofile 65535" >> /etc/security/limits.conf

# Systemd servisi için artır
# /etc/systemd/system/vortexui.service içinde
[Service]
LimitNOFILE=65535
```

### Ağ Ayarlama

```bash
# /etc/sysctl.conf
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.ip_local_port_range = 1024 65535
net.ipv4.tcp_tw_reuse = 1
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
```

Uygula: `sysctl -p`

### Redis Ayarlama

```
# /etc/redis/redis.conf
maxmemory 256mb
maxmemory-policy allkeys-lru
```

### PostgreSQL Ayarlama

VortexUI iş yükleri için temel ayarlar:

```
# postgresql.conf
shared_buffers = 512MB          # RAM'in %25'i
effective_cache_size = 1536MB   # RAM'in %75'i
work_mem = 16MB
maintenance_work_mem = 128MB
random_page_cost = 1.1          # SSD
```

---

## Systemd Servisleri

### Panel Servisi

```ini
# /etc/systemd/system/vortexui.service
[Unit]
Description=VortexUI Panel
After=postgresql.service redis.service

[Service]
Type=simple
User=vortex
WorkingDirectory=/opt/vortexui
ExecStart=/opt/vortexui/vortexui serve
Restart=always
RestartSec=5
LimitNOFILE=65535
EnvironmentFile=/opt/vortexui/.env

[Install]
WantedBy=multi-user.target
```

### Düğüm Ajanı Servisi

```ini
# /etc/systemd/system/vortex-node.service
[Unit]
Description=VortexUI Node Agent
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/vortex-node
ExecStart=/opt/vortex-node/vortex-node
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

### Yaygın Komutlar

```bash
sudo systemctl start vortexui
sudo systemctl stop vortexui
sudo systemctl restart vortexui
sudo systemctl status vortexui
sudo systemctl enable vortexui   # başlangıçta başlat
```
