package tkPresentationMiddlewareHoneypot

import (
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type HoneypotPathMapping struct {
	activePaths map[tkValueObject.UrlPath]tkValueObject.HoneypotPathClass
}

func NewHoneypotPathMapping(
	activePaths map[tkValueObject.UrlPath]tkValueObject.HoneypotPathClass,
) *HoneypotPathMapping {
	return &HoneypotPathMapping{activePaths: activePaths}
}

func (mapping *HoneypotPathMapping) Resolve(
	requestPath tkValueObject.UrlPath,
) (tkValueObject.HoneypotPathClass, bool) {
	pathClass, isHoneypot := mapping.activePaths[requestPath]
	return pathClass, isHoneypot
}
