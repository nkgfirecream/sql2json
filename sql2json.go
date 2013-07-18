package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"strings"
)

func Convert(r io.Reader) []byte {
	tables := parse(r)
	return tables.ToJSON()
}

type DummyTable struct {
	Name    string
	Columns []string
	Rows    [][]string
}

func (d *DummyTable) RowMap(r int) map[string]string {
	row := d.Rows[r]

	m := make(map[string]string)
	for i := 0; i < len(row); i++ {
		m[d.Columns[i]] = row[i]
	}

	return m
}

type DummyTables []*DummyTable

func (d DummyTables) ToJSON() []byte {
	jsonTables := make(map[string][]map[string]string)

	for _, table := range d {
		for i := 0; i < len(table.Rows); i++ {
			jsonTables[table.Name] = append(jsonTables[table.Name], table.RowMap(i))
		}
	}

	data, err := json.MarshalIndent(jsonTables, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	return data
}

func parse(r io.Reader) DummyTables {

	data, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	dummyTables := make(DummyTables, 0)

	str := string(data)

	// TODO: Lots of ugly string splitting that needs to be taken care of.
	tables := strings.Split(str, "-- Table structure for table `")[1:]
	for _, table := range tables {
		tableName := strings.SplitN(table, "`", 2)[0]

		dummyTable := DummyTable{tableName, make([]string, 0), make([][]string, 0)}
		dummyTables = append(dummyTables, &dummyTable)

		columnsRaw := strings.Split(table, "  `")
		for _, columnRaw := range columnsRaw {
			column := strings.SplitN(columnRaw, "`", 2)[0]
			dummyTable.Columns = append(dummyTable.Columns, column)
		}
		dummyTable.Columns = dummyTable.Columns[1:]

		valuesData := strings.Split(table, "` VALUES (")
		if len(valuesData) < 2 {
			continue
		}
		values := valuesData[1]

		newRow := make([]string, 0)

		open := false
		openString := false
		start := 0

		i := 0
		for {
			if i >= len(values) {
				break
			}

			if values[i] == ')' {
				// Special case if number and last element
				if open && !openString {
					// Close and add as usual
					newRow = append(newRow, values[start:i])
					open = false
					i++

					// Add row
					dummyTable.Rows = append(dummyTable.Rows, newRow)
					newRow = make([]string, 0)
					continue
				} else if !open {
					// Means we're ending after a string, add row
					dummyTable.Rows = append(dummyTable.Rows, newRow)
					newRow = make([]string, 0)
					i++
					continue
				}
			}

			// Null
			if !open && values[i:i+4] == "NULL" {
				newRow = append(newRow, values[i:i+4])
				i = i + 4
				continue
			}

			//// Numbers
			// Open number
			numbers := "0123456789-."
			if !open && strings.Contains(numbers, string(values[i])) {
				start = i
				open = true
				i++
				continue
			}

			// End number
			if open && !openString && values[i] == ',' {
				newRow = append(newRow, values[start:i])
				open = false
				i++
				continue
			}

			//// Strings
			// Open string
			if !openString && values[i] == '\'' {
				start = i + 1
				open = true
				openString = true
				i++
				continue
			}

			// End string
			if openString && values[i] == '\'' {
				newRow = append(newRow, values[start:i])
				open = false
				openString = false
				i++
				continue
			}
			i++
		}

	}

	return dummyTables
}
