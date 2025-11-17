package tkValueObject

import (
	"errors"
	"log/slog"
	"regexp"

	tkVoUtil "github.com/goinfinite/tk/src/domain/valueObject/util"
)

var systemResourceIdentifierRegex = regexp.MustCompile(`^sri://(?P<accountId>\d{1,64}):(?P<resourceType>[a-zA-Z][\w-]{0,255})\/(?P<resourceId>([a-zA-Z0-9][\w\.\-]{0,511}|\*))$`)

// SystemResourceIdentifier is a string that represents the complete identifier
// of a system resource.
//
// It has the following format: sri://<accountId>:<resourceType>/<resourceId>
//
// - accountId: the account id of the resource, 0 if it's a system resource. It must be
// a positive integer.
//
// - resourceType: the type of the resource. It must start with a letter and can only
// contain letters, numbers, underscores, and hyphens. It has a maximum length of
// 256 characters.
//
// - resourceId: the id of the resource. It must start with a letter or number and can
// only contain letters, numbers, underscores, hyphens, and periods OR be a single
// asterisk ("*") to represent all resources of the type. It has a maximum length of
// 512 characters.
type SystemResourceIdentifier string

func NewSystemResourceIdentifier(value any) (sri SystemResourceIdentifier, err error) {
	stringValue, err := tkVoUtil.InterfaceToString(value)
	if err != nil {
		return sri, errors.New("SystemResourceIdentifierMustBeString")
	}

	if !systemResourceIdentifierRegex.MatchString(stringValue) {
		return sri, errors.New("InvalidSystemResourceIdentifier")
	}

	return SystemResourceIdentifier(stringValue), nil
}

// Note: This function is solely used for developer auditing to warn about the misuse of the
// SystemResourceIdentifier Value Object. It is not intended for user input validation and
// should not be used as a model for other Value Objects. It is not a standard and should not
// be followed for the development of other Value Objects.
func NewSystemResourceIdentifierMustCreate(value any) SystemResourceIdentifier {
	sri, err := NewSystemResourceIdentifier(value)
	if err != nil {
		panicMessage := "UnexpectedSystemResourceIdentifierCreationError"
		slog.Debug(panicMessage, slog.Any("value", value), slog.String("err", err.Error()))
		panic(panicMessage)
	}

	return sri
}

func (vo SystemResourceIdentifier) String() string {
	return string(vo)
}

func (vo SystemResourceIdentifier) readComponents() []string {
	return systemResourceIdentifierRegex.FindStringSubmatch(vo.String())
}

func (vo SystemResourceIdentifier) ReadAccountId() (accountId AccountId, err error) {
	sriComponents := vo.readComponents()
	if len(sriComponents) < 2 {
		return accountId, errors.New("SystemResourceIdentifierHasNoAccountId")
	}
	return NewAccountId(sriComponents[1])
}

func (vo SystemResourceIdentifier) ReadResourceType() (resourceType SystemResourceType, err error) {
	sriComponents := vo.readComponents()
	if len(sriComponents) < 3 {
		return resourceType, errors.New("SystemResourceIdentifierHasNoResourceType")
	}
	return NewSystemResourceType(sriComponents[2])
}

func (vo SystemResourceIdentifier) ReadResourceId() (resourceId SystemResourceId, err error) {
	sriComponents := vo.readComponents()
	if len(sriComponents) < 4 {
		return resourceId, errors.New("SystemResourceIdentifierHasNoResourceId")
	}
	return NewSystemResourceId(sriComponents[3])
}

func NewSriAccount(accountId AccountId) SystemResourceIdentifier {
	return NewSystemResourceIdentifierMustCreate(
		"sri://0:account/" + accountId.String(),
	)
}
