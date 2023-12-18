package redmine

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"golang.org/x/sys/unix"
)

// storeSilences will persist the data in JSON form to the given file.
// It requires that the basedir of the stateFile is writeable.
func storeSilences(stateFile string, data map[string]labels) error {
	dir, err := os.Open(path.Dir(stateFile))
	if err != nil {
		return err
	}
	// Close directory in any case
	defer func() {
		_ = dir.Close()
	}()

	if err := unix.Access(dir.Name(), unix.W_OK); err != nil {
		return fmt.Errorf("state-file directory %s not writeable: %w", dir.Name(), err)
	}

	// Create a new hidden temporary file to hold new state
	tmpfile, err := os.CreateTemp(dir.Name(), "."+path.Base(stateFile)+"-*.json")
	if err != nil {
		return err
	}

	// Remove temporary file in any case
	defer func() {
		// In case the state file has been written correctly, this will fail
		_ = os.Remove(tmpfile.Name())
	}()

	encoder := json.NewEncoder(tmpfile)
	if err := encoder.Encode(data); err != nil {
		// if data can't be written in full, cancel processing and remove the new invalid state
		return err
	}

	// Data is now in kernel's page cache, flush page cache to disk
	if err := tmpfile.Sync(); err != nil {
		// if data can't be flushed to disk, cancel processing and remove the new invalid state
		return err
	}

	if err := tmpfile.Close(); err != nil {
		return err
	}

	// replace the old state file
	if err := os.Rename(tmpfile.Name(), stateFile); err != nil {
		return err
	}

	// New state file has just been created, ensure the directory entry is flushed to disk
	if err := dir.Sync(); err != nil {
		return err
	}

	return nil
}

// loadSilences will read the data from the given JSON file.
func loadSilences(stateFile string, data *map[string]labels) error {
	file, err := os.Open(stateFile)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(data); err != nil {
		return err
	}

	return nil
}
