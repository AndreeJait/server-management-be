package echo

import (
	"net/http"
	"sync"

	"github.com/AndreeJait/server-management-be/port/inbound/app"
	"github.com/AndreeJait/server-management-be/port/inbound/deployment"
	"github.com/AndreeJait/go-utility/v2/authw"
	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func terminalHandler(
	deployUC deployment.UseCase,
	appUC app.UseCase,
	authenticator authw.Authenticator,
	rbac *authw.RBAC,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Authenticate via query param token
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		// Create a new request with the token in the Authorization header for auth
		authReq := r.Clone(r.Context())
		authReq.Header.Set("Authorization", "Bearer "+token)
		result, err := authenticator.Authenticate(authReq)
		if err != nil || result == nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Check apps:deploy permission
		userID := result.GetUserID()
		ok, _ := rbac.CheckPermission(r.Context(), userID, "apps:deploy")
		if !ok {
			http.Error(w, "insufficient permissions", http.StatusForbidden)
			return
		}

		// Extract app ID from query param
		appIDStr := r.URL.Query().Get("appId")
		if appIDStr == "" {
			http.Error(w, "missing appId", http.StatusBadRequest)
			return
		}

		// Find running container
		containerID, err := deployUC.GetRunningContainerID(r.Context(), appIDStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		// Determine shell to use
		shell := r.URL.Query().Get("shell")
		if shell == "" {
			shell = "/bin/sh"
		}

		// Upgrade to WebSocket
		wsConn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer wsConn.Close()

		conn, reader, err := deployUC.ExecInteractive(r.Context(), appIDStr, containerID, shell)
		if err != nil {
			// If requested shell fails, try /bin/sh as fallback
			if shell != "/bin/sh" {
				conn, reader, err = deployUC.ExecInteractive(r.Context(), appIDStr, containerID, "/bin/sh")
			}
			if err != nil {
				wsConn.WriteMessage(websocket.TextMessage, []byte("Error: failed to create terminal session\r\n"))
				return
			}
		}
		defer conn.Close()

		wsConn.WriteMessage(websocket.TextMessage, []byte("\x1b[32mConnected to container terminal.\x1b[0m\r\n"))

		// Pump data between WebSocket and Docker
		var once sync.Once
		done := make(chan struct{})

		// Docker -> WebSocket
		go func() {
			defer once.Do(func() { close(done) })
			buf := make([]byte, 4096)
			for {
				n, err := reader.Read(buf)
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

		// WebSocket -> Docker
		go func() {
			defer once.Do(func() { close(done) })
			for {
				_, msg, err := wsConn.ReadMessage()
				if err != nil {
					return
				}
				if len(msg) > 0 && msg[0] == '{' {
					// Resize event — skip for now
					continue
				}
				if _, err := conn.Write(msg); err != nil {
					return
				}
			}
		}()

		<-done
	}
}