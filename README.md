# Upbit-Bitget Auto Trading Bot

Upbit borsasında yeni listelenen coinleri otomatik algılayan ve Bitget borsasında long pozisyon açan otomatik trading sistemi.

## 🚀 Özellikler

- ⚡ **Ultra Hızlı**: 16ms detection coverage ile yeni listing yakalama (0.016s - test edildi: 19 proxy)
- 🔄 **Çoklu Proxy Rotasyon**: Intelligent proxy rotation ile 24/7 monitoring (3s cooldown, random interval, %0 429 riski)
- 🤖 **Telegram Bot Arayüzü**: Çoklu kullanıcı yönetimi ve inline keyboard UI
- 🔐 **Güvenli Credential Yönetimi**: Şifreli API key saklama
- 📊 **Otomatik P&L Takibi**: 5, 30, 60 dakika ve 6 saatte bir bildirim
- 🛡️ **Duplicate Prevention**: Her coin sadece bir kez trade edilir
- ⚙️ **Kişiselleştirilebilir**: Kullanıcı bazında margin, leverage ve risk ayarları
- 🛡️ **Bot Tespit Koruması**: 11 farklı User-Agent, gerçek browser header'ları
- 📝 **JSONL Format**: Ultra hızlı append-only log sistemi

## 📋 Sistem Gereksinimleri

- Ubuntu 22.04 LTS (64-bit)
- Root access
- En az 512MB RAM
- 1GB disk alanı
- İnternet bağlantısı (24/7)

---

## 🔧 1. Ubuntu 22.04 Kurulum (Root User)

### 1.1 Sistem Güncellemesi
```bash
apt update && apt upgrade -y
```

### 1.2 Gerekli Paketlerin Kurulumu
```bash
# Temel paketler
apt install -y curl wget git nano htop screen

# SSL/TLS sertifikaları
apt install -y ca-certificates
update-ca-certificates
```

---

## 🐹 2. Go Kurulumu

### 2.1 Go 1.22+ İndirme ve Kurulum
```bash
# Go 1.22.6 indirme (en son sürümü https://go.dev/dl/ adresinden kontrol edin)
cd /root
wget https://go.dev/dl/go1.22.6.linux-amd64.tar.gz

# Eski Go sürümünü kaldırma (varsa)
rm -rf /usr/local/go

# Yeni Go'yu kurma
tar -C /usr/local -xzf go1.22.6.linux-amd64.tar.gz

# İndirilen arşivi silme
rm go1.22.6.linux-amd64.tar.gz
```

### 2.2 PATH Ayarlama
```bash
# .bashrc dosyasına Go path ekleme
echo 'export PATH=$PATH:/usr/local/go/bin' >> /root/.bashrc
echo 'export GOPATH=/root/go' >> /root/.bashrc

# Değişiklikleri uygulama
source /root/.bashrc
```

### 2.3 Go Kurulumunu Doğrulama
```bash
go version
# Çıktı: go version go1.22.6 linux/amd64
```

---

## 📥 3. Projeyi GitHub'dan Klonlama

### 3.1 Proje Klasörü Oluşturma
```bash
cd /root
git clone https://github.com/0xmtnslk/upbit-trade.git
cd upbit-trade
```

### 3.2 Go Dependencies Kurulumu
```bash
go mod download
go mod tidy
```

---

## ⚙️ 4. Yapılandırma (.env Dosyası)

### 4.1 .env Dosyası Oluşturma
```bash
cp .env.example .env
nano .env
```

### 4.2 .env Dosyasını Doldurma

**Gerekli Değişkenler:**

```bash
# Telegram Bot Token (BotFather'dan alınır)
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here

# 24 SOCKS5 Proxy Sunucuları (format: username:password@ip:port)
# CRITICAL: Proxy #1-2 MUST be Seoul-based for lowest latency!
# With 24 proxies + 8s interval = 333ms coverage (0.333s) - TOTAL: 3 req/sec (SAFE)
UPBIT_PROXY_1=proxy1_user:proxy1_pass@ip1:1080
UPBIT_PROXY_2=proxy2_user:proxy2_pass@ip2:1080
UPBIT_PROXY_3=proxy3_user:proxy3_pass@ip3:1080
UPBIT_PROXY_4=proxy4_user:proxy4_pass@ip4:1080
UPBIT_PROXY_5=proxy5_user:proxy5_pass@ip5:1080
UPBIT_PROXY_6=proxy6_user:proxy6_pass@ip6:1080
UPBIT_PROXY_7=proxy7_user:proxy7_pass@ip7:1080
UPBIT_PROXY_8=proxy8_user:proxy8_pass@ip8:1080
UPBIT_PROXY_9=proxy9_user:proxy9_pass@ip9:1080
UPBIT_PROXY_10=proxy10_user:proxy10_pass@ip10:1080
UPBIT_PROXY_11=proxy11_user:proxy11_pass@ip11:1080
UPBIT_PROXY_12=proxy12_user:proxy12_pass@ip12:1080
UPBIT_PROXY_13=proxy13_user:proxy13_pass@ip13:1080
UPBIT_PROXY_14=proxy14_user:proxy14_pass@ip14:1080
UPBIT_PROXY_15=proxy15_user:proxy15_pass@ip15:1080
UPBIT_PROXY_16=proxy16_user:proxy16_pass@ip16:1080
UPBIT_PROXY_17=proxy17_user:proxy17_pass@ip17:1080
UPBIT_PROXY_18=proxy18_user:proxy18_pass@ip18:1080
UPBIT_PROXY_19=proxy19_user:proxy19_pass@ip19:1080
UPBIT_PROXY_20=proxy20_user:proxy20_pass@ip20:1080
UPBIT_PROXY_21=proxy21_user:proxy21_pass@ip21:1080

# Şifreleme anahtarı (32 karakter)
BOT_ENCRYPTION_KEY=your_32_character_encryption_key_here_12345
```

**Not:** `.env` dosyasını kaydetmek için `Ctrl+O` sonra `Enter`, çıkmak için `Ctrl+X`

### 4.3 Telegram Bot Token Alma

1. Telegram'da [@BotFather](https://t.me/BotFather) ile konuşma başlat
2. `/newbot` komutunu gönder
3. Bot adını ve username'ini belirle
4. Verilen token'ı kopyala ve `.env`'ye yapıştır

### 4.4 Proxy Servisleri

**Önerilen Proxy Sağlayıcılar:**
- [Webshare.io](https://webshare.io) - SOCKS5 proxy
- [ProxyScrape](https://proxyscrape.com) - Rotating proxies
- [IPRoyal](https://iproyal.com) - Datacenter proxies

**Proxy Formatı:**
```
username:password@ip_address:port
```

---

## 🧪 4.5 Test Araçları (ÖNEMLİ)

**Test klasöründe kapsamlı test araçları bulunmaktadır:**

### Kapsamlı Sistem Testi
```bash
cd /workspace
./test/test_simulation.sh
```
**Test eder:** Proxy bağlantısı, zaman sync, bot detection, ETag sistemi, rate limit

### Canlı Simülasyon (30 saniye)
```bash
cd /workspace
./test/live_simulation.sh
```
**Test eder:** Gerçek proxy'lerle 30 saniyelik canlı monitoring

### Rate Limit Testi
```bash
cd /workspace/test
go run test_rate_limit.go
```
**Test süresi:** ~7-10 dakika  
**Amaç:** Farklı interval'larda (0.5s, 1s, 2s, 3s, 3.3s, 4s, 5s) test yaparak güvenli limiti bulur

### Kuru Çalıştırma (Dry Run)
```bash
cd /workspace/test
go run dry_run.go
```
**Test eder:** Tüm sistemlerin bağımsız olarak çalışmasını kontrol eder

### Test Raporları
- `test/TEST_REPORT.md` - Detaylı test sonuçları
- `test/SIMULATION_SUMMARY.md` - Hızlı özet rapor

**Son test sonuçları (2025-10-23):**
- ✅ 19/19 proxy çalışıyor
- ✅ 16ms detection coverage
- ✅ Bot detection bypass başarılı
- ✅ Rate limit güvenli (3 req/sec)

---

## 🔨 5. Build (Derleme)

### 5.1 Binary Oluşturma
```bash
cd /root/upbit-trade
go build -o upbit-bitget-bot .
```

### 5.2 Executable Permission Verme
```bash
chmod +x upbit-bitget-bot
```

### 5.3 Manuel Test (Opsiyonel)
```bash
# Önce screen oturumu açın (Ctrl+A+D ile detach edilebilir)
screen -S trading-bot

# Botu çalıştır
./upbit-bitget-bot

# Detach: Ctrl+A sonra D
# Re-attach: screen -r trading-bot
```

---

## 🔄 6. Systemd Service Kurulumu (Arka Plan Çalışma)

### 6.1 Service Dosyası Oluşturma
```bash
nano /etc/systemd/system/upbit-bitget-bot.service
```

### 6.2 Service İçeriği

Aşağıdaki içeriği yapıştırın:

```ini
[Unit]
Description=Upbit-Bitget Auto Trading Bot
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/root/upbit-trade
ExecStart=/root/upbit-trade/upbit-bitget-bot
Restart=always
RestartSec=10
StandardOutput=append:/var/log/upbit-bitget-bot.log
StandardError=append:/var/log/upbit-bitget-bot-error.log

# Environment dosyası
EnvironmentFile=/root/upbit-trade/.env

# Kaynak limitleri
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

**Kaydet ve çık:** `Ctrl+O` → `Enter` → `Ctrl+X`

### 6.3 Service'i Aktifleştirme
```bash
# Systemd'yi yeniden yükle
systemctl daemon-reload

# Service'i aktif et
systemctl enable upbit-bitget-bot.service

# Service'i başlat
systemctl start upbit-bitget-bot.service
```

### 6.4 Durum Kontrolü
```bash
# Service durumu
systemctl status upbit-bitget-bot.service

# Logları izleme (canlı)
tail -f /var/log/upbit-bitget-bot.log

# Hata logları
tail -f /var/log/upbit-bitget-bot-error.log

# Son 100 satır log
journalctl -u upbit-bitget-bot.service -n 100
```

### 6.5 Service Komutları

```bash
# Başlat
systemctl start upbit-bitget-bot.service

# Durdur
systemctl stop upbit-bitget-bot.service

# Yeniden başlat
systemctl restart upbit-bitget-bot.service

# Otomatik başlatmayı kapat
systemctl disable upbit-bitget-bot.service

# Durumu kontrol et
systemctl status upbit-bitget-bot.service
```

---

## 📊 7. Bot Kullanımı (Telegram)

### 7.1 Telegram Bot'u Başlatma

1. Telegram'da botunuzu bulun (BotFather'da verdiğiniz username)
2. `/start` komutunu gönderin
3. Ana menü görünecek

### 7.2 Kullanıcı Kaydı ve Ayarlar

**İlk Kurulum:**
```
1. ⚙️ Ayarlar → API Ayarları
2. Bitget API Key, Secret, Passphrase girin
3. Margin miktarını belirleyin (örn: 100 USDT)
4. Leverage seçin (örn: 10x)
5. ✅ Botu Aktif Et
```

**Hızlı Güncelleme:**
```
⚙️ Ayarlar → Margin/Leverage güncelleme
(API bilgileri korunur)
```

### 7.3 Komutlar

- `/start` - Botu başlat ve menüyü göster
- `/status` - Aktif pozisyonlar ve durum
- `/settings` - Ayarları göster/düzenle
- `/activate` - Botu aktif et
- `/deactivate` - Botu pasif et

---

## 🔄 8. Güncelleme ve Bakım

### 8.1 Kod Güncellemesi (GitHub'dan)

**Adım adım güncelleme:**

```bash
# 1. Service'i durdur
systemctl stop upbit-bitget-bot.service

# 2. Proje klasörüne git
cd /root/upbit-trade

# 3. Güncellemeleri çek
git pull origin main

# 4. Dependencies güncelle
go mod download
go mod tidy

# 5. Yeniden derle
go build -o upbit-bitget-bot .

# 6. Zaman senkronizasyonu kontrolü
make checksync

# 7. (Opsiyonel) Upbit ile zaman sync (eğer offset > 1s ise)
make synctime

# 8. (Opsiyonel) System-wide tool kurulumu
make install-tools
# Artık "checksync" ve "synctime" komutlarını her yerden kullanabilirsin

# 9. Service'i başlat
systemctl start upbit-bitget-bot.service

# 10. Durumu kontrol et
systemctl status upbit-bitget-bot.service

# 11. Logları kontrol et (ilk 30 satır)
tail -n 30 /var/log/upbit-bitget-bot.log
```

**Zaman Senkronizasyonu Kontrolü:**

Güncelleme sonrası sistem zamanının doğru olduğundan emin olun:

```bash
# Upbit ve Bitget ile saat farkını kontrol et
cd /root/upbit-trade
make checksync
```

**Beklenen çıktı:**
```
📡 UPBIT TIME SYNC:
   • Clock Offset: -428ms
   ✅ Clock sync OK (offset < 1s)

📡 BITGET TIME SYNC:
   • Clock Offset: -34ms
   ✅ Clock sync OK (offset < 1s)
```

**Eğer offset > 1 saniye görürseniz:**

```bash
# Upbit server zamanıyla sistem saatini sync et
make synctime

# Veya direkt script ile:
./sync_upbit_time.sh

# Kontrol et
make checksync
```

### 8.2 Otomatik Güncelleme Script'i (Opsiyonel)

```bash
# Güncelleme script'i oluştur
nano /root/update-bot.sh
```

**Script içeriği:**
```bash
#!/bin/bash

echo "🔄 Upbit-Bitget Bot güncelleniyor..."
echo ""

# Service durdur
echo "1️⃣ Stopping service..."
systemctl stop upbit-bitget-bot.service

# Güncellemeleri çek
echo "2️⃣ Pulling latest code..."
cd /root/upbit-trade
git pull origin main

# Dependencies
echo "3️⃣ Updating dependencies..."
go mod download
go mod tidy

# Build
echo "4️⃣ Building..."
go build -o upbit-bitget-bot .

# Zaman sync kontrolü
echo "5️⃣ Checking time synchronization..."
make checksync

# Eğer offset > 1s ise uyar
echo ""
echo "⚠️  Eğer yukarıda 'WARNING: Clock offset > 1s' görüyorsanız:"
echo "    make synctime komutuyla zamanı sync edin!"
echo ""

# Service başlat
echo "6️⃣ Starting service..."
systemctl start upbit-bitget-bot.service

echo ""
echo "✅ Güncelleme tamamlandı!"
echo ""
systemctl status upbit-bitget-bot.service --no-pager
echo ""
echo "📊 Log izlemek için: tail -f /var/log/upbit-bitget-bot.log"
```

**Executable yap:**
```bash
chmod +x /root/update-bot.sh
```

**Kullanım:**
```bash
/root/update-bot.sh
```

### 8.3 Yeni Komutlar (Make Kullanımı)

Bot artık **Makefile** ile daha kolay yönetilebiliyor:

```bash
# Bot'u çalıştır (development)
make run

# Zaman senkronizasyonu kontrol et
make checksync

# Upbit server zamanıyla sistem sync et
make synctime

# Binary oluştur
make build

# Helper tool'ları system-wide kur (opsiyonel)
make install-tools

# Build dosyalarını temizle
make clean
```

**System-wide Kurulum (Opsiyonel):**

```bash
# Tool'ları sistem genelinde kullanılabilir yap
make install-tools

# Artık her yerden kullanabilirsin:
checksync    # /usr/local/bin/checksync
synctime     # /usr/local/bin/synctime
```

**Zaman Senkronizasyonu Detayları:**

```bash
# 1. Kontrol et
cd /root/upbit-trade
make checksync
```

**Çıktı örneği:**
```
⏰ Checking time synchronization with exchanges...

📡 UPBIT TIME SYNC:
   • Server Time:     2025-10-15 13:29:17.058
   • Local Time:      2025-10-15 13:29:17.486
   • Clock Offset:    -428ms
   • Network Latency: 58ms
   ✅ Clock sync OK (offset < 1s)

📡 BITGET TIME SYNC:
   • Server Time:     2025-10-15 13:29:17.707
   • Local Time:      2025-10-15 13:29:17.742
   • Clock Offset:    -34ms
   • Network Latency: 127ms
   ✅ Clock sync OK (offset < 1s)
```

```bash
# 2. Eğer offset > 1s ise sync et
make synctime
```

**Sync çıktısı:**
```
⏰ Syncing system time with Upbit server...

📡 Upbit Server Time: Tue, 15 Oct 2025 13:29:17 GMT
🔧 Setting system time to: 2025-10-15 13:29:17

✅ System time synchronized!
```

**⚠️ Önemli Notlar:**

- `make synctime` çalıştırdıktan sonra NTP otomatik sync kapatılır
- Sistem zamanı Upbit server zamanıyla eşitlenir
- Trade timing hassasiyeti için kritik önem taşır
- Günde 1-2 kere kontrol etmek önerilir

### 8.4 Log Rotasyonu (Disk Tasarrufu)

```bash
# Logrotate yapılandırması
nano /etc/logrotate.d/upbit-bitget-bot
```

**İçerik:**
```
/var/log/upbit-bitget-bot.log
/var/log/upbit-bitget-bot-error.log
{
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 root root
    postrotate
        systemctl reload upbit-bitget-bot.service > /dev/null 2>&1 || true
    endscript
}
```

---

## 🛡️ 9. Güvenlik Önerileri

### 9.1 Firewall Ayarları (UFW)

```bash
# UFW kur
apt install -y ufw

# Temel kurallar
ufw default deny incoming
ufw default allow outgoing

# SSH port (değiştirdiyseniz ona göre ayarlayın)
ufw allow 22/tcp

# Aktif et
ufw enable

# Durumu kontrol et
ufw status
```

### 9.2 Fail2Ban (SSH Koruması)

```bash
# Fail2ban kur
apt install -y fail2ban

# Başlat ve aktif et
systemctl start fail2ban
systemctl enable fail2ban
```

### 9.3 .env Dosyası İzinleri

```bash
# Sadece root okuyabilsin
chmod 600 /root/upbit-trade/.env
```

### 9.4 API Key Güvenliği

**Bitget API:**
- ✅ Sadece futures/spot trading iznini aktif edin
- ✅ Withdrawal iznini KESİNLİKLE kapalı tutun
- ✅ IP whitelist ekleyin (sunucu IP'niz)
- ✅ API key'leri düzenli olarak rotate edin

---

## 🔍 10. Sorun Giderme

### 10.1 Bot Çalışmıyor

**Kontrol adımları:**
```bash
# Service durumu
systemctl status upbit-bitget-bot.service

# Log kontrolü
tail -n 50 /var/log/upbit-bitget-bot.log
tail -n 50 /var/log/upbit-bitget-bot-error.log

# .env dosyası doğru mu?
cat /root/upbit-trade/.env

# Binary çalıştırılabilir mi?
ls -la /root/upbit-trade/upbit-bitget-bot
```

### 10.2 Proxy Bağlantı Hataları

**Test:**
```bash
# Proxy test (örnek)
curl --socks5 username:password@ip:port https://api.upbit.com/v1/status/wallet

# Tüm proxyleri test et
nano /root/test-proxies.sh
```

### 10.3 Telegram Bağlantı Hatası

**Kontrol:**
```bash
# Token doğru mu?
grep TELEGRAM_BOT_TOKEN /root/upbit-trade/.env

# İnternet bağlantısı var mı?
ping -c 4 api.telegram.org
```

### 10.4 Go Build Hataları

```bash
# Go modules temizle
cd /root/upbit-trade
go clean -modcache
go mod download
go mod tidy

# Yeniden derle
go build -o upbit-bitget-bot .
```

### 10.5 Disk Dolu

```bash
# Log dosyalarını temizle
> /var/log/upbit-bitget-bot.log
> /var/log/upbit-bitget-bot-error.log

# Journal logları sınırla
journalctl --vacuum-time=2d
```

---

## 📈 11. İzleme ve Monitoring

### 11.1 Gerçek Zamanlı Log İzleme

```bash
# Tüm loglar (renkli)
tail -f /var/log/upbit-bitget-bot.log | grep --color=always -E "🔥|⚡|ERROR|WARN|✅"

# Sadece trade işlemleri
tail -f /var/log/upbit-bitget-bot.log | grep "FAST TRACK\|pozisyon açıldı"

# Sadece hatalar
tail -f /var/log/upbit-bitget-bot-error.log
```

### 11.2 Performans İzleme

```bash
# Bot kaynak kullanımı
ps aux | grep upbit-bitget-bot

# Sistem kaynakları
htop

# Network bağlantıları
netstat -tunlp | grep upbit-bitget-bot
```

### 11.3 Otomatik Restart (Crash Durumunda)

Service dosyasında zaten var:
```ini
Restart=always
RestartSec=10
```

Bot crash olursa 10 saniye sonra otomatik restart olur.

---

## 📝 12. Veri Dosyaları

### 12.1 Önemli Dosyalar

```bash
# Kullanıcı veritabanı
/root/upbit-trade/bot_users.json

# Tespit edilen listeler
/root/upbit-trade/upbit_new.json

# Aktif pozisyonlar
/root/upbit-trade/active_positions.json
```

### 12.2 Yedekleme (Backup)

```bash
# Yedekleme script'i
nano /root/backup-bot.sh
```

**Script:**
```bash
#!/bin/bash

BACKUP_DIR="/root/bot-backups"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# Veri dosyalarını yedekle
cp /root/upbit-trade/bot_users.json $BACKUP_DIR/bot_users_$DATE.json
cp /root/upbit-trade/upbit_new.json $BACKUP_DIR/upbit_new_$DATE.json
cp /root/upbit-trade/active_positions.json $BACKUP_DIR/active_positions_$DATE.json

# .env yedekle (GÜVENLİ SAKLAYIN!)
cp /root/upbit-trade/.env $BACKUP_DIR/env_$DATE.backup

echo "✅ Yedekleme tamamlandı: $BACKUP_DIR"

# 30 günden eski yedekleri sil
find $BACKUP_DIR -type f -mtime +30 -delete
```

**Executable yap:**
```bash
chmod +x /root/backup-bot.sh
```

**Cron ile otomatik yedekleme (her gün 03:00):**
```bash
crontab -e

# Şunu ekle:
0 3 * * * /root/backup-bot.sh
```

---

## 🚨 13. Acil Durum Prosedürleri

### 13.1 Tüm Trade'leri Durdurma

```bash
# Botu durdur
systemctl stop upbit-bitget-bot.service

# Telegram'dan tüm kullanıcıları pasif et
# Bot üzerinden: ⚙️ Ayarlar → ❌ Botu Deaktif Et
```

### 13.2 Factory Reset

```bash
# Service durdur
systemctl stop upbit-bitget-bot.service

# Veri dosyalarını yedekle
cp /root/upbit-trade/bot_users.json /root/bot_users_backup.json

# Veri dosyalarını temizle
rm /root/upbit-trade/bot_users.json
rm /root/upbit-trade/upbit_new.json
rm /root/upbit-trade/active_positions.json

# Service başlat (temiz başlangıç)
systemctl start upbit-bitget-bot.service
```

---

## 📞 14. Destek ve İletişim

- **GitHub Issues:** https://github.com/0xmtnslk/upbit-trade/issues
- **Telegram:** Bot üzerinden destek talebi

---

## ⚖️ Yasal Uyarı

**DİKKAT:** Bu bot otomatik trading yapar ve finansal riskler içerir:

- ⚠️ Yüksek leverage kullanımı sermaye kaybına yol açabilir
- ⚠️ Kripto piyasalar 7/24 volatildir
- ⚠️ Bot yazılımındaki hatalar zarara sebep olabilir
- ⚠️ API key güvenliği tamamen kullanıcının sorumluluğundadır

**Kullanmadan önce:**
- ✅ Küçük miktarlarla test edin
- ✅ Leverage'ı düşük tutun (5x-10x önerilir)
- ✅ Stop-loss stratejinizi belirleyin
- ✅ Kaybetmeyi göze alabileceğiniz sermaye kullanın

**Bu yazılımı kullanarak, tüm riskleri kabul ettiğinizi beyan edersiniz.**

---

## 📄 Lisans

MIT License - Detaylar için `LICENSE` dosyasına bakın.

---

## 🎯 Hızlı Başlangıç Özet

```bash
# 1. Sistem hazırlığı
apt update && apt upgrade -y
apt install -y curl wget git nano

# 2. Go kurulumu
wget https://go.dev/dl/go1.22.6.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.22.6.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /root/.bashrc
source /root/.bashrc

# 3. Projeyi klonla
cd /root
git clone https://github.com/0xmtnslk/upbit-trade.git
cd upbit-trade

# 4. Dependencies
go mod download && go mod tidy

# 5. .env dosyası
cp .env.example .env
nano .env  # Doldurun

# 6. Build
go build -o upbit-bitget-bot .

# 7. Systemd service
nano /etc/systemd/system/upbit-bitget-bot.service  # Yukarıdaki içeriği yapıştır
systemctl daemon-reload
systemctl enable upbit-bitget-bot.service
systemctl start upbit-bitget-bot.service

# 8. Kontrol
systemctl status upbit-bitget-bot.service
tail -f /var/log/upbit-bitget-bot.log
```

**✅ Bot çalışıyor! Telegram'dan /start ile başlayın.**

---

*Son güncelleme: 2025-10-15*

---

## 🆕 Yeni Özellikler (v2.0)

### ⏰ Zaman Senkronizasyonu Sistemi

- **Otomatik Kontrol**: Her bot restart'ında Upbit ve Bitget server zamanları kontrol edilir
- **Clock Offset Uyarısı**: > 1 saniye sapma varsa otomatik uyarı
- **Manuel Sync**: `make synctime` ile Upbit zamanına göre sistem senkronizasyonu
- **Trade Accuracy**: Zaman hassasiyeti trade execution için kritik

### 📊 Trade Execution Logging

- **4 Kritik Timestamp**: Detection, file save, order sent, order confirmed
- **Latency Breakdown**: Her aşamanın süre analizi
- **Microsecond Precision**: Milisaniye hassasiyetinde kayıt
- **Log Dosyası**: `trade_execution_log.json` (JSONL format)

### 🚀 Performance & Optimizasyon

- **16ms Detection Coverage**: 19 proxy ile ultra hızlı tespit (test edildi: 2025-10-23)
- **JSONL Format**: Append-only logging, %90+ disk I/O azalması
- **Intelligent Proxy Rotation**: 3s proactive cooldown + random intervals
- **Bot Detection Bypass**: 11 User-Agent rotation, gerçek browser headers
- **~3 req/sec**: Güvenli rate limit (%70 altında kullanım)
- **KST Timezone**: Upbit server zamanı ile tam uyumlu

### 📁 Proje Yapısı

```
/workspace/
├── main.go              # Ana uygulama
├── upbit_monitor.go     # Upbit monitoring sistemi
├── telegram_bot.go      # Telegram bot arayüzü
├── bitget.go            # Bitget API entegrasyonu
├── test/                # Test dosyaları
│   ├── TEST_REPORT.md          # Detaylı test raporu
│   ├── SIMULATION_SUMMARY.md   # Hızlı özet
│   ├── test_simulation.sh      # Sistem testi
│   ├── live_simulation.sh      # Canlı simülasyon
│   ├── test_rate_limit.go      # Rate limit testi
│   └── dry_run.go              # Kuru çalıştırma
├── tools/               # Yardımcı araçlar
│   ├── check_upbit_time.sh
│   └── checksync.go
└── OPTIMIZATION_SUMMARY.md  # Optimizasyon detayları
```

*Son güncelleme: 2025-10-23*
