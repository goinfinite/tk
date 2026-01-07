package tkVoUtil

import (
	"regexp"
)

func NamedGroupsExtractor(regex *regexp.Regexp, input string) map[string]string {
	namedGroupsValueMap := make(map[string]string)

	matches := regex.FindStringSubmatch(input)
	if len(matches) == 0 {
		return namedGroupsValueMap
	}
	namedGroups := regex.SubexpNames()
	if len(namedGroups) == 0 {
		return namedGroupsValueMap
	}

	for namedGroupIndex, namedGroup := range namedGroups {
		if namedGroupIndex == 0 || namedGroup == "" || matches[namedGroupIndex] == "" {
			continue
		}
		namedGroupsValueMap[namedGroup] = matches[namedGroupIndex]
	}

	return namedGroupsValueMap
}
