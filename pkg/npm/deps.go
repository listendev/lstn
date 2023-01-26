/*
Copyright © 2022 The listen.dev team <engineering@garnet.ai>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package npm

// Deps gets you the package lock dependencies.
func (p *packageLockJSON) Deps() map[string]PackageLockDependency {
	switch p.LockfileVersion.Value {
	case 2:
		return p.packageLockJSONVersion2.Dependencies
	case 3:
		return p.packageLockJSONVersion3.Dependencies
	case 1:
		fallthrough
	default:
		return nil
	}
}
