// Copyright 2022 Livesport TV s.r.o. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package util

import (
	"fmt"
	"math/rand"
	"time"
)

var uniqueIDGenerator = rand.New(rand.NewSource(time.Now().Unix()))

func GenerateUniqueID() string {
	return fmt.Sprintf("%16x", uniqueIDGenerator.Uint64())
}
