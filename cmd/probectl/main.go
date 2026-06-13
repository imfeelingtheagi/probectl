// SPDX-License-Identifier: LicenseRef-probectl-TBD

// Command probectl is the probectl command-line interface for test and agent
// management against the control-plane /v1 API. It covers the test
// {list,get,create,delete} and agent {list,get,delete} resources today — not yet
// the full /v1 surface. See `probectl help`.
//
// Configuration comes from flags or PROBECTL_API_URL / PROBECTL_API_TOKEN /
// PROBECTL_TENANT. The implementation lives in internal/cli (so it is testable).
package main

import (
	"os"

	"github.com/imfeelingtheagi/probectl/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Getenv, os.Stdout, os.Stderr))
}
