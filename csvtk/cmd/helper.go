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
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/brentp/xopen"
	"github.com/shenwei356/util/stringutil"
	"github.com/spf13/cobra"
)

func checkError(err error) {
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
}

func getFileList(args []string) []string {
	files := []string{}
	if len(args) == 0 {
		files = append(files, "-")
	} else {
		for _, file := range files {
			if file == "-" {
				continue
			}
			if _, err := os.Stat(file); os.IsNotExist(err) {
				checkError(err)
			}
		}
		files = args
	}
	return files
}

func getFlagInt(cmd *cobra.Command, flag string) int {
	value, err := cmd.Flags().GetInt(flag)
	checkError(err)
	return value
}

func getFlagPositiveInt(cmd *cobra.Command, flag string) int {
	value, err := cmd.Flags().GetInt(flag)
	checkError(err)
	if value <= 0 {
		checkError(fmt.Errorf("value of flag --%s should be greater than 0", flag))
	}
	return value
}

func getFlagNonNegativeInt(cmd *cobra.Command, flag string) int {
	value, err := cmd.Flags().GetInt(flag)
	checkError(err)
	if value < 0 {
		checkError(fmt.Errorf("value of flag --%s should be greater than 0", flag))
	}
	return value
}

func getFlagBool(cmd *cobra.Command, flag string) bool {
	value, err := cmd.Flags().GetBool(flag)
	checkError(err)
	return value
}

func getFlagString(cmd *cobra.Command, flag string) string {
	value, err := cmd.Flags().GetString(flag)
	checkError(err)
	return value
}

func getFlagCommaSeparatedStrings(cmd *cobra.Command, flag string) []string {
	value, err := cmd.Flags().GetString(flag)
	checkError(err)
	return stringutil.Split(value, ",")
}

func getFlagSemicolonSeparatedStrings(cmd *cobra.Command, flag string) []string {
	value, err := cmd.Flags().GetString(flag)
	checkError(err)
	return stringutil.Split(value, ";")
}

func getFlagCommaSeparatedInts(cmd *cobra.Command, flag string) []int {
	filedsStrList := getFlagCommaSeparatedStrings(cmd, flag)
	fields := make([]int, len(filedsStrList))
	for i, value := range filedsStrList {
		v, err := strconv.Atoi(value)
		if err != nil {
			checkError(fmt.Errorf("value of flag --%s should be comma separated integers", flag))
		}
		fields[i] = v
	}
	return fields
}

func getFlagRune(cmd *cobra.Command, flag string) rune {
	value, err := cmd.Flags().GetString(flag)
	checkError(err)
	if len(value) > 1 {
		checkError(fmt.Errorf("value of flag --%s should has length of 1", flag))
	}
	var v rune
	for _, r := range value {
		v = r
		break
	}
	return v
}

func getFlagFloat64(cmd *cobra.Command, flag string) float64 {
	value, err := cmd.Flags().GetFloat64(flag)
	checkError(err)
	return value
}

func getFlagInt64(cmd *cobra.Command, flag string) int64 {
	value, err := cmd.Flags().GetInt64(flag)
	checkError(err)
	return value
}

func getFlagStringSlice(cmd *cobra.Command, flag string) []string {
	value, err := cmd.Flags().GetStringSlice(flag)
	checkError(err)
	return value
}

// Config is the struct containing all gloabl flags
type Config struct {
	ChunkSize int
	NumCPUs   int

	Delimiter    rune
	OutDelimiter rune
	// QuoteChar   rune
	CommentChar rune
	LazyQuotes  bool

	Tabs        bool
	OutTabs     bool
	NoHeaderRow bool

	OutFile string
}

func getConfigs(cmd *cobra.Command) Config {
	return Config{
		ChunkSize: getFlagPositiveInt(cmd, "chunk-size"),
		NumCPUs:   getFlagPositiveInt(cmd, "num-cpus"),

		Delimiter:    getFlagRune(cmd, "delimiter"),
		OutDelimiter: getFlagRune(cmd, "out-delimiter"),
		// QuoteChar:   getFlagRune(cmd, "quote-char"),
		CommentChar: getFlagRune(cmd, "comment-char"),
		LazyQuotes:  getFlagBool(cmd, "lazy-quotes"),

		Tabs:        getFlagBool(cmd, "tabs"),
		OutTabs:     getFlagBool(cmd, "out-tabs"),
		NoHeaderRow: getFlagBool(cmd, "no-header-row"),

		OutFile: getFlagString(cmd, "out-file"),
	}
}

func newCSVReaderByConfig(config Config, file string) (*CSVReader, error) {
	reader, err := NewCSVReader(file, config.NumCPUs, config.ChunkSize)
	if err != nil {
		return nil, err
	}
	if config.Tabs {
		reader.Reader.Comma = '\t'
	} else {
		reader.Reader.Comma = config.Delimiter
	}
	reader.Reader.Comment = config.CommentChar
	reader.Reader.LazyQuotes = config.LazyQuotes

	return reader, nil
}

// NewCSVWriterChanByConfig returns a chanel which you can send record to write
func NewCSVWriterChanByConfig(config Config) (chan []string, error) {
	outfh, err := xopen.Wopen(config.OutFile)
	if err != nil {
		return nil, err
	}

	ch := make(chan []string, config.NumCPUs)

	writer := csv.NewWriter(outfh)
	if config.OutTabs {
		writer.Comma = '\t'
	} else {
		writer.Comma = config.OutDelimiter
	}
	go func() {
		defer outfh.Close()
		for record := range ch {
			if err := writer.Write(record); err != nil {
				log.Fatal("error writing record to csv:", err)
			}
		}
		writer.Flush()
		if err := writer.Error(); err != nil {
			log.Fatal(err)
		}

	}()

	return ch, nil
}

var reFields = regexp.MustCompile(`([^,]+)(,[^,]+)*,?`)
var reDigitals = regexp.MustCompile(`^[\-\d\.e,E\+]+$`)
var reDigitalRange = regexp.MustCompile(`^([\-\d]+?)\-([\-\d]+?)$`)

func getFlagFields(cmd *cobra.Command, flag string) string {
	fieldsStr, err := cmd.Flags().GetString(flag)
	checkError(err)
	if fieldsStr == "" {
		checkError(fmt.Errorf("flag --%s needed", flag))
	}
	if !reFields.MatchString(fieldsStr) {
		checkError(fmt.Errorf("invalid value of flag %s", flag))
	}
	return fieldsStr
}

func parseFields(cmd *cobra.Command,
	fieldsStr string,
	noHeaderRow bool) ([]int, []string, bool, bool) {

	var fields []int
	var colnames []string
	var parseHeaderRow bool
	var negativeFields bool
	firstField := reFields.FindAllStringSubmatch(fieldsStr, -1)[0][1]
	if reDigitals.MatchString(firstField) {
		fields = []int{}
		fieldsStrs := strings.Split(fieldsStr, ",")
		for _, s := range fieldsStrs {
			found := reDigitalRange.FindAllStringSubmatch(s, -1)
			if len(found) > 0 { // field range
				start, err := strconv.Atoi(found[0][1])
				checkError(err)
				end, err := strconv.Atoi(found[0][2])
				checkError(err)
				if start == 0 || end == 0 {
					checkError(fmt.Errorf("no 0 allowed in field range: %s", s))
				}
				if start >= end {
					checkError(fmt.Errorf("invalid field range: %s. start (%d) should be less than end (%d)", s, start, end))
				}
				for i := start; i <= end; i++ {
					fields = append(fields, i)
				}
			} else {
				field, err := strconv.Atoi(s)
				checkError(err)
				fields = append(fields, field)
			}
		}

		for _, f := range fields {
			if f == 0 {
				checkError(fmt.Errorf(`field should not be 0`))
			} else if f < 0 {
				negativeFields = true
			} else {
				if negativeFields {
					checkError(fmt.Errorf(`filed should not fixed with positive and negative fields`))
				}
			}
		}
		// 2 pass check
		if negativeFields {
			for _, f := range fields {
				if f > 0 {
					checkError(fmt.Errorf(`filed should not fixed with positive and negative fields`))
				}
			}
		}

		if !noHeaderRow {
			parseHeaderRow = true
		}
	} else {
		colnames = strings.Split(fieldsStr, ",")
		for _, f := range colnames {
			if f[0] == '-' {
				negativeFields = true
			} else {
				if negativeFields {
					checkError(fmt.Errorf(`filed should not fixed with positive and negative fields`))
				}
			}
		}
		// 2 pass check
		if negativeFields {
			for _, f := range colnames {
				if f[0] != '-' {
					checkError(fmt.Errorf(`filed should not fixed with positive and negative fields`))
				}
			}
		}
		if getFlagBool(cmd, "no-header-row") {
			log.Warningf("colnames detected, flag -H (--no-header-row) ignored")
		}
		parseHeaderRow = true
	}
	return fields, colnames, negativeFields, parseHeaderRow
}

func fuzzyField2Regexp(field string) *regexp.Regexp {
	if strings.IndexAny(field, "*") >= 0 {
		field = strings.Replace(field, "*", ".*?", -1)
	}

	field = "^" + field + "$"
	re, err := regexp.Compile(field)
	checkError(err)
	return re
}

func parseCSVfile(cmd *cobra.Command, config Config, file string,
	fieldStr string, fuzzyFields bool) ([]string, [][]string, []int) {
	fields, colnames, negativeFields, needParseHeaderRow := parseFields(cmd, fieldStr, config.NoHeaderRow)
	var fieldsMap map[int]struct{}
	if len(fields) > 0 {
		fields2 := make([]int, len(fields))
		fieldsMap = make(map[int]struct{}, len(fields))
		for i, f := range fields {
			if negativeFields {
				fieldsMap[f*-1] = struct{}{}
				fields2[i] = f * -1
			} else {
				fieldsMap[f] = struct{}{}
				fields2[i] = f
			}
		}
		fields = fields2
	}

	csvReader, err := newCSVReaderByConfig(config, file)
	checkError(err)
	csvReader.Run()

	parseHeaderRow := needParseHeaderRow // parsing header row
	var colnames2fileds map[string]int   // column name -> field
	var colnamesMap map[string]*regexp.Regexp

	var HeaderRow []string
	var Data [][]string

	checkFields := true

	for chunk := range csvReader.Ch {
		checkError(chunk.Err)

		for _, record := range chunk.Data {
			if parseHeaderRow { // parsing header row
				colnames2fileds = make(map[string]int, len(record))
				for i, col := range record {
					colnames2fileds[col] = i + 1
				}
				colnamesMap = make(map[string]*regexp.Regexp, len(colnames))
				for _, col := range colnames {
					if negativeFields {
						colnamesMap[col[1:]] = fuzzyField2Regexp(col)
					} else {
						colnamesMap[col] = fuzzyField2Regexp(col)
					}
				}

				if len(fields) == 0 { // user gives the colnames
					fields = []int{}
					for _, col := range record {
						var ok bool
						if fuzzyFields {
							for _, re := range colnamesMap {
								if re.MatchString(col) {
									ok = true
									break
								}
							}
						} else {
							_, ok = colnamesMap[col]
						}
						if (negativeFields && !ok) || (!negativeFields && ok) {
							fields = append(fields, colnames2fileds[col])
						}
					}
				}

				fieldsMap = make(map[int]struct{}, len(fields))
				for _, f := range fields {
					fieldsMap[f] = struct{}{}
				}

				parseHeaderRow = false
				HeaderRow = record
				continue
			}
			if checkFields {
				fields2 := []int{}
				for f := range record {
					_, ok := fieldsMap[f+1]
					if negativeFields {
						if !ok {
							fields2 = append(fields2, f+1)
						}
					} else {
						if ok {
							fields2 = append(fields2, f+1)
						}
					}
				}
				fields = fields2
				if len(fields) == 0 {
					checkError(fmt.Errorf("no fields matched in file: %s", file))
				}
				checkFields = false
			}

			Data = append(Data, record)
		}
	}
	return HeaderRow, Data, fields
}

func removeComma(s string) string {
	newSlice := []byte{}
	for i:=0; i<len(s); i++ {
		switch s[i] {
		case ',':
		default:
			newSlice = append(newSlice, s[i])
		}
	}
	return string(newSlice)
}
