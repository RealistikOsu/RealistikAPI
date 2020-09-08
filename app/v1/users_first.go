package v1

import (
	"database/sql"
	"strings"

	"gopkg.in/thehowl/go-osuapi.v1"
	"github.com/RealistikOsu/RealistikAPI/common"
	"zxq.co/x/getrank"
)

// Score is a score done.
type tuser struct {
	common.ResponseBase
  Total     string      `json:"total"`
	Scores		[]userScore `json:"scores"`
}

func UserFirstGET(md common.MethodData) common.CodeMessager {
	id := common.Int(md.Query("id"))
	if id == 0 {
		return ErrMissingField("id")
	}
	mode := 0
	m := common.Int(md.Query("mode"))
	if m != 0 {
		mode = m
	}
  //the worst queries i ever done and its fact
  query := "SELECT scores.id, scores.beatmap_md5, scores.score, scores.max_combo, scores.full_combo, scores.mods, scores.300_count, scores.100_count, scores.50_count, scores.katus_count, scores.gekis_count, scores.misses_count, scores.time, scores.play_mode, scores.accuracy, scores.pp, scores.completed, beatmaps.beatmap_id, beatmaps.beatmapset_id, beatmaps.beatmap_md5, beatmaps.song_name, beatmaps.ar, beatmaps.od, beatmaps.difficulty_std, beatmaps.difficulty_std, beatmaps.difficulty_taiko, beatmaps.difficulty_ctb, beatmaps.difficulty_mania, beatmaps.max_combo, beatmaps.hit_length, beatmaps.ranked, beatmaps.ranked_status_freezed, beatmaps.latest_update FROM first_places, scores, beatmaps WHERE first_places.scoreid = scores.id AND scores.beatmap_md5 = beatmaps.beatmap_md5 AND scores.completed = 3 AND first_places.userid = ? AND scores.play_mode = ? AND first_places.relax = 0 ORDER BY scores.time DESC "
  rx := common.Int(md.Query("rx"))
  if rx == 1 {
    query = "SELECT scores_relax.id, scores_relax.beatmap_md5, scores_relax.score, scores_relax.max_combo, scores_relax.full_combo, scores_relax.mods, scores_relax.300_count, scores_relax.100_count, scores_relax.50_count, scores_relax.katus_count, scores_relax.gekis_count, scores_relax.misses_count, scores_relax.time, scores_relax.play_mode, scores_relax.accuracy, scores_relax.pp, scores_relax.completed, beatmaps.beatmap_id, beatmaps.beatmapset_id, beatmaps.beatmap_md5, beatmaps.song_name, beatmaps.ar, beatmaps.od, beatmaps.difficulty_std, beatmaps.difficulty_std, beatmaps.difficulty_taiko, beatmaps.difficulty_ctb, beatmaps.difficulty_mania, beatmaps.max_combo, beatmaps.hit_length, beatmaps.ranked, beatmaps.ranked_status_freezed, beatmaps.latest_update FROM first_places, scores_relax, beatmaps WHERE first_places.score_id = scores_relax.id AND scores_relax.completed = 3 AND scores_relax.beatmap_md5 = beatmaps.beatmap_md5 AND first_places.user_id = ? AND scores_relax.play_mode = ? AND first_places.relax = 1 ORDER BY scores_relax.time DESC"
  } else if rx == 2 {
    query = "SELECT scores_ap.id, scores_ap.beatmap_md5, scores_ap.score, scores_ap.max_combo, scores_ap.full_combo, scores_ap.mods, scores_ap.300_count, scores_ap.100_count, scores_ap.50_count, scores_ap.katus_count, scores_ap.gekis_count, scores_ap.misses_count, scores_ap.time, scores_ap.play_mode, scores_ap.accuracy, scores_ap.pp, scores_ap.completed, beatmaps.beatmap_id, beatmaps.beatmapset_id, beatmaps.beatmap_md5, beatmaps.song_name, beatmaps.ar, beatmaps.od, beatmaps.difficulty_std, beatmaps.difficulty_std, beatmaps.difficulty_taiko, beatmaps.difficulty_ctb, beatmaps.difficulty_mania, beatmaps.max_combo, beatmaps.hit_length, beatmaps.ranked, beatmaps.ranked_status_freezed, beatmaps.latest_update FROM first_places, scores_ap, beatmaps WHERE first_places.score_id = scores_ap.id AND scores_ap.completed = 3 AND scores_ap.beatmap_md5 = beatmaps.beatmap_md5 AND first_places.user_id = ? AND scores_ap.play_mode = ? AND first_places.relax = 2 ORDER BY scores_ap.time DESC"
  } else {
      query = "SELECT scores.id, scores.beatmap_md5, scores.score, scores.max_combo, scores.full_combo, scores.mods, scores.300_count, scores.100_count, scores.50_count, scores.katus_count, scores.gekis_count, scores.misses_count, scores.time, scores.play_mode, scores.accuracy, scores.pp, scores.completed, beatmaps.beatmap_id, beatmaps.beatmapset_id, beatmaps.beatmap_md5, beatmaps.song_name, beatmaps.ar, beatmaps.od, beatmaps.difficulty_std, beatmaps.difficulty_std, beatmaps.difficulty_taiko, beatmaps.difficulty_ctb, beatmaps.difficulty_mania, beatmaps.max_combo, beatmaps.hit_length, beatmaps.ranked, beatmaps.ranked_status_freezed, beatmaps.latest_update FROM first_places, scores, beatmaps WHERE first_places.score_id = scores.id AND scores.beatmap_md5 = beatmaps.beatmap_md5 AND scores.completed = 3 AND first_places.user_id = ? AND scores.play_mode = ? AND first_places.relax = 0 ORDER BY scores.time DESC "
}
	var (
		r    tuser
		rows *sql.Rows
		err  error
	)

	// Fetch all score from users
  md.DB.Get(&r.Total, "SELECT COUNT(score_id) FROM first_places WHERE user_id = ? AND mode = ? AND relax = ?", id, mode, rx)
	rows, err = md.DB.Query(query + common.Paginate(md.Query("p"), md.Query("l"), 50), id, mode)
	if err != nil {
		md.Err(err)
		return Err500
	}
	defer rows.Close()
	for rows.Next() {
		nc := userScore{}
		err = rows.Scan(&nc.Score.ID, &nc.Score.BeatmapMD5, &nc.Score.Score, &nc.Score.MaxCombo, &nc.Score.FullCombo, &nc.Score.Mods, &nc.Score.Count300, &nc.Score.Count100, &nc.Score.Count50, &nc.Score.CountKatu, &nc.Score.CountGeki, &nc.Score.CountMiss, &nc.Score.Time, &nc.Score.PlayMode, &nc.Score.Accuracy, &nc.Score.PP, &nc.Score.Completed, &nc.Beatmap.BeatmapID, &nc.Beatmap.BeatmapsetID, &nc.Beatmap.BeatmapMD5, &nc.Beatmap.SongName, &nc.Beatmap.AR, &nc.Beatmap.OD, &nc.Beatmap.Difficulty, &nc.Beatmap.Diff2.STD, &nc.Beatmap.Diff2.Taiko, &nc.Beatmap.Diff2.CTB, &nc.Beatmap.Diff2.Mania, &nc.Beatmap.MaxCombo, &nc.Beatmap.HitLength, &nc.Beatmap.Ranked, &nc.Beatmap.RankedStatusFrozen, &nc.Beatmap.LatestUpdate)
		if err != nil {
			md.Err(err)
		}
		nc.Rank = strings.ToUpper(getrank.GetRank(
			osuapi.Mode(nc.PlayMode),
			osuapi.Mods(nc.Mods),
			nc.Accuracy,
			nc.Count300,
			nc.Count100,
			nc.Count50,
			nc.CountMiss,
		))

		if err != nil {
			md.Err(err)
		}

		r.Scores = append(r.Scores, nc)
	}

	r.ResponseBase.Code = 200
	return r
}
