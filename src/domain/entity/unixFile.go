package entity

import (
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type UnixFile struct {
	Name        tkValueObject.UnixFileName         `json:"name"`
	Path        tkValueObject.UnixAbsoluteFilePath `json:"path"`
	MimeType    tkValueObject.MimeType             `json:"mimeType"`
	Permissions tkValueObject.UnixFilePermissions  `json:"permissions"`
	Size        tkValueObject.Byte                 `json:"size"`
	Extension   *tkValueObject.UnixFileExtension   `json:"extension"`
	Uid         tkValueObject.UnixUserId           `json:"uid"`
	Owner       tkValueObject.UnixUsername         `json:"owner"`
	Gid         tkValueObject.UnixGroupId          `json:"gid"`
	Group       tkValueObject.UnixGroupName        `json:"group"`
	UpdatedAt   tkValueObject.UnixTime             `json:"updatedAt"`
	IsSymlink   bool                               `json:"isSymlink"`
}

func NewUnixFile(
	name tkValueObject.UnixFileName,
	path tkValueObject.UnixAbsoluteFilePath,
	mimeType tkValueObject.MimeType,
	permissions tkValueObject.UnixFilePermissions,
	size tkValueObject.Byte,
	extension *tkValueObject.UnixFileExtension,
	uid tkValueObject.UnixUserId,
	owner tkValueObject.UnixUsername,
	gid tkValueObject.UnixGroupId,
	group tkValueObject.UnixGroupName,
	updatedAt tkValueObject.UnixTime,
	isSymlink bool,
) UnixFile {
	return UnixFile{
		Name:        name,
		Path:        path,
		MimeType:    mimeType,
		Permissions: permissions,
		Size:        size,
		Extension:   extension,
		Uid:         uid,
		Owner:       owner,
		Gid:         gid,
		Group:       group,
		UpdatedAt:   updatedAt,
		IsSymlink:   isSymlink,
	}
}
