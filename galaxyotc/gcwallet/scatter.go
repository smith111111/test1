package gcwallet

import (
	"github.com/googollee/go-socket.io"
	"fmt"
	"gcwallet/scatter/api"
	"net/http"
	"gcwallet/log"
)

type Scatter struct {
	wsServer *socketio.Server
}

func (this *Scatter) Start() bool {
	log.WriteLog("Scatter, Start...")

	var err error

	this.wsServer, err = socketio.NewServer(nil)
	if err != nil {
		log.WriteLog(err.Error())
		return false
	}

	//WsServer.SetPingTimeout(100000000000000000)

	//nameSpace := WsServer.Of("/scatter")
	this.wsServer.On("connection", func(so socketio.Socket) {
		log.WriteLog(fmt.Sprintf("connection, sessionId: %v", so.Id()))

		so.Emit("connected")

		so.On("api", func(comm *api.CommonReq) {
			log.WriteLog(fmt.Sprintf("api, comm:%v", comm))

			req := &api.Request{}
			err := req.Unmarshal(comm)
			if err != nil {
				log.WriteLog(err.Error())
				return
			}

			log.WriteLog(fmt.Sprintf("api, req:%v", req))

			if req.Type == api.GET_OR_REQUEST_IDENTITY {
				accounts := make([]*api.IdentityAccount, 0)
				accounts = append(accounts, &api.IdentityAccount{
					Name:       "monopoly1111",
					Authority:  "active",
					PublicKey:  "PublicKey11111111111111",
					Blockchain: "eos",
					ChainId:"aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906",
					IsHardware: false,
				})

				rsp := &api.Response{}
				rsp.Id = req.Id

				result := &api.IdentityRsp{}
				result.Hash = "baec1db69091f81ddc9a0bdfb3920d9b0b39516c4c1ddbeff10063635638e5e8"
				result.PublicKey = "EOS7ZUKLadJTtp5veFUHQBibutKtTTU39BAFYdg1kAKvs5bzmA4bH"
				result.Name = "MyFirstIdentity"
				result.Kyc = false
				result.Accounts = accounts

				rsp.Result = result

				so.Emit("api", rsp)
			} else if req.Type == api.IDENTITY_FROM_PERMISSIONS {

			}
		})

		so.On("rekeyed", func(comm *api.CommonReq) {
			log.WriteLog(fmt.Sprintf("rekeyed, comm:%v", comm))

			req := &api.RekeyedReq{}
			err := req.Unmarshal(comm)
			if err != nil {
				log.WriteLog(err.Error())
				return
			}

			log.WriteLog(fmt.Sprintf("rekeyed, req:%v", req))

			so.Emit("rekeyed", "rekeyed ok")
		})

		so.On("pair", func(comm *api.CommonReq) {
			log.WriteLog(fmt.Sprintf("pair, comm:%v", comm))

			req := &api.PairReq{}
			err := req.Unmarshal(comm)
			if err != nil {
				log.WriteLog(err.Error())
				return
			}

			log.WriteLog(fmt.Sprintf("pair, comm:%v", req))

			if req.Passthrough {
				so.Emit("paired", nil)
			} else {
				so.Emit("paired", true)
			}
		})

		so.On("disconnection", func() {
			log.WriteLog(fmt.Sprintf("disconnection, sessionId: %v", so.Id()))
		})
	})

	this.wsServer.On("error", func(so socketio.Socket, err error) {
		log.WriteLog(err.Error())
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//origin := r.Header.Get("Origin")
		//w.Header().Set("Access-Control-Allow-Origin", origin)
		//w.Header().Set("Access-Control-Allow-Credentials", "true")

		log.WriteLog(fmt.Sprintf("url: %v", r.URL.String()))

		this.wsServer.ServeHTTP(w, r)
	})

	wsPort := ":50005"
	log.WriteLog(fmt.Sprintf("Start WebSocket %v", wsPort))

	http.Handle("/socket.io/", this.wsServer)
	err = http.ListenAndServe(wsPort, nil)
	if err != nil {
		log.WriteLog(err.Error())
		return false
	}

	return true
}

func (this *Scatter) Stop() bool {
	log.WriteLog("Scatter, Stop...")
	return true
}