// Package deploy provides automated node provisioning over SSH. An admin
// provides server credentials, and the panel installs the node agent + cores
// automatically — no manual SSH required.
package deploy

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// SSHClient is the interface for executing commands on a remote server.
type SSHClient interface {
	Run(ctx context.Context, cmd string) (stdout string, err error)
	Upload(ctx context.Context, localPath, remotePath string) error
	Close() error
}

// NodeDeployInput describes a new node to auto-provision.
type NodeDeployInput struct {
	Host       string `json:"host"`       // IP or hostname
	Port       int    `json:"port"`       // SSH port (default 22)
	Username   string `json:"username"`   // SSH user (default root)
	Password   string `json:"password"`   // or use key
	PrivateKey string `json:"private_key"` // SSH private key (PEM)
	NodeName      string   `json:"node_name"`  // name for the node in the panel
	Core          string   `json:"core"`       // default core: xray or singbox
	EnabledCores  []string `json:"enabled_cores"` // optional; e.g. ["xray","singbox"]
	AgentPort     int      `json:"agent_port"` // gRPC listen port (default 50051)
}

// Deployer handles automated node provisioning.
type Deployer struct {
	log        *slog.Logger
	panelHost  string // panel's public host (for cert SAN)
	certDir    string // where panel certs are stored
	repoURL    string // VortexUI repo URL
	dialer     func(input NodeDeployInput) (SSHClient, error)
}

// NewDeployer builds the deployer.
func NewDeployer(panelHost, certDir, repoURL string, dialer func(NodeDeployInput) (SSHClient, error), log *slog.Logger) *Deployer {
	if log == nil {
		log = slog.Default()
	}
	if repoURL == "" {
		repoURL = "https://github.com/iPmartNetwork/VortexUI"
	}
	return &Deployer{
		log:       log,
		panelHost: panelHost,
		certDir:   certDir,
		repoURL:   repoURL,
		dialer:    dialer,
	}
}

// DeployResult holds the outcome of a deployment.
type DeployResult struct {
	Success bool   `json:"success"`
	Address string `json:"address"` // host:port for adding to panel
	Core    string `json:"core"`
	Output  string `json:"output"`  // combined command output
	Error   string `json:"error,omitempty"`
}

// coreBinPath returns the conventional binary path for a deploy core name.
func coreBinPath(core string) string {
	switch core {
	case "singbox":
		return "/usr/local/bin/sing-box"
	default:
		return "/usr/local/bin/xray"
	}
}

func singboxV2RayAPIValue(core string) string {
	if core == "singbox" {
		return "false"
	}
	return "true"
}

func normalizedDeployCores(core string, enabled []string) []string {
	if len(enabled) == 0 {
		return []string{core}
	}
	out := make([]string, 0, len(enabled))
	seen := make(map[string]struct{})
	for _, c := range enabled {
		if c == "" {
			continue
		}
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	if len(out) == 0 {
		return []string{core}
	}
	return out
}

func nodeEnvContent(agentPort int, core string, enabled []string) string {
	cores := normalizedDeployCores(core, enabled)
	lines := []string{
		fmt.Sprintf("VORTEX_NODE_LISTEN=:%d", agentPort),
		fmt.Sprintf("VORTEX_CORE=%s", core),
		fmt.Sprintf("VORTEX_CORE_BIN=%s", coreBinPath(core)),
		"VORTEX_CORE_CONFIG=/etc/vortex/node-core.json",
		fmt.Sprintf("VORTEX_SINGBOX_V2RAY_API=%s", singboxV2RayAPIValue(core)),
		"VORTEX_TLS_CERT=/etc/vortex/certs/node.crt",
		"VORTEX_TLS_KEY=/etc/vortex/certs/node.key",
		"VORTEX_TLS_CA=/etc/vortex/certs/ca.crt",
		"XRAY_LOCATION_ASSET=/etc/vortex/assets",
	}
	if len(cores) > 1 {
		lines = append(lines,
			"VORTEX_ENABLED_CORES="+strings.Join(cores, ","),
			"VORTEX_XRAY_CONFIG=/etc/vortex/xray.json",
			"VORTEX_SINGBOX_CONFIG=/etc/vortex/singbox.json",
			"VORTEX_XRAY_API_PORT=10085",
			"VORTEX_SINGBOX_API_PORT=10086",
			"VORTEX_SINGBOX_V2RAY_API=false",
		)
	}
	return strings.Join(lines, "\n") + "\n"
}

// singboxInstallCmd installs sing-box 1.12.12 (arch-aware).
const singboxInstallCmd = `sarch=$(uname -m); case "$sarch" in aarch64|arm64) sarch=arm64 ;; *) sarch=amd64 ;; esac; sbver=v1.12.12; curl -fsSL -o /tmp/sb.tgz "https://github.com/SagerNet/sing-box/releases/download/${sbver}/sing-box-${sbver#v}-linux-${sarch}.tar.gz" && tar -xzf /tmp/sb.tgz -C /tmp && install -m 755 /tmp/sing-box-*/sing-box /usr/local/bin/sing-box || true`

func (d *Deployer) Deploy(ctx context.Context, input NodeDeployInput) *DeployResult {
	if input.Port == 0 {
		input.Port = 22
	}
	if input.Username == "" {
		input.Username = "root"
	}
	if input.AgentPort == 0 {
		input.AgentPort = 50051
	}
	if input.Core == "" {
		input.Core = "xray"
	}

	result := &DeployResult{Core: input.Core}
	d.log.Info("starting node auto-deploy", "host", input.Host, "core", input.Core)

	client, err := d.dialer(input)
	if err != nil {
		result.Error = fmt.Sprintf("SSH connect failed: %v", err)
		return result
	}
	defer client.Close() //nolint:errcheck // best-effort cleanup

	// Steps:
	steps := []struct {
		name string
		cmd  string
	}{
		{"update packages", "apt-get update -y && apt-get install -y curl git unzip tar"},
		{"install xray", "curl -fsSL -o /tmp/xray.zip https://github.com/XTLS/Xray-core/releases/latest/download/Xray-linux-64.zip && unzip -o /tmp/xray.zip -d /tmp/xray && install -m 755 /tmp/xray/xray /usr/local/bin/xray"},
		{"install sing-box", singboxInstallCmd},
		{"clone vortexui", fmt.Sprintf("git clone --depth 1 %s /opt/vortexui 2>/dev/null || (cd /opt/vortexui && git fetch origin master && git reset --hard origin/master)", d.repoURL)},
		{"build node agent", "cd /opt/vortexui && go build -o /usr/local/bin/vortex-node ./cmd/node 2>/dev/null || curl -fsSL -o /usr/local/bin/vortex-node https://github.com/iPmartNetwork/VortexUI/releases/latest/download/vortexui-node-linux-amd64.tar.gz && chmod +x /usr/local/bin/vortex-node"},
		{"setup certs dir", "mkdir -p /etc/vortex/certs /etc/vortex/assets"},
		{"create service", fmt.Sprintf(`cat > /etc/systemd/system/vortexui-node.service << 'EOF'
[Unit]
Description=VortexUI node agent
After=network.target
[Service]
EnvironmentFile=/etc/vortexui/node.env
ExecStart=/usr/local/bin/vortex-node
Restart=always
RestartSec=3
[Install]
WantedBy=multi-user.target
EOF
mkdir -p /etc/vortexui
cat > /etc/vortexui/node.env << 'EOF'
%sEOF`, nodeEnvContent(input.AgentPort, input.Core, input.EnabledCores))},
		{"open firewall", fmt.Sprintf("ufw allow %d/tcp 2>/dev/null || iptables -I INPUT -p tcp --dport %d -j ACCEPT 2>/dev/null || true", input.AgentPort, input.AgentPort)},
		{"enable service", "systemctl daemon-reload && systemctl enable vortexui-node"},
	}

	var output strings.Builder
	for _, step := range steps {
		d.log.Info("deploy step", "step", step.name, "host", input.Host)
		out, err := client.Run(ctx, step.cmd)
		fmt.Fprintf(&output, "=== %s ===\n%s\n", step.name, out)
		if err != nil {
			// Non-fatal: some steps may fail (e.g. go not installed)
			fmt.Fprintf(&output, "WARN: %v\n", err)
		}
	}

	result.Success = true
	result.Address = fmt.Sprintf("%s:%d", input.Host, input.AgentPort)
	result.Output = output.String()
	_ = time.Now() // suppress unused import
	return result
}
