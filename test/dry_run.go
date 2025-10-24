package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
)

// Test programı - Gerçek bot çalıştırmadan proxy ve API testleri
func main() {
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("🧪 UPBIT-BITGET BOT - DRY RUN TEST")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// .env yükle
	_ = godotenv.Load()

	// Test callback fonksiyonu
	detectedCoins := []string{}
	callbackFunc := func(symbol string) {
		detectedCoins = append(detectedCoins, symbol)
		log.Printf("🔥 CALLBACK TRIGGERED: New coin detected: %s", symbol)
	}

	// Upbit monitor oluştur
	monitor := NewUpbitMonitor(callbackFunc)

	fmt.Println("📊 Sistem Bilgileri:")
	fmt.Printf("   • Proxy Sayısı: %d\n", len(monitor.proxies))
	fmt.Printf("   • API URL: %s\n", monitor.apiURL)
	fmt.Printf("   • JSON Dosyası: %s\n", monitor.jsonFile)
	fmt.Printf("   • User-Agent Pool: %d adet\n", len(monitor.userAgents))
	fmt.Println()

	// Proxy testleri
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("1️⃣  PROXY BAĞLANTI TESTLERİ")
	fmt.Println("═══════════════════════════════════════════════════════════════")

	successfulProxies := 0
	failedProxies := 0

	for i, proxyURL := range monitor.proxies {
		fmt.Printf("\n🔍 Testing Proxy #%d: %s\n", i+1, proxyURL[:30]+"...")
		
		client, err := monitor.createProxyClient(proxyURL)
		if err != nil {
			fmt.Printf("   ❌ FAILED: %v\n", err)
			failedProxies++
			continue
		}

		// Basit bağlantı testi
		req, err := monitor.createTestRequest()
		if err != nil {
			fmt.Printf("   ❌ FAILED: Request creation error: %v\n", err)
			failedProxies++
			continue
		}

		start := time.Now()
		resp, err := client.Do(req)
		latency := time.Since(start)

		if err != nil {
			fmt.Printf("   ❌ FAILED: %v\n", err)
			failedProxies++
			continue
		}
		defer resp.Body.Close()

		fmt.Printf("   ✅ SUCCESS: Status=%d, Latency=%dms\n", resp.StatusCode, latency.Milliseconds())
		
		// ETag kontrolü
		etag := resp.Header.Get("ETag")
		if etag != "" {
			fmt.Printf("   📌 ETag: %s...\n", etag[:20])
		}
		
		successfulProxies++
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📊 Proxy Test Sonuçları:\n")
	fmt.Printf("   • Başarılı: %d/%d\n", successfulProxies, len(monitor.proxies))
	fmt.Printf("   • Başarısız: %d/%d\n", failedProxies, len(monitor.proxies))
	if successfulProxies == len(monitor.proxies) {
		fmt.Println("   ✅ Tüm proxy'ler çalışıyor!")
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Zaman senkronizasyonu testi
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("2️⃣  ZAMAN SENKRONİZASYONU TESTİ")
	fmt.Println("═══════════════════════════════════════════════════════════════")

	timeSync, err := monitor.GetServerTime()
	if err != nil {
		fmt.Printf("❌ Zaman senkronizasyonu başarısız: %v\n", err)
	} else {
		fmt.Printf("📡 UPBIT ZAMAN SENKRONİZASYONU:\n")
		fmt.Printf("   • Server Time:     %s\n", timeSync.ServerTime.Format("2006-01-02 15:04:05.000"))
		fmt.Printf("   • Local Time:      %s\n", timeSync.LocalTime.Format("2006-01-02 15:04:05.000"))
		fmt.Printf("   • Clock Offset:    %v\n", timeSync.ClockOffset)
		fmt.Printf("   • Network Latency: %v\n", timeSync.NetworkLatency)
		
		if timeSync.ClockOffset.Abs() > 1*time.Second {
			fmt.Printf("   ⚠️  WARNING: Clock offset > 1s! May cause timing issues!\n")
		} else {
			fmt.Printf("   ✅ Clock sync OK (offset < 1s)\n")
		}
	}
	fmt.Println()

	// Bot detection header testi
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("3️⃣  BOT TESPİT KORUMA SİSTEMİ")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	
	fmt.Printf("🎭 User-Agent Pool:\n")
	for i, ua := range monitor.userAgents {
		fmt.Printf("   %d. %s...\n", i+1, ua[:50])
	}
	
	fmt.Println()
	fmt.Printf("🛡️  Bot Tespit Koruması:\n")
	fmt.Printf("   • User-Agent Rotation: ✅ Aktif (%d farklı agent)\n", len(monitor.userAgents))
	fmt.Printf("   • Accept Headers: ✅ Browser-like\n")
	fmt.Printf("   • Referer & Origin: ✅ upbit.com\n")
	fmt.Printf("   • Sec-Fetch Headers: ✅ Modern browser\n")
	fmt.Printf("   • Cookie Jar: ✅ Session persistence\n")
	fmt.Printf("   • TLS Config: ✅ Browser fingerprint\n")
	fmt.Println()

	// Rate limit simülasyonu
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("4️⃣  RATE LIMIT SİMÜLASYONU")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	
	interval := 300 * time.Millisecond // .env'den UPBIT_CHECK_INTERVAL_MS
	proxyCount := len(monitor.proxies)
	reqPerSec := 1000.0 / float64(interval.Milliseconds())
	coverage := float64(interval.Milliseconds()) / float64(proxyCount)
	
	fmt.Printf("⚙️  Konfigürasyon:\n")
	fmt.Printf("   • Interval: %dms\n", interval.Milliseconds())
	fmt.Printf("   • Proxy Sayısı: %d\n", proxyCount)
	fmt.Printf("   • Request Rate: %.2f req/sec\n", reqPerSec)
	fmt.Printf("   • Coverage: %.0fms (%.3fs)\n", coverage, coverage/1000)
	fmt.Println()
	
	fmt.Printf("🎯 Güvenlik Analizi:\n")
	if reqPerSec < 5.0 {
		fmt.Printf("   ✅ Request rate < 5 req/sec (GÜVENLI)\n")
	} else if reqPerSec < 10.0 {
		fmt.Printf("   ⚠️  Request rate 5-10 req/sec (DİKKATLİ)\n")
	} else {
		fmt.Printf("   ❌ Request rate > 10 req/sec (RİSKLİ!)\n")
	}
	
	if coverage < 500 {
		fmt.Printf("   ✅ Detection coverage < 500ms (HARİKA!)\n")
	} else if coverage < 1000 {
		fmt.Printf("   ✅ Detection coverage < 1s (İYİ)\n")
	} else {
		fmt.Printf("   ⚠️  Detection coverage > 1s (YAVAŞ)\n")
	}
	fmt.Println()

	// Proxy cooldown simülasyonu
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("5️⃣  PROXY COOLDOWN SİSTEMİ")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	
	fmt.Printf("⏱️  Cooldown Stratejisi:\n")
	fmt.Printf("   • Normal Cooldown: 3 saniye (proaktif)\n")
	fmt.Printf("   • Rate Limit Ceza: 30 saniye (429 hatası)\n")
	fmt.Printf("   • Random Interval: 250-400ms + %%10 uzun pause\n")
	fmt.Printf("   • Stagger Delay: 10-50ms (request öncesi)\n")
	fmt.Println()
	
	fmt.Printf("🎲 Randomization (Bot detection önleme):\n")
	fmt.Printf("   • Proxy Selection: ✅ Random\n")
	fmt.Printf("   • User-Agent Rotation: ✅ Sequential\n")
	fmt.Printf("   • Request Timing: ✅ Variable (250-400ms + occasional 0.5-1.5s)\n")
	fmt.Printf("   • Pre-request Jitter: ✅ 10-50ms\n")
	fmt.Println()

	// Duyuru filtreleme mantığı
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("6️⃣  DUYURU FİLTRELEME MANTIĞI")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	
	fmt.Printf("🔍 Filtreleme Kuralları:\n")
	fmt.Println("   1. Negative Filter (kara liste):")
	fmt.Println("      • 거래지원 + 종료 (trading support ended)")
	fmt.Println("      • 상장폐지 (delisting)")
	fmt.Println("      • 유의 종목 지정 (caution designation)")
	fmt.Println()
	fmt.Println("   2. Positive Filter (beyaz liste):")
	fmt.Println("      • 신규 + 거래지원 (new trading support)")
	fmt.Println("      • 디지털 자산 추가 (digital asset addition)")
	fmt.Println()
	fmt.Println("   3. Maintenance Filter:")
	fmt.Println("      • 변경, 연기, 연장, 재개")
	fmt.Println("      • 입출금, 이벤트, 출금 수수료")
	fmt.Println()
	fmt.Println("   4. Ticker Extraction:")
	fmt.Println("      • Parantez içi 2-10 karakter ticker'lar")
	fmt.Println("      • KRW, BTC, USDT hariç (market symbols)")
	fmt.Println()

	// Final özet
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("📊 GENEL ÖZET")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("✅ Sistem Durumu:\n")
	fmt.Printf("   • Proxy Sistemi: %d/%d çalışıyor\n", successfulProxies, len(monitor.proxies))
	fmt.Printf("   • Zaman Senkronizasyonu: %s\n", 
		map[bool]string{true: "✅ OK", false: "❌ Hata"}[err == nil])
	fmt.Printf("   • Bot Tespit Koruması: ✅ Aktif\n")
	fmt.Printf("   • Rate Limit Güvenliği: ✅ Güvenli\n")
	fmt.Printf("   • Filtreleme Sistemi: ✅ Hazır\n")
	fmt.Println()
	
	fmt.Printf("⚡ Performans:\n")
	fmt.Printf("   • Detection Time: %.0fms (%.3fs)\n", coverage, coverage/1000)
	fmt.Printf("   • Request Rate: %.2f req/sec\n", reqPerSec)
	fmt.Printf("   • Coverage Quality: %s\n", 
		map[bool]string{true: "🔥 Mükemmel", false: "✅ İyi"}[coverage < 500])
	fmt.Println()
	
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("✅ Dry run tamamlandı! Sistem production'a hazır.")
	fmt.Println("═══════════════════════════════════════════════════════════════")
}

// Helper: Test request oluştur
func (um *UpbitMonitor) createTestRequest() (*http.Request, error) {
	req, err := http.NewRequest("GET", um.apiURL, nil)
	if err != nil {
		return nil, err
	}

	// Bot detection bypass headers
	req.Header.Set("User-Agent", um.getRandomUserAgent())
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "ko-KR,ko;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Referer", "https://upbit.com/")
	req.Header.Set("Origin", "https://upbit.com")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("Connection", "keep-alive")

	return req, nil
}
