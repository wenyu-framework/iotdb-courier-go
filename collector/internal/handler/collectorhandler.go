package handler

import (
	"io"
	"iotdb-courier/collector/internal/types"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"iotdb-courier/collector/internal/logic"
	"iotdb-courier/collector/internal/svc"
)

func CollectorHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json, err := io.ReadAll(r.Body)
		if err != nil {
			return
		}

		var req types.AcceptRequest
		if err := httpx.ParseForm(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewCollectorLogic(r.Context(), svcCtx)
		resp, err := l.Accept(&req, &json)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
