package conf

import (
	"encoding/hex"
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"os"
)

func Test_saveAndLoad(t *testing.T) {
	assert := assert.New(t)

	tmpFile := TempFileName("keys_test", ".tmp")
	keys_created := createAndSaveNewKeys(tmpFile)
	keys_loaded := readKeysFromFile(tmpFile)
	assert.Equal(keys_created, keys_loaded)
}

// TempFileName generates a temporary filename for use in testing or whatever
func TempFileName(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes)+suffix)
}
