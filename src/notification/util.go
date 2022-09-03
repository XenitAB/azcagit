// Orignial copyright from Flux (commit: c904061): https://github.com/fluxcd/notification-controller/blob/main/internal/notifier/util.go
// /*
// Copyright 2020 The Flux authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

package notification

import (
	"fmt"
	"strings"

	giturls "github.com/whilp/git-urls"
)

func toPtr[T any](a T) *T {
	return &a
}

func parseGitAddress(s string) (string, string, error) {
	u, err := giturls.Parse(s)
	if err != nil {
		return "", "", nil
	}

	scheme := u.Scheme
	if u.Scheme == "ssh" {
		scheme = "https"
	}

	id := strings.TrimLeft(u.Path, "/")
	id = strings.TrimSuffix(id, ".git")
	host := fmt.Sprintf("%s://%s", scheme, u.Host)
	return host, id, nil
}
