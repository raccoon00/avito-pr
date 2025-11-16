package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Team struct {
	Name    string       `json:"team_name"`
	Members []TeamMember `json:"members"`
}

type TeamMember struct {
	Id       string `json:"user_id"`
	Name     string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type CreatePullRequestRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

type MergePullRequestRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type StressTestResult struct {
	TestName      string
	TotalRequests int32
	SuccessCount  int32
	ErrorCount    int32
	TotalTime     time.Duration
	RPS           float64
	AvgLatency    time.Duration
	P95Latency    time.Duration
	P99Latency    time.Duration
}

type TestConfig struct {
	BaseURL      string
	Duration     time.Duration
	Concurrency  int
	TeamsCount   int
	UsersPerTeam int
}

type LatencyTracker struct {
	latencies []time.Duration
	mu        sync.Mutex
}

func NewLatencyTracker() *LatencyTracker {
	return &LatencyTracker{
		latencies: make([]time.Duration, 0),
	}
}

func (lt *LatencyTracker) Add(latency time.Duration) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	lt.latencies = append(lt.latencies, latency)
}

func (lt *LatencyTracker) CalculatePercentile(percentile float64) time.Duration {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	if len(lt.latencies) == 0 {
		return 0
	}

	index := int(float64(len(lt.latencies)-1) * percentile / 100.0)
	index = max(index, 0)
	if index >= len(lt.latencies) {
		index = len(lt.latencies) - 1
	}

	return lt.latencies[index]
}

func initializeTestDatabase(config TestConfig) {
	log.Printf("Initializing test database with %d teams and %d users...", config.TeamsCount, config.TeamsCount*config.UsersPerTeam)

	for i := range config.TeamsCount {
		teamNum := i + 1
		teamName := "stress-team-" + strconv.Itoa(teamNum)
		members := make([]TeamMember, 0, config.UsersPerTeam)

		for j := range config.UsersPerTeam {
			userNum := j + 1
			userID := fmt.Sprintf("stress-user-%d-%d", teamNum, userNum)
			userName := fmt.Sprintf("Stress User %d-%d", teamNum, userNum)

			// Делаем первые 80% юзеров активными
			isActive := userNum <= int(float64(config.UsersPerTeam)*0.8)

			members = append(members, TeamMember{
				Id:       userID,
				Name:     userName,
				IsActive: isActive,
			})
		}

		team := Team{
			Name:    teamName,
			Members: members,
		}

		teamJSON, _ := json.Marshal(team)
		resp, err := http.Post(config.BaseURL+"/team/add", "application/json", bytes.NewBuffer(teamJSON))
		if err == nil && resp.StatusCode == http.StatusCreated {
			resp.Body.Close()
		}

		if teamNum%5 == 0 || teamNum == config.TeamsCount {
			log.Printf("Created %d/%d teams...", teamNum, config.TeamsCount)
		}
	}

	log.Println("Database initialization completed!")
}

func stressTestCreatePR(config TestConfig) StressTestResult {
	var totalRequests, successCount, errorCount int32
	latencyTracker := NewLatencyTracker()

	startTime := time.Now()
	endTime := startTime.Add(config.Duration)

	var wg sync.WaitGroup
	for i := range config.Concurrency {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			prCounter := 0
			for time.Now().Before(endTime) {
				prID := fmt.Sprintf("stress-pr-%d-%d", workerID, prCounter)

				teamNum := rand.Intn(config.TeamsCount) + 1
				userNum := rand.Intn(int(float64(config.UsersPerTeam)*0.8)) + 1
				authorID := fmt.Sprintf("stress-user-%d-%d", teamNum, userNum)

				createPRReq := CreatePullRequestRequest{
					PullRequestID:   prID,
					PullRequestName: "Stress Test PR",
					AuthorID:        authorID,
				}

				reqJSON, _ := json.Marshal(createPRReq)

				requestStart := time.Now()
				resp, err := http.Post(config.BaseURL+"/pullRequest/create", "application/json", bytes.NewBuffer(reqJSON))
				latency := time.Since(requestStart)

				atomic.AddInt32(&totalRequests, 1)
				latencyTracker.Add(latency)

				if err != nil || resp.StatusCode != http.StatusCreated {
					atomic.AddInt32(&errorCount, 1)
					if resp != nil {
						resp.Body.Close()
					}
				} else {
					atomic.AddInt32(&successCount, 1)
					resp.Body.Close()
				}

				prCounter++
				time.Sleep(10 * time.Millisecond) // Small delay to avoid overwhelming
			}
		}(i)
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	return StressTestResult{
		TestName:      "Create PR",
		TotalRequests: totalRequests,
		SuccessCount:  successCount,
		ErrorCount:    errorCount,
		TotalTime:     totalTime,
		RPS:           float64(totalRequests) / totalTime.Seconds(),
		AvgLatency:    latencyTracker.CalculatePercentile(50),
		P95Latency:    latencyTracker.CalculatePercentile(95),
		P99Latency:    latencyTracker.CalculatePercentile(99),
	}
}

func stressTestReassign(config TestConfig) StressTestResult {
	var totalRequests, successCount, errorCount int32
	latencyTracker := NewLatencyTracker()

	for i := range 5 {
		teamNum := rand.Intn(config.TeamsCount) + 1
		userNum := rand.Intn(int(float64(config.UsersPerTeam)*0.8)) + 1
		authorID := fmt.Sprintf("stress-user-%d-%d", teamNum, userNum)

		createPRReq := CreatePullRequestRequest{
			PullRequestID:   fmt.Sprintf("stress-reassign-pr-%d", i),
			PullRequestName: "Stress Reassign PR",
			AuthorID:        authorID,
		}
		reqJSON, _ := json.Marshal(createPRReq)
		resp, err := http.Post(config.BaseURL+"/pullRequest/create", "application/json", bytes.NewBuffer(reqJSON))
		if err == nil && resp.StatusCode == http.StatusCreated {
			resp.Body.Close()
		}
	}

	startTime := time.Now()
	endTime := startTime.Add(config.Duration)

	var wg sync.WaitGroup
	for i := range config.Concurrency {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			requestCounter := 0
			for time.Now().Before(endTime) {
				prNum := rand.Intn(5)
				teamNum := rand.Intn(config.TeamsCount) + 1
				oldUserNum := rand.Intn(int(float64(config.UsersPerTeam)*0.8)-1) + 2 // Start from 2 to avoid author

				reassignReq := ReassignReviewerRequest{
					PullRequestID: fmt.Sprintf("stress-reassign-pr-%d", prNum),
					OldUserID:     fmt.Sprintf("stress-user-%d-%d", teamNum, oldUserNum),
				}

				reqJSON, _ := json.Marshal(reassignReq)

				requestStart := time.Now()
				resp, err := http.Post(config.BaseURL+"/pullRequest/reassign", "application/json", bytes.NewBuffer(reqJSON))
				latency := time.Since(requestStart)

				atomic.AddInt32(&totalRequests, 1)
				latencyTracker.Add(latency)

				if err != nil || (resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict) {
					atomic.AddInt32(&errorCount, 1)
					if resp != nil {
						resp.Body.Close()
					}
				} else {
					atomic.AddInt32(&successCount, 1)
					resp.Body.Close()
				}

				requestCounter++
				time.Sleep(50 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	return StressTestResult{
		TestName:      "Reassign Reviewer",
		TotalRequests: totalRequests,
		SuccessCount:  successCount,
		ErrorCount:    errorCount,
		TotalTime:     totalTime,
		RPS:           float64(totalRequests) / totalTime.Seconds(),
		AvgLatency:    latencyTracker.CalculatePercentile(50),
		P95Latency:    latencyTracker.CalculatePercentile(95),
		P99Latency:    latencyTracker.CalculatePercentile(99),
	}
}

func stressTestGetTeam(config TestConfig) StressTestResult {
	var totalRequests, successCount, errorCount int32
	latencyTracker := NewLatencyTracker()

	startTime := time.Now()
	endTime := startTime.Add(config.Duration)

	var wg sync.WaitGroup
	for i := range config.Concurrency {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for time.Now().Before(endTime) {
				teamNum := rand.Intn(config.TeamsCount) + 1
				requestStart := time.Now()
				resp, err := http.Get(config.BaseURL + "/team/get?team_name=stress-team-" + strconv.Itoa(teamNum))
				latency := time.Since(requestStart)

				atomic.AddInt32(&totalRequests, 1)
				latencyTracker.Add(latency)

				if err != nil || resp.StatusCode != http.StatusOK {
					atomic.AddInt32(&errorCount, 1)
					if resp != nil {
						resp.Body.Close()
					}
				} else {
					atomic.AddInt32(&successCount, 1)
					resp.Body.Close()
				}

				time.Sleep(5 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	return StressTestResult{
		TestName:      "Get Team",
		TotalRequests: totalRequests,
		SuccessCount:  successCount,
		ErrorCount:    errorCount,
		TotalTime:     totalTime,
		RPS:           float64(totalRequests) / totalTime.Seconds(),
		AvgLatency:    latencyTracker.CalculatePercentile(50),
		P95Latency:    latencyTracker.CalculatePercentile(95),
		P99Latency:    latencyTracker.CalculatePercentile(99),
	}
}

func printResults(results []StressTestResult) {
	fmt.Println("\nStress Test Results:")
	fmt.Printf("%-20s %12s %10s %8s %8s %12s %12s %12s\n",
		"Test", "Requests", "Success", "Errors", "RPS", "Avg Latency", "P95 Latency", "P99 Latency")
	fmt.Println("--------------------------------------------------------------------------------------------")

	for _, result := range results {
		successRate := float64(result.SuccessCount) / float64(result.TotalRequests) * 100
		fmt.Printf("%-20s %12d %9.1f%% %8d %8.1f %12s %12s %12s\n",
			result.TestName,
			result.TotalRequests,
			successRate,
			result.ErrorCount,
			result.RPS,
			result.AvgLatency.Round(time.Millisecond),
			result.P95Latency.Round(time.Millisecond),
			result.P99Latency.Round(time.Millisecond))
	}
}

func main() {
	// Parse command line arguments
	baseURL := flag.String("url", "http://localhost:8080", "Base URL of the service")
	testDuration := flag.Duration("duration", 30*time.Second, "Duration of each stress test")
	concurrency := flag.Int("concurrency", 10, "Number of concurrent workers")
	teamsCount := flag.Int("teams", 20, "Number of teams to create")
	usersPerTeam := flag.Int("users", 10, "Number of users per team")
	flag.Parse()

	log.Println("Starting stress tests...")
	log.Println("Make sure the service is running on", *baseURL)

	// Create test configuration
	config := TestConfig{
		BaseURL:      *baseURL,
		Duration:     *testDuration,
		Concurrency:  *concurrency,
		TeamsCount:   *teamsCount,
		UsersPerTeam: *usersPerTeam,
	}

	// Initialize test database first
	initializeTestDatabase(config)

	results := []StressTestResult{}

	log.Printf("Running Create PR stress test for %v with %d concurrent workers...", *testDuration, *concurrency)
	results = append(results, stressTestCreatePR(config))

	log.Printf("Running Reassign Reviewer stress test for %v with %d concurrent workers...", *testDuration, *concurrency)
	results = append(results, stressTestReassign(config))

	log.Printf("Running Get Team stress test for %v with %d concurrent workers...", *testDuration, *concurrency)
	results = append(results, stressTestGetTeam(config))

	printResults(results)

	// Check if requirements are met
	log.Println("\nRequirements Check")
	for _, result := range results {
		if result.AvgLatency > 300*time.Millisecond {
			log.Printf("NOT OK: %s: Average latency %.0fms exceeds 300ms requirement",
				result.TestName, result.AvgLatency.Seconds()*1000)
		} else {
			log.Printf("OK: %s: Average latency %.0fms meets 300ms requirement",
				result.TestName, result.AvgLatency.Seconds()*1000)
		}

		successRate := float64(result.SuccessCount) / float64(result.TotalRequests) * 100
		if successRate < 99.9 {
			log.Printf("NOT OK: %s: Success rate %.2f%% below 99.9%% requirement", result.TestName, successRate)
		} else {
			log.Printf("OK: %s: Success rate %.2f%% meets 99.9%% requirement", result.TestName, successRate)
		}
	}
}
