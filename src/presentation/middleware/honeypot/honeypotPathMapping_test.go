package tkPresentationMiddlewareHoneypot

import (
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestHoneypotPathMapping(t *testing.T) {
	pathPool := NewHoneypotPathPool()
	allPaths := make(
		map[tkValueObject.UrlPath]tkValueObject.HoneypotPathClass,
	)
	for _, class := range []tkValueObject.HoneypotPathClass{
		tkValueObject.HoneypotPathClassStaticVulnerability,
		tkValueObject.HoneypotPathClassBandwidthExhaust,
		tkValueObject.HoneypotPathClassAiTrap,
	} {
		for _, path := range pathPool.PathsByClass(class) {
			allPaths[path] = class
		}
	}

	mapping := NewHoneypotPathMapping(allPaths)

	t.Run("ResolvesHoneypotPath", func(t *testing.T) {
		wpConfigPath, _ := tkValueObject.NewUrlPath("/wp-config.php")
		pathClass, isHoneypot := mapping.Resolve(wpConfigPath)

		if !isHoneypot {
			t.Errorf("ExpectedHoneypotPath: '%s'", wpConfigPath.String())
		}
		if pathClass != tkValueObject.HoneypotPathClassStaticVulnerability {
			t.Errorf(
				"PathClassMismatch: Expected='staticVulnerability', Actual='%s'",
				pathClass.String(),
			)
		}
	})

	t.Run("NonHoneypotPathReturnsFalse", func(t *testing.T) {
		legitPath, _ := tkValueObject.NewUrlPath("/api/users")
		_, isHoneypot := mapping.Resolve(legitPath)

		if isHoneypot {
			t.Errorf(
				"UnexpectedHoneypotPath: '%s'",
				legitPath.String(),
			)
		}
	})

	t.Run("ResolvesDotEnvPath", func(t *testing.T) {
		envPath, _ := tkValueObject.NewUrlPath("/.env")
		pathClass, isHoneypot := mapping.Resolve(envPath)

		if !isHoneypot {
			t.Errorf("ExpectedHoneypotPath: '%s'", envPath.String())
		}
		if pathClass != tkValueObject.HoneypotPathClassStaticVulnerability {
			t.Errorf(
				"PathClassMismatch: Expected='staticVulnerability', Actual='%s'",
				pathClass.String(),
			)
		}
	})
}
