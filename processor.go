package processor

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"fw/pkg/config"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
)

// Read the file under the given "filename" from the given "offset" to its end.
// When logrotate truncates the file, it restart reading from the beginning.
// Updates the given "offset" variable to reflects the new cursor position.
func ProcessFileChange(config config.Config, path string, offset *int64) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	fileStat, _ := os.Stat(path)
	if fileStat.Size() < *offset {
		*offset = 0
		log.Printf("Read offset reinitialization, '%s' was truncated by logrotate.\n", path)
	} else {
		_, _ = file.Seek(*offset, io.SeekStart)
	}
	log.Printf("Reading '%s' from offset %d to the end.\n", path, *offset)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		*offset += int64(len(line) + len("\n"))
		matchAndProcessIP(config, path, line)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

// Check a given line of a file against regular expressions.
// Add the IP to the appropriate set (Accept or Reject)
// Applies a mask to reject multiple IPs at once depending on configuration
func matchAndProcessIP(config config.Config, path string, line string) {
	for _, re := range config.AcceptRegexps[path] {
		if match := re.FindStringSubmatch(line); match != nil {
			ip := match[1]
			addIPToSet(config, "accept", ip+"/32")
			return
		}
	}
	for _, re := range config.RejectRegexps[path] {
		if match := re.FindStringSubmatch(line); match != nil {
			ip := net.ParseIP(match[1])
			num := binary.BigEndian.Uint32(ip.To4())
			num = num & generateBitSequence(config.RejectBitLen)
			ip = net.IPv4(byte(num>>24), byte(num>>16), byte(num>>8), byte(num))
			addIPToSet(config, "reject", fmt.Sprintf("%s/%d", ip.String(), config.RejectBitLen))
			return
		}
	}
}

// Provide a 32-bit number with all bits set to 0 except the leftmost "n"
func generateBitSequence(n int) uint32 {
	if n < 0 || n > 32 {
		panic("N must be between 0 and 32")
	}
	mask := uint32((1 << n) - 1)
	return mask << (32 - n)
}

// Execute the "ipset" command which accepts or rejects an IP address
func addIPToSet(config config.Config, setName string, ip string) {
	fwName := config.Name + "-" + setName
	cmd := exec.Command("ipset", "-exist", "add", fwName, ip)
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error adding IP to set: %v", err)
	}
	log.Printf("- %sing %s.\n", setName, ip)
}
