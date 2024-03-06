package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"text/tabwriter"
	"time"

	"log/slog"
)

func main() {
	var (
		file   string
		asJson bool
		notify bool
	)

	flag.StringVar(&file, "s", "./services.json", "path to services file")
	flag.StringVar(&file, "services-file", "./services.json", "path to services file")
	flag.BoolVar(&asJson, "j", false, "whether to display output as JSON")
	flag.BoolVar(&asJson, "json", false, "whether to display output as JSON")
	flag.BoolVar(&notify, "notify", false, "whether to display service issues as notifications")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	data, err := os.ReadFile(file)
	if err != nil {
		logger.Error("unable to open services file",
			"file", file,
			"error", err)
		os.Exit(1)
	}

	// service represents a service definition from the configuration file.
	// It is also used for output.
	type service struct {
		Name    string            `json:"name"`
		URL     string            `json:"url"`
		Headers map[string]string `json:"headers,omitempty"`
		Status  int               `json:"status"`
	}

	// Read contents of services file.
	var services []*service
	err = json.Unmarshal(data, &services)
	if err != nil {
		logger.Error("unable to parse services file",
			"file", file,
			"error", err)
		os.Exit(1)
	}

	// Attempt to get all specfied URLs.
	var wg sync.WaitGroup
	wg.Add(len(services))
	for _, svc := range services {
		go func(s *service) {
			defer wg.Done()
			client := http.Client{
				Timeout: 2 * time.Second,
			}
			req, err := http.NewRequest(http.MethodGet, s.URL, nil)
			if err != nil {
				logger.Error("error creating new request",
					"url", s.URL,
					"error", err)
				return
			}
			if s.Headers != nil {
				for k, v := range s.Headers {
					req.Header.Add(k, v)
				}
			}
			resp, err := client.Do(req)
			if err != nil {
				logger.Error("error getting URL",
					"url", s.URL,
					"error", err)
				// Server error response OK for now; just need
				// to indicate a problem.
				s.Status = http.StatusServiceUnavailable
				return
			}
			s.Status = resp.StatusCode
		}(svc)
	}
	wg.Wait()

	// Output results.
	switch {
	case asJson:
		b, err := json.Marshal(services)
		if err != nil {
			logger.Error("unable to marshal responses", "error", err)
			os.Exit(1)
		}
		fmt.Printf("%s", string(b))
	case notify:
		for _, s := range services {
			if s.Status != http.StatusOK {
				n := fmt.Sprintf("display notification \"%s\" with title \"%s\"",
					http.StatusText(s.Status), s.Name)
				err := exec.Command("osascript", "-e", n).Run()
				if err != nil {
					logger.Error("could not execute 'osascript': %s\n",
						"error", err)
					os.Exit(1)
				}
			}
		}
	default:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.StripEscape)
		fmt.Fprintln(w, "SERVICE\tURL\tSTATUS")
		for _, s := range services {
			fmt.Fprintf(w, "%s\t%s\t%s\n", s.Name, s.URL, http.StatusText(s.Status))
		}
		w.Flush()
	}
}
