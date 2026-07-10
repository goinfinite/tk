package tkUseCase

import (
	"log/slog"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkRepository "github.com/goinfinite/tk/src/domain/repository"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func resolveBanThreshold(
	mode tkValueObject.HoneypotAggressivenessMode,
) uint64 {
	switch mode {
	case tkValueObject.HoneypotAggressivenessModeImmediate:
		return 1
	case tkValueObject.HoneypotAggressivenessModeBalanced:
		return 3
	case tkValueObject.HoneypotAggressivenessModeTolerant:
		return 10
	case tkValueObject.HoneypotAggressivenessModeObserve:
		return 0
	default:
		return 0
	}
}

func resolveSuggestedAction(
	mode tkValueObject.HoneypotAggressivenessMode,
	hitCount uint64,
	threshold uint64,
) tkValueObject.HoneypotSuggestedAction {
	if threshold > 0 && hitCount >= threshold {
		return tkValueObject.HoneypotSuggestedActionBan
	}

	switch mode {
	case tkValueObject.HoneypotAggressivenessModeImmediate:
		return tkValueObject.HoneypotSuggestedActionServePayload
	case tkValueObject.HoneypotAggressivenessModeBalanced:
		return tkValueObject.HoneypotSuggestedActionServeMixed
	case tkValueObject.HoneypotAggressivenessModeTolerant:
		return tkValueObject.HoneypotSuggestedActionServeStream
	case tkValueObject.HoneypotAggressivenessModeObserve:
		return tkValueObject.HoneypotSuggestedActionPassthrough
	default:
		return tkValueObject.HoneypotSuggestedActionPassthrough
	}
}

func ReadHoneypotBanDecision(
	honeypotQueryRepo tkRepository.HoneypotQueryRepo,
	requestDto tkDto.ReadHoneypotBanDecisionRequest,
	settings tkDto.HoneypotSettings,
) (tkDto.ReadHoneypotBanDecisionResponse, error) {
	responseDto, err := honeypotQueryRepo.ReadBanDecision(requestDto)
	if err != nil {
		slog.Error(
			"ReadHoneypotBanDecisionInfraError",
			slog.String("err", err.Error()),
		)
		return responseDto, err
	}

	mode := settings.AggressivenessMode
	threshold := resolveBanThreshold(mode)
	isBanned := threshold > 0 && responseDto.HitCount >= threshold
	suggestedAction := resolveSuggestedAction(
		mode, responseDto.HitCount, threshold,
	)

	responseDto.IsBanned = isBanned
	responseDto.SuggestedAction = suggestedAction

	return responseDto, nil
}
