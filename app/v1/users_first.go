package v1

import (
	"database/sql"
	//"strconv"
	"strings"

	"fmt"

	"github.com/RealistikOsu/RealistikAPI/common"
	"gopkg.in/thehowl/go-osuapi.v1"
	"zxq.co/x/getrank"
)

// Score is a score done.
type tuser struct {
	common.ResponseBase
	Total  string      `json:"total"`
	Scores []userScore `json:"scores"`
}

type FirstPlaceScore struct {
	Score
	Beatmap  beatmap `json:"beatmap"`
	UserID   uint32
	Username string
}

type FirstPlaceResonse struct {
	common.ResponseBase
	FirstPlaces []FirstPlaceScore `json:"first_places"`
	Total       uint32            `json:"total"`
}

const const_query = `
SELECT
	s.id, s.beatmap_md5, s.score, s.max_combo,
	s.full_combo, s.mods, s.300_count, s.100_count,
	s.50_count, s.katus_count, s.gekis_count, s.misses_count,
	s.time, s.play_mode, s.accuracy, s.pp, s.completed,
	b.beatmap_id, b.beatmapset_id, b.beatmap_md5, b.song_name, b.ar, b.od,
	b.difficulty_std, b.difficulty_std, b.difficulty_taiko, b.difficulty_ctb, b.difficulty_mania,
	b.max_combo, b.hit_length, b.ranked, b.ranked_status_freezed, b.latest_update,
	u.id, u.username
FROM 
	first_places fp,
	%s s,
	beatmaps b,
	users u
WHERE 
	fp.score_id = s.id
	AND s.completed = 3
	AND s.beatmap_md5 = b.beatmap_md5
	AND u.id = s.userid
	AND s.play_mode = ?
	AND fp.relax = ?
	AND u.privileges & 2
LIMIT %d
OFFSET %d
`

const count_first_places = `
SELECT
	COUNT(*)
FROM
	first_places f
INNER JOIN
	users u
ON u.id = first_places.userid
WHERE
	f.relax = ?
	AND u.privileges & 2
`

const page_size = 25

func OldestFirstGET(md common.MethodData) common.CodeMessager {
	page := common.Int(md.Query("page"))
	mode := common.Int(md.Query("m"))
	c_mode := common.Int(md.Query("c_mode"))

	offset := page_size * page

	table := "scores"

	switch c_mode {
	case 2:
		table = "scores_ap"
	case 1:
		table = "scores_relax"
	}

	query := fmt.Sprintf(const_query, table, page_size, offset)

	var (
		rows *sql.Rows
		err  error
		resp FirstPlaceResonse
	)

	md.DB.Get(&resp.Total, count_first_places, c_mode)

	rows, err = md.DB.Query(
		query,
		mode,
		c_mode,
	)
	if err != nil {
		md.Err(err)
		return Err500
	}
	defer rows.Close()

	for rows.Next() {
		nc := FirstPlaceScore{}
		err = rows.Scan(
			&nc.Score.ID,
			&nc.Score.BeatmapMD5,
			&nc.Score.Score,
			&nc.Score.MaxCombo,
			&nc.Score.FullCombo,
			&nc.Score.Mods,
			&nc.Score.Count300,
			&nc.Score.Count100,
			&nc.Score.Count50,
			&nc.Score.CountKatu,
			&nc.Score.CountGeki,
			&nc.Score.CountMiss,
			&nc.Score.Time,
			&nc.Score.PlayMode,
			&nc.Score.Accuracy,
			&nc.Score.PP,
			&nc.Score.Completed,
			&nc.Beatmap.BeatmapID,
			&nc.Beatmap.BeatmapsetID,
			&nc.Beatmap.BeatmapMD5,
			&nc.Beatmap.SongName,
			&nc.Beatmap.AR,
			&nc.Beatmap.OD,
			&nc.Beatmap.Difficulty,
			&nc.Beatmap.Diff2.STD,
			&nc.Beatmap.Diff2.Taiko,
			&nc.Beatmap.Diff2.CTB,
			&nc.Beatmap.Diff2.Mania,
			&nc.Beatmap.MaxCombo,
			&nc.Beatmap.HitLength,
			&nc.Beatmap.Ranked,
			&nc.Beatmap.RankedStatusFrozen,
			&nc.Beatmap.LatestUpdate,
			&nc.UserID,
			&nc.Username,
		)
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
		resp.FirstPlaces = append(resp.FirstPlaces, nc)
	}

	return resp
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
	rx := common.Int(md.Query("rx"))
	if rx != 1 && rx != 2 {
		rx = 0
	}
	table := "scores" + rxTableSuffix(rx)
	query := fmt.Sprintf(
		"SELECT %[1]s.id, %[1]s.beatmap_md5, %[1]s.score, %[1]s.max_combo, %[1]s.full_combo, %[1]s.mods, "+
			"%[1]s.300_count, %[1]s.100_count, %[1]s.50_count, %[1]s.katus_count, %[1]s.gekis_count, %[1]s.misses_count, "+
			"%[1]s.time, %[1]s.play_mode, %[1]s.accuracy, %[1]s.pp, %[1]s.completed, "+
			"beatmaps.beatmap_id, beatmaps.beatmapset_id, beatmaps.beatmap_md5, beatmaps.song_name, beatmaps.ar, beatmaps.od, "+
			"beatmaps.difficulty_std, beatmaps.difficulty_std, beatmaps.difficulty_taiko, beatmaps.difficulty_ctb, beatmaps.difficulty_mania, "+
			"beatmaps.max_combo, beatmaps.hit_length, beatmaps.ranked, beatmaps.ranked_status_freezed, beatmaps.latest_update "+
			"FROM first_places, %[1]s, beatmaps "+
			"WHERE first_places.score_id = %[1]s.id AND %[1]s.completed = 3 AND %[1]s.beatmap_md5 = beatmaps.beatmap_md5 "+
			"AND first_places.user_id = ? AND %[1]s.play_mode = ? AND first_places.relax = %[2]d "+
			"ORDER BY %[1]s.time DESC ",
		table, rx,
	)
	var (
		r    tuser
		rows *sql.Rows
		err  error
	)

	// Fetch all score from users
	md.DB.Get(&r.Total, "SELECT COUNT(*) FROM first_places WHERE user_id = ? AND mode = ? AND relax = ?", id, mode, rx)
	rows, err = md.DB.Query(query+common.Paginate(md.Query("p"), md.Query("l"), 50), id, mode)
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

	//r.Total = strconv.Itoa(len(r.Scores))
	r.ResponseBase.Code = 200
	return r
}
