// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gitlab

import (
	"regexp"
)

var k8sTagVersionPattern = regexp.MustCompile(`^v[0-9]+(?:(?:alpha|beta)[0-9]+)?$`)

func isK8S(version string) bool {
	return k8sTagVersionPattern.MatchString(version)
}
