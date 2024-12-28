package http

import "net/http"

func (s Server) getCurrentUserInfo(r *http.Request) (int, string) {
	return s.sessionManager.GetInt(r.Context(), authenticatedUserIdSessionKey), s.sessionManager.GetString(r.Context(), authenticatedUserNameSessionKey)
}
