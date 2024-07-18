package main

import (
	"fw/pkg/config"
	"fw/pkg/watcher"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// Shortcut : the term "__IP__" will be replaced by this string in regular expressions
const IPRegexString = "(?:(?:25[0-5]|(2[0-4]|1\\d|[1-9]|)\\d)\\.?\\b){4}"

func main() {
	cfg := config.ReadConfig("config.json")
	initRegexps(&cfg)
	setupIPSetsAndTables(cfg)
	var wg sync.WaitGroup
	for path := range cfg.Files {
		wg.Add(1)
		go watcher.WatchFile(cfg, path, &wg)
	}
	wg.Wait()
}

// Fill the provided config accordingly to the provided literal regular expressions
func initRegexps(config *config.Config) {
	config.AcceptRegexps = make(map[string][]*regexp.Regexp)
	config.RejectRegexps = make(map[string][]*regexp.Regexp)
	for fn, fc := range config.Files {
		for _, pattern := range fc.Accept {
			config.AcceptRegexps[fn] = append(config.AcceptRegexps[fn], prepareRegex(pattern))
		}
		for _, pattern := range fc.Reject {
			config.RejectRegexps[fn] = append(config.RejectRegexps[fn], prepareRegex(pattern))
		}
	}
	log.Printf("Regular Expression compiled.\n")
}

// Return the compiled version of a single literal regular expression
func prepareRegex(pattern string) *regexp.Regexp {
	re, err := regexp.Compile(strings.Replace(pattern, "__IP__", IPRegexString, 1))
	if err != nil {
		log.Fatalf("Error compiling regexp %v", err)
	}
	return re
}

// Initialize the system by inserting 2 new "iptables" rules, and creating 2 new "ipset" sets
// If a path is provided in config, it will be used to access "iptables" and "ipset"
// This function is useful when the system restart, and do not affect it otherwise.
func setupIPSetsAndTables(config config.Config) {
	if len(config.Path) != 0 {
		path := os.Getenv("PATH")
		newPath := strings.TrimRight(config.Path, ":") + ":" + path
		_ = os.Setenv("PATH", newPath)
		log.Printf("PATH for iptables/ipset is now '%s' accordingly to config.\n", newPath)
	}
	cmd := exec.Command("iptables-save")
	out, _ := cmd.CombinedOutput()
	if strings.Contains(string(out), config.Name+"-"+"accept") {
		log.Printf("iptables/ipset initialization ignored (already exists).\n")
	} else {
		cmd := exec.Command("ipset", "create", config.Name+"-"+"reject", "hash:net", "timeout", strconv.Itoa(config.RejectTimeout))
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
		cmd = exec.Command("iptables", "-A", "INPUT", "-m", "set", "--match-set", config.Name+"-"+"reject", "src", "-j", "DROP")
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
		cmd = exec.Command("ipset", "create", config.Name+"-"+"accept", "hash:net")
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
		cmd = exec.Command("iptables", "-A", "INPUT", "-m", "set", "--match-set", config.Name+"-"+"accept", "src", "-j", "ACCEPT")
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
		log.Printf("iptables/ipset initialization using '%s-*' namespace completed.\n", config.Name)
	}
}
