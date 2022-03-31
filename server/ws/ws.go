package ws

import (
	"context"
	"io"
	"log"
	"net/http"
	pb "sehyoung/pb/gen"

	"github.com/gorilla/websocket"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/shserver/gopackage/shlog"
	"google.golang.org/grpc"
)

const (
	publicServiceAddress = "0.0.0.0:50002"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func RegisterWebsocketMux(mux *runtime.ServeMux) error {
	err := mux.HandlePath("GET", "/public/counsel", PublicCounsel)
	if err != nil {
		return err
	}
	return nil
}

func PublicCounsel(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	shlog.Logf("INFO", "START")
	// grpc connection
	rpcConn, err := grpc.Dial(publicServiceAddress, grpc.WithInsecure())
	if err != nil {
		shlog.Logf("ERROR", "Failed to connect to public service... %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rpcConn.Close()

	clnt := pb.NewPublicServiceClient(rpcConn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rpcStream, err := clnt.Counsel(ctx)
	if err != nil {
		shlog.Logf("ERROR", "Failed to request counsel to public service... %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// websocket connection
	wsConn, err := upgrader.Upgrade(w, r, nil)
	defer wsConn.Close()

	if err != nil {
		log.Println(err)
		return
	}

	wait := make(chan struct{})
	go func() {
		for {
			_, p, err := wsConn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					shlog.Logf("INFO", "ws connection normal close")
				} else {
					shlog.Logf("ERROR", "ws ReadMessage from client: %v", err)
				}
				rpcStream.CloseSend()
				cancel()
				break
			}
			if err := rpcStream.Send(&pb.RequestCounsel{Message: string(p)}); err != nil {
				shlog.Logf("ERROR", "Send to public service: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				break
			}
		}
	}()
	go func() {
		defer close(wait)
		for {
			res, err := rpcStream.Recv()

			select {
			case <-rpcStream.Context().Done():
				return
			default:
			}

			if err == io.EOF {
				shlog.Logf("INFO", "stream Recv EOF from public service")
				break
			} else if err != nil {
				shlog.Logf("ERROR", "recv from public service : ", err)
				w.WriteHeader(http.StatusInternalServerError)
				break
			}
			p := []byte(res.GetMessage())
			if err := wsConn.WriteMessage(websocket.TextMessage, p); err != nil {
				shlog.Logf("ERROR", "WriteMessage to client: ", err)
				return
			}
		}
	}()
	<-wait
	shlog.Logf("INFO", "END")
}
