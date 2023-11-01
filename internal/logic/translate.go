package logic

import (
	"log"
	"openai/internal/service/openaiex"
)

func transToEngEx(original string) (trans string, err error) {
	for _, vendor := range aiVendors {
		trans, err = openaiex.TransToEng(original, vendor)
		if err == nil {
			break
		}
		log.Printf("openaiex.transToEng(%s) failed %v", vendor, err)
	}
	return
}
