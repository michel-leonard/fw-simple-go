
# Simple Firewall

## Overview

This project monitors specified files for changes and processes IP addresses according to configuration rules. The software configures **ipset** and **iptables** if the configuration does not already exist. When a regular expression in the configuration file matches a new line in the log file, it captures the IP address and updates the ipset accordingly to :
- **accept** an IP
- **reject** an IP using, depending on the configuration, a bitmask and a timeout

The default configuration works with IPv4 on a Debian server, iptables and ipset must be installed beforehand.

## Configuration

Configuration is provided in a JSON file with the following structure:

````json{
  "firewall-name": "fw-simple",
  "path-iptables-ipset": "/sbin:/usr/sbin",
  "reject-timeout": 7200,
  "reject-bitmask-length": 24,
  "files": {
    "/var/log/log-file-1": {
      "accept": [ "regex accepting (__IP__) in file 1" ],
      "reject": [  "regex rejecting (__IP__) in file 1"]
    },
    "/var/log/log-file-2": {
      "accept": [ "regex accepting (__IP__) in file 2" ],
      "reject": [  "regex rejecting (__IP__) in file 2"]
    }
  }
}
