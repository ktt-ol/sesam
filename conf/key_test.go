package conf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"log"
)

func Test_saveAndLoad(t *testing.T) {
	assert := assert.New(t)

	tmpFile, err := ioutil.TempFile(os.TempDir(), "keys_test")
	if err != nil {
		log.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	keys_created := createAndSaveNewKeys(tmpFile.Name())
	keys_loaded := readKeysFromFile(tmpFile.Name())
	assert.Equal(keys_created, keys_loaded)
}
