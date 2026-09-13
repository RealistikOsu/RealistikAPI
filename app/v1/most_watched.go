package v1

import (
	"fmt"
	"strings"

	"github.com/RealistikOsu/RealistikAPI/common"
	"gopkg.in/thehowl/go-osuapi.v1"
	"zxq.co/x/getrank"
)

type mostWatchedScore struct {
	Score
	WatchedCount int     `json:"watched_count"`
	Beatmap      beatmap `json:"beatmap"`
}

type mostWatchedScoresResponse struct {
	common.ResponseBase
	Scores []mostWatchedScore `json:"scores"`
}

const mostWatchedSelectBase = `
		SELECT
			scores%[1]s.id, scores%[1]s.beatmap_md5, scores%[1]s.score,
			scores%[1]s.max_combo, scores%[1]s.full_combo, scores%[1]s.mods,
			scores%[1]s.300_count, scores%[1]s.100_count, scores%[1]s.50_count,
			scores%[1]s.gekis_count, scores%[1]s.katus_count, scores%[1]s.misses_count,
			scores%[1]s.time, scores%[1]s.play_mode, scores%[1]s.accuracy, scores%[1]s.pp,
			scores%[1]s.completed, scores%[1]s.playback_rate, scores%[1]s.watched_count,

			beatmaps.beatmap_id, beatmaps.beatmapset_id, beatmaps.beatmap_md5,
			beatmaps.song_name, beatmaps.ar, beatmaps.od, beatmaps.difficulty_std,
			beatmaps.difficulty_taiko, beatmaps.difficulty_ctb, beatmaps.difficulty_mania,
			beatmaps.max_combo, beatmaps.hit_length, beatmaps.ranked,
			beatmaps.ranked_status_freezed, beatmaps.latest_update
		FROM scores%[1]s
		INNER JOIN beatmaps ON beatmaps.beatmap_md5 = scores%[1]s.beatmap_md5
		INNER JOIN users ON users.id = scores%[1]s.userid
		`

// MostWatchedScoresGET retrieves a single user's scores whose replays have
// been watched the most - a profile card, same shape as
// UserScoresBestGET/UserScoresRecentGET (id/name-scoped via
// whereClauseUser, mode-scoped via genModeClause, rx picks the table).
func MostWatchedScoresGET(md common.MethodData) common.CodeMessager {
	cm, wc, param := whereClauseUser(md, "users")
	if cm != nil {
		return *cm
	}

	rx := common.Int(md.Query("rx"))
	table := "scores" + rxTableSuffix(rx)
	mc := genModeClause(md)

	rows, err := md.DB.Query(fmt.Sprintf(
		mostWatchedSelectBase+`
		WHERE %s.completed = 3
			AND beatmaps.ranked IN (2,3)
			AND %s.watched_count > 0
			AND %s
			%s
			AND `+md.User.OnlyUserPublic(true)+`
		ORDER BY %s.watched_count DESC, %s.pp DESC %s`,
		rxTableSuffix(rx), table, table, wc, mc, table, table,
		common.Paginate(md.Query("p"), md.Query("l"), 100),
	), param)
	if err != nil {
		md.Err(err)
		return Err500
	}
	defer rows.Close()

	var scores []mostWatchedScore
	for rows.Next() {
		var (
			ms mostWatchedScore
			b  beatmap
		)
		err = rows.Scan(
			&ms.ID, &ms.BeatmapMD5, &ms.Score.Score,
			&ms.MaxCombo, &ms.FullCombo, &ms.Mods,
			&ms.Count300, &ms.Count100, &ms.Count50,
			&ms.CountGeki, &ms.CountKatu, &ms.CountMiss,
			&ms.Time, &ms.PlayMode, &ms.Accuracy, &ms.PP,
			&ms.Completed, &ms.PlaybackRate, &ms.WatchedCount,

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
		ms.Beatmap = b
		ms.Rank = strings.ToUpper(getrank.GetRank(
			osuapi.Mode(ms.PlayMode),
			osuapi.Mods(ms.Mods),
			ms.Accuracy,
			ms.Count300,
			ms.Count100,
			ms.Count50,
			ms.CountMiss,
		))
		scores = append(scores, ms)
	}

	r := mostWatchedScoresResponse{}
	r.Code = 200
	r.Scores = scores
	return r
}
