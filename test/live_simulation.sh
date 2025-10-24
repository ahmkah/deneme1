#!/bin/bash

echo "═══════════════════════════════════════════════════════════════"
echo "🚀 CANLI SİMÜLASYON - 30 SANİYE"
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo "⏱️  Bu simülasyon 30 saniye boyunca gerçek proxy'lerle"
echo "    Upbit API'sine istek atacak ve sistemin çalışmasını test edecek."
echo ""
echo "📊 Test Edilecekler:"
echo "   • Proxy rotation (random selection)"
echo "   • ETag değişiklik tespiti"
echo "   • Rate limit kontrolü"
echo "   • Bot tespit bypass"
echo "   • Response time tracking"
echo ""

# Geçici simülasyon dosyası
TEMP_LOG="/tmp/upbit_simulation_$$.log"

# Simülasyonu arka planda başlat ve 30 saniye çalıştır
echo "▶️  Simülasyon başlatılıyor..."
echo ""

# Go programını arka planda çalıştır
timeout 30 go run main.go > "$TEMP_LOG" 2>&1 &
SIM_PID=$!

echo "🔄 Simülasyon çalışıyor (PID: $SIM_PID)..."
echo ""

# Progress bar
for i in {1..30}; do
    printf "\r⏳ İlerleme: [%-30s] %d/30s" $(printf '#%.0s' $(seq 1 $i)) $i
    sleep 1
done

echo ""
echo ""

# Simülasyonu durdur
kill $SIM_PID 2>/dev/null
wait $SIM_PID 2>/dev/null

echo "✅ Simülasyon tamamlandı!"
echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "📊 SİMÜLASYON ANALİZİ"
echo "═══════════════════════════════════════════════════════════════"
echo ""

# Log analizleri
if [ -f "$TEMP_LOG" ]; then
    # Proxy kullanımları
    echo "📡 Proxy Kullanımı:"
    proxy_count=$(grep -c "Proxy #" "$TEMP_LOG" 2>/dev/null || echo "0")
    echo "   • Toplam istek: $proxy_count"
    
    # Başarılı istekler
    success_count=$(grep -c "✅\|SUCCESS\|200 OK" "$TEMP_LOG" 2>/dev/null || echo "0")
    echo "   • Başarılı: $success_count"
    
    # Rate limit kontrolü
    rate_limit=$(grep -c "429\|Too Many Requests" "$TEMP_LOG" 2>/dev/null || echo "0")
    if [ "$rate_limit" -gt 0 ]; then
        echo "   ⚠️  Rate limit: $rate_limit kez"
    else
        echo "   ✅ Rate limit: Yok"
    fi
    
    echo ""
    
    # ETag değişiklikleri
    echo "🔄 ETag Değişiklikleri:"
    etag_changes=$(grep -c "ETag change\|FIRST TO DETECT" "$TEMP_LOG" 2>/dev/null || echo "0")
    echo "   • Tespit edilen: $etag_changes"
    
    echo ""
    
    # Yeni coin tespiti
    echo "🔥 Yeni Coin Tespiti:"
    new_coins=$(grep -c "YENİ LİSTELEME\|NEW LISTING" "$TEMP_LOG" 2>/dev/null || echo "0")
    if [ "$new_coins" -gt 0 ]; then
        echo "   🎉 $new_coins yeni coin tespit edildi!"
        grep "YENİ LİSTELEME\|NEW LISTING" "$TEMP_LOG" 2>/dev/null | head -5
    else
        echo "   ℹ️  Yeni coin yok (normal)"
    fi
    
    echo ""
    
    # Hata analizi
    echo "❌ Hata Analizi:"
    error_count=$(grep -c "ERROR\|FAILED\|❌" "$TEMP_LOG" 2>/dev/null || echo "0")
    if [ "$error_count" -gt 0 ]; then
        echo "   ⚠️  $error_count hata tespit edildi"
        echo "   Son hatalar:"
        grep "ERROR\|FAILED" "$TEMP_LOG" 2>/dev/null | tail -3
    else
        echo "   ✅ Hata yok"
    fi
    
    echo ""
    
    # Log dosyası bilgisi
    echo "📝 Detaylı log dosyası: $TEMP_LOG"
    echo "   (İncelemek için: cat $TEMP_LOG)"
else
    echo "⚠️  Log dosyası oluşturulamadı"
fi

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "✅ Analiz tamamlandı!"
echo "═══════════════════════════════════════════════════════════════"
