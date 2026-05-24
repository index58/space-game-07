package data

import (
	"bytes"
	"encoding/json"
	"os"
	"sort"
	"strconv"
)

// Сохраняет таблицу так, чтобы ключи основного хранилища шли по числовому порядку.
func saveTableWithOrderedItems[T any](path string, maxID int64, items map[int64]*T) error {
	content, err := marshalTableWithOrderedItems(maxID, items)
	if err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o600)
}

// Формирует JSON с устойчивым порядком записей по их числовому идентификатору.
func marshalTableWithOrderedItems[T any](maxID int64, items map[int64]*T) ([]byte, error) {
	var buffer bytes.Buffer

	buffer.WriteString("{\n")
	buffer.WriteString(`  "MaxID": `)
	buffer.WriteString(strconv.FormatInt(maxID, 10))
	buffer.WriteString(",\n")
	buffer.WriteString(`  "Items": `)

	ids := sortedTableItemIDs(items)
	if len(ids) == 0 {
		buffer.WriteString("{}\n")
		buffer.WriteString("}")
		return buffer.Bytes(), nil
	}

	buffer.WriteString("{\n")
	for index, id := range ids {
		itemContent, err := json.MarshalIndent(items[id], "    ", "  ")
		if err != nil {
			return nil, err
		}

		buffer.WriteString(`    "`)
		buffer.WriteString(strconv.FormatInt(id, 10))
		buffer.WriteString(`": `)
		buffer.Write(itemContent)
		if index < len(ids)-1 {
			buffer.WriteString(",")
		}
		buffer.WriteString("\n")
	}
	buffer.WriteString("  }\n")
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

// Возвращает ключи основного хранилища по возрастанию.
func sortedTableItemIDs[T any](items map[int64]*T) []int64 {
	ids := make([]int64, 0, len(items))
	for id := range items {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool {
		return ids[left] < ids[right]
	})
	return ids
}
