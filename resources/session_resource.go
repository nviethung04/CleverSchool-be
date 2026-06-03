package resources

import (
	"be-lms/prot"
	"be-lms/redis"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type SessionResource interface {
	FormatSession(session *redis.SessionInfo) *prot.SessionInfo
	FormatSessions(sessions []*redis.SessionInfo) []*prot.SessionInfo
}

type SessionResourceImpl struct{}

func NewSessionResource() SessionResource {
	return &SessionResourceImpl{}
}

func (r *SessionResourceImpl) FormatSession(session *redis.SessionInfo) *prot.SessionInfo {
	if session == nil {
		return nil
	}

	return &prot.SessionInfo{
		Id:        session.Id,
		Token:     session.Token,
		Ip:        session.IP,
		UserAgent: session.UserAgent,
		Expires:   timestamppb.New(session.Expires),
		LoginAt:   timestamppb.New(session.LoginAt),
		IsCurrent: session.IsCurrent,
	}
}

func (r *SessionResourceImpl) FormatSessions(sessions []*redis.SessionInfo) []*prot.SessionInfo {
	result := make([]*prot.SessionInfo, 0, len(sessions))
	for _, s := range sessions {
		if formatted := r.FormatSession(s); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}
