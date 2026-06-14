// Command gencerts writes a development mTLS trust chain to disk: ca.crt,
// panel.crt/panel.key (the client the panel uses to dial nodes), and
// node.crt/node.key (the server each node agent presents). For production, use
// your own CA instead.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vortexui/vortexui/internal/pki"
)

func main() {
	out := flag.String("out", "deploy/certs", "output directory")
	sans := flag.String("san", "localhost,127.0.0.1", "comma-separated node SANs (hostnames/IPs)")
	flag.Parse()

	if err := run(*out, strings.Split(*sans, ",")); err != nil {
		fmt.Fprintln(os.Stderr, "gencerts:", err)
		os.Exit(1)
	}
}

func run(out string, sans []string) error {
	chain, err := pki.Generate(sans)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}

	files := map[string][]byte{
		"ca.crt":    chain.CA.CertPEM,
		"ca.key":    chain.CA.KeyPEM,
		"node.crt":  chain.Server.CertPEM,
		"node.key":  chain.Server.KeyPEM,
		"panel.crt": chain.Client.CertPEM,
		"panel.key": chain.Client.KeyPEM,
	}
	for name, data := range files {
		mode := os.FileMode(0o644)
		if strings.HasSuffix(name, ".key") {
			mode = 0o600 // private keys are owner-only
		}
		if err := os.WriteFile(filepath.Join(out, name), data, mode); err != nil {
			return err
		}
	}
	fmt.Printf("wrote dev mTLS chain to %s (node SANs: %s)\n", out, strings.Join(sans, ", "))
	return nil
}
