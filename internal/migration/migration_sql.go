package migration

import (
	"strings"
)

func parseSQLMigration(data []byte, name string) (*Migration, error) {
	content := string(data)
	upPart := extractSQLPart(content, "Up")
	downPart := extractSQLPart(content, "Down")

	upFunc := func() error {
		_, err := GetStore().Exec(upPart)
		return err
	}

	downFunc := func() error {
		_, err := GetStore().Exec(downPart)
		return err
	}

	return NewMigration(name, upFunc, downFunc), nil
}

func extractSQLPart(content, part string) string {
	startMarker := "-- +migrate " + part
	endIndex := strings.Index(content, startMarker)
	if endIndex == -1 {
		return ""
	}

	startIndex := endIndex + len(startMarker)
	endMarker := "-- +migrate "
	endIndex = strings.Index(content[startIndex:], endMarker)
	if endIndex == -1 {
		endIndex = len(content)
	} else {
		endIndex += startIndex
	}

	return strings.TrimSpace(content[startIndex:endIndex])
}
