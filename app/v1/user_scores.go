package v1

import (
	"fmt"
	"strings"

	"github.com/RealistikOsu/RealistikAPI/common"
	"gopkg.in/thehowl/go-osuapi.v1"
	"zxq.co/x/getrank"
)

type userScore struct {
	Score
	Beatmap beatmap `json:"beatmap"`
}

type userScoresResponse struct {
	common.ResponseBase
	//Total  string      `json:"total"`
	Scores []userScore `json:"scores"`
}

const userScoreSelectBase = `
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
		FROM scores%[1]s
		INNER JOIN beatmaps ON beatmaps.beatmap_md5 = scores%[1]s.beatmap_md5
		INNER JOIN users ON users.id = scores%[1]s.userid
		`

// UserScoresBestGET retrieves the best scores of an user, sorted by PP if
// mode is standard and sorted by ranked score otherwise.
func UserScoresBestGET(md common.MethodData) common.CodeMessager {
	cm, wc, param := whereClauseUser(md, "users")
	if cm != nil {
		return *cm
	}

	rx := common.Int(md.Query("rx"))
	table := "scores" + rxTableSuffix(rx)
	// For all modes that have PP, we leave out 0 PP scores.
	mc := genModeClause(md)

	return userScoresPuts(md, rx, fmt.Sprintf(
		`WHERE
			%s.completed = '3'
			AND beatmaps.ranked IN (2,3)
			AND %s
			%s
			AND `+md.User.OnlyUserPublic(true)+`
		ORDER BY %s.pp DESC, %s.score DESC %s`,
		table, wc, mc, table, table, common.Paginate(md.Query("p"), md.Query("l"), 100),
	), param)
}

// UserScoresRecentGET retrieves an user's latest scores.
func UserScoresRecentGET(md common.MethodData) common.CodeMessager {
	cm, wc, param := whereClauseUser(md, "users")
	if cm != nil {
		return *cm
	}

	rx := common.Int(md.Query("rx"))
	table := "scores" + rxTableSuffix(rx)

	recentClause := ""
	if md.Query("filter") == "recent" {
		recentClause = fmt.Sprintf(" AND %s.completed > 1 AND %s.time > UNIX_TIMESTAMP() - 86400 ", table, table) // 24 hours
	}

	return userScoresPuts(md, rx, fmt.Sprintf(
		`WHERE
			%s
			%s
			%s
			AND `+md.User.OnlyUserPublic(true)+`
		ORDER BY %s.id DESC %s`,
		wc, genModeClause(md), recentClause, table, common.Paginate(md.Query("p"), md.Query("l"), 100),
	), param)
}

func userScoresPuts(md common.MethodData, rx int, whereClause string, params ...interface{}) common.CodeMessager {
	rows, err := md.DB.Query(fmt.Sprintf(userScoreSelectBase, rxTableSuffix(rx))+whereClause, params...)
	if err != nil {
		md.Err(err)
		return Err500
	}
	defer rows.Close()
	var scores []userScore
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
		b.Difficulty = b.Diff2.STD
		us.Beatmap = b
		us.Rank = strings.ToUpper(getrank.GetRank(
			osuapi.Mode(us.PlayMode),
			osuapi.Mods(us.Mods),
			us.Accuracy,
			us.Count300,
			us.Count100,
			us.Count50,
			us.CountMiss,
		))
		scores = append(scores, us)
	}
	r := userScoresResponse{}
	r.Code = 200
	r.Scores = scores
	return r
}
