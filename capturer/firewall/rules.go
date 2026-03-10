package firewall

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/nrf24l01/sniffly/capturer/core"
	capture_grpc "github.com/nrf24l01/sniffly/capturer/grpc"
	pb "github.com/nrf24l01/sniffly/capturer/proto"
)

const parentChain = "FORWARD"

func runIptables(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "iptables", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		trimmed := strings.TrimSpace(string(output))
		if trimmed != "" {
			return fmt.Errorf("%w: %s", err, trimmed)
		}
		return err
	}
	return nil
}

func ensureChain(ctx context.Context, chain string) error {
	err := runIptables(ctx, "-N", chain)
	if err != nil && !strings.Contains(err.Error(), "Chain already exists") {
		return fmt.Errorf("create chain %s: %w", chain, err)
	}

	err = runIptables(ctx, "-C", parentChain, "-j", chain)
	if err != nil {
		err = runIptables(ctx, "-I", parentChain, "1", "-j", chain)
		if err != nil {
			return fmt.Errorf("attach chain %s: %w", chain, err)
		}
	}

	return nil
}

func buildRuleCommands(chain string, rule *pb.BlockRule) ([][]string, error) {
	device_mac := strings.TrimSpace(rule.GetDeviceMac())
	if device_mac == "" {
		return nil, fmt.Errorf("device MAC is required")
	}

	switch rule.GetTargetType() {
	case pb.BlockRuleTargetType_BLOCK_RULE_TARGET_TYPE_IP, pb.BlockRuleTargetType_BLOCK_RULE_TARGET_TYPE_CIDR:
		return [][]string{{"-A", chain, "-m", "mac", "--mac-source", device_mac, "-d", strings.TrimSpace(rule.GetTargetValue()), "-j", "DROP"}}, nil
	case pb.BlockRuleTargetType_BLOCK_RULE_TARGET_TYPE_SNI:
		target_value := strings.TrimSpace(rule.GetTargetValue())
		if target_value == "" {
			return nil, fmt.Errorf("target value is required")
		}
		return [][]string{{"-A", chain, "-p", "tcp", "-m", "mac", "--mac-source", device_mac, "-m", "multiport", "--dports", "80,443", "-m", "string", "--algo", "bm", "--string", target_value, "-j", "DROP"}}, nil
	default:
		return nil, fmt.Errorf("unsupported target type %s", rule.GetTargetType().String())
	}
}

func applyRules(ctx context.Context, cfg *core.Config, client pb.PacketGatewayClient) error {
	rules, err := capture_grpc.FetchBlockRules(ctx, client, cfg)
	if err != nil {
		return err
	}

	err = ensureChain(ctx, cfg.IptablesChain)
	if err != nil {
		return err
	}

	err = runIptables(ctx, "-F", cfg.IptablesChain)
	if err != nil {
		return fmt.Errorf("flush chain: %w", err)
	}

	for _, rule := range rules {
		if rule == nil || !rule.Enabled || strings.TrimSpace(rule.GetDeviceMac()) == "" {
			continue
		}

		commands, err := buildRuleCommands(cfg.IptablesChain, rule)
		if err != nil {
			log.Printf("skip invalid firewall rule %q for device %s: %v", rule.GetId(), rule.GetDeviceMac(), err)
			continue
		}

		for _, args := range commands {
			err = runIptables(ctx, args...)
			if err != nil {
				return fmt.Errorf("apply firewall rule %q: %w", rule.GetId(), err)
			}
		}
	}

	log.Printf("firewall sync applied %d active rules", len(rules))
	return nil
}

func SyncLoop(ctx context.Context, cfg *core.Config, client pb.PacketGatewayClient) {
	sync_interval := time.Duration(cfg.BlockRulesSyncIntervalSeconds) * time.Second
	if sync_interval <= 0 {
		sync_interval = 30 * time.Second
	}

	ticker := time.NewTicker(sync_interval)
	defer ticker.Stop()

	for {
		if err := applyRules(ctx, cfg, client); err != nil {
			log.Printf("firewall sync failed: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
