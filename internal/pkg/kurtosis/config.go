package kurtosis

import (
	"attacknet/cmd/internal/pkg/network"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Participants        []*Participant `yaml:"participants"`
	NetParams           GenesisConfig  `yaml:"network_params"`
	AdditionalServices  []string       `yaml:"additional_services"`
	ParallelKeystoreGen bool           `yaml:"parallel_keystore_generation"`
	Persistent          bool           `yaml:"persistent"`
	DisablePeerScoring  bool           `yaml:"disable_peer_scoring"`
}

type GenesisConfig struct {
	PreregisteredValidatorKeysMnemonic *string `yaml:"preregistered_validator_keys_mnemonic,omitempty"`
	PreregisteredValidatorCount        *int    `yaml:"preregistered_validator_count,omitempty"`
	NetworkId                          *int    `yaml:"network_id,omitempty"`
	DepositContractAddress             *string `yaml:"deposit_contract_address,omitempty"`
	SecondsPerSlot                     *int    `yaml:"seconds_per_slot,omitempty"`
	GenesisDelay                       *int    `yaml:"genesis_delay,omitempty"`
	MaxChurn                           *uint64 `yaml:"max_churn,omitempty"`
	EjectionBalance                    *uint64 `yaml:"ejection_balance,omitempty"`
	Eth1FollowDistance                 *int    `yaml:"eth1_follow_distance,omitempty"`
	CapellaForkEpoch                   *int    `yaml:"capella_fork_epoch,omitempty"`
	DenebForkEpoch                     *int    `yaml:"deneb_fork_epoch,omitempty"`
	ElectraForkEpoch                   *int    `yaml:"electra_fork_epoch,omitempty"`
	NumValKeysPerNode                  int     `yaml:"num_validator_keys_per_node"`
}

// Participant represents a participant in a kurtosis config. Note we use pointers for omitempty fields,
// so we can determine if they were set.
type Participant struct {
	// required fields
	ElClientType network.ExecutionClientType `yaml:"el_type"`
	ClClientType network.ConsensusClientType `yaml:"cl_type"`

	Count int `yaml:"count"`
	// all other fields are optional and will be populated with defaults defined in kurtosis
	ElClientImage *string           `yaml:"el_image,omitempty"`
	ELExtraLabels map[string]string `yaml:"el_extra_labels,omitempty"`

	ClClientImage *string           `yaml:"cl_image,omitempty"`
	CLExtraLabels map[string]string `yaml:"cl_extra_labels,omitempty"`

	//Sidecars, these are optional
	CLSeparateVC     *bool                        `yaml:"use_separate_vc,omitempty"`
	CLValidatorType  *network.ConsensusClientType `yaml:"vc_type,omitempty"`
	CLValidatorImage *string                      `yaml:"vc_image,omitempty"`
	VCExtraLabels    map[string]string            `yaml:"vc_extra_labels,omitempty"`

	// resource requirements are optional
	ElMinCpu    *int `yaml:"el_min_cpu,omitempty"`
	ElMaxCpu    *int `yaml:"el_max_cpu,omitempty"`
	ElMinMemory *int `yaml:"el_min_mem,omitempty"`
	ElMaxMemory *int `yaml:"el_max_mem,omitempty"`

	ClMinCpu    *int `yaml:"cl_min_cpu,omitempty"`
	ClMaxCpu    *int `yaml:"cl_max_cpu,omitempty"`
	ClMinMemory *int `yaml:"cl_min_mem,omitempty"`
	ClMaxMemory *int `yaml:"cl_max_mem,omitempty"`

	ValMinCpu    *int `yaml:"vc_min_cpu,omitempty"`
	ValMaxCpu    *int `yaml:"vc_max_cpu,omitempty"`
	ValMinMemory *int `yaml:"vc_min_mem,omitempty"`
	ValMaxMemory *int `yaml:"vc_max_mem,omitempty"`
}

func (n Config) String() string {
	bytes, _ := yaml.Marshal(n)
	return string(bytes)
}
