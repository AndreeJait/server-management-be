package echo

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	sshInbound "github.com/AndreeJait/server-management-be/port/inbound/ssh"
	"github.com/AndreeJait/go-utility/v2/authw"
	"github.com/gorilla/websocket"
)

func sshTerminalHandler(
	sshUC sshInbound.UseCase,
	authenticator authw.Authenticator,
	rbac *authw.RBAC,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		authReq := r.Clone(r.Context())
		authReq.Header.Set("Authorization", "Bearer "+token)
		result, err := authenticator.Authenticate(authReq)
		if err != nil || result == nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		userIDStr := result.GetUserID()
		ok, _ := rbac.CheckPermission(r.Context(), userIDStr, "ssh:connect")
		if !ok {
			http.Error(w, "insufficient permissions", http.StatusForbidden)
			return
		}

		hostIDStr := r.URL.Query().Get("hostId")
		if hostIDStr == "" {
			// Try path param as fallback
			hostIDStr = r.PathValue("hostId")
		}
		if hostIDStr == "" {
			http.Error(w, "missing hostId", http.StatusBadRequest)
			return
		}
		hostID, err := strconv.ParseUint(hostIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid hostId", http.StatusBadRequest)
			return
		}

		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid user id", http.StatusBadRequest)
			return
		}

		wsConn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer wsConn.Close()

		session, err := sshUC.Connect(r.Context(), uint(hostID), uint(userID))
		if err != nil {
			wsConn.WriteMessage(websocket.TextMessage, []byte("Error: failed to connect SSH session\r\n"))
			return
		}
		defer session.Close()

		wsConn.WriteMessage(websocket.TextMessage, []byte("\x1b[32mConnected to SSH terminal.\x1b[0m\r\n"))

		var once sync.Once
		done := make(chan struct{})

		// SSH stdout -> WebSocket
		go func() {
			defer once.Do(func() { close(done) })
			buf := make([]byte, 4096)
			for {
				n, err := session.Stdout().Read(buf)
				if err != nil {
					return
				}
				if n > 0 {
					if err := wsConn.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
						return
					}
				}
			}
		}()

		// WebSocket -> SSH stdin
		go func() {
			defer once.Do(func() { close(done) })
			for {
				_, msg, err := wsConn.ReadMessage()
				if err != nil {
					return
				}
				if len(msg) > 0 && msg[0] == '{' {
					var resize struct {
						Rows int `json:"rows"`
						Cols int `json:"cols"`
					}
					if json.Unmarshal(msg, &resize) == nil && resize.Rows > 0 && resize.Cols > 0 {
						_ = session.Resize(resize.Rows, resize.Cols)
					}
					continue
				}
				if _, err := session.Stdin().Write(msg); err != nil {
					return
				}
			}
		}()

		<-done
	}
}