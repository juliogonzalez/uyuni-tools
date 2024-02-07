// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/uyuni-project/uyuni-tools/mgradm/shared/templates"
	"github.com/uyuni-project/uyuni-tools/shared"
	"github.com/uyuni-project/uyuni-tools/shared/kubernetes"
	"github.com/uyuni-project/uyuni-tools/shared/utils"
)

func ExecCommand(logLevel zerolog.Level, cnx *shared.Connection, args ...string) error {
	podName, err := cnx.GetPodName()
	if err != nil {
		log.Fatal().Err(err)
	}

	commandArgs := []string{"exec", podName}

	command, err := cnx.GetCommand()
	if err != nil {
		log.Fatal().Err(err)
	}

	if command == "kubectl" {
		commandArgs = append(commandArgs, "-c", "uyuni", "--")
	}

	commandArgs = append(commandArgs, "sh", "-c", strings.Join(args, " "))

	runCmd := exec.Command(command, commandArgs...)
	logger := utils.OutputLogWriter{Logger: log.Logger, LogLevel: logLevel}
	runCmd.Stdout = logger
	runCmd.Stderr = logger
	return runCmd.Run()
}

func GeneratePgMigrationScript(scriptDir string, oldPgVersion string, newPgVersion string, kubernetes bool) (string, error) {
	data := templates.MigratePostgresVersionTemplateData{
		OldVersion: oldPgVersion,
		NewVersion: newPgVersion,
		Kubernetes: kubernetes,
	}

	scriptName := "migrate_pgsql.sh"
	scriptPath := filepath.Join(scriptDir, scriptName)
	if err := utils.WriteTemplateToFile(data, scriptPath, 0555, true); err != nil {
		return "", fmt.Errorf("Failed to generate %s", scriptName)
	}
	return scriptName, nil
}

func GenerateFinalizePostgresMigrationScript(scriptDir string, RunAutotune bool, RunReindex bool, RunSchemaUpdate bool, RunDistroMigration bool, kubernetes bool) (string, error) {
	data := templates.FinalizePostgresTemplateData{
		RunAutotune:        RunAutotune,
		RunReindex:         RunReindex,
		RunSchemaUpdate:    RunSchemaUpdate,
		RunDistroMigration: RunDistroMigration,
		Kubernetes:         kubernetes,
	}

	scriptName := "finalize_pgsql.sh"
	scriptPath := filepath.Join(scriptDir, scriptName)
	if err := utils.WriteTemplateToFile(data, scriptPath, 0555, true); err != nil {
		return "", fmt.Errorf("Failed to generate %s", scriptName)
	}
	return scriptName, nil
}

func ReadContainerData(scriptDir string) (string, string, string) {
	data, err := os.ReadFile(filepath.Join(scriptDir, "data"))
	if err != nil {
		log.Fatal().Msgf("Failed to read data extracted from source host")
	}
	viper.SetConfigType("env")
	viper.ReadConfig(bytes.NewBuffer(data))
	return viper.GetString("Timezone"), viper.GetString("old_pg_version"), viper.GetString("new_pg_version")
}

func RunMigration(cnx *shared.Connection, tmpPath string, scriptName string) {
	log.Info().Msg("Migrating server")
	err := ExecCommand(zerolog.InfoLevel, cnx, "/var/lib/uyuni-tools/"+scriptName)
	if err != nil {
		log.Fatal().Err(err).Msg("error running the migration script")
	}
}

func GenerateMigrationScript(sourceFqdn string, kubernetes bool) string {
	scriptDir, err := os.MkdirTemp("", "mgradm-*")
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to create temporary directory")
	}

	data := templates.MigrateScriptTemplateData{
		Volumes:    utils.VOLUMES,
		SourceFqdn: sourceFqdn,
		Kubernetes: kubernetes,
	}

	scriptPath := filepath.Join(scriptDir, "migrate.sh")
	if err = utils.WriteTemplateToFile(data, scriptPath, 0555, true); err != nil {
		log.Fatal().Err(err).Msgf("Failed to generate migration script")
	}

	return scriptDir
}

func RunningImage(cnx *shared.Connection, containerName string) (string, error) {
	command, err := cnx.GetCommand()

	switch command {
	case "podman":
		args := []string{"ps", "--format", "{{.Image}}", "--noheading"}
		image, err := utils.RunCmdOutput(zerolog.DebugLevel, "podman", args...)
		if err != nil {
			return "", err
		}
		return strings.Trim(string(image), "\n"), nil

	case "kubectl":

		//FIXME this will work until containers 0 is uyuni. Then jsonpath should be something like
		// {.items[0].spec.containers[?(@.name=="` + containerName + `")].image but there are problems
		// using RunCmdOutput with an arguments with round brackets
		args := []string{"get", "pods", kubernetes.ServerFilter, "-o", "jsonpath={.items[0].spec.containers[0].image}"}
		image, err := utils.RunCmdOutput(zerolog.DebugLevel, "kubectl", args...)

		log.Info().Msgf("image is: %s", image)
		if err != nil {
			return "", err
		}
		return strings.Trim(string(image), "\n"), nil
	}

	return command, err
}
