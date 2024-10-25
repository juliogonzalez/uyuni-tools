// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/uyuni-project/uyuni-tools/shared/test_utils"
	"github.com/uyuni-project/uyuni-tools/shared/test_utils/flags_tests"
	"github.com/uyuni-project/uyuni-tools/shared/types"
)

func TestParamsParsing(t *testing.T) {
	args := []string{
		"config.tar.gz",
	}

	args = append(args, flags_tests.ImageProxyFlagsTestArgs...)
	args = append(args, flags_tests.ProxyHelmFlagsTestArgs...)
	args = append(args, flags_tests.SccFlagTestArgs...)

	// Test function asserting that the args are properly parsed
	tester := func(globalFlags *types.GlobalFlags, flags *kubernetesProxyInstallFlags,
		cmd *cobra.Command, args []string,
	) error {
		flags_tests.AssertProxyImageFlags(t, cmd, &flags.ProxyImageFlags)
		flags_tests.AssertProxyHelmFlags(t, cmd, &flags.Helm)
		flags_tests.AssertSccFlag(t, cmd, &flags.Scc)
		return nil
	}

	globalFlags := types.GlobalFlags{}
	cmd := newCmd(&globalFlags, tester)

	test_utils.AssertHasAllFlags(t, cmd, args)

	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Errorf("command failed with error: %s", err)
	}
}
