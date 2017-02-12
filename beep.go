// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package main

import "log"

// BUG(brainman): MessageBeep Windows api is broken on Windows 7,
// so this example does not beep when runs as service on Windows 7.

func beep() {
	log.Printf("beeep")
}
