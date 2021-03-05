// Copyright © 2021 Jeffrey H. Johnson. <trnsz@pobox.com>
// Copyright © 2021 Gridfinity, LLC.
// Copyright © 2020 The Go Authors.
//
// All rights reserved.
//
// Use of this source code is governed by the BSD-style
// license that can be found in the LICENSE file.

// +build leaktest

package testutil

import (
	"fmt"
	"testing"
)

// LeakVerifyNone -> enabled (wrapper function)
func LeakVerifyNone(
	t *testing.T,
) error {
	err := leakVerifyNone(
		t,
		true,
	)
	if err != nil {
		return fmt.Errorf(
			fmt.Sprintf(
				"	%v",
				err,
			),
		)
	}
	return nil
}
