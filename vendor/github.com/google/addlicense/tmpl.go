// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"text/template"
	"unicode"
)

var licenseTemplate = map[string]string{
	"Apache-2.0": tmplApache,
	"MIT":        tmplMIT,
	"bsd":        tmplBSD,
	"MPL-2.0":    tmplMPL,
}

// maintain backwards compatibility by mapping legacy license types to their
// SPDX equivalents.
var legacyLicenseTypes = map[string]string{
	"apache": "Apache-2.0",
	"mit":    "MIT",
	"mpl":    "MPL-2.0",
}

// licenseData specifies the data used to fill out a license template.
type licenseData struct {
	Year   string // Copyright year(s).
	Holder string // Name of the copyright holder.
	SPDXID string // SPDX Identifier
}

// fetchTemplate returns the license template for the specified license and
// optional templateFile. If templateFile is provided, the license is read
// from the specified file. Otherwise, a template is loaded for the specified
// license, if recognized.
func fetchTemplate(license string, templateFile string, spdx spdxFlag) (string, error) {
	var t string
	if spdx == spdxOnly {
		t = tmplSPDX
	} else if templateFile != "" {
		d, err := ioutil.ReadFile(templateFile)
		if err != nil {
			return "", fmt.Errorf("license file: %w", err)
		}

		t = string(d)
	} else {
		t = licenseTemplate[license]
		if t == "" {
			if spdx == spdxOn {
				// unknown license, but SPDX headers requested
				t = tmplSPDX
			} else {
				return "", fmt.Errorf("unknown license: %q. Include the '-s' flag to request SPDX style headers using this license", license)
			}
		} else if spdx == spdxOn {
			// append spdx headers to recognized license
			t = t + spdxSuffix
		}
	}

	return t, nil
}

// executeTemplate will execute a license template t with data d
// and prefix the result with top, middle and bottom.
func executeTemplate(t *template.Template, d licenseData, top, mid, bot string) ([]byte, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, d); err != nil {
		return nil, err
	}
	var out bytes.Buffer
	if top != "" {
		fmt.Fprintln(&out, top)
	}
	s := bufio.NewScanner(&buf)
	for s.Scan() {
		fmt.Fprintln(&out, strings.TrimRightFunc(mid+s.Text(), unicode.IsSpace))
	}
	if bot != "" {
		fmt.Fprintln(&out, bot)
	}
	fmt.Fprintln(&out)
	return out.Bytes(), nil
}

const tmplApache = `Copyright{{ if .Year }} {{.Year}}{{ end }} {{.Holder}}

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.`

const tmplBSD = `Copyright (c){{ if .Year }} {{.Year}}{{ end }} {{.Holder}} All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.`

const tmplMIT = `Copyright (c){{ if .Year }} {{.Year}}{{ end }} {{.Holder}}

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.`

const tmplMPL = `This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.`

const tmplSPDX = `{{ if .Holder }}Copyright{{ if .Year }} {{.Year}}{{ end }} {{.Holder}}
{{ end }}SPDX-License-Identifier: {{.SPDXID}}`

const spdxSuffix = "\n\nSPDX-License-Identifier: {{.SPDXID}}"
