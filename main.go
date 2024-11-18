package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

func banner() {
	page := `
┬ ┬┌┬┐┌┬┐┌─┐┌─┐┬ ┬┌─┐┌─┐┬┌─
├─┤ │  │ ├─┘│  ├─┤├┤ │  ├┴┐
┴ ┴ ┴  ┴ ┴  └─┘┴ ┴└─┘└─┘┴ ┴                                    
Author: Al0neme
`
	fmt.Println(page)
}

func readtarget(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	targets := make([]string, 0)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "://") {
			targets = append(targets, line)
		} else {
			targets = append(targets, "https://"+line)
			targets = append(targets, "http://"+line)
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return targets
}

func randomuseragent() string {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Safari/605.1.15",
		"Mozilla/5.0 (Linux; Android 10; Pixel 3 XL) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Mobile Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Gecko/20100101 Firefox/89.0",
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	randomIndex := r.Intn(len(userAgents))
	return userAgents[randomIndex]
}

func checklive(url string) (islive bool, result map[string]string) {
	islive = true
	result = make(map[string]string)
	result["url"] = url

	host := ""
	if strings.Contains(url, "://") {
		host = strings.Split(url, "://")[1]
	} else {
		host = url
	}
	if idx := strings.Index(host, "/"); idx != -1 {
		host = host[:idx]
	}
	host = strings.Split(host, ":")[0]

	ip := ""
	ips, err := net.LookupIP(host)
	if err != nil {

		islive = false
		return
	} else {
		for _, i := range ips {
			ip = i.String()
		}
	}
	result["ip"] = ip

	client := &http.Client{
		Timeout: 7 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", randomuseragent())
	req.Header.Set("Origin", url)
	req.Header.Set("Referer", url)

	resp, err := client.Do(req)
	if err != nil {
		islive = false

		return
	}
	defer resp.Body.Close()

	status := fmt.Sprintf("%d", resp.StatusCode)
	result["status"] = status

	location := resp.Header.Get("Location")

	if location != "" {
		result["location"] = location
	}

	length := fmt.Sprintf("%d", resp.ContentLength)
	result["length"] = length

	scanner := bufio.NewScanner(resp.Body)
	var body strings.Builder
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		body.WriteString(line)
	}
	if err := scanner.Err(); err != nil {

	}
	title := "[]"
	re := regexp.MustCompile("(?i)<title>(.*?)</title>")
	matches := re.FindStringSubmatch(body.String())
	if len(matches) > 1 {
		title = "[" + matches[1] + "]"
		result["title"] = title

	}

	return islive, result
}

func getresult(result map[string]string) string {
	outresult := ""
	for k, v := range result {
		outresult = outresult + k + " : " + v + "\n"
	}
	fmt.Println(outresult)
	return outresult
}

func saveresult(result string) {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	file, err := os.OpenFile("result.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	if _, err := file.WriteString(result + "\n--------------------\n"); err != nil {
		return
	}
}

func syncrun(targets []string, concurrencylimit int) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrencylimit)
	for _, url := range targets {
		wg.Add(1)
		sem <- struct{}{}
		go func(u string) {
			defer wg.Done()
			islive, result := checklive(url)
			if islive {
				saveresult(getresult(result))
			}
			<-sem
		}(url)
	}
	wg.Wait()
}

func main() {
	banner()
	filename := flag.String("f", "", "file include targets")
	thread := flag.Int("t", 10, "number of thread, default 10")

	flag.Parse()
	if *filename == "" {
		fmt.Println("usage: httpcheck -f target.txt -t 5")
		return
	}
	targets := readtarget(*filename)
	if *thread != 10 {
		syncrun(targets, *thread)

	} else {
		syncrun(targets, 10)
	}
	fmt.Println("all task done, the live target saved to result.txt")

}
