package db

import (
	"fmt"
)

func GetShardTable(userID int64) string {
	return fmt.Sprintf("users_%d", userID%3)
}
