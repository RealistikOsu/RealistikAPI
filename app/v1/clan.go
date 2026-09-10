package v1

import (
	"database/sql"
	"fmt"
	"math"
	"sort"

	"github.com/jmoiron/sqlx"

	"github.com/RealistikOsu/RealistikAPI/common"
)

type singleClan struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Tag         string `json:"tag"`
	Icon        string `json:"icon"`
}

type multiClanData struct {
	common.ResponseBase
	Clans []singleClan `json:"clans"`
}

// clansGET retrieves all the clans on this ripple instance.
func ClansGET(md common.MethodData) common.CodeMessager {
	var (
		r    multiClanData
		rows *sql.Rows
		err  error
	)
	if md.Query("id") != "" {
		rows, err = md.DB.Query("SELECT id, name, description, tag, icon FROM clans WHERE id = ? LIMIT 1", md.Query("id"))
	} else {
		rows, err = md.DB.Query("SELECT id, name, description, tag, icon FROM clans " + common.Paginate(md.Query("p"), md.Query("l"), 50))
	}
	if err != nil {
		md.Err(err)
		return Err500
	}
	defer rows.Close()
	for rows.Next() {
		nc := singleClan{}
		err = rows.Scan(&nc.ID, &nc.Name, &nc.Description, &nc.Tag, &nc.Icon)
		if err != nil {
			md.Err(err)
		}
		r.Clans = append(r.Clans, nc)
	}
	if err := rows.Err(); err != nil {
		md.Err(err)
	}
	r.ResponseBase.Code = 200
	return r
}

type clanMembersData struct {
	common.ResponseBase
	Members []userNotFullResponse `json:"members"`
}

// get total stats of clan. later.
type totalStats struct {
	common.ResponseBase
	ClanID     int      `json:"id"`
	ChosenMode modeData `json:"chosen_mode"`
	Rank       int      `json:"rank"`
}
type clanLbSingle struct {
	ID          int      `json:"id,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tag         string   `json:"tag"`
	Icon        string   `json:"icon"`
	ChosenMode  modeData `json:"chosen_mode"`
	Rank        int      `json:"rank"`
}

type megaStats struct {
	common.ResponseBase
	Clans []clanLbSingle `json:"clans"`
}

const RXClanQuery = `SELECT uc.clan AS clan_id, users.id, users.username, users.register_datetime, users.privileges,
latest_activity, rx_stats.username_aka,

users.country, rx_stats.user_color,
rx_stats.ranked_score_std, rx_stats.total_score_std, rx_stats.pp_std, rx_stats.playcount_std, rx_stats.replays_watched_std, rx_stats.total_hits_std,
rx_stats.ranked_score_taiko, rx_stats.total_score_taiko, rx_stats.pp_taiko, rx_stats.playcount_taiko, rx_stats.replays_watched_taiko,rx_stats.total_hits_taiko,
rx_stats.ranked_score_ctb, rx_stats.total_score_ctb, rx_stats.pp_ctb, rx_stats.playcount_ctb, rx_stats.replays_watched_ctb, rx_stats.total_hits_ctb,
rx_stats.ranked_score_mania, rx_stats.total_score_mania, rx_stats.pp_mania, rx_stats.playcount_mania, rx_stats.replays_watched_mania, rx_stats.total_hits_mania

FROM user_clans uc
INNER JOIN users
ON users.id = uc.user
INNER JOIN rx_stats ON rx_stats.id = uc.user
WHERE uc.clan IN (?) AND privileges & 1 = 1
`

const VNClanQuery = `SELECT uc.clan AS clan_id, users.id, users.username, users.register_datetime, users.privileges,
latest_activity, users_stats.username_aka,

users.country, users_stats.user_color,
users_stats.ranked_score_std, users_stats.total_score_std, users_stats.pp_std, users_stats.playcount_std, users_stats.replays_watched_std, users_stats.total_hits_std,
users_stats.ranked_score_taiko, users_stats.total_score_taiko, users_stats.pp_taiko, users_stats.playcount_taiko, users_stats.replays_watched_taiko,users_stats.total_hits_taiko,
users_stats.ranked_score_ctb, users_stats.total_score_ctb, users_stats.pp_ctb, users_stats.playcount_ctb, users_stats.replays_watched_ctb, users_stats.total_hits_ctb,
users_stats.ranked_score_mania, users_stats.total_score_mania, users_stats.pp_mania, users_stats.playcount_mania, users_stats.replays_watched_mania, users_stats.total_hits_mania

FROM user_clans uc
INNER JOIN users
ON users.id = uc.user
INNER JOIN users_stats ON users_stats.id = uc.user
WHERE uc.clan IN (?) AND privileges & 1 = 1
`

const APClanQuery = `SELECT uc.clan AS clan_id, users.id, users.username, users.register_datetime, users.privileges,
latest_activity, ap_stats.username_aka,

users.country, ap_stats.user_color,
ap_stats.ranked_score_std, ap_stats.total_score_std, ap_stats.pp_std, ap_stats.playcount_std, ap_stats.replays_watched_std, ap_stats.total_hits_std,
ap_stats.ranked_score_taiko, ap_stats.total_score_taiko, ap_stats.pp_taiko, ap_stats.playcount_taiko, ap_stats.replays_watched_taiko,ap_stats.total_hits_taiko,
ap_stats.ranked_score_ctb, ap_stats.total_score_ctb, ap_stats.pp_ctb, ap_stats.playcount_ctb, ap_stats.replays_watched_ctb, ap_stats.total_hits_ctb,
ap_stats.ranked_score_mania, ap_stats.total_score_mania, ap_stats.pp_mania, ap_stats.playcount_mania, ap_stats.replays_watched_mania, ap_stats.total_hits_mania

FROM user_clans uc
INNER JOIN users
ON users.id = uc.user
INNER JOIN ap_stats ON ap_stats.id = uc.user
WHERE uc.clan IN (?) AND privileges & 1 = 1
`

// clanStatsMember is a clan member row enriched with the clan they belong
// to, so all clans' members can be fetched in a single query instead of
// issuing one query per clan (the previous behaviour: an O(clans) fan-out of
// blocking DB round-trips on every /clans/stats/all request).
type clanStatsMember struct {
	userNotFullResponse
	ClanID int `db:"clan_id"`
}

// clanQueryForRx returns the member-stats query for the given relax/autopilot mode.
func clanQueryForRx(rx int) string {
	switch rx {
	case 1:
		return RXClanQuery
	case 2:
		return APClanQuery
	default:
		return VNClanQuery
	}
}

// loadClanStats fetches every member of the given clans in one query and
// aggregates each clan's weighted pp (0.95 decay per rank position, matching
// the per-clan pp weighting used elsewhere) plus ranked score/total
// score/playcount totals for the given mode.
func loadClanStats(md common.MethodData, clanIDs []int, rx int, mode string) (map[int]modeData, error) {
	stats := make(map[int]modeData, len(clanIDs))
	if len(clanIDs) == 0 {
		return stats, nil
	}

	query, params, err := sqlx.In(clanQueryForRx(rx), clanIDs)
	if err != nil {
		return nil, err
	}
	var members []clanStatsMember
	if err := md.DB.Select(&members, md.DB.Rebind(query), params...); err != nil {
		return nil, err
	}

	byClan := make(map[int][]clanStatsMember)
	for _, m := range members {
		byClan[m.ClanID] = append(byClan[m.ClanID], m)
	}

	for clanID, cm := range byClan {
		switch mode {
		case "taiko":
			sort.Slice(cm, func(i, j int) bool { return cm[i].PpTaiko > cm[j].PpTaiko })
		case "ctb":
			sort.Slice(cm, func(i, j int) bool { return cm[i].PpCtb > cm[j].PpCtb })
		case "mania":
			sort.Slice(cm, func(i, j int) bool { return cm[i].PpMania > cm[j].PpMania })
		default:
			sort.Slice(cm, func(i, j int) bool { return cm[i].PpStd > cm[j].PpStd })
		}

		var d modeData
		for idx, u := range cm {
			var pp, playcount int
			var rankedScore, totalScore uint64
			switch mode {
			case "taiko":
				pp, rankedScore, totalScore, playcount = u.PpTaiko, u.RankedScoreTaiko, u.TotalScoreTaiko, u.PlaycountTaiko
			case "ctb":
				pp, rankedScore, totalScore, playcount = u.PpCtb, u.RankedScoreCtb, u.TotalScoreCtb, u.PlaycountCtb
			case "mania":
				pp, rankedScore, totalScore, playcount = u.PpMania, u.RankedScoreMania, u.TotalScoreMania, u.PlaycountMania
			default:
				pp, rankedScore, totalScore, playcount = u.PpStd, u.RankedScoreStd, u.TotalScoreStd, u.PlaycountStd
			}
			d.PP += int(float64(pp) * math.Pow(0.95, float64(idx)))
			d.RankedScore += rankedScore
			d.TotalScore += totalScore
			d.PlayCount += playcount
		}
		stats[clanID] = d
	}
	return stats, nil
}

func AllClanStatsGET(md common.MethodData) common.CodeMessager {
	var (
		r    megaStats
		rows *sql.Rows
		err  error
	)
	p := common.Int(md.Query("p")) - 1
	if p < 0 {
		p = 0
	}
	l := common.InString(1, md.Query("l"), 500, 50)
	rows, err = md.DB.Query("SELECT id, name, description, tag, icon FROM clans")

	if err != nil {
		md.Err(err)
		return Err500
	}
	defer rows.Close()
	for rows.Next() {
		nc := clanLbSingle{}
		err = rows.Scan(&nc.ID, &nc.Name, &nc.Description, &nc.Tag, &nc.Icon)
		if err != nil {
			md.Err(err)
		}
		nc.ChosenMode.PP = 0
		r.Clans = append(r.Clans, nc)
	}
	if err := rows.Err(); err != nil {
		md.Err(err)
	}
	r.ResponseBase.Code = 200

	mode := common.Int(md.Query("m"))
	rx := common.Int(md.Query("rx"))
	n := modeName(mode)

	clanIDs := make([]int, len(r.Clans))
	for i, c := range r.Clans {
		clanIDs[i] = c.ID
	}
	stats, err := loadClanStats(md, clanIDs, rx, n)
	if err != nil {
		md.Err(err)
		return Err500
	}
	for i := range r.Clans {
		r.Clans[i].ChosenMode = stats[r.Clans[i].ID]
	}

	sort.Slice(r.Clans, func(i, j int) bool {
		return r.Clans[i].ChosenMode.PP > r.Clans[j].ChosenMode.PP
	})

	for i := 0; i < len(r.Clans); i++ {
		r.Clans[i].Rank = i + 1
	}

	start := p * l
	if start > len(r.Clans) {
		start = len(r.Clans)
	}
	end := start + l
	if end > len(r.Clans) {
		end = len(r.Clans)
	}
	r.Clans = r.Clans[start:end]
	return r
}

func TotalClanStatsGET(md common.MethodData) common.CodeMessager {
	var (
		r    megaStats
		rows *sql.Rows
		err  error
	)
	rows, err = md.DB.Query("SELECT id, name, description, icon FROM clans")

	if err != nil {
		md.Err(err)
		return Err500
	}
	defer rows.Close()
	for rows.Next() {
		nc := clanLbSingle{}
		err = rows.Scan(&nc.ID, &nc.Name, &nc.Description, &nc.Icon)
		if err != nil {
			md.Err(err)
		}
		nc.ChosenMode.PP = 0
		r.Clans = append(r.Clans, nc)
	}
	if err := rows.Err(); err != nil {
		md.Err(err)
	}
	r.ResponseBase.Code = 200

	id := common.Int(md.Query("id"))
	if id == 0 {
		return ErrMissingField("id")
	}

	m := common.Int(md.Query("m"))
	rx := common.Int(md.Query("rx"))
	n := modeName(m)

	clanIDs := make([]int, len(r.Clans))
	for i, c := range r.Clans {
		clanIDs[i] = c.ID
	}
	stats, err := loadClanStats(md, clanIDs, rx, n)
	if err != nil {
		md.Err(err)
		return Err500
	}
	for i := range r.Clans {
		r.Clans[i].ChosenMode = stats[r.Clans[i].ID]
	}

	sort.Slice(r.Clans, func(i, j int) bool {
		return r.Clans[i].ChosenMode.PP > r.Clans[j].ChosenMode.PP
	})

	for i := 0; i < len(r.Clans); i++ {
		r.Clans[i].Rank = i + 1
	}
	b := totalStats{}
	for i := 0; i < len(r.Clans); i++ {
		if r.Clans[i].ID == id {
			b.ClanID = id
			b.ChosenMode.PP = r.Clans[i].ChosenMode.PP
			b.ChosenMode.RankedScore = r.Clans[i].ChosenMode.RankedScore
			b.ChosenMode.TotalScore = r.Clans[i].ChosenMode.TotalScore
			b.ChosenMode.PlayCount = r.Clans[i].ChosenMode.PlayCount
			b.ChosenMode.ReplaysWatched = r.Clans[i].ChosenMode.ReplaysWatched
			b.ChosenMode.TotalHits = r.Clans[i].ChosenMode.TotalHits
			b.Rank = r.Clans[i].Rank
			b.Code = 200
		}
	}

	return b
}

type isClanData struct {
	Clan  int `json:"clan"`
	User  int `json:"user"`
	Perms int `json:"perms"`
}

type isClan struct {
	common.ResponseBase
	Clan isClanData `json:"clan"`
}

func IsInClanGET(md common.MethodData) common.CodeMessager {
	ui := md.Query("uid")

	if ui == "0" {
		return ErrMissingField("uid")
	}

	var r isClan
	rows, err := md.DB.Query("SELECT user, clan, perms FROM user_clans WHERE user = ?", ui)

	if err != nil {
		md.Err(err)
		return Err500
	}

	defer rows.Close()
	for rows.Next() {
		nc := isClanData{}
		err = rows.Scan(&nc.User, &nc.Clan, &nc.Perms)
		if err != nil {
			md.Err(err)
		}
		r.Clan = nc
	}
	if err := rows.Err(); err != nil {
		md.Err(err)
	}
	r.ResponseBase.Code = 200
	return r
}

type imFoolish struct {
	common.ResponseBase
	Invite string `json:"invite"`
}
type adminClan struct {
	Id    int `json:"user"`
	Perms int `json:"perms"`
}

func ClanInviteGET(md common.MethodData) common.CodeMessager {
	// big perms check lol ok
	n := common.Int(md.Query("id"))
	adminFoolish := adminClan{}

	var r imFoolish
	var clan int
	// get user clan, then get invite
	md.DB.QueryRow("SELECT user, clan, perms FROM user_clans WHERE user = ? LIMIT 1", n).Scan(&adminFoolish.Id, &clan, &adminFoolish.Perms)
	if adminFoolish.Perms < 8 || adminFoolish.Id != md.ID() {
		return common.SimpleResponse(500, "You are not the admin of the clan")
	}
	if err := md.DB.QueryRow("SELECT invite FROM clans_invites WHERE clan = ? LIMIT 1", clan).Scan(&r.Invite); err != nil {
		md.Err(err)
	}
	return r
}

const clanMembersSelectBase = `SELECT users.id, users.username, users.register_datetime, users.privileges,
	latest_activity, users_stats.username_aka,

	users.country, users_stats.user_color,
	users_stats.ranked_score_std, users_stats.total_score_std, users_stats.pp_std, users_stats.playcount_std, users_stats.replays_watched_std, users_stats.total_hits_std,
	users_stats.ranked_score_taiko, users_stats.total_score_taiko, users_stats.pp_taiko, users_stats.playcount_taiko, users_stats.replays_watched_taiko, users_stats.total_hits_taiko

FROM user_clans uc
INNER JOIN users
ON users.id = uc.user
INNER JOIN users_stats ON users_stats.id = uc.user
WHERE clan = ?%s
ORDER BY id ASC `

// ClanMembersGET retrieves the people who are in a certain clan. If r
// (perms) is given, only members with that permission level are returned.
func ClanMembersGET(md common.MethodData) common.CodeMessager {
	i := common.Int(md.Query("id"))
	if i == 0 {
		return ErrMissingField("id")
	}

	var members clanMembersData
	var err error
	if r := common.Int(md.Query("r")); r == 0 {
		err = md.DB.Select(&members.Members, fmt.Sprintf(clanMembersSelectBase, ""), i)
	} else {
		err = md.DB.Select(&members.Members, fmt.Sprintf(clanMembersSelectBase, " AND perms = ?"), i, r)
	}
	if err != nil {
		md.Err(err)
		return Err500
	}

	members.Code = 200
	return members
}

// Zunhapan likes this.
