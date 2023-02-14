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
	"os"
	"testing"

	internaltesting "github.com/listendev/lstn/internal/testing"
)

func TestMain(m *testing.M) {
	behavior := os.Getenv("TEST_NPM_BEHAVIOR")
	switch behavior {
	case "":
		os.Exit(m.Run())
	case "npm-lt-6x":
		if err := internaltesting.StubNpm(internaltesting.NPM{
			Version: "4.6.3",
		}); err != nil {
			os.Exit(1)
		}
	case "npm-gt-6x":
		if err := internaltesting.StubNpm(internaltesting.NPM{
			Version: "8.19.2",
		}); err != nil {
			os.Exit(1)
		}
	case "npm-non-semver":
		if err := internaltesting.StubNpm(internaltesting.NPM{
			Version: "non-semver",
		}); err != nil {
			os.Exit(1)
		}
	case "nvm-gt-6x":
		if err := internaltesting.StubNpm(internaltesting.NPM{
			Version: "8.19.1",
			WithNVM: true,
		}); err != nil {
			os.Exit(1)
		}
	case "nvm-lt-6x":
		if err := internaltesting.StubNpm(internaltesting.NPM{
			Version: "4.6.0",
			WithNVM: true,
		}); err != nil {
			os.Exit(1)
		}
	case "nvm-gt-6x-no-use":
		if err := internaltesting.StubNpm(internaltesting.NPM{
			Version:      "8.19.2",
			WithNVM:      true,
			WithNVMNoUse: true,
		}); err != nil {
			os.Exit(1)
		}
	default:
		os.Exit(1)
	}
}
