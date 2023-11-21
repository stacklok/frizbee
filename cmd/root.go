//
// Copyright 2023 Stacklok, Inc.
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

// Package cmd provides the boomerang command line interface.
package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/stacklok/boomerang/cmd/ghactions"
)

// Execute runs the root command.
func Execute() {
	var rootCmd = &cobra.Command{
		Use:   "boomerang",
		Short: "boomerang is a tool you may throw a tag at and it comes back with a checksum",
	}

	rootCmd.AddCommand(ghactions.CmdGHActions())
	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		os.Exit(1)
	}
}
