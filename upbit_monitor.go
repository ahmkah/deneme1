package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	json "github.com/json-iterator/go"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

type UpbitAPIResponse struct {
        Success bool       `json:"success"`
        Data    UpbitData2 `json:"data"`
}

type UpbitData2 struct {
        Notices []Announcement `json:"notices"`
}

type Announcement struct {
        ID    int    `json:"id"`
        Title string `json:"title"`
}

type ListingEntry struct {
        Symbol     string `json:"symbol"`
        Timestamp  string `json:"timestamp"`
        DetectedAt string `json:"detected_at"`
}

// Type aliases for compatibility with telegram_bot.go
type CoinDetection = ListingEntry
type UpbitDetection = ListingEntry

type TradeExecutionLog struct {
        Ticker               string                 `json:"ticker"`
        UpbitDetectedAt      string                 `json:"upbit_detected_at"`
        SavedToFileAt        string                 `json:"saved_to_file_at"`
        UserID               int64                  `json:"user_id"`
        BitgetOrderSentAt    string                 `json:"bitget_order_sent_at"`
        BitgetOrderConfirmed string                 `json:"bitget_order_confirmed_at"`
        LatencyBreakdown     map[string]interface{} `json:"latency_breakdown"`
}

type ETagChangeLog struct {
        ProxyIndex     int    `json:"proxy_index"`
        ProxyName      string `json:"proxy_name"`
        DetectedAt     string `json:"detected_at"`
        ServerTime     string `json:"server_time"`
        OldETag        string `json:"old_etag"`
        NewETag        string `json:"new_etag"`
        ResponseTimeMs int64  `json:"response_time_ms"`
}


	type UpbitMonitor struct {
	apiURL           string
	proxies          []string
	tickerRegex      *regexp.Regexp
	cachedTickers    map[string]bool
	proxyETags       map[int]string // Each proxy has its own ETag
	etagMu           sync.RWMutex   // Separate mutex for ETag operations
	proxyIndex       int
	mu               sync.Mutex
	jsonFile         string
	onNewListing     func(symbol string) // Callback for new listings
	executionLogFile string
	etagLogFile      string // ETag change detection log
	currentLogEntry  *TradeExecutionLog
	logMu            sync.Mutex
	// Intelligent Proxy Pool (Cooldowns for all proxies)
	proxyCooldowns   map[int]time.Time // proxy index -> cooldown expire time
	cooldownMu       sync.RWMutex
	// Timezone-based Scheduling
	pauseEnabled     bool
	pauseStart       int // Minutes since midnight (e.g., 13:00 = 780)
	pauseEnd         int // Minutes since midnight (e.g., 03:00 = 180)
	timezone         *time.Location
	isPaused         bool
	pauseMu          sync.Mutex
	// KST timezone for timestamps
	kstLocation      *time.Location
	// ETag processing control
	lastProcessedETag string
	etagProcessMu     sync.Mutex
	// Bot detection bypass
	userAgents        []string
	userAgentMu       sync.Mutex
	userAgentIndex    int
}

func NewUpbitMonitor(onNewListing func(string)) *UpbitMonitor {
        var proxies []string
        
        // Load up to 24 proxies (Proxy #1-2 should be Seoul for lowest latency)
        for i := 1; i <= 24; i++ {
                proxyEnv := os.Getenv(fmt.Sprintf("UPBIT_PROXY_%d", i))
                if proxyEnv != "" {
                        proxies = append(proxies, proxyEnv)
                }
        }

        if len(proxies) == 0 {
                proxies = []string{
                        "socks5://doproxy1:DigitalOcean55@143.198.221.194:1080",
                        "socks5://doproxy2:DigitalOcean55@159.223.68.49:1080",
                        "socks5://doproxy3:DigitalOcean55@104.248.147.230:1080",
                }
                log.Printf("⚠️ UPBIT_PROXY environment variables not set, using %d default proxies", len(proxies))
        } else {
                log.Printf("✅ Loaded %d proxies from environment variables", len(proxies))
        }

        // Load pause configuration
        pauseEnabled := os.Getenv("UPBIT_MONITOR_PAUSE_ENABLED") == "true"
        pauseStart := parseTimeToMinutes(os.Getenv("UPBIT_MONITOR_PAUSE_START"), 780)   // Default: 13:00
        pauseEnd := parseTimeToMinutes(os.Getenv("UPBIT_MONITOR_PAUSE_END"), 180)       // Default: 03:00
        tzName := os.Getenv("UPBIT_MONITOR_TZ")
        if tzName == "" {
                tzName = "Europe/Istanbul" // Default: Turkey time (UTC+3)
        }
        
        timezone, err := time.LoadLocation(tzName)
        if err != nil {
                log.Printf("⚠️ Invalid timezone '%s', using UTC", tzName)
                timezone = time.UTC
        }

        // Load KST timezone for Upbit timestamps
        kstLocation, err := time.LoadLocation("Asia/Seoul")
        if err != nil {
                log.Printf("⚠️ Failed to load KST timezone, using UTC: %v", err)
                kstLocation = time.UTC
        }

	// Realistic User-Agent pool (latest browsers)
	userAgents := []string{
		// Chrome on Windows
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
		// Chrome on macOS
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		// Firefox on Windows
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
		// Firefox on macOS
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0",
		// Safari on macOS
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
		// Edge on Windows
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
	}

	return &UpbitMonitor{
		apiURL:           "https://api-manager.upbit.com/api/v1/announcements?os=web&page=1&per_page=20&category=overall",
		proxies:          proxies,
		tickerRegex:      regexp.MustCompile(`\(([A-Z]{2,6})\)`), // Only 2-6 uppercase letters (valid tickers)
		cachedTickers:    make(map[string]bool),
		proxyETags:       make(map[int]string), // Initialize ETag map for each proxy
		proxyIndex:       0,
		jsonFile:         "upbit_new.json",
		executionLogFile: "trade_execution_log.json",
		proxyCooldowns:   make(map[int]time.Time), // Initialize cooldowns
		etagLogFile:      "etag_news.json",
		onNewListing:     onNewListing,
		pauseEnabled:     pauseEnabled,
		pauseStart:       pauseStart,
		pauseEnd:         pauseEnd,
		timezone:         timezone,
		isPaused:         false,
		kstLocation:      kstLocation,
		userAgents:       userAgents,
		userAgentIndex:   0,
	}
}

// parseTimeToMinutes converts "HH:MM" to minutes since midnight
func parseTimeToMinutes(timeStr string, defaultMinutes int) int {
        if timeStr == "" {
                return defaultMinutes
        }
        
        parts := regexp.MustCompile(`^(\d{1,2}):(\d{2})$`).FindStringSubmatch(timeStr)
        if len(parts) != 3 {
                log.Printf("⚠️ Invalid time format '%s', using default", timeStr)
                return defaultMinutes
        }
        
        var hour, minute int
        fmt.Sscanf(parts[1], "%d", &hour)
        fmt.Sscanf(parts[2], "%d", &minute)
        
        if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
                log.Printf("⚠️ Invalid time values in '%s', using default", timeStr)
                return defaultMinutes
        }
        
        return hour*60 + minute
}

func (um *UpbitMonitor) createProxyClient(proxyURL string) (*http.Client, error) {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("proxy URL'si ayrıştırılamadı: %w", err)
	}

	dialer, err := proxy.FromURL(parsedURL, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("proxy dialer oluşturulamadı: %w", err)
	}

	// TLS configuration to mimic real browsers and avoid fingerprinting
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
		InsecureSkipVerify: false, // Keep certificate validation
		// Cipher suites matching modern browsers
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		},
	}

	transport := &http.Transport{
		Dial:              dialer.Dial,
		TLSClientConfig:   tlsConfig,
		DisableKeepAlives: false, // Enable keep-alive like real browsers
		MaxIdleConns:      100,
		IdleConnTimeout:   90 * time.Second,
	}

	// Create cookie jar for session persistence
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("cookie jar oluşturulamadı: %w", err)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
		Jar:       jar, // Enable cookie handling
	}

	return client, nil
}

// getRandomUserAgent returns a random User-Agent from the pool
func (um *UpbitMonitor) getRandomUserAgent() string {
	um.userAgentMu.Lock()
	defer um.userAgentMu.Unlock()
	
	// Rotate through user agents
	userAgent := um.userAgents[um.userAgentIndex]
	um.userAgentIndex = (um.userAgentIndex + 1) % len(um.userAgents)
	
	return userAgent
}

func (um *UpbitMonitor) loadExistingData() error {
        if _, err := os.Stat(um.jsonFile); os.IsNotExist(err) {
                return nil
        }

        file, err := os.Open(um.jsonFile)
        if err != nil {
                return fmt.Errorf("error opening JSON file: %v", err)
        }
        defer file.Close()

        scanner := bufio.NewScanner(file)
        count := 0
        for scanner.Scan() {
                line := strings.TrimSpace(scanner.Text())
                if line == "" {
                        continue
                }
                
                var entry ListingEntry
                if err := json.Unmarshal([]byte(line), &entry); err != nil {
                        log.Printf("⚠️ Skipping invalid JSON line: %v", err)
                        continue
                }
                
                um.cachedTickers[entry.Symbol] = true
                count++
        }

        if err := scanner.Err(); err != nil {
                return fmt.Errorf("error reading JSON file: %v", err)
        }

        log.Printf("Loaded %d existing symbols from %s (JSONL format)", count, um.jsonFile)
        return nil
}

func (um *UpbitMonitor) saveToJSON(symbol string) error {
        // DUPLICATE CHECK: If symbol already exists in cache, skip saving
        if um.cachedTickers[symbol] {
                log.Printf("⚠️ DUPLICATE PREVENTED: %s already exists in cache, skipping save", symbol)
                return nil // Not an error, just skip
        }

        // Record detection timestamp for trade log
        detectedAt := time.Now()
        
        now := time.Now()
        newEntry := ListingEntry{
                Symbol:     symbol,
                Timestamp:  now.In(um.kstLocation).Format(time.RFC3339),
                DetectedAt: now.In(um.kstLocation).Format("2006-01-02 15:04:05 KST"),
        }

        // Append to JSONL file (O_APPEND mode)
        file, err := os.OpenFile(um.jsonFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
                return fmt.Errorf("error opening JSON file for append: %v", err)
        }
        defer file.Close()

        jsonData, err := json.Marshal(newEntry)
        if err != nil {
                return fmt.Errorf("error marshaling JSON: %v", err)
        }

        // Write JSON line + newline
        if _, err := file.Write(append(jsonData, '\n')); err != nil {
                return fmt.Errorf("error writing to JSON file: %v", err)
        }

        savedAt := time.Now()
        
        // Initialize trade execution log entry
        um.logMu.Lock()
        um.currentLogEntry = &TradeExecutionLog{
                Ticker:          symbol,
                UpbitDetectedAt: detectedAt.In(um.kstLocation).Format("2006-01-02 15:04:05.000000 KST"),
                SavedToFileAt:   savedAt.In(um.kstLocation).Format("2006-01-02 15:04:05.000000 KST"),
                LatencyBreakdown: make(map[string]interface{}),
        }
        um.logMu.Unlock()

        log.Printf("✅ Successfully saved NEW listing %s to %s (JSONL format)", symbol, um.jsonFile)
        return nil
}

// normalizeText: Unicode normalization and whitespace cleanup
func normalizeText(text string) string {
        // Remove punctuation and emojis, normalize whitespace
        reg := regexp.MustCompile(`[\p{P}\p{S}\p{Z}]+`)
        normalized := reg.ReplaceAllString(text, " ")
        normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")
        return regexp.MustCompile(`\s+`).ReplaceAllString(normalized, "")
}

// containsAll: Check if text contains all words (order independent)
func containsAll(text string, words []string) bool {
        normalized := normalizeText(text)
        for _, word := range words {
                if !regexp.MustCompile(normalizeText(word)).MatchString(normalized) {
                        return false
                }
        }
        return true
}

// containsAny: Check if text contains any word
func containsAny(text string, words []string) bool {
        normalized := normalizeText(text)
        for _, word := range words {
                if regexp.MustCompile(normalizeText(word)).MatchString(normalized) {
                        return true
                }
        }
        return false
}

// isNegativeFiltered: Rule 2 - Negative filtering (highest priority)
func isNegativeFiltered(title string) bool {
        negativeRules := [][]string{
                {"거래지원", "종료"},           // trading support ended
                {"상장폐지"},                   // delisting
                {"유의", "종목", "지정"},       // caution designation
                {"투자", "유의", "촉구"},       // investment caution warning
                {"유의", "촉구"},               // caution warning
                {"유의", "종목", "지정", "해제"}, // caution designation removal
        }
        
        for _, rule := range negativeRules {
                if containsAll(title, rule) {
                        return true
                }
        }
        return false
}

// isPositiveFiltered: Rule 3 - Positive filtering
func isPositiveFiltered(title string) bool {
        positiveRules := [][]string{
                {"신규", "거래지원"},     // new trading support
                {"디지털", "자산", "추가"}, // digital asset addition
        }
        
        for _, rule := range positiveRules {
                if containsAll(title, rule) {
                        return true
                }
        }
        return false
}

// isMaintenanceUpdate: Rule 4 - Maintenance/Update filter
func isMaintenanceUpdate(title string) bool {
        updateKeywords := []string{
                "변경", "연기", "연장", "재개", 
                "입출금", "이벤트", "출금 수수료",
        }
        
        if containsAny(title, updateKeywords) {
                return true
        }
        return false
}

// extractTickers: Rule 5 - Extract tickers from title
func extractTickers(title string) []string {
        var tickers []string
        tickerMap := make(map[string]bool)
        
        // Find all parentheses content
        parenRegex := regexp.MustCompile(`\(([^)]+)\)`)
        matches := parenRegex.FindAllStringSubmatch(title, -1)
        
        for _, match := range matches {
                content := match[1]
                
                // Skip if contains "마켓" (market indicator)
                if regexp.MustCompile(`마켓`).MatchString(content) {
                        continue
                }
                
                // Split by comma, trim, uppercase
                parts := regexp.MustCompile(`[,\s]+`).Split(content, -1)
                for _, part := range parts {
                        part = regexp.MustCompile(`\s+`).ReplaceAllString(part, "")
                        part = regexp.MustCompile(`[^A-Z0-9]`).ReplaceAllString(part, "")
                        
                        // Exclude market symbols
                        if part == "KRW" || part == "BTC" || part == "USDT" {
                                continue
                        }
                        
                        // Validate pattern [A-Z0-9]{1,10}
                        if regexp.MustCompile(`^[A-Z0-9]{1,10}$`).MatchString(part) {
                                if !tickerMap[part] {
                                        tickerMap[part] = true
                                        tickers = append(tickers, part)
                                }
                        }
                }
        }
        
        return tickers
}

func (um *UpbitMonitor) processAnnouncements(body io.Reader) {
        var response UpbitAPIResponse
        if err := json.NewDecoder(body).Decode(&response); err != nil {
                log.Printf("JSON verisi işlenemedi: %v", err)
                return
        }

        newTickers := make(map[string]bool)
        var newTickersList []string

        for _, announcement := range response.Data.Notices {
                title := announcement.Title
                
                // Rule 2: Negative filtering (highest priority - skips everything)
                if isNegativeFiltered(title) {
                        continue
                }
                
                // Rule 3: Positive filtering (must pass)
                if !isPositiveFiltered(title) {
                        continue
                }
                
                // Rule 4: Maintenance/Update filter
                if isMaintenanceUpdate(title) {
                        continue
                }
                
                // Rule 5: Extract tickers
                tickers := extractTickers(title)
                if len(tickers) > 0 {
                        for _, ticker := range tickers {
                                newTickers[ticker] = true
                                newTickersList = append(newTickersList, ticker)
                        }
                }
        }

        um.mu.Lock()
        defer um.mu.Unlock()

        var newlyAdded []string
        for ticker := range newTickers {
                if !um.cachedTickers[ticker] {
                        newlyAdded = append(newlyAdded, ticker)
                }
        }

        if len(newlyAdded) > 0 {
                fmt.Printf("\n🔥🔥🔥 YENİ LİSTELEME TESPİT EDİLDİ: %v 🔥🔥🔥\n", newlyAdded)
                for _, ticker := range newlyAdded {
                        um.cachedTickers[ticker] = true
                        if err := um.saveToJSON(ticker); err != nil {
                                log.Printf("Error saving ticker %s: %v", ticker, err)
                        }
                        if um.onNewListing != nil {
                                go um.onNewListing(ticker)
                        }
                }
        }

        // MERGE newTickers into cachedTickers (don't replace!)
        for ticker := range newTickers {
                um.cachedTickers[ticker] = true
        }
}

// checkProxy performs a single API check with one proxy
func (um *UpbitMonitor) checkProxy(proxyURL string, proxyIndex int) {
        client, err := um.createProxyClient(proxyURL)
        if err != nil {
                log.Printf("❌ Proxy #%d: Client creation failed: %v", proxyIndex+1, err)
                return
        }

        requestStart := time.Now()
        
	req, err := http.NewRequest("GET", um.apiURL, nil)
	if err != nil {
		log.Printf("❌ Proxy #%d: Request creation failed: %v", proxyIndex+1, err)
		return
	}

	// ============================================
	// COMPREHENSIVE BOT DETECTION BYPASS HEADERS
	// ============================================
	
	// 1. Realistic User-Agent (rotated from pool)
	req.Header.Set("User-Agent", um.getRandomUserAgent())
	
	// 2. Accept headers (matching real browser behavior)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "ko-KR,ko;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	
	// 3. Referer and Origin (simulate coming from Upbit website)
	req.Header.Set("Referer", "https://upbit.com/")
	req.Header.Set("Origin", "https://upbit.com")
	
	// 4. Sec-Fetch-* headers (modern browser security features)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	
	// 5. Connection settings (like real browsers)
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	
	// 6. Additional browser-like headers
	req.Header.Set("Sec-Ch-Ua", "\"Not_A Brand\";v=\"8\", \"Chromium\";v=\"120\", \"Google Chrome\";v=\"120\"")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", "\"Windows\"")
        
        // Each proxy uses its own ETag for independent caching
        um.etagMu.RLock()
        oldETag := um.proxyETags[proxyIndex]
        if oldETag != "" {
                req.Header.Set("If-None-Match", oldETag)
        }
        um.etagMu.RUnlock()

        resp, err := client.Do(req)
        responseTime := time.Since(requestStart).Milliseconds()
        
        if err != nil {
                log.Printf("❌ Proxy #%d: API request failed: %v", proxyIndex+1, err)
                return
        }

        switch resp.StatusCode {
        case http.StatusOK:
                newETag := resp.Header.Get("ETag")
                
                // Check if this ETag change was already processed by another proxy
                um.etagProcessMu.Lock()
                if um.lastProcessedETag == newETag {
                        // Already processed by another proxy, just update local ETag silently
                        um.etagMu.Lock()
                        um.proxyETags[proxyIndex] = newETag
                        um.etagMu.Unlock()
                        um.etagProcessMu.Unlock()
                        resp.Body.Close()
                        return
                }
                
                // This proxy is FIRST TO DETECT the change
                um.lastProcessedETag = newETag
                um.etagProcessMu.Unlock()
                
                log.Printf("🔥 Proxy #%d: FIRST TO DETECT ETag change! Processing...", proxyIndex+1)
                
                // Save ETag for this specific proxy and log the change atomically
                um.etagMu.Lock()
                oldETagValue := um.proxyETags[proxyIndex]
                um.proxyETags[proxyIndex] = newETag
                um.etagMu.Unlock()
                
                // Log ETag change to etag_news.json (async, with captured oldETag)
                go um.logETagChange(proxyIndex, oldETagValue, newETag, responseTime)
                
                um.processAnnouncements(resp.Body)
                resp.Body.Close()

        case http.StatusNotModified:
                resp.Body.Close()

        case http.StatusTooManyRequests: // 429 - Rate Limited
                log.Printf("⚠️ Proxy #%d: RATE LIMITED (429) - Cooldown for 30s", proxyIndex+1)
                resp.Body.Close()
                
                // Add to cooldown for 30 seconds
                um.cooldownMu.Lock()
                um.proxyCooldowns[proxyIndex] = time.Now().Add(30 * time.Second)
                um.cooldownMu.Unlock()

        default:
                log.Printf("⚠️ Proxy #%d: Unexpected status %d", proxyIndex+1, resp.StatusCode)
                resp.Body.Close()
        }
}

func (um *UpbitMonitor) Start() {
        log.Println("🚀 Upbit Monitor Starting with OPTIMIZED PROXY ROTATION...")

        if err := um.loadExistingData(); err != nil {
                log.Printf("⚠️ Warning: %v", err)
        }

        proxyCount := len(um.proxies)
        if proxyCount == 0 {
                log.Fatal("❌ No proxies configured! Please add UPBIT_PROXY_* to .env file")
        }

        log.Printf("📊 OPTIMIZED PROXY ROTATION CONFIGURATION:")
        log.Printf("   • Total Proxies: %d (rotating pool)", proxyCount)
        log.Printf("   • Strategy: 3s proactive cooldown + 30s rate limit penalty")
        log.Printf("   • Interval: 250-350ms random stagger")
        log.Printf("⚡ PERFORMANCE:")
        log.Printf("   • Detection Target: <500ms")
        log.Printf("   • Rate: ~3 req/sec (SAFE under Upbit's limit)")
        log.Printf("🎯 STRATEGY:")
        log.Printf("   • Proactive 3s cooldown per proxy")
        log.Printf("   • Random 250-350ms intervals")
        log.Printf("   • Auto-skip cooling down proxies")

        rand.Seed(time.Now().UnixNano())

        // Log pause configuration if enabled
        if um.pauseEnabled {
                log.Printf("⏸️  PAUSE SCHEDULE ENABLED:")
                log.Printf("   • Timezone: %s", um.timezone.String())
                log.Printf("   • Pause: %02d:%02d - %02d:%02d", 
                        um.pauseStart/60, um.pauseStart%60,
                        um.pauseEnd/60, um.pauseEnd%60)
        }

        log.Println("🚀 Optimized proxy rotation started!")

        for {
                // Check if we should pause (timezone-based scheduling)
                if um.pauseEnabled && um.shouldPauseNow() {
                        um.pauseMu.Lock()
                        if !um.isPaused {
                                um.isPaused = true
                                now := time.Now().In(um.timezone)
                                log.Printf("⏸️  PAUSING monitor (quiet hours) - Current time: %s %s", 
                                        now.Format("15:04:05"), um.timezone.String())
                                log.Printf("   Will resume at %02d:%02d %s", 
                                        um.pauseEnd/60, um.pauseEnd%60, um.timezone.String())
				}
			um.pauseMu.Unlock()
			// During pause, sleep with more variation
			time.Sleep(time.Duration(5000+rand.Intn(5000)) * time.Millisecond) // 5-10 seconds
			continue
                }

                // Check if we just resumed
                um.pauseMu.Lock()
                if um.isPaused {
                        um.isPaused = false
                        now := time.Now().In(um.timezone)
                        log.Printf("▶️  RESUMING monitor - Current time: %s %s", 
                                now.Format("15:04:05"), um.timezone.String())
                }
                um.pauseMu.Unlock()

		// Get available (non-cooling down) proxies
		availableIndices := um.getAvailableProxies()
		
		if len(availableIndices) == 0 {
			// No proxies available, wait with randomization
			time.Sleep(time.Duration(250+rand.Intn(150)) * time.Millisecond)
			continue
		}

                // Pick random proxy from available pool
                randomIndex := availableIndices[rand.Intn(len(availableIndices))]
                proxyURL := um.proxies[randomIndex]
                
                // PROACTIVE 3-second cooldown (Rule #3)
                um.cooldownMu.Lock()
                um.proxyCooldowns[randomIndex] = time.Now().Add(3 * time.Second)
                um.cooldownMu.Unlock()
                
		// Add random pre-request delay (human-like behavior)
		// Small jitter before request: 10-50ms
		preDelay := time.Duration(10+rand.Intn(40)) * time.Millisecond
		time.Sleep(preDelay)
		
		// Perform check with selected proxy
		um.checkProxy(proxyURL, randomIndex)
		
		// Random stagger with more variation: 250-400ms (more human-like)
		// Occasionally add longer pauses to mimic human behavior
		baseDelay := 250 + rand.Intn(150) // 250-400ms
		
		// 10% chance of longer pause (0.5-1.5 seconds) to mimic human reading/thinking
		if rand.Float32() < 0.10 {
			baseDelay = 500 + rand.Intn(1000) // 500-1500ms
		}
		
		time.Sleep(time.Duration(baseDelay) * time.Millisecond)
        }
}

// shouldPauseNow checks if current time is within pause window
func (um *UpbitMonitor) shouldPauseNow() bool {
        now := time.Now().In(um.timezone)
        currentMinutes := now.Hour()*60 + now.Minute()

        // Handle overnight window (e.g., 13:00-03:00 = 780-180)
        if um.pauseStart > um.pauseEnd {
                // Overnight: pause if >= start OR < end
                return currentMinutes >= um.pauseStart || currentMinutes < um.pauseEnd
        }
        
        // Same-day window (e.g., 01:00-05:00 = 60-300)
        return currentMinutes >= um.pauseStart && currentMinutes < um.pauseEnd
}

// getAvailableProxies returns indices of proxies that are not in cooldown
func (um *UpbitMonitor) getAvailableProxies() []int {
        um.cooldownMu.Lock()
        defer um.cooldownMu.Unlock()

        now := time.Now()
        var available []int
        var expired []int

        // First pass: collect available and expired
        for i := range um.proxies {
                expireTime, isInCooldown := um.proxyCooldowns[i]
                if !isInCooldown {
                        available = append(available, i)
                } else if now.After(expireTime) {
                        // Cooldown expired
                        expired = append(expired, i)
                        available = append(available, i)
                }
        }

        // Clean up expired cooldown entries
        for _, i := range expired {
                delete(um.proxyCooldowns, i)
        }

        return available
}

// appendTradeLog appends a trade execution log entry to the JSONL file
func (um *UpbitMonitor) appendTradeLog(logEntry *TradeExecutionLog) error {
        um.logMu.Lock()
        defer um.logMu.Unlock()

        // Append to JSONL file (O_APPEND mode)
        file, err := os.OpenFile(um.executionLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
                return fmt.Errorf("error opening execution log file for append: %v", err)
        }
        defer file.Close()

        jsonData, err := json.Marshal(logEntry)
        if err != nil {
                return fmt.Errorf("error marshaling execution log: %v", err)
        }

        // Write JSON line + newline
        if _, err := file.Write(append(jsonData, '\n')); err != nil {
                return fmt.Errorf("error writing to execution log file: %v", err)
        }

        log.Printf("📊 Trade execution log saved for %s", logEntry.Ticker)
        return nil
}

// GetCurrentLogEntry returns the current log entry (for use in ExecuteTrade)
func (um *UpbitMonitor) GetCurrentLogEntry(ticker string) *TradeExecutionLog {
        um.logMu.Lock()
        defer um.logMu.Unlock()
        
        if um.currentLogEntry != nil && um.currentLogEntry.Ticker == ticker {
                return um.currentLogEntry
        }
        return nil
}

// GetServerTime retrieves Upbit server time from HTTP response headers
func (um *UpbitMonitor) GetServerTime() (*TimeSyncResult, error) {
        localTimeBefore := time.Now()

        // Use any lightweight public endpoint
        client, err := um.createProxyClient(um.proxies[0])
        if err != nil {
                // Fallback to default client if proxy fails
                client = &http.Client{Timeout: 10 * time.Second}
        }

	req, err := http.NewRequest("GET", um.apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Apply comprehensive bot detection bypass headers
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

        resp, err := client.Do(req)
        if err != nil {
                return nil, fmt.Errorf("request failed: %w", err)
        }
        defer resp.Body.Close()

        localTimeAfter := time.Now()

        // Parse Date header (RFC1123 format)
        dateHeader := resp.Header.Get("Date")
        if dateHeader == "" {
                return nil, fmt.Errorf("no Date header in response")
        }

        serverTime, err := time.Parse(time.RFC1123, dateHeader)
        if err != nil {
                return nil, fmt.Errorf("failed to parse Date header: %w", err)
        }

        // Calculate network latency (round-trip time / 2)
        roundTripTime := localTimeAfter.Sub(localTimeBefore)
        networkLatency := roundTripTime / 2

        // Adjust server time for network latency
        adjustedServerTime := serverTime.Add(networkLatency)

        // Calculate clock offset
        clockOffset := adjustedServerTime.Sub(localTimeAfter)

        return &TimeSyncResult{
                ServerTime:     adjustedServerTime,
                LocalTime:      localTimeAfter,
                ClockOffset:    clockOffset,
                NetworkLatency: networkLatency,
        }, nil
}

// logETagChange logs ETag change detection events to etag_news.json (JSONL format)
func (um *UpbitMonitor) logETagChange(proxyIndex int, oldETag, newETag string, responseTimeMs int64) error {
        um.logMu.Lock()
        defer um.logMu.Unlock()

        // Create new log entry
        now := time.Now()
        proxyName := fmt.Sprintf("Proxy #%d", proxyIndex+1)
        if proxyIndex < 2 {
                proxyName += " (Seoul)"
        }
        
        logEntry := ETagChangeLog{
                ProxyIndex:     proxyIndex + 1,
                ProxyName:      proxyName,
                DetectedAt:     now.In(um.kstLocation).Format("2006-01-02 15:04:05.000 KST"),
                ServerTime:     now.UTC().Format(time.RFC3339Nano),
                OldETag:        oldETag,
                NewETag:        newETag,
                ResponseTimeMs: responseTimeMs,
        }

        // Append to JSONL file (O_APPEND mode)
        file, err := os.OpenFile(um.etagLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
                return fmt.Errorf("error opening etag log file for append: %v", err)
        }
        defer file.Close()

        jsonData, err := json.Marshal(logEntry)
        if err != nil {
                return fmt.Errorf("error marshaling etag log: %v", err)
        }

        // Write JSON line + newline
        if _, err := file.Write(append(jsonData, '\n')); err != nil {
                return fmt.Errorf("error writing to etag log file: %v", err)
        }

        // Safely truncate ETags for logging
        oldETagShort := "empty"
        if len(oldETag) >= 8 {
                oldETagShort = oldETag[:8]
        } else if len(oldETag) > 0 {
                oldETagShort = oldETag
        }
        
        newETagShort := "unknown"
        if len(newETag) >= 8 {
                newETagShort = newETag[:8]
        } else if len(newETag) > 0 {
                newETagShort = newETag
        }
        
        log.Printf("📝 ETag change logged: Proxy #%d, %s -> %s", proxyIndex+1, oldETagShort, newETagShort)
        return nil
}
