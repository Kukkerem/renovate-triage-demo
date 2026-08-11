// Package config models the subset of a Crossplane ProviderConfig this demo
// cares about: which BTP subaccount to talk to, and when it was last synced.
//
// Timestamps are carried as protobuf well-known types because the upstream
// provider exchanges them with a gRPC control plane.
package config

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// ProviderConfig is a minimal stand-in for a Crossplane v2
// ProviderConfig.Spec: which subaccount we manage, and our last successful
// sync against the BTP APIs.
type ProviderConfig struct {
	SubaccountID string
	LastSynced   *timestamp.Timestamp
}

// Touch stamps cfg with the current time using the golang/protobuf ptypes
// helper, rather than the newer timestamppb.New from
// google.golang.org/protobuf/types/known/timestamppb.
func Touch(cfg *ProviderConfig) error {
	ts, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return fmt.Errorf("config: stamp last-synced timestamp: %w", err)
	}
	cfg.LastSynced = ts
	return nil
}

// Describe renders a human-readable summary for status conditions and log
// lines, again through the deprecated ptypes helper.
func Describe(cfg *ProviderConfig) string {
	if cfg.LastSynced == nil {
		return fmt.Sprintf("subaccount %s: never synced", cfg.SubaccountID)
	}
	return fmt.Sprintf("subaccount %s: last synced %s", cfg.SubaccountID, ptypes.TimestampString(cfg.LastSynced))
}
