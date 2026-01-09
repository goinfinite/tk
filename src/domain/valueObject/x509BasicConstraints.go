package tkValueObject

import "errors"

type X509BasicConstraints struct {
	IsAuthority   bool `json:"isAuthority"`
	MaxPathLength *int `json:"maxPathLength"`
}

func NewX509BasicConstraints(
	isAuthority bool, maxPathLength *int,
) (constraints X509BasicConstraints, err error) {
	if !isAuthority && maxPathLength != nil {
		return constraints, errors.New(
			"InvalidX509BasicConstraintsMaxPathLengthForNonCA",
		)
	}

	if maxPathLength != nil && *maxPathLength < 0 {
		return constraints, errors.New("InvalidX509BasicConstraintsMaxPathLength")
	}

	return X509BasicConstraints{
		IsAuthority:   isAuthority,
		MaxPathLength: maxPathLength,
	}, nil
}
