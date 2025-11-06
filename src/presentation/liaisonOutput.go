package tkPresentation

type LiaisonOutputStatus string

const (
	LiaisonOutputSuccess      LiaisonOutputStatus = "success"
	LiaisonOutputCreated      LiaisonOutputStatus = "created"
	LiaisonOutputMultiStatus  LiaisonOutputStatus = "multiStatus"
	LiaisonOutputUserError    LiaisonOutputStatus = "userError"
	LiaisonOutputUnauthorized LiaisonOutputStatus = "unauthorized"
	LiaisonOutputForbidden    LiaisonOutputStatus = "forbidden"
	LiaisonOutputInfraError   LiaisonOutputStatus = "infraError"
	LiaisonOutputUnknownError LiaisonOutputStatus = "unknownError"
)

type LiaisonOutput struct {
	Status LiaisonOutputStatus `json:"status"`
	Body   any                 `json:"body"`
}

func NewLiaisonOutput(status LiaisonOutputStatus, body any) LiaisonOutput {
	return LiaisonOutput{
		Status: status,
		Body:   body,
	}
}
