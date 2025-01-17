// Copyright © 2016 Wei Shen <shenwei356@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "csvtk",
	Short: "Another cross-platform, efficient and practical CSV/TSV toolkit",
	Long: `Another cross-platform, efficient and practical CSV/TSV toolkit

Version: 0.3

Author: Wei Shen <shenwei356@gmail.com>

Documents  : http://shenwei356.github.io/csvtk
Source code: https://github.com/shenwei356/csvtk

Attention:

  1. The CSV parser requires all the lines have same number of fields/columns.
     Even lines with spaces will cause error.
  2. By default, csvtk think your files have header row, if not, use "-H".
  3. By default, lines starting with '#' will be ignored, if the header row
     starts with '#', please assign "-C" another rare symbol, e.g. '$'.
  4. By default, csvtk handles CSV files, use "-t" for tab-delimited files.

`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().IntP("chunk-size", "c", 50, `chunk size of CSV reader`)
	RootCmd.PersistentFlags().IntP("num-cpus", "j", runtime.NumCPU(), `number of CPUs to use (default value depends on your computer)`)

	RootCmd.PersistentFlags().StringP("delimiter", "d", ",", `delimiting character of the input CSV file`)
	RootCmd.PersistentFlags().StringP("out-delimiter", "D", ",", `delimiting character of the input CSV file`)
	// RootCmd.PersistentFlags().StringP("quote-char", "q", `"`, `character used to quote strings in the input CSV file`)
	RootCmd.PersistentFlags().StringP("comment-char", "C", `#`, "lines starting with commment-character will be ignored. "+
		`if your header row starts with '#', please assign "-C" another rare symbol, e.g. '$'`)
	RootCmd.PersistentFlags().BoolP("lazy-quotes", "l", false, `if given, a quote may appear in an unquoted field and a non-doubled quote may appear in a quoted field`)

	RootCmd.PersistentFlags().BoolP("tabs", "t", false, `specifies that the input CSV file is delimited with tabs. Overrides "-d"`)
	RootCmd.PersistentFlags().BoolP("out-tabs", "T", false, `specifies that the output is delimited with tabs. Overrides "-D"`)
	RootCmd.PersistentFlags().BoolP("no-header-row", "H", false, `specifies that the input CSV file does not have header row`)
	RootCmd.PersistentFlags().StringP("out-file", "o", "-", `out file ("-" for stdout, suffix .gz for gzipped out)`)
}
