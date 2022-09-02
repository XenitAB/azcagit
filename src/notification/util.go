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
	"crypto/sha1"
	"encoding/base64"
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

func parseRevision(rev string) (string, error) {
	comp := strings.Split(rev, "/")
	if len(comp) < 2 {
		return "", fmt.Errorf("Revision string format incorrect: %v", rev)
	}
	sha := comp[len(comp)-1]
	if sha == "" {
		return "", fmt.Errorf("Commit Sha cannot be empty: %v", rev)
	}
	return sha, nil
}

func isCommitStatus(meta map[string]string, status string) bool {
	if val, ok := meta["commit_status"]; ok && val == status {
		return true
	}
	return false
}

func sha1String(str string) string {
	bs := []byte(str)
	return fmt.Sprintf("%x", sha1.Sum(bs))
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}