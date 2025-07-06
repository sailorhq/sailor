package backup

import (
	"fmt"
	"strings"
	"time"
)

// BackupRawSails backs up the core sails of sailer to S3
func BackupRawSails(bucket, region, accessKey, secretKey string) error {
	zipName := getFileName()
	filePath := "sails/" + zipName

	// TODO :: configs folder should be fetched from environment variable, because docker can volume
	// mount custom folders
	zipFileBytes, err := zipDir("./configs")
	if err != nil {
		return err
	}

	if len(zipFileBytes) == 0 {
		return fmt.Errorf("config folder is empty or the zipped content len is: %d", 0)
	}

	return uploadToS3(zipFileBytes, bucket, region, accessKey, secretKey, filePath)
}

func getFileName() string {
	spaceReplaced := strings.ReplaceAll(time.Now().Format(time.DateTime), " ", "-")
	colonReplaced := strings.ReplaceAll(spaceReplaced, ":", "-")
	return fmt.Sprintf("sailor-%s.zip", colonReplaced)
}
