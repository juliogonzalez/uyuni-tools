// SPDX-FileCopyrightText: 2024 SUSE LLC
//
// SPDX-License-Identifier: Apache-2.0

package templates

import (
	"io"
	"text/template"

	"github.com/uyuni-project/uyuni-tools/shared/types"
)

const serviceTemplate = `# uyuni-server.service, generated by mgradm
# Use an uyuni-server.service.d/local.conf file to override

[Unit]
Description=Uyuni server image container service
Wants=network.target
After=network-online.target
RequiresMountsFor=%t/containers

[Service]
Environment=PODMAN_SYSTEMD_UNIT=%n
Restart=on-failure
ExecStartPre=/bin/rm -f %t/uyuni-server.pid %t/%n.ctr-id
ExecStartPre=/usr/bin/podman rm --ignore --force -t 10 {{ .NamePrefix }}-server
ExecStart=/bin/sh -c '/usr/bin/podman run \
	--conmon-pidfile %t/uyuni-server.pid \
	--cidfile=%t/%n.ctr-id \
	--cgroups=no-conmon \
	--shm-size=0 \
	--shm-size-systemd=0 \
	--sdnotify=conmon \
	-d \
	--name {{ .NamePrefix }}-server \
	--hostname {{ .NamePrefix }}-server.mgr.internal \
	{{ .Args }} \
	{{- range .Ports }}
	-p {{ .Exposed }}:{{ .Port }}{{if .Protocol}}/{{ .Protocol }}{{end}} \
	{{- end }}
	{{- range .Volumes }}
	-v {{ .Name }}:{{ .MountPath }} \
	{{- end }}
	-e TZ=${TZ} \
	--network {{ .Network }} \
	${PODMAN_EXTRA_ARGS} ${UYUNI_IMAGE}'
ExecStop=/usr/bin/podman exec \
    uyuni-server \
    /bin/bash -c 'spacewalk-service stop && systemctl stop postgresql'
ExecStop=/usr/bin/podman stop \
	--ignore -t 10 \
	--cidfile=%t/%n.ctr-id
ExecStopPost=/usr/bin/podman rm \
	-f \
	--ignore -t 10 \
	--cidfile=%t/%n.ctr-id

PIDFile=%t/uyuni-server.pid
TimeoutStopSec=180
TimeoutStartSec=900
Type=forking

[Install]
WantedBy=multi-user.target default.target
`

// PodmanServiceTemplateData POD information to create systemd file.
type PodmanServiceTemplateData struct {
	Volumes    []types.VolumeMount
	NamePrefix string
	Args       string
	Ports      []types.PortMap
	Image      string
	Network    string
}

// Render will create the systemd configuration file.
func (data PodmanServiceTemplateData) Render(wr io.Writer) error {
	t := template.Must(template.New("service").Parse(serviceTemplate))
	return t.Execute(wr, data)
}
