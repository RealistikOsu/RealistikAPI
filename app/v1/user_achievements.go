package v1

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/RealistikOsu/RealistikAPI/common"
	"github.com/jmoiron/sqlx"
)

// Achievement represents an achievement in the database.
type Achievement struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
	File string `json:"file"`
}

// LoadAchievementsEvery reloads the achievements in the database every given
// amount of time.
func LoadAchievementsEvery(db *sqlx.DB, d time.Duration) {
	for {
		achievs = nil
		err := db.Select(&achievs,
			"SELECT `id`, `name`, `desc`, `file` FROM `ussr_achievements` ORDER BY `id` ASC")
		if err != nil {
			fmt.Println("LoadAchievements error", err)
			common.GenericError(err)
		}
		time.Sleep(d)
	}
}

var achievs []Achievement

type userAchievement struct {
	Achievement
	Achived bool            `json:"achived"`
	Info    AchievementUser `json:"info"`
}

type userAchievementsResponse struct {
	common.ResponseBase
	Achievements []userAchievement `json:"achievements"`
}

type AchievementUser struct {
	ID   int `json:"id"`
	Time int `json:"time"`
}

// UserAchievementsGET handles requests for retrieving the achievements of a
// given user.
func UserAchievementsGET(md common.MethodData) common.CodeMessager {
	shouldRet, _, _ := whereClauseUser(md, "users")
	if shouldRet != nil {
		return *shouldRet
	}

	var users_achi []AchievementUser
	err := md.DB.Select(&users_achi, "SELECT achievement_id AS id, time FROM users_achievements WHERE user_id = ? ORDER by achievement_id ASC", md.Query("id"))
	switch {
	case err == sql.ErrNoRows:
		return common.SimpleResponse(404, "No such user!")
	case err != nil:
		md.Err(err)
		return Err500
	}

	var achi_ids2 []int
	for _, achi := range users_achi {
		achi_ids2 = append(achi_ids2, achi.ID)
	}

	resp := userAchievementsResponse{Achievements: make([]userAchievement, 0, len(achievs))}
	for _, achi := range achievs {
		// for _, achi2 := range users_achi {
		// 	// if inInt(achi.ID, achi2) {
		// 	// 	resp.Achievements = append(resp.Achievements, userAchievement{achi, achi2})
		// 	// } else {
		// 	// 	resp.Achievements = append(resp.Achievements, userAchievement{achi, AchievementUser{achi.ID, 0}})
		// 	// }

		// }
		achived, curr_achi := inInt(achi.ID, users_achi)
		resp.Achievements = append(resp.Achievements, userAchievement{achi, achived, curr_achi})
		// if achived {
		// 	resp.Achievements = append(resp.Achievements, userAchievement{achi, curr_achi})
		// } else {
		// 	resp.Achievements = append(resp.Achievements, userAchievement{achi, AchievementUser{achi.ID, 0}})
		// }
		// if inInt(achi.ID, users_achi) {
		// 	for _, in achi
		// }
	}

	resp.Code = 200
	return resp

	// 	var ids []int
	// 	err := md.DB.Select(&ids, `SELECT ua.achievement_id FROM users_achievements ua
	// INNER JOIN users ON users.id = ua.user_id
	// WHERE `+whereClause+` ORDER BY ua.achievement_id ASC`, param)
	// 	switch {
	// 	case err == sql.ErrNoRows:
	// 		return common.SimpleResponse(404, "No such user!")
	// 	case err != nil:
	// 		md.Err(err)
	// 		return Err500
	// 	}
	// 	all := md.HasQuery("all")
	// 	resp := userAchievementsResponse{Achievements: make([]userAchievement, 0, len(achievs))}
	// 	for _, ach := range achievs {
	// 		achieved := inInt(ach.ID, ids)
	// 		if all || achieved {
	// 			resp.Achievements = append(resp.Achievements, userAchievement{ach, achieved})
	// 		}
	// 	}
	// 	resp.Code = 200
	// 	return resp
}

func inInt(i int, js []AchievementUser) (bool, AchievementUser) {
	for _, j := range js {
		if i == j.ID {
			return true, j
		}
	}
	return false, AchievementUser{i, 0}
}
