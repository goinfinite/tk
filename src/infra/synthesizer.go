package tkInfra

import (
	"math/rand"
	"strings"
)

const (
	CharsetLowercaseLetters string = "abcdefghijklmnopqrstuvwxyz"
	CharsetUppercaseLetters string = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharsetNumbers          string = "0123456789"
	CharsetSymbols          string = "!@#$%^&*()_+"
)

type Synthesizer struct {
}

func (synth *Synthesizer) CharsetPresenceGuarantor(
	originalString []byte,
	charset string,
) []byte {
	if strings.ContainsAny(string(originalString), charset) {
		return originalString
	}

	randomStringIndex := rand.Intn(len(originalString))
	isFirstChar := randomStringIndex == 0
	if isFirstChar {
		randomStringIndex++
	}
	isLastChar := randomStringIndex == len(originalString)-1
	if isLastChar {
		randomStringIndex--
	}
	if randomStringIndex >= len(originalString) {
		randomStringIndex = len(originalString) - 1
	}

	randomCharsetIndex := rand.Intn(len(charset))
	originalString[randomStringIndex] = charset[randomCharsetIndex]

	return originalString
}

func (synth *Synthesizer) PasswordFactory(
	desiredLength int,
	shouldIncludeSymbols bool,
) string {
	alphanumericCharset := CharsetLowercaseLetters + CharsetUppercaseLetters + CharsetNumbers
	alphanumericCharsetLength := len(alphanumericCharset)

	passwordBytes := make([]byte, desiredLength)
	for charIdx := 0; charIdx < desiredLength; charIdx++ {
		passwordBytes[charIdx] = alphanumericCharset[rand.Intn(alphanumericCharsetLength)]
	}

	if desiredLength > 4 {
		passwordBytes = synth.CharsetPresenceGuarantor(passwordBytes, CharsetLowercaseLetters)
		passwordBytes = synth.CharsetPresenceGuarantor(passwordBytes, CharsetUppercaseLetters)
		passwordBytes = synth.CharsetPresenceGuarantor(passwordBytes, CharsetNumbers)
	}

	if shouldIncludeSymbols {
		passwordBytes = synth.CharsetPresenceGuarantor(passwordBytes, CharsetSymbols)
	}

	return string(passwordBytes)
}

func (synth *Synthesizer) UsernameFactory() string {
	dummyUsernames := []string{
		"pike", "spock", "kirk", "scotty", "bones", "uhura", "sulu", "chekov",
	}
	return dummyUsernames[rand.Intn(len(dummyUsernames))]
}

func (synth *Synthesizer) MailAddressFactory(username *string) string {
	if username == nil {
		dummyUsername := synth.UsernameFactory()
		username = &dummyUsername
	}

	atDomains := []string{
		"@ufp.gov", "@starfleet.gov", "@academy.edu", "@terran.gov",
	}
	return *username + atDomains[rand.Intn(len(atDomains))]
}
