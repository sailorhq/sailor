// sailor
// Copyright (C) 2025 SailorHQ and Ashish Shekar (codekidX)

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
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
	zipFileBytes, err := zipDir("./sails")
	if err != nil {
		return err
	}

	if len(zipFileBytes) == 0 {
		return fmt.Errorf("config folder is empty or the zipped content len is: %d", 0)
	}

	return UploadToS3(zipFileBytes, bucket, region, accessKey, secretKey, filePath)
}

func getFileName() string {
	spaceReplaced := strings.ReplaceAll(time.Now().Format(time.DateTime), " ", "-")
	colonReplaced := strings.ReplaceAll(spaceReplaced, ":", "-")
	return fmt.Sprintf("sailor-%s.zip", colonReplaced)
}
