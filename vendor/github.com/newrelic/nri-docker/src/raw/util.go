package raw

import (
	"bufio"
	"io"
	"os"
	"path"
	"strings"
)

const (
	procPathEnvVarName = "HOST_PROC"
	defaultProcPath    = "/proc"
	defaultMountsPath  = "/mounts"
)

// returns a path that is located on the root folder of the host and the `/host` folder
// on the integrations. If they existed in both root and /host, returns the /host path,
// assuming the integration is running in a container
func containerToHost(hostFolder, hostPath string) string {
	insideContainerPath := path.Join(hostFolder, hostPath)
	var err error
	if _, err = os.Stat(insideContainerPath); err == nil {
		return insideContainerPath
	}
	return hostPath
}

// getEnv will get the environment variable and return a string containing all extra values provided
// joined with the environment variable. If environment variable is not set, a default value will be used.
func getEnv(name, defaultValue string, combineWith ...string) string {
	value := os.Getenv(name)
	if value == "" {
		value = defaultValue
	}

	if len(combineWith) > 0 {
		value += path.Join(combineWith...)
	}

	return value
}

// getFirstExistingNonEmptyPath will return the first path in the array that can be accessed.
func getFirstExistingNonEmptyPath(paths []string) (result string, found bool) {
	for _, path := range paths {
		isEmpty, err := isDirEmpty(path)
		if err != nil || isEmpty {
			continue
		}

		result = path
		found = true
		break
	}
	return
}
func isDirEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	maxResults := 1
	_, err = f.Readdirnames(maxResults)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

type mount struct {
	Device     string
	MountPoint string
	FSType     string
	Options    string
}

// getMounts will parse the provided mounts file into mount structure.
func getMounts(file io.Reader) ([]*mount, error) {
	var result []*mount

	sc := bufio.NewScanner(file)
	for sc.Scan() {
		line := sc.Text()
		fields := strings.Fields(line)
		mount := &mount{
			Device:     fields[0],
			MountPoint: fields[1],
			FSType:     fields[2],
			Options:    fields[3],
		}
		result = append(result, mount)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// GetMountsFilePath will return the path to the system mounts file.
func GetMountsFilePath() string {
	return getEnv(procPathEnvVarName, defaultProcPath, defaultMountsPath)
}
