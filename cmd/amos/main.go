package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func main() {
	start := time.Now()

	cmd := exec.Command("./emos", os.Args[1])
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("unable to run emos: %v", err)
	}
	out := bytes.NewBuffer(output)

	elapsed := time.Since(start)
	fmt.Fprintln(os.Stderr, "execution took: ", elapsed)

	fmt.Printf(`<?xml version="1.0"?>`)
	fmt.Printf(`<items>`)
	scanner := bufio.NewScanner(out)
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " - ")

		wg.Add(1)
		go func(name, url string) {
			defer wg.Done()
			_, err = os.Stat("/tmp/emos")
			if err != nil {
				os.Mkdir("/tmp/emos", 0755)
			}
			target := fmt.Sprintf("/tmp/emos/%s.jpeg", parts[0])
			_, err = os.Stat(target)
			if err != nil {
				resp, err := http.Get(url)
				if err != nil {
					fmt.Fprintln(os.Stderr, "failed downloading image:", err)
				}

				f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, 0755)
				if err != nil {
					fmt.Fprintln(os.Stderr, "failed downloading image:", err)
					return
				}
				defer f.Close()
				io.Copy(f, resp.Body)
			}

			mu.Lock()
			fmt.Printf(`<item uid="%s" arg="%s" valid="yes">
            <title>%s</title>
            <icon>%s</icon>
        </item>`, name, url, name, target)
			mu.Unlock()
		}(parts[0], parts[1])
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "failed reading input:", err)
	}
	wg.Wait()
	fmt.Printf(`</items>`)
}
