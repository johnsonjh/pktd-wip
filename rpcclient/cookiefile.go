// Copyrgith © 2021 Jeffrey H. Johnson. <trnsz@pobox.com>
// Copyright © 2021 Gridfinity, LLC.
// Copyright © 2017 The Namecoin developers
// Copyright © 2019 The btcsuite developers
//
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpcclient

import (
	"bufio"
	"os"
	"strings"

	"github.com/pkt-cash/pktd/btcutil/er"
)

func readCookieFile(path string) (username, password string, err er.R) {
	f, errr := os.Open(path)
	if errr != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan()
	errr = scanner.Err()
	if errr != nil {
		return
	}
	s := scanner.Text()

	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		err := er.E(errr)
		err.AddMessage("Corrupt or malformed pktcookie file")
		return "", "", err
	}

	username, password = parts[0], parts[1]
	return
}
