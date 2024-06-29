package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type (
	ForwardingMode string

	ConfigGroup struct {
		GroupId            string         `json:"group_id"`
		IsEnabled          bool           `json:"is_enabled"`
		ForwardingMode     ForwardingMode `json:"forwarding_mode"`
		ReceiversGroupIds  []string       `json:"receivers_group_ids"`
		BotSpecialAddonMsg string         `json:"bot_special_addon_msg,omitempty"`
		ReactionMark       string         `json:"reaction_mark,omitempty"`
		SenderNames        []string       `json:"sender_names,omitempty"`
		SenderUUIDs        []string       `json:"sender_uuids,omitempty"`
		StartsWith         []string       `json:"starts_with,omitempty"` //to filter messages, that starts with given patterns
		Contains           []string       `json:"contains,omitempty"`    //to filter messages, that contains given patterns
	}
	Config struct {
		CLIAddress          string        `json:"cli_address"`
		SelfNumber          string        `json:"self_number"`
		IgnoreOlderMessages uint64        `json:"ignore_older_messages"`
		IsSendingEnabled    bool          `json:"is_sending_enabled"`
		IsPrintMessages     bool          `json:"is_print_messages"`
		EnableDebugMessages bool          `json:"enable_debug_messages"`
		Forwarding          []ConfigGroup `json:"forwarding"`
	}
)

const (
	FwModeAttachments ForwardingMode = "attachments"
	FwModeMessages    ForwardingMode = "messages"
	FwModeAll         ForwardingMode = "all"
)

func (fm ForwardingMode) Validate() error {
	switch fm {
	case "attachments":
		fallthrough
	case "messages":
		fallthrough
	case "all":
		return nil
	default:
		return fmt.Errorf("invalid forwarding mode: %s", fm)
	}
}

func LoadConfig(filePath string) (c *Config, err error) {
	fileContent, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer fileContent.Close()

	byteResult, _ := io.ReadAll(fileContent)

	err = json.Unmarshal(byteResult, &c)
	if err != nil {
		return nil, err
	}

	return c, c.Validate()
}

func (c *Config) Validate() error {
	c.CLIAddress = strings.TrimSpace(c.CLIAddress)
	if len(c.CLIAddress) == 0 {
		return fmt.Errorf("CLI address is required")
	}

	c.SelfNumber = strings.TrimSpace(c.SelfNumber)
	if len(c.SelfNumber) == 0 {
		return fmt.Errorf("self number is required")
	}

	if len(c.Forwarding) > 0 {
		for i, group := range c.Forwarding {
			c.Forwarding[i].GroupId = strings.TrimSpace(group.GroupId)
			if c.Forwarding[i].IsEnabled && len(c.Forwarding[i].GroupId) == 0 {
				return fmt.Errorf("forwarding group id is required when record is enabled")
			}
			if c.Forwarding[i].IsEnabled && len(c.Forwarding[i].ReceiversGroupIds) == 0 {
				return fmt.Errorf("forwarding at leat one receivers group id is required when record is enabled")
			}

			if c.Forwarding[i].IsEnabled && len(c.Forwarding[i].ForwardingMode) == 0 {
				c.Forwarding[i].ForwardingMode = FwModeAll
			}
			if err := c.Forwarding[i].ForwardingMode.Validate(); err != nil {
				return err
			}

			if c.Forwarding[i].IsEnabled && len(c.Forwarding[i].ReceiversGroupIds) > 0 {
				for j := range c.Forwarding[i].ReceiversGroupIds {
					c.Forwarding[i].ReceiversGroupIds[j] = strings.TrimSpace(c.Forwarding[i].ReceiversGroupIds[j])
				}
				if len(c.Forwarding[i].ReceiversGroupIds[0]) == 0 {
					return fmt.Errorf("forwarding at leat one receivers group id is required not to be empty when record is enabled")
				}
			}
		}
	}

	return nil
}
