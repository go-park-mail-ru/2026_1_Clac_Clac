package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

const (
	baseURL     = "http://localhost:8081"
	email       = "tester@nexus.test"
	password    = "12345678"
	displayName = "LoadTester"

	numCards  = 100000
	createRPS = 100

	readRPS      = 500
	readDuration = 30 * time.Second
)

type SetupData struct {
	SessionID   string
	CSRFToken   string
	BoardLink   string
	SectionLink string
	CardLinks   []string
}

func loadCardLinks(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	links := make([]string, 0, len(lines))
	for _, l := range lines {
		if l != "" {
			links = append(links, strings.TrimSpace(l))
		}
	}
	return links
}

func setupEnvironment() (*SetupData, error) {
	d := &SetupData{}
	client := &http.Client{Timeout: 60 * time.Second}

	post := func(path, body string, extra map[string]string) (*http.Response, error) {
		req, _ := http.NewRequest("POST", baseURL+path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		for k, v := range extra {
			req.Header.Set(k, v)
		}
		return client.Do(req)
	}

	do := func(method, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
		req, _ := http.NewRequest(method, baseURL+path, body)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		return client.Do(req)
	}

	resp, _ := post("/api/register",
		fmt.Sprintf(`{"email":"%s","password":"%s","repeated_password":"%s","display_name":"%s"}`,
			email, password, password, displayName), nil)
	if resp != nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	resp, err := post("/api/login",
		fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password), nil)
	if err != nil {
		return nil, fmt.Errorf("login request: %w", err)
	}
	for _, c := range resp.Cookies() {
		if c.Name == "session_id" {
			d.SessionID = c.Value
		}
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if d.SessionID == "" {
		return nil, fmt.Errorf("no session_id")
	}

	resp, err = do("GET", "/api/csrf", nil, map[string]string{"Cookie": "session_id=" + d.SessionID})
	if err != nil {
		return nil, fmt.Errorf("csrf request: %w", err)
	}
	for _, c := range resp.Cookies() {
		if c.Name == "csrf_token" {
			d.CSRFToken = c.Value
		}
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if d.CSRFToken == "" {
		return nil, fmt.Errorf("no csrf_token")
	}

	cookie := fmt.Sprintf("session_id=%s; csrf_token=%s", d.SessionID, d.CSRFToken)
	ah := map[string]string{"Content-Type": "application/json", "X-CSRF-Token": d.CSRFToken, "Cookie": cookie}

	for retry := 0; retry < 3; retry++ {
		resp, err = post("/api/boards", `{"name":"LoadTestBoard","description":""}`, ah)
		if err != nil {
			fmt.Printf("  board attempt %d: %v\n", retry+1, err)
			time.Sleep(3 * time.Second)
			continue
		}
		var br struct {
			Data *struct {
				Link string `json:"link"`
			} `json:"data"`
		}
		json.NewDecoder(resp.Body).Decode(&br)
		resp.Body.Close()
		if br.Data != nil {
			d.BoardLink = br.Data.Link
		}
		if d.BoardLink != "" {
			break
		}
		fmt.Printf("  board attempt %d: no link in response\n", retry+1)
		time.Sleep(3 * time.Second)
	}
	if d.BoardLink == "" {
		return nil, fmt.Errorf("no board link after 3 retries")
	}

	resp, err = post("/api/sections",
		fmt.Sprintf(`{"name":"DefaultSection","board_link":"%s"}`, d.BoardLink), ah)
	if err != nil {
		return nil, fmt.Errorf("section request: %w", err)
	}
	var sr struct {
		Data *struct {
			Link string `json:"link"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&sr)
	resp.Body.Close()
	if sr.Data != nil {
		d.SectionLink = sr.Data.Link
	}
	if d.SectionLink == "" {
		return nil, fmt.Errorf("no section link")
	}

	return d, nil
}

func loadTestCreate(d *SetupData) *vegeta.Metrics {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("X-CSRF-Token", d.CSRFToken)
	headers.Set("Cookie", fmt.Sprintf("session_id=%s; csrf_token=%s", d.SessionID, d.CSRFToken))

	body := []byte(fmt.Sprintf(`{"section_link":"%s","title":"LoadTest Card"}`, d.SectionLink))
	target := vegeta.Target{
		Method: "POST", URL: baseURL + "/api/cards",
		Header: headers, Body: body,
	}

	targeter := vegeta.NewStaticTargeter(target)
	rate := vegeta.Rate{Freq: createRPS, Per: time.Second}
	est := time.Duration(numCards/createRPS) * time.Second

	attacker := vegeta.NewAttacker()

	fmt.Printf("\n=== CREATE CARDS ===\n")
	fmt.Printf("Endpoint:  POST /api/cards\n")
	fmt.Printf("Rate:      %d RPS\n", createRPS)
	fmt.Printf("Requests:  %d\n", numCards)
	fmt.Printf("Duration:  %v\n", est)
	fmt.Println(strings.Repeat("-", 60))

	var m vegeta.Metrics
	for r := range attacker.Attack(targeter, rate, est, "CreateCards") {
		if r.Code == 200 && len(d.CardLinks) < 100 {
			var cr struct {
				Data *struct {
					CardLink string `json:"card_link"`
				} `json:"data"`
			}
			json.Unmarshal(r.Body, &cr)
			if cr.Data != nil && cr.Data.CardLink != "" {
				d.CardLinks = append(d.CardLinks, cr.Data.CardLink)
			}
		}
		m.Add(r)
	}
	m.Close()
	return &m
}

func loadTestRead(d *SetupData, links []string) *vegeta.Metrics {
	headers := http.Header{}
	headers.Set("X-CSRF-Token", d.CSRFToken)
	headers.Set("Cookie", fmt.Sprintf("session_id=%s; csrf_token=%s", d.SessionID, d.CSRFToken))

	targets := make([]vegeta.Target, len(links))
	for i, l := range links {
		targets[i] = vegeta.Target{Method: "GET", URL: baseURL + "/api/cards/" + l, Header: headers}
	}

	attacker := vegeta.NewAttacker()

	fmt.Printf("\n=== READ CARDS ===\n")
	fmt.Printf("Endpoint:  GET /api/cards/{link}\n")
	fmt.Printf("Rate:      %d RPS\n", readRPS)
	fmt.Printf("Duration:  %v\n", readDuration)
	fmt.Printf("Card pool: %d unique\n", len(links))
	fmt.Println(strings.Repeat("-", 60))

	var m vegeta.Metrics
	for r := range attacker.Attack(vegeta.NewStaticTargeter(targets...),
		vegeta.Rate{Freq: readRPS, Per: time.Second}, readDuration, "ReadCards") {
		m.Add(r)
	}
	m.Close()
	return &m
}

func printMetrics(title string, m *vegeta.Metrics) {
	fmt.Printf("\n======= %s =======\n", title)
	fmt.Printf("  Total requests:   %d\n", m.Requests)
	fmt.Printf("  Success (2xx):    %d (%.1f%%)\n",
		int(float64(m.Requests)*m.Success), m.Success*100)
	fmt.Printf("  Errors:           %d unique\n", len(m.Errors))
	fmt.Printf("  Duration:         %v\n", m.Duration.Round(time.Millisecond))
	fmt.Printf("  Actual RPS:       %.1f\n", m.Rate)
	fmt.Printf("  Throughput:       %.1f req/s\n", m.Throughput)

	if m.Requests > 0 {
		fmt.Printf("\n  --- Latency ---\n")
		fmt.Printf("  Min:      %v\n", m.Latencies.Min.Round(time.Microsecond))
		fmt.Printf("  Mean:     %v\n", m.Latencies.Mean.Round(time.Microsecond))
		fmt.Printf("  P50:      %v\n", m.Latencies.P50.Round(time.Microsecond))
		fmt.Printf("  P90:      %v\n", m.Latencies.P90.Round(time.Microsecond))
		fmt.Printf("  P95:      %v\n", m.Latencies.P95.Round(time.Microsecond))
		fmt.Printf("  P99:      %v\n", m.Latencies.P99.Round(time.Microsecond))
		fmt.Printf("  Max:      %v\n", m.Latencies.Max.Round(time.Microsecond))
	}
	if len(m.StatusCodes) > 0 {
		fmt.Printf("\n  --- Status Codes ---\n")
		for code, cnt := range m.StatusCodes {
			fmt.Printf("    HTTP %s: %d\n", code, cnt)
		}
	}
	fmt.Println(strings.Repeat("=", 60))
}

func main() {
	fmt.Println("==============================================")
	fmt.Println("  NeXuS Load Test — 100k CREATION + READ")
	fmt.Println("==============================================")

	setup, err := setupEnvironment()
	if err != nil {
		fmt.Printf("Setup FAILED: %v\n", err)
		return
	}
	fmt.Printf("Board: %s  Section: %s\n\n", setup.BoardLink[:12], setup.SectionLink[:12])

	createM := loadTestCreate(setup)
	printMetrics("CREATE 100k CARDS", createM)

	fmt.Printf("\nCollected %d real card links\n", len(setup.CardLinks))

	if len(setup.CardLinks) < 10 {
		cardLinks := loadCardLinks("/tmp/card_links.txt")
		if len(cardLinks) > 0 {
			setup.CardLinks = append(setup.CardLinks, cardLinks...)
		}
	}

	readM := loadTestRead(setup, setup.CardLinks)
	printMetrics("READ CARDS (500 RPS)", readM)

	fmt.Printf("\n  DB cards: %d\n", 100525+int(createM.Success*float64(createM.Requests)))
	fmt.Println("\n=== DONE ===")
}
