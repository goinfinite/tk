package tkValueObject

import (
	"errors"
	"os"
	"regexp"
	"strconv"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var unixFilePermissionsRegex = regexp.MustCompile(`^[0-7]{3,4}$`)

type UnixFilePermissions string

func NewUnixFilePermissions(value any) (
	unixFilePermissions UnixFilePermissions, err error,
) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return unixFilePermissions, errors.New("UnixFilePermissionsMustBeString")
	}

	if !unixFilePermissionsRegex.MatchString(stringValue) {
		return unixFilePermissions, errors.New("InvalidUnixFilePermissions")
	}

	return UnixFilePermissions(stringValue), nil
}

func (vo UnixFilePermissions) ToFileMode() (os.FileMode, error) {
	const (
		setuidPermissionBit = 0o4000
		setgidPermissionBit = 0o2000
		stickyPermissionBit = 0o1000
	)

	permissionBits, parseErr := strconv.ParseUint(string(vo), 8, 12)
	if parseErr != nil {
		return 0, errors.New("InvalidUnixFilePermissions")
	}

	fileMode := os.FileMode(permissionBits).Perm()
	if permissionBits&setuidPermissionBit != 0 {
		fileMode |= os.ModeSetuid
	}
	if permissionBits&setgidPermissionBit != 0 {
		fileMode |= os.ModeSetgid
	}
	if permissionBits&stickyPermissionBit != 0 {
		fileMode |= os.ModeSticky
	}

	return fileMode, nil
}

func (vo UnixFilePermissions) String() string {
	return string(vo)
}
