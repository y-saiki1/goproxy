// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gitlab

import (
	"fmt"

	"go.lstv.dev/goproxy/source"
)

type tagNotFoundError struct {
	tag string
}

func newTagNotFoundError(tag string) error {
	return source.NewVersionNotFoundError(&tagNotFoundError{
		tag: tag,
	})
}

func (t *tagNotFoundError) Error() string {
	return fmt.Sprintf("tag %q not found", t.tag)
}
