package v1

import (
	"database/sql"
	"time"

	"github.com/RealistikOsu/RealistikAPI/common"
)

type pinnedScore struct {
	common.ResponseBase
	Scores []struct {
		Score

		Beatmap beatmap `json:"beatmap"`
	} `json:"scores"`
}

const pinnedQuery = `
SELECT
    scores.id, scores.beatmap_md5, scores.score,
    scores.max_combo, scores.full_combo, scores.mods,
    scores.300_count, scores.100_count, scores.50_count,
    scores.gekis_count, scores.katus_count, scores.misses_count,
    scores.time, scores.play_mode, scores.accuracy, scores.pp,
    scores.completed, users_pinned.pin_date
    beatmaps.beatmap_id, beatmaps.beatmapset_id, beatmaps.beatmap_md5,
    beatmaps.song_name, beatmaps.ar, beatmaps.od, beatmaps.difficulty_std,
    beatmaps.difficulty_taiko, beatmaps.difficulty_ctb, beatmaps.difficulty_mania,
    beatmaps.max_combo, beatmaps.hit_length, beatmaps.ranked,
    beatmaps.ranked_status_freezed, beatmaps.latest_update
FROM user_pinned
JOIN scores ON scores.id = scoreid
JOIN beatmaps on scores.beatmap_md5 = beatmaps.beatmap_md5
WHERE user_pinned.userid = 1000 AND scores.mode = 1
`

func UserPinnedGET() {
	panic("unimplemented")
}

func UserUnpinDELETE() {
	panic("unimplemented")
}

func UserPinnedPOST(md common.MethodData) common.CodeMessager {
	var scoreExists bool
	var scorePinned bool
	var pinDatetime int64 = time.Now().Unix()

	mode := common.Int(md.Query("playmode"))
	exists := []string{
		"SELECT EXISTS(SELECT 1 FROM scores WHERE id = ? AND completed != 0 AND userid = ?);",
		"SELECT EXISTS(SELECT 1 FROM scores_relax WHERE id = ? AND completed != 0 AND userid = ?);",
		"SELECT EXISTS(SELECT 1 FROM scores_ap WHERE id = ? AND completed != 0 AND userid = ?);",
	}

	if mode < 0 || mode > 2 {
		return common.SimpleResponse(400, "Invalid mode")
	}

	res := common.ResponseBase{}
	scoreid := common.Int(md.Query("score_id"))
	err := md.DB.QueryRow(exists[mode], scoreid, md.User.UserID).Scan(&scoreExists)

	if err != nil && err != sql.ErrNoRows {
		md.Err(err)
		return Err500
	} else if !scoreExists {
		return common.SimpleResponse(400, "Can't pin a score that does not exist.")
	}

	// score already pinned?
	err = md.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM user_pinned WHERE scoreid = ?)", scoreid).Scan(&scorePinned)
	if err != nil && err != sql.ErrNoRows {
		md.Err(err)
		return Err500
	} else if scorePinned {
		return common.SimpleResponse(400, "Already pinned...")
	}

	_, err = md.DB.Exec("INSERT INTO user_pinned (userid, scoreid, pin_date) VALUES (?, ?, ?)", md.User.UserID, scoreid, pinDatetime)
	if err != nil {
		md.Err(err)
		return Err500
	}

	res.Message = "success!"
	return res
}
