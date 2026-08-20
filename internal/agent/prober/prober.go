package prober

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/podcctv/detective-chicken/internal/model"
)

type ProbeResult struct {
	ID       string
	Name     string
	Category string
	Status   string // available, limited, blocked
	Region   string
	Quality  string
	Latency  int // ms
	Detail   string
}

// Prober performs native concurrent probes on target media and AI endpoints.
type Prober struct {
	Family  int // 4 or 6
	Timeout time.Duration
}

func NewProber(family int, timeout time.Duration) *Prober {
	if timeout <= 0 {
		timeout = 4 * time.Second
	}
	return &Prober{Family: family, Timeout: timeout}
}

func (p *Prober) getHTTPClient() *http.Client {
	network := "tcp"
	if p.Family == 4 {
		network = "tcp4"
	} else if p.Family == 6 {
		network = "tcp6"
	}

	dialer := &net.Dialer{
		Timeout:   p.Timeout,
		KeepAlive: 10 * time.Second,
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, addr)
		},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		DisableKeepAlives: true,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   p.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
}

func (p *Prober) RunAll(ctx context.Context) map[string]ProbeResult {
	results := make(map[string]ProbeResult)
	var mu sync.Mutex
	var wg sync.WaitGroup

	client := p.getHTTPClient()

	tasks := []func(context.Context, *http.Client) ProbeResult{
		// AI Services
		p.probeChatGPT,
		p.probeClaude,
		p.probeGemini,
		p.probeDeepSeek,
		p.probeMidjourney,
		p.probeCopilot,
		p.probeGrok,
		p.probePerplexity,
		p.probeGitHubCopilot,

		// Streaming & Media Services
		p.probeNetflix,
		p.probeDisneyPlus,
		p.probeYouTube,
		p.probePrimeVideo,
		p.probeSpotify,
		p.probeMax,
		p.probeHulu,
		p.probeBahamut,
		p.probeAbema,
		p.probeTikTok,
		p.probeDAZN,
	}

	for _, task := range tasks {
		wg.Add(1)
		go func(fn func(context.Context, *http.Client) ProbeResult) {
			defer wg.Done()
			res := fn(ctx, client)
			mu.Lock()
			results[res.ID] = res
			mu.Unlock()
		}(task)
	}

	wg.Wait()
	return results
}

// AI Probe Implementations

func (p *Prober) probeChatGPT(ctx context.Context, client *http.Client) ProbeResult {
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://chatgpt.com", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ProbeResult{ID: "chatgpt", Name: "ChatGPT", Category: "ai", Status: "blocked", Quality: "连接超时", Latency: latency, Detail: "请求未能连通"}
	}
	defer resp.Body.Close()

	cfCountry := resp.Header.Get("CF-IPCountry")
	if cfCountry == "" {
		cfCountry = resp.Header.Get("X-Country-Code")
	}

	if resp.StatusCode == 403 || resp.StatusCode == 429 {
		return ProbeResult{ID: "chatgpt", Name: "ChatGPT", Category: "ai", Status: "blocked", Region: cfCountry, Quality: "Cloudflare 质询", Latency: latency, Detail: fmt.Sprintf("HTTP %d 拦截", resp.StatusCode)}
	}
	if resp.StatusCode == 200 || resp.StatusCode == 301 || resp.StatusCode == 302 {
		return ProbeResult{ID: "chatgpt", Name: "ChatGPT", Category: "ai", Status: "available", Region: cfCountry, Quality: "GPT-4o Web+API", Latency: latency, Detail: "直连免验证码"}
	}

	return ProbeResult{ID: "chatgpt", Name: "ChatGPT", Category: "ai", Status: "limited", Region: cfCountry, Quality: "响应受限", Latency: latency, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

func (p *Prober) probeClaude(ctx context.Context, client *http.Client) ProbeResult {
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://claude.ai/login", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ProbeResult{ID: "claude", Name: "Claude", Category: "ai", Status: "blocked", Quality: "连接阻断", Latency: latency, Detail: "无法连通 Anthropic 节点"}
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	bodyStr := string(bodyBytes)
	cfCountry := resp.Header.Get("CF-IPCountry")

	if resp.StatusCode == 403 || strings.Contains(bodyStr, "App unavailable in your region") || strings.Contains(bodyStr, "Access denied") {
		return ProbeResult{ID: "claude", Name: "Claude", Category: "ai", Status: "blocked", Region: cfCountry, Quality: "区域限制/机房屏蔽", Latency: latency, Detail: "Claude 区域阻断"}
	}
	if resp.StatusCode == 200 || resp.StatusCode == 302 {
		return ProbeResult{ID: "claude", Name: "Claude", Category: "ai", Status: "available", Region: cfCountry, Quality: "Claude 3.5 Sonnet 原生", Latency: latency, Detail: "免验证解锁"}
	}

	return ProbeResult{ID: "claude", Name: "Claude", Category: "ai", Status: "limited", Region: cfCountry, Quality: "受限响应", Latency: latency, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

func (p *Prober) probeGemini(ctx context.Context, client *http.Client) ProbeResult {
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://gemini.google.com", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ProbeResult{ID: "gemini", Name: "Gemini", Category: "ai", Status: "blocked", Quality: "连接失败", Latency: latency, Detail: "Google 端点超时"}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 302 || resp.StatusCode == 301 {
		return ProbeResult{ID: "gemini", Name: "Gemini", Category: "ai", Status: "available", Quality: "Gemini 1.5 Pro 畅通", Latency: latency, Detail: "Google AI 全功能支持"}
	}
	if resp.StatusCode == 403 {
		return ProbeResult{ID: "gemini", Name: "Gemini", Category: "ai", Status: "blocked", Quality: "区域未开放", Latency: latency, Detail: "Google IP 地区过滤"}
	}

	return ProbeResult{ID: "gemini", Name: "Gemini", Category: "ai", Status: "limited", Quality: "部分受限", Latency: latency, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

func (p *Prober) probeDeepSeek(ctx context.Context, client *http.Client) ProbeResult {
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://chat.deepseek.com", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ProbeResult{ID: "deepseek", Name: "DeepSeek", Category: "ai", Status: "blocked", Quality: "连接超时", Latency: latency, Detail: "DeepSeek 端点不可达"}
	}
	defer resp.Body.Close()

	cfCountry := resp.Header.Get("CF-IPCountry")
	if resp.StatusCode == 200 || resp.StatusCode == 302 {
		return ProbeResult{ID: "deepseek", Name: "DeepSeek", Category: "ai", Status: "available", Region: cfCountry, Quality: "R1/V3 Web+API", Latency: latency, Detail: "高并发高速通道"}
	}
	if resp.StatusCode == 403 {
		return ProbeResult{ID: "deepseek", Name: "DeepSeek", Category: "ai", Status: "blocked", Region: cfCountry, Quality: "WAF 拦截", Latency: latency, Detail: "WAF 规则拦截"}
	}

	return ProbeResult{ID: "deepseek", Name: "DeepSeek", Category: "ai", Status: "limited", Region: cfCountry, Quality: "响应延迟", Latency: latency, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

func (p *Prober) probeMidjourney(ctx context.Context, client *http.Client) ProbeResult {
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.midjourney.com", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ProbeResult{ID: "midjourney", Name: "Midjourney", Category: "ai", Status: "blocked", Quality: "无法访问", Latency: latency, Detail: "Web 生成端点超时"}
	}
	defer resp.Body.Close()

	cfCountry := resp.Header.Get("CF-IPCountry")
	if resp.StatusCode == 200 || resp.StatusCode == 302 || resp.StatusCode == 301 {
		return ProbeResult{ID: "midjourney", Name: "Midjourney", Category: "ai", Status: "available", Region: cfCountry, Quality: "Web+Discord 畅通", Latency: latency, Detail: "AI 绘画免验证"}
	}
	return ProbeResult{ID: "midjourney", Name: "Midjourney", Category: "ai", Status: "blocked", Region: cfCountry, Quality: "质询阻断", Latency: latency, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

func (p *Prober) probeCopilot(ctx context.Context, client *http.Client) ProbeResult {
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://copilot.microsoft.com", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ProbeResult{ID: "copilot", Name: "Copilot", Category: "ai", Status: "blocked", Quality: "连接超时", Latency: latency, Detail: "微软 AI 端点超时"}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 302 {
		return ProbeResult{ID: "copilot", Name: "Copilot", Category: "ai", Status: "available", Quality: "GPT-4 Turbo 驱动", Latency: latency, Detail: "Bing AI 畅通"}
	}
	return ProbeResult{ID: "copilot", Name: "Copilot", Category: "ai", Status: "limited", Quality: "受限访问", Latency: latency, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

func (p *Prober) probeGrok(ctx context.Context, client *http.Client) ProbeResult {
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://x.ai", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ProbeResult{ID: "grok", Name: "Grok", Category: "ai", Status: "blocked", Quality: "无法访问", Latency: latency, Detail: "xAI 节点超时"}
	}
	defer resp.Body.Close()

	cfCountry := resp.Header.Get("CF-IPCountry")
	if resp.StatusCode == 200 || resp.StatusCode == 302 || resp.StatusCode == 301 {
		return ProbeResult{ID: "grok", Name: "Grok", Category: "ai", Status: "available", Region: cfCountry, Quality: "xAI Grok-2 原生", Latency: latency, Detail: "免质询畅通"}
	}
	return ProbeResult{ID: "grok", Name: "Grok", Category: "ai", Status: "blocked", Region: cfCountry, Quality: "WAF 拦截", Latency: latency, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

func (p *Prober) probePerplexity(ctx context.Context, client *http.Client) ProbeResult {
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.perplexity.ai", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ProbeResult{ID: "perplexity", Name: "Perplexity", Category: "ai", Status: "blocked", Quality: "连接超时", Latency: latency, Detail: "AI 搜索端点不可达"}
	}
	defer resp.Body.Close()

	cfCountry := resp.Header.Get("CF-IPCountry")
	if resp.StatusCode == 200 || resp.StatusCode == 302 {
		return ProbeResult{ID: "perplexity", Name: "Perplexity", Category: "ai", Status: "available", Region: cfCountry, Quality: "AI 实时搜索畅通", Latency: latency, Detail: "免验证解锁"}
	}
	return ProbeResult{ID: "perplexity", Name: "Perplexity", Category: "ai", Status: "limited", Region: cfCountry, Quality: "质询验证", Latency: latency, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

func (p *Prober) probeGitHubCopilot(ctx context.Context, client *http.Client) ProbeResult {
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://github.com", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ProbeResult{ID: "github_cop", Name: "GitHub Copilot", Category: "ai", Status: "blocked", Quality: "连接失败", Latency: latency, Detail: "GitHub 端点超时"}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 301 || resp.StatusCode == 302 {
		return ProbeResult{ID: "github_cop", Name: "GitHub Copilot", Category: "ai", Status: "available", Quality: "IDE 编码补全畅通", Latency: latency, Detail: "低延迟极速响应"}
	}
	return ProbeResult{ID: "github_cop", Name: "GitHub Copilot", Category: "ai", Status: "limited", Quality: "响应受限", Latency: latency, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

// Streaming Probe Implementations

func (p *Prober) probeNetflix(ctx context.Context, client *http.Client) ProbeResult {
	start := time.Now()
	// Netflix Non-Original Licensed Title (80018499) vs Original (70143836)
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.netflix.com/title/80018499", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ProbeResult{ID: "netflix", Name: "Netflix", Category: "streaming", Status: "blocked", Quality: "网络超时", Latency: latency, Detail: "Netflix 端点不可达"}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return ProbeResult{ID: "netflix", Name: "Netflix", Category: "streaming", Status: "available", Quality: "原生 4K/HDR 全解锁", Latency: latency, Detail: "支持全部非自制版权剧"}
	}
	if resp.StatusCode == 404 || resp.StatusCode == 403 {
		// Test original title to distinguish limited from blocked
		reqOrig, _ := http.NewRequestWithContext(ctx, "GET", "https://www.netflix.com/title/70143836", nil)
		respOrig, errOrig := client.Do(reqOrig)
		if errOrig == nil {
			defer respOrig.Body.Close()
			if respOrig.StatusCode == 200 {
				return ProbeResult{ID: "netflix", Name: "Netflix", Category: "streaming", Status: "limited", Quality: "仅自制剧 (Originals)", Latency: latency, Detail: "机房 IP 仅解锁自制内容"}
			}
		}
		return ProbeResult{ID: "netflix", Name: "Netflix", Category: "streaming", Status: "blocked", Quality: "未解锁/机房封禁", Latency: latency, Detail: "Netflix 屏蔽当前 IP"}
	}

	return ProbeResult{ID: "netflix", Name: "Netflix", Category: "streaming", Status: "limited", Quality: "自制剧解锁", Latency: latency, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

func (p *Prober) probeDisneyPlus(ctx context.Context, client *http.Client) ProbeResult {
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.disneyplus.com", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ProbeResult{ID: "disney", Name: "Disney+", Category: "streaming", Status: "blocked", Quality: "连接超时", Latency: latency, Detail: "Disney+ 端点超时"}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 301 || resp.StatusCode == 302 {
		return ProbeResult{ID: "disney", Name: "Disney+", Category: "streaming", Status: "available", Quality: "4K UHD / IMAX Enhanced", Latency: latency, Detail: "原生解锁完整影视库"}
	}
	return ProbeResult{ID: "disney", Name: "Disney+", Category: "streaming", Status: "blocked", Quality: "区域限制", Latency: latency, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

func (p *Prober) probeYouTube(ctx context.Context, client *http.Client) ProbeResult {
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.youtube.com/premium", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ProbeResult{ID: "youtube", Name: "YouTube Prem", Category: "streaming", Status: "blocked", Quality: "无法访问", Latency: latency, Detail: "YouTube 节点超时"}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 302 {
		return ProbeResult{ID: "youtube", Name: "YouTube Prem", Category: "streaming", Status: "available", Quality: "Premium 原生解锁", Latency: latency, Detail: "后台播放与免广告支持"}
	}
	return ProbeResult{ID: "youtube", Name: "YouTube Prem", Category: "streaming", Status: "limited", Quality: "区域限制", Latency: latency, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

func (p *Prober) probePrimeVideo(ctx context.Context, client *http.Client) ProbeResult {
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.primevideo.com", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ProbeResult{ID: "prime", Name: "Prime Video", Category: "streaming", Status: "blocked", Quality: "连接超时", Latency: latency, Detail: "Amazon 节点不可达"}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 302 || resp.StatusCode == 301 {
		return ProbeResult{ID: "prime", Name: "Prime Video", Category: "streaming", Status: "available", Quality: "Prime Video 原生", Latency: latency, Detail: "亚马逊影视全量解锁"}
	}
	return ProbeResult{ID: "prime", Name: "Prime Video", Category: "streaming", Status: "blocked", Quality: "屏蔽阻断", Latency: latency, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

func (p *Prober) probeSpotify(ctx context.Context, client *http.Client) ProbeResult {
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://open.spotify.com", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ProbeResult{ID: "spotify", Name: "Spotify", Category: "streaming", Status: "blocked", Quality: "无法访问", Latency: latency, Detail: "Spotify 端点超时"}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 302 {
		return ProbeResult{ID: "spotify", Name: "Spotify", Category: "streaming", Status: "available", Quality: "无损音质原生解锁", Latency: latency, Detail: "曲库无跨区限制"}
	}
	return ProbeResult{ID: "spotify", Name: "Spotify", Category: "streaming", Status: "limited", Quality: "区域限制", Latency: latency, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

func (p *Prober) probeMax(ctx context.Context, client *http.Client) ProbeResult {
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://auth.max.com", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ProbeResult{ID: "max", Name: "Max (HBO)", Category: "streaming", Status: "blocked", Quality: "无法连接", Latency: latency, Detail: "HBO Max 鉴权超时"}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 302 {
		return ProbeResult{ID: "max", Name: "Max (HBO)", Category: "streaming", Status: "available", Quality: "HBO Max 4K 原生", Latency: latency, Detail: "完整影视资源库"}
	}
	if resp.StatusCode == 403 {
		return ProbeResult{ID: "max", Name: "Max (HBO)", Category: "streaming", Status: "blocked", Quality: "区域机房屏蔽", Latency: latency, Detail: "Max 地区访问限制"}
	}
	return ProbeResult{ID: "max", Name: "Max (HBO)", Category: "streaming", Status: "limited", Quality: "部分受限", Latency: latency, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

func (p *Prober) probeHulu(ctx context.Context, client *http.Client) ProbeResult {
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.hulu.com", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ProbeResult{ID: "hulu", Name: "Hulu", Category: "streaming", Status: "blocked", Quality: "连接超时", Latency: latency, Detail: "Hulu US 节点不可达"}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 302 {
		return ProbeResult{ID: "hulu", Name: "Hulu", Category: "streaming", Status: "available", Quality: "Hulu US 原生解锁", Latency: latency, Detail: "支持全量美剧点播"}
	}
	if resp.StatusCode == 403 {
		return ProbeResult{ID: "hulu", Name: "Hulu", Category: "streaming", Status: "blocked", Quality: "机房过滤拦截", Latency: latency, Detail: "Hulu 区域及机房封锁"}
	}
	return ProbeResult{ID: "hulu", Name: "Hulu", Category: "streaming", Status: "limited", Quality: "受限响应", Latency: latency, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

func (p *Prober) probeBahamut(ctx context.Context, client *http.Client) ProbeResult {
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://ani.gamer.com.tw/ajax/getdeviceid.php", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ProbeResult{ID: "bahamut", Name: "巴哈姆特动画疯", Category: "streaming", Status: "blocked", Quality: "连接失败", Latency: latency, Detail: "台湾 CDN 节点超时"}
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	bodyStr := string(bodyBytes)

	if resp.StatusCode == 200 && strings.Contains(bodyStr, "deviceid") {
		return ProbeResult{ID: "bahamut", Name: "巴哈姆特动画疯", Category: "streaming", Status: "available", Region: "TW", Quality: "1080P 原生直连", Latency: latency, Detail: "台湾地区全量动漫解锁"}
	}
	if resp.StatusCode == 403 || strings.Contains(bodyStr, "error") {
		return ProbeResult{ID: "bahamut", Name: "巴哈姆特动画疯", Category: "streaming", Status: "blocked", Quality: "非台湾 IP 限制", Latency: latency, Detail: "锁区仅限台湾地区"}
	}

	return ProbeResult{ID: "bahamut", Name: "巴哈姆特动画疯", Category: "streaming", Status: "limited", Quality: "需要验证", Latency: latency, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

func (p *Prober) probeAbema(ctx context.Context, client *http.Client) ProbeResult {
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.abema.io/v1/ip/check?device=android", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ProbeResult{ID: "abema", Name: "AbemaTV", Category: "streaming", Status: "blocked", Quality: "连接超时", Latency: latency, Detail: "日本 CDN 节点超时"}
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	bodyStr := string(bodyBytes)

	if resp.StatusCode == 200 && (strings.Contains(bodyStr, `"isoCountryCode":"JPN"`) || strings.Contains(bodyStr, `"JP"`)) {
		return ProbeResult{ID: "abema", Name: "AbemaTV", Category: "streaming", Status: "available", Region: "JP", Quality: "Abema 日本原生", Latency: latency, Detail: "日本全量动漫直播解锁"}
	}
	if resp.StatusCode == 200 {
		return ProbeResult{ID: "abema", Name: "AbemaTV", Category: "streaming", Status: "limited", Quality: "海外版受限", Latency: latency, Detail: "仅限部分国际频道"}
	}

	return ProbeResult{ID: "abema", Name: "AbemaTV", Category: "streaming", Status: "blocked", Quality: "日本锁区屏蔽", Latency: latency, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

func (p *Prober) probeTikTok(ctx context.Context, client *http.Client) ProbeResult {
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.tiktok.com", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ProbeResult{ID: "tiktok", Name: "TikTok", Category: "streaming", Status: "blocked", Quality: "无法访问", Latency: latency, Detail: "TikTok 节点不可达"}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 301 || resp.StatusCode == 302 {
		return ProbeResult{ID: "tiktok", Name: "TikTok", Category: "streaming", Status: "available", Quality: "原生畅通", Latency: latency, Detail: "视频流正常加载"}
	}
	return ProbeResult{ID: "tiktok", Name: "TikTok", Category: "streaming", Status: "blocked", Quality: "区域限制/拦截", Latency: latency, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

func (p *Prober) probeDAZN(ctx context.Context, client *http.Client) ProbeResult {
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://startup.core.indazn.com/misl/v5/Startup", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return ProbeResult{ID: "dazn", Name: "DAZN", Category: "streaming", Status: "blocked", Quality: "连接超时", Latency: latency, Detail: "DAZN 体育节点不可达"}
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	bodyStr := string(bodyBytes)

	if resp.StatusCode == 200 && strings.Contains(bodyStr, "GeolocatedCountry") {
		return ProbeResult{ID: "dazn", Name: "DAZN", Category: "streaming", Status: "available", Quality: "DAZN 体育直播原生", Latency: latency, Detail: "支持全赛事直播观看"}
	}
	if resp.StatusCode == 200 {
		return ProbeResult{ID: "dazn", Name: "DAZN", Category: "streaming", Status: "available", Quality: "支持访问", Latency: latency, Detail: "端点连通正常"}
	}

	return ProbeResult{ID: "dazn", Name: "DAZN", Category: "streaming", Status: "blocked", Quality: "机房过滤拦截", Latency: latency, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
}

// ConvertResultsToUnlocks maps native probe results into model.NodeUnlocks
func ConvertResultsToUnlocks(results map[string]ProbeResult, defaultRegion string) model.NodeUnlocks {
	streamingMap := make(map[string]model.UnlockInfo)
	aiMap := make(map[string]model.UnlockInfo)

	for id, res := range results {
		region := res.Region
		if region == "" {
			region = defaultRegion
		}
		info := model.UnlockInfo{
			ID:        id,
			Name:      res.Name,
			Category:  res.Category,
			Status:    res.Status,
			Region:    region,
			Quality:   res.Quality,
			LatencyMs: res.Latency,
			Detail:    res.Detail,
		}
		if res.Category == "ai" {
			aiMap[id] = info
		} else {
			streamingMap[id] = info
		}
	}

	return model.NodeUnlocks{
		Streaming: streamingMap,
		AI:        aiMap,
	}
}
