package internal

// We want to be conservative about what columns we add into the table because space is limited.
// This function checks for columns that are never populated, and removes those columns from
// both the headers and the rows.
func FilterEmptyColumns(preFilteredHeaderRow []string, preFilteredRows [][]string) (filteredHeaders []string, filteredRows [][]string) {
	// these two slices are what will actually be set into the table
	var headersWithNonemptyColumns []string
	var rowsWithEmptyColumnsFiltered [][]string

	// for the headers that are NOT populated in any row in the table,
	// set the name of that header to an empty string as a marker that it should not be used
	for columnNum, header := range preFilteredHeaderRow {
		allCellsEmpty := true
		for _, row := range preFilteredRows {
			allCellsEmpty = allCellsEmpty && row[columnNum] == ""
		}

		if allCellsEmpty {
			preFilteredHeaderRow[columnNum] = ""
		} else {
			headersWithNonemptyColumns = append(headersWithNonemptyColumns, header)
		}
	}

	// for each row,
	//   for each column,
	//     if the header corresponding to that column has NOT been set to empty string,
	//       then include this value from the row
	for _, row := range preFilteredRows {
		filteredRow := []string{}
		for colNum, value := range row {
			if preFilteredHeaderRow[colNum] == "" {
				continue
			}

			filteredRow = append(filteredRow, value)
		}

		rowsWithEmptyColumnsFiltered = append(rowsWithEmptyColumnsFiltered, filteredRow)
	}

	return headersWithNonemptyColumns, rowsWithEmptyColumnsFiltered
}
