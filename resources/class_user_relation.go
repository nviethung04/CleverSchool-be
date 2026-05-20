package resources

import (
	"be-cleverschool/dto"
	"be-cleverschool/prot"
)

func ClassUserRelationResponseResource(response dto.ClassUserRelationResponse) *prot.ClassUserRelationResponse {
	return &prot.ClassUserRelationResponse{
		Success: response.Success,
		Message: response.Message,
	}
} 
