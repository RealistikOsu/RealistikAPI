package v1

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/RealistikOsu/RealistikAPI/common"
)

func UserPinnedGET(md common.MethodData) common.CodeMessager {
	scoreResponse := userScoresResponse{}
	dbs := []string{"", "_relax", "_ap"}

	playmode := common.Int(md.Query("playmode"))
	mode := common.Int(md.Query("mode"))
	userid := common.Int(md.Query("id"))

	if playmode < 0 || playmode > 2 {
		return common.SimpleResponse(400, "Invalid playmode")
	}

	if mode < 0 || mode > 3 {
		return common.SimpleResponse(400, "Invalid mode")
	}

	query := fmt.Sprintf(`
		SELECT
			scores%[1]s.id, scores%[1]s.beatmap_md5, scores%[1]s.score,
			scores%[1]s.max_combo, scores%[1]s.full_combo, scores%[1]s.mods,
			scores%[1]s.300_count, scores%[1]s.100_count, scores%[1]s.50_count,
			scores%[1]s.gekis_count, scores%[1]s.katus_count, scores%[1]s.misses_count,
			scores%[1]s.time, scores%[1]s.play_mode, scores%[1]s.accuracy, scores%[1]s.pp,
			scores%[1]s.completed,

			beatmaps.beatmap_id, beatmaps.beatmapset_id, beatmaps.beatmap_md5,
			beatmaps.song_name, beatmaps.ar, beatmaps.od, beatmaps.difficulty_std,
			beatmaps.difficulty_taiko, beatmaps.difficulty_ctb, beatmaps.difficulty_mania,
			beatmaps.max_combo, beatmaps.hit_length, beatmaps.ranked,
			beatmaps.ranked_status_freezed, beatmaps.latest_update
		FROM user_pinned
		JOIN scores%[1]s ON scores%[1]s.id = scoreid
		JOIN beatmaps on scores%[1]s.beatmap_md5 = beatmaps.beatmap_md5
		WHERE user_pinned.userid = ? AND scores%[1]s.play_mode = ?
	`, dbs[playmode]) + common.Paginate(md.Query("p"), md.Query("l"), 100)

	rows, err := md.DB.Query(query, userid, mode)

	if err != nil {
		md.Err(err)
		return Err500
	}

	for rows.Next() {
		var (
			us userScore
			b  beatmap
		)

		err = rows.Scan(
			&us.ID, &us.BeatmapMD5, &us.Score.Score,
			&us.MaxCombo, &us.FullCombo, &us.Mods,
			&us.Count300, &us.Count100, &us.Count50,
			&us.CountGeki, &us.CountKatu, &us.CountMiss,
			&us.Time, &us.PlayMode, &us.Accuracy, &us.PP,
			&us.Completed,

			&b.BeatmapID, &b.BeatmapsetID, &b.BeatmapMD5,
			&b.SongName, &b.AR, &b.OD, &b.Diff2.STD,
			&b.Diff2.Taiko, &b.Diff2.CTB, &b.Diff2.Mania,
			&b.MaxCombo, &b.HitLength, &b.Ranked,
			&b.RankedStatusFrozen, &b.LatestUpdate,
		)

		if err != nil {
			md.Err(err)
			return Err500
		}

		us.Beatmap = b
		scoreResponse.Scores = append(scoreResponse.Scores, us)
	}

	scoreResponse.Code = 200
	return scoreResponse
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
		return common.SimpleResponse(400, "Invalid playmode")
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
