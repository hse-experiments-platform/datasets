package domain

import (
	"fmt"
)

func GetBucketName(userID int64) string {
	return fmt.Sprintf("user-%d", userID)
}

func GetObjectName(datasetID int64) string {
	return fmt.Sprintf("%d.csv", datasetID)
}
