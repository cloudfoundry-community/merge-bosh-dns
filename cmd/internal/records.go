package internal

import (
	"fmt"
	"os"

	"github.com/cloudfoundry-community/merge-bosh-dns/merge"
)

func MergeAndWriteConfigs(local, remote *merge.RecordsConfig, destFilepath string) error {
	versionNum := uint64(0)

	destConfigFile, err := os.OpenFile(
		destFilepath,
		os.O_CREATE|os.O_RDWR,
		0640,
	)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Error opening destination file `%s': %s", destFilepath, err)
	}

	//If we have the file open, actually try to get the version number from the
	// file and then... we'll use the next number
	if err == nil {
		//If the file was just created, it will have a size of 0 and the JSON
		//unmarshaler will fail as a result. Therefore, we only want to invoke the
		//JSON marshaler if the file isn't empty.
		destConfigFileInfo, err := destConfigFile.Stat()
		if err != nil {
			return fmt.Errorf("Error when statting destination file `%s': %s", destFilepath, err)
		}

		if destConfigFileInfo.Size() > 0 {
			destConfigCurrent, err := merge.DecodeRecordsConfig(destConfigFile)
			if err != nil {
				return fmt.Errorf("Error parsing current destination config `%s': %s", destFilepath, err)
			}

			versionNum = destConfigCurrent.Version
		}
	}

	mergedConfig := merge.MergeRecordsConfigs(local, remote, versionNum)

	err = destConfigFile.Truncate(0)
	if err != nil {
		return fmt.Errorf("Error truncating current destination config `%s': %s", destFilepath, err)
	}

	fileContents, err := mergedConfig.Serialize()
	if err != nil {
		return fmt.Errorf("Error marshalling destination config into JSON to export: %s", err)
	}

	_, err = destConfigFile.WriteAt(fileContents, 0)
	if err != nil {
		return fmt.Errorf("Error writing JSON into destination config file `%s': %s", destFilepath, err)
	}

	return nil
}
