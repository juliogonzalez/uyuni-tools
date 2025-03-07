// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package podman

import (
	"path"
	"testing"

	"github.com/uyuni-project/uyuni-tools/shared/testutils"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

func TestHostInspectorGenerate(t *testing.T) {
	testDir := t.TempDir()

	inspector := NewHostInspector(testDir)
	if err := inspector.GenerateScript(); err != nil {
		t.Errorf("Unexpected error %s", err)
	}

	dataPath := inspector.GetDataPath()

	//nolint:lll
	expected := `#!/bin/bash
# inspect.sh, generated by mgradm
echo "scc_username=$(cat /etc/zypp/credentials.d/SCCcredentials 2>&1 /dev/null | grep username | cut -d= -f2 || true)" >> ` + dataPath + `
echo "scc_password=$(cat /etc/zypp/credentials.d/SCCcredentials 2>&1 /dev/null | grep password | cut -d= -f2 || true)" >> ` + dataPath + `
echo "has_uyuni_server=$(systemctl list-unit-files uyuni-server.service >/dev/null && echo true || echo false)" >> ` + dataPath + `
exit 0
`

	actual := testutils.ReadFile(t, path.Join(testDir, utils.InspectScriptFilename))
	testutils.AssertEquals(t, "Wrongly generated script", expected, actual)
}

func TestHostInspectorParse(t *testing.T) {
	testDir := t.TempDir()

	inspector := NewHostInspector(testDir)

	content := `
scc_username=myuser
scc_password=mysecret
has_uyuni_server=true
`
	testutils.WriteFile(t, inspector.GetDataPath(), content)

	actual, err := inspector.ReadInspectData()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	testutils.AssertEquals(t, "Invalid SCC username", "myuser", actual.SCCUsername)
	testutils.AssertEquals(t, "Invalid SCC password", "mysecret", actual.SCCPassword)
	testutils.AssertTrue(t, "HasUyuniServer should be true", actual.HasUyuniServer)
}
