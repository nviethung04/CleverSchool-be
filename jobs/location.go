package jobs

import (
	"be-cleverschool/config"
	"time"
)

func Location() *time.Location {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		config.Log.Warn("Không load được timezone Asia/Ho_Chi_Minh, dùng mặc định:", err)
		loc = time.Local
	}

	return loc
}

