// SPDX-FileCopyrightText: 2025 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/shared"
	. "github.com/uyuni-project/uyuni-tools/shared/l10n"
	"github.com/uyuni-project/uyuni-tools/shared/types"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

type cacheClearFlags struct {
}

// NewClearCmd creates the command to clear the cache.
func NewClearCmd(globalFlags *types.GlobalFlags) *cobra.Command {
	var clearCmd = &cobra.Command{
		Use:   "clear",
		Short: L("Clear the cache"),
		Long:  L("Clear the cache"),
		RunE: func(cmd *cobra.Command, args []string) error {
			var flags cacheClearFlags
			return utils.CommandHelper(globalFlags, cmd, args, &flags, nil, clearCmd)
		},
	}

	return clearCmd
}

func clearCmd(globalFlags *types.GlobalFlags, flags *cacheClearFlags, cmd *cobra.Command, args []string) error {
	fn, err := shared.ChooseProxyPodmanOrKubernetes(cmd.Flags(), podmanCacheClear, kubernetesCacheClear)
	if err != nil {
		return err
	}

	return fn(globalFlags, flags, cmd, args)
}
