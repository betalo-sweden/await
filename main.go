package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"context"
)

const retryDelay = 500 * time.Millisecond

type resource interface {
	fmt.Stringer
	Await(context.Context) error
}

type unavailableError struct {
	Reason error
}

// Error implements the error interface.
func (e *unavailableError) Error() string {
	return e.Reason.Error()
}

func main() {
	var (
		forceFlag   = flag.Bool("f", false, "Force running the command even after giving up")
		timeoutFlag = flag.Duration("t", 1*time.Minute, "Timeout duration before giving up")
		verboseFlag = flag.Bool("v", false, "Set verbose output")
		quietFlag   = flag.Bool("q", false, "Set quiet mode")
	)
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: await [options...] <res>... [ -- <cmd>]")
		fmt.Fprintln(os.Stderr, "Await availability of resources.")
		fmt.Fprintln(os.Stderr)
		flag.PrintDefaults()
	}
	flag.Parse()

	logLevel := errorLevel
	switch {
	case *quietFlag:
		logLevel = silentLevel
	case *verboseFlag:
		logLevel = infoLevel
	}
	log := NewLogger(logLevel)

	resArgs, cmdArgs := splitArgs(flag.Args())
	ress, err := parseResources(resArgs)
	if err != nil {
		log.Fatalln("Error: failed to parse resources: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeoutFlag)
	go func() {
		for i := 0; i < len(ress); {
			select {
			case <-ctx.Done(): // Exceeded timeout
				return
			default:
				res, err := identifyResource(ress[i])
				if err != nil { // Permanent error
					log.Fatalf("Error: %v", err)
				}

				log.Infof("Awaiting resource: %s", res)
				if err := res.Await(ctx); err != nil {
					if e, ok := err.(*unavailableError); ok { // transient error
						log.Infof("Resource unavailable: %v", e)
					} else { // Maybe transient error
						log.Errorf("Error: failed to await resource: %v", err)
					}
					time.Sleep(retryDelay)
				} else {
					i++ // Next resource
				}
			}
		}

		cancel() // All resources are available
	}()

	switch <-ctx.Done(); ctx.Err() {
	case context.Canceled:
		log.Infoln("All resources available")
	case context.DeadlineExceeded:
		log.Infoln("Timeout exceeded")
		if !*forceFlag {
			os.Exit(1)
		}
	}

	if len(cmdArgs) > 0 {
		log.Infof("Runnning command: %v", cmdArgs)
		if err := execCmd(cmdArgs); err != nil {
			log.Fatalf("Error: failed to execute command: %v", err)
		}
	}
}

func splitArgs(args []string) ([]string, []string) {
	for i, a := range args {
		if a == "--" {
			return args[0:i], args[i+1:]
		}
	}
	return args, []string{}
}

func parseResources(urlArgs []string) ([]url.URL, error) {
	var urls []url.URL
	for _, urlArg := range urlArgs {
		// Leveraging the fact the Go's URL parser matches e.g. `curl -s
		// http://example.com` as url.Path instead of throwing an error.
		u, err := url.Parse(urlArg)
		if err != nil {
			return urls, err
		}
		urls = append(urls, *u)
	}
	return urls, nil
}

func identifyResource(u url.URL) (resource, error) {
	switch u.Scheme {
	case "http", "https":
		return &httpResource{u}, nil
	case "ws", "wss":
		return &websocketResource{u}, nil
	case "tcp", "tcp4", "tcp6":
		return &tcpResource{u}, nil
	case "file":
		return &fileResource{u}, nil
	case "postgres":
		return &postgresqlResource{u}, nil
	case "mysql":
		return &mysqlResource{u}, nil
	case "":
		return &commandResource{u}, nil
	default:
		return nil, fmt.Errorf("unsupported resource scheme: %v", u.Scheme)
	}
}

func execCmd(cmdArgs []string) error {
	path, err := exec.LookPath(cmdArgs[0])
	if err != nil {
		return err
	}
	return syscall.Exec(path, cmdArgs, os.Environ())
}

func parseTags(tag string) map[string]string {
	tags := map[string]string{}
	tagParts := strings.Split(tag, "&")
	for _, t := range tagParts {
		kv := strings.SplitN(t, "=", 2)
		k := kv[0]
		if len(kv) == 1 {
			tags[k] = ""
		} else {
			tags[k] = kv[1]
		}
	}
	return tags
}
