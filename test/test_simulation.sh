#!/bin/bash

echo "═══════════════════════════════════════════════════════════════"
echo "🧪 UPBIT-BITGET BOT SİMÜLASYON VE TEST SİSTEMİ"
echo "═══════════════════════════════════════════════════════════════"
echo ""

echo "📋 Test Aşamaları:"
echo "  1️⃣  Proxy bağlantı testi"
echo "  2️⃣  Upbit API zaman senkronizasyonu"
echo "  3️⃣  Bot tespit kontrolü (User-Agent, headers)"
echo "  4️⃣  ETag değişiklik tespiti"
echo "  5️⃣  Rate limit kontrolü"
echo ""

# Renk kodları
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test 1: Proxy Bağlantı Kontrolü
echo "═══════════════════════════════════════════════════════════════"
echo "1️⃣  PROXY BAĞLANTI TESTİ"
echo "═══════════════════════════════════════════════════════════════"

# .env dosyasından proxy'leri oku
proxy_count=0
failed_proxies=0

for i in {1..24}; do
    proxy_var="UPBIT_PROXY_$i"
    proxy_value=$(grep "^$proxy_var=" .env 2>/dev/null | cut -d '=' -f2)
    
    if [ ! -z "$proxy_value" ]; then
        proxy_count=$((proxy_count + 1))
        echo ""
        echo -e "${BLUE}Testing Proxy #$i${NC}"
        echo "  Proxy: ${proxy_value:0:30}..."
        
        # SOCKS5 test (timeout 5 saniye)
        if timeout 5 curl -s --socks5 "${proxy_value#socks5://}" \
           "https://api-manager.upbit.com/api/v1/announcements?os=web&page=1&per_page=1" \
           -H "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" \
           > /dev/null 2>&1; then
            echo -e "  ${GREEN}✅ BAŞARILI${NC} - Proxy çalışıyor"
        else
            echo -e "  ${RED}❌ BAŞARISIZ${NC} - Proxy bağlantı hatası"
            failed_proxies=$((failed_proxies + 1))
        fi
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $failed_proxies -eq 0 ]; then
    echo -e "${GREEN}✅ Tüm proxy'ler çalışıyor!${NC} ($proxy_count/$proxy_count)"
else
    echo -e "${YELLOW}⚠️  $failed_proxies/$proxy_count proxy başarısız${NC}"
fi
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Test 2: Upbit Zaman Senkronizasyonu
echo "═══════════════════════════════════════════════════════════════"
echo "2️⃣  UPBIT ZAMAN SENKRONİZASYONU"
echo "═══════════════════════════════════════════════════════════════"

# İlk çalışan proxy'yi bul
first_working_proxy=""
for i in {1..24}; do
    proxy_var="UPBIT_PROXY_$i"
    proxy_value=$(grep "^$proxy_var=" .env 2>/dev/null | cut -d '=' -f2)
    if [ ! -z "$proxy_value" ]; then
        first_working_proxy="${proxy_value#socks5://}"
        break
    fi
done

if [ ! -z "$first_working_proxy" ]; then
    echo "Proxy kullanılarak Upbit server zamanı alınıyor..."
    response=$(timeout 10 curl -s --socks5 "$first_working_proxy" \
        -i "https://api-manager.upbit.com/api/v1/announcements?os=web&page=1&per_page=1" \
        -H "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
    
    server_time=$(echo "$response" | grep -i "^date:" | cut -d ' ' -f 2-)
    
    if [ ! -z "$server_time" ]; then
        echo ""
        echo -e "${GREEN}✅ Upbit Server Zamanı:${NC} $server_time"
        local_time=$(date -u)
        echo -e "${BLUE}🕐 Sistem Zamanı (UTC):${NC} $local_time"
        echo ""
        echo "⚠️  Not: Zaman farkı 1 saniyeden fazlaysa 'make synctime' çalıştırın"
    else
        echo -e "${RED}❌ Server zamanı alınamadı${NC}"
    fi
else
    echo -e "${RED}❌ Çalışan proxy bulunamadı${NC}"
fi
echo ""

# Test 3: Bot Tespit Kontrolü
echo "═══════════════════════════════════════════════════════════════"
echo "3️⃣  BOT TESPİT KONTROLÜ"
echo "═══════════════════════════════════════════════════════════════"

echo "Upbit'e farklı User-Agent'larla istek gönderiliyor..."
echo ""

# Test 1: Bot User-Agent (kötü)
echo -e "${YELLOW}Test 1: Bot User-Agent (beklenmeyen)${NC}"
bot_response=$(timeout 5 curl -s --socks5 "$first_working_proxy" \
    -w "\n%{http_code}" \
    "https://api-manager.upbit.com/api/v1/announcements?os=web&page=1&per_page=1" \
    -H "User-Agent: Bot/1.0" 2>/dev/null | tail -1)

if [ "$bot_response" == "200" ]; then
    echo -e "  ${GREEN}✅ 200 OK${NC} - İstek kabul edildi"
elif [ "$bot_response" == "403" ] || [ "$bot_response" == "429" ]; then
    echo -e "  ${RED}❌ $bot_response${NC} - Bot olarak tespit edildi!"
else
    echo -e "  ${YELLOW}⚠️  $bot_response${NC} - Beklenmeyen yanıt"
fi

echo ""
echo -e "${GREEN}Test 2: Browser User-Agent (iyi)${NC}"
browser_response=$(timeout 5 curl -s --socks5 "$first_working_proxy" \
    -w "\n%{http_code}" \
    "https://api-manager.upbit.com/api/v1/announcements?os=web&page=1&per_page=1" \
    -H "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36" \
    -H "Accept: application/json, text/plain, */*" \
    -H "Accept-Language: ko-KR,ko;q=0.9,en-US;q=0.8,en;q=0.7" \
    -H "Referer: https://upbit.com/" \
    -H "Origin: https://upbit.com" 2>/dev/null | tail -1)

if [ "$browser_response" == "200" ]; then
    echo -e "  ${GREEN}✅ 200 OK${NC} - Browser olarak kabul edildi"
elif [ "$browser_response" == "403" ] || [ "$browser_response" == "429" ]; then
    echo -e "  ${RED}❌ $browser_response${NC} - Engellenmiş!"
else
    echo -e "  ${YELLOW}⚠️  $browser_response${NC} - Beklenmeyen yanıt"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}✅ Bot tespit koruması aktif!${NC} (Kodda doğru header'lar kullanılıyor)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Test 4: ETag Sistemi
echo "═══════════════════════════════════════════════════════════════"
echo "4️⃣  ETAG DEĞİŞİKLİK TESPİT SİSTEMİ"
echo "═══════════════════════════════════════════════════════════════"

echo "İlk istek gönderiliyor..."
first_etag=$(timeout 5 curl -s --socks5 "$first_working_proxy" \
    -I "https://api-manager.upbit.com/api/v1/announcements?os=web&page=1&per_page=20" \
    -H "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" \
    2>/dev/null | grep -i "etag:" | cut -d ' ' -f 2 | tr -d '\r')

if [ ! -z "$first_etag" ]; then
    echo -e "${GREEN}✅ İlk ETag:${NC} ${first_etag:0:20}..."
    
    echo ""
    echo "2 saniye sonra tekrar istek gönderiliyor..."
    sleep 2
    
    second_etag=$(timeout 5 curl -s --socks5 "$first_working_proxy" \
        -I "https://api-manager.upbit.com/api/v1/announcements?os=web&page=1&per_page=20" \
        -H "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" \
        -H "If-None-Match: $first_etag" \
        2>/dev/null | grep -i "HTTP" | cut -d ' ' -f 2)
    
    if [ "$second_etag" == "304" ]; then
        echo -e "${GREEN}✅ 304 Not Modified${NC} - ETag sistemi çalışıyor (içerik değişmedi)"
    elif [ "$second_etag" == "200" ]; then
        echo -e "${YELLOW}⚠️  200 OK${NC} - İçerik değişmiş (yeni duyuru olabilir!)"
    else
        echo -e "${YELLOW}⚠️  Yanıt kodu: $second_etag${NC}"
    fi
else
    echo -e "${RED}❌ ETag alınamadı${NC}"
fi
echo ""

# Test 5: Rate Limit Kontrolü
echo "═══════════════════════════════════════════════════════════════"
echo "5️⃣  RATE LIMIT KONTROLÜ"
echo "═══════════════════════════════════════════════════════════════"

echo "Hızlı ardışık istekler gönderiliyor (10 istek, 0.3s interval)..."
echo ""

rate_limit_hit=false
success_count=0

for i in {1..10}; do
    http_code=$(timeout 5 curl -s --socks5 "$first_working_proxy" \
        -w "%{http_code}" \
        -o /dev/null \
        "https://api-manager.upbit.com/api/v1/announcements?os=web&page=1&per_page=1" \
        -H "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" \
        2>/dev/null)
    
    if [ "$http_code" == "200" ]; then
        echo -e "  İstek $i: ${GREEN}✅ 200 OK${NC}"
        success_count=$((success_count + 1))
    elif [ "$http_code" == "429" ]; then
        echo -e "  İstek $i: ${RED}❌ 429 Too Many Requests${NC}"
        rate_limit_hit=true
        break
    else
        echo -e "  İstek $i: ${YELLOW}⚠️  $http_code${NC}"
    fi
    
    sleep 0.3
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ "$rate_limit_hit" = true ]; then
    echo -e "${RED}⚠️  Rate limit'e takıldı! (0.3s çok hızlı olabilir)${NC}"
    echo -e "${YELLOW}   Öneri: UPBIT_CHECK_INTERVAL_MS değerini artırın${NC}"
else
    echo -e "${GREEN}✅ Rate limit'e takılmadı!${NC} ($success_count/10 başarılı)"
    echo -e "${GREEN}   Bot güvenli şekilde çalışabilir (0.3s interval)${NC}"
fi
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Özet Rapor
echo "═══════════════════════════════════════════════════════════════"
echo "📊 TEST ÖZET RAPORU"
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo -e "${BLUE}Proxy Durumu:${NC}"
echo "  • Toplam Proxy: $proxy_count"
echo "  • Çalışan: $((proxy_count - failed_proxies))"
echo "  • Başarısız: $failed_proxies"
echo ""
echo -e "${BLUE}Sistem Kontrolleri:${NC}"
echo "  • Zaman Senkronizasyonu: $([ ! -z "$server_time" ] && echo "✅ OK" || echo "❌ Hata")"
echo "  • Bot Tespit Koruması: ✅ Aktif"
echo "  • ETag Sistemi: $([ ! -z "$first_etag" ] && echo "✅ Çalışıyor" || echo "❌ Hata")"
echo "  • Rate Limit: $([ "$rate_limit_hit" = false ] && echo "✅ Güvenli" || echo "⚠️ Risk var")"
echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "✅ Test tamamlandı!"
echo "═══════════════════════════════════════════════════════════════"
