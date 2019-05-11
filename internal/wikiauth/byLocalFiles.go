package wikiauth

import (
	"bufio"
	"github.com/ktt-ol/sesam/internal/conf"
	"github.com/sirupsen/logrus"
	"gopkg.in/hlandau/passlib.v1"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
)

var logger = logrus.WithField("where", "localFilesAuth")

type localFilesAuth struct {
	// contains only user in the group file
	// name -> pwHash
	nameToHashMap map[string]string
	// email address -> name
	emailToNameMap map[string]string
}

func NewLocalFilesAuth(config *conf.AuthLocal) WikiAuth {
	wd := localFilesAuth{make(map[string]string), make(map[string]string)}
	memberNames := wd.loadGroupData(config.GroupPageFile)
	wd.loadUserDataDir(config.UserDirectory, memberNames)
	return &wd
}

// CheckPassword checks the password for the given email or name
// if the login was successful (sucess = true) the userName is returned, even if the user logged in with an email
func (w *localFilesAuth) CheckPassword(emailOrName string, password string) (userName string, authError *AuthError) {
	userName = emailOrName
	if strings.Contains(emailOrName, "@") {
		nameFromMap, ok := w.emailToNameMap[emailOrName]
		if !ok {
			//errResult = errors.New("email doesn't exist (or not in Member group)")
			//loginNotFound = true
			return
		}
		userName = nameFromMap
	}

	pwHash, ok := w.nameToHashMap[userName]
	if !ok {
		//errResult = errors.New("user doesn't exist (or not in Member group)")
		//loginNotFound = true
		return
	}

	err := passlib.VerifyNoUpgrade(password, pwHash)
	if err != nil {
		// incorrect password, malformed hash, etc.
		// either way, reject
		//errResult = errors.New(fmt.Sprintf("invalid password (lib said: '%s'", err))
		return
	}

	return
}

func (w *localFilesAuth) loadGroupData(groupPageFile string) map[string]struct{} {
	file, err := os.Open(groupPageFile)
	defer file.Close()
	if err != nil {
		logger.WithError(err).WithField("groupPageFile", groupPageFile).Fatal("Can't open group file")
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	r := regexp.MustCompile(`^\s*\*\s*([\w\W]+)$`)
	memberNameSet := make(map[string]struct{})
	for scanner.Scan() {
		line := scanner.Text()
		match := r.FindStringSubmatch(line)
		if len(match) == 0 {
			continue
		}
		memberNameSet[match[1]] = struct{}{}
	}

	return memberNameSet
}

func (w *localFilesAuth) loadUserDataDir(userDirectory string, memberNames map[string]struct{}) {
	dirList, err := ioutil.ReadDir(userDirectory)
	if err != nil {
		logger.WithError(err).WithField("userDirectory", userDirectory).Fatal("Can't read simpleUser directory")
	}

	for _, entry := range dirList {
		if entry.IsDir() {
			continue
		}

		w.readUserFile(path.Join(userDirectory, entry.Name()), memberNames)
	}
}

func (w *localFilesAuth) readUserFile(userFile string, memberNames map[string]struct{}) {
	file, err := os.Open(userFile)
	defer file.Close()
	if err != nil {
		logger.WithError(err).WithField("userFile", userFile).Fatal("Can't open simpleUser file.")
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	const emailPrefix = "email="
	const pwPrefix = "enc_password={PASSLIB}"
	const namePrefix = "name="

	email := ""
	hash := ""
	name := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, emailPrefix) {
			email = line[len(emailPrefix):]
		} else if strings.HasPrefix(line, pwPrefix) {
			hash = line[len(pwPrefix):]
		} else if strings.HasPrefix(line, namePrefix) {
			name = line[len(namePrefix):]
		}
	}
	if len(email) == 0 || len(hash) == 0 || len(name) == 0 {
		//log.Println("Missing email/pw/name entry for ", file)
		return
	}

	_, isMember := memberNames[name]
	if isMember {
		w.nameToHashMap[name] = hash
		w.emailToNameMap[email] = name
	}
}
