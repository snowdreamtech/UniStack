// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package ansible

import (
	"embed"
)

//go:embed inventory playbooks roles ansible.cfg requirements.txt requirements.yml
var FS embed.FS
