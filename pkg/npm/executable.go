// SPDX-License-Identifier: Apache-2.0
//
// Copyright © 2023 The listen.dev team <engineering@garnet.ai>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package npm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

func getNPM(ctx context.Context) (*exec.Cmd, error) {
	exe, err := exec.LookPath("npm")
	if err != nil {
		return nil, fmt.Errorf("couldn't find the npm executable in the PATH")
	}

	return exec.CommandContext(ctx, exe), nil
}

func getNPMFromNVM(ctx context.Context) (*exec.Cmd, error) {
	nvmDir := os.Getenv("NVM_DIR")
	if nvmDir == "" {
		return nil, fmt.Errorf("couldn't detect the nvm directory")
	}
	bashExe, err := exec.LookPath("bash")
	if err != nil {
		return nil, fmt.Errorf("couldn't find bash in the PATH")
	}

	cmdline := fmt.Sprintf("source %s/nvm.sh", nvmDir)

	nvmNoUse := os.Getenv("NVM_NO_USE")
	if nvmNoUse == "true" {
		cmdline += " --no-use"
	}

	// TODO: detect whether nvm has already a Node version installed (`nvm ls current` or `nvm which current`)
	// in case it doesn't have any we can:
	// 1. automatically install the latest lts (`source $NVM_DIR/nvm.sh --install`)
	// 2. exit and warn the user to install and use a Node version via nvm

	return exec.CommandContext(ctx, bashExe, "-c", fmt.Sprintf("%s && npm", cmdline)), nil
}
