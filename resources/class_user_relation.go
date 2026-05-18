package resources

import (
	"be-Clever School/dto"
	"be-Clever School/prot"
)

func ClassUserRelationResponseResource(response dto.ClassUserRelationResponse) *prot.ClassUserRelationResponse {
	return &prot.ClassUserRelationResponse{
		Success: response.Success,
		Message: response.Message,
	}
} 