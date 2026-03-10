package postgres

import "github.com/nrf24l01/go-web-utils/pg_kit"

const (
	DeviceBlockRuleTargetIP   = "ip"
	DeviceBlockRuleTargetCIDR = "cidr"
	DeviceBlockRuleTargetSNI  = "sni"
)

type DeviceBlockRule struct {
	pg_kit.BaseModel

	DeviceID    string `gorm:"type:uuid;not null;index;uniqueIndex:idx_device_block_rule_unique"`
	TargetType  string `gorm:"type:varchar(16);not null;index;uniqueIndex:idx_device_block_rule_unique"`
	TargetValue string `gorm:"type:text;not null;uniqueIndex:idx_device_block_rule_unique"`
	Enabled     bool   `gorm:"not null;default:true;index"`
}

func (DeviceBlockRule) TableName() string {
	return "device_block_rules"
}
