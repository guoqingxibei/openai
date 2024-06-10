package logic

import (
	"log/slog"
	"openai/internal/service/openaiex"
)

func transToEngEx(original string) (trans string, err error) {
	for _, vendor := range aiVendors {
		trans, err = openaiex.TransToEng(original, vendor)
		if err == nil {
			break
		}
		slog.Error("openaiex.TransToEng() failed", "vendor", vendor, "error", err)
	}
	return
}
