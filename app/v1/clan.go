package v1

import (
	"database/sql"
	"fmt"
	"math"
	"sort"

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

const RXClanQuery = `SELECT users.id, users.username, users.register_datetime, users.privileges,
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
WHERE clan = ? AND privileges & 1 = 1
`

const VNClanQuery = `SELECT users.id, users.username, users.register_datetime, users.privileges,
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
WHERE clan = ? AND privileges & 1 = 1
`

const APClanQuery = `SELECT users.id, users.username, users.register_datetime, users.privileges,
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
WHERE clan = ? AND privileges & 1 = 1
`

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

	n := "std"
	if mode == 1 {
		n = "taiko"
	} else if mode == 2 {
		n = "ctb"
	} else if mode == 3 {
		n = "mania"
	} else {
		n = "std"
	}

	selectedQuery := VNClanQuery
	if rx == 1 {
		selectedQuery = RXClanQuery
	} else if rx == 2 {
		selectedQuery = APClanQuery
	}

	for i := 0; i < len(r.Clans); i++ {
		var members clanMembersData

		rid := r.Clans[i].ID

		err := md.DB.Select(&members.Members, selectedQuery, rid)

		if err != nil {
			fmt.Println(err)
		}

		members.Code = 200

		if n == "std" {
			sort.Slice(members.Members, func(i, j int) bool {
				return members.Members[i].PpStd > members.Members[j].PpStd
			})

			for idx, u := range members.Members {
				r.Clans[i].ChosenMode.PP = r.Clans[i].ChosenMode.PP + int(float64(u.PpStd)*math.Pow(0.95, float64(idx)))
				r.Clans[i].ChosenMode.RankedScore = r.Clans[i].ChosenMode.RankedScore + u.RankedScoreStd
				r.Clans[i].ChosenMode.TotalScore = r.Clans[i].ChosenMode.TotalScore + u.TotalScoreStd
				r.Clans[i].ChosenMode.PlayCount = r.Clans[i].ChosenMode.PlayCount + u.PlaycountStd
			}

		} else if n == "taiko" {
			sort.Slice(members.Members, func(i, j int) bool {
				return members.Members[i].PpTaiko > members.Members[j].PpTaiko
			})

			for idx, u := range members.Members {
				r.Clans[i].ChosenMode.PP = r.Clans[i].ChosenMode.PP + int(float64(u.PpTaiko)*math.Pow(0.95, float64(idx)))
				r.Clans[i].ChosenMode.RankedScore = r.Clans[i].ChosenMode.RankedScore + u.RankedScoreTaiko
				r.Clans[i].ChosenMode.TotalScore = r.Clans[i].ChosenMode.TotalScore + u.TotalScoreTaiko
				r.Clans[i].ChosenMode.PlayCount = r.Clans[i].ChosenMode.PlayCount + u.PlaycountTaiko
			}

		} else if n == "ctb" {
			sort.Slice(members.Members, func(i, j int) bool {
				return members.Members[i].PpCtb > members.Members[j].PpCtb
			})

			for idx, u := range members.Members {
				r.Clans[i].ChosenMode.PP = r.Clans[i].ChosenMode.PP + int(float64(u.PpCtb)*math.Pow(0.95, float64(idx)))
				r.Clans[i].ChosenMode.RankedScore = r.Clans[i].ChosenMode.RankedScore + u.RankedScoreCtb
				r.Clans[i].ChosenMode.TotalScore = r.Clans[i].ChosenMode.TotalScore + u.TotalScoreCtb
				r.Clans[i].ChosenMode.PlayCount = r.Clans[i].ChosenMode.PlayCount + u.PlaycountCtb
			}

		} else if n == "mania" {
			sort.Slice(members.Members, func(i, j int) bool {
				return members.Members[i].PpMania > members.Members[j].PpMania
			})

			for idx, u := range members.Members {
				r.Clans[i].ChosenMode.PP = r.Clans[i].ChosenMode.PP + int(float64(u.PpMania)*math.Pow(0.95, float64(idx)))
				r.Clans[i].ChosenMode.RankedScore = r.Clans[i].ChosenMode.RankedScore + u.RankedScoreMania
				r.Clans[i].ChosenMode.TotalScore = r.Clans[i].ChosenMode.TotalScore + u.TotalScoreMania
				r.Clans[i].ChosenMode.PlayCount = r.Clans[i].ChosenMode.PlayCount + u.PlaycountMania
			}

		}
	}

	sort.Slice(r.Clans, func(i, j int) bool {
		return r.Clans[i].ChosenMode.PP > r.Clans[j].ChosenMode.PP
	})

	for i := 0; i < len(r.Clans); i++ {
		r.Clans[i].Rank = (p * l) + i + 1
	}

	r.Clans = r.Clans[p*l : l]
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

	n := "std"
	if m == 1 {
		n = "taiko"
	} else if m == 2 {
		n = "ctb"
	} else if m == 3 {
		n = "mania"
	} else {
		n = "std"
	}

	selectedQuery := VNClanQuery
	if rx == 1 {
		selectedQuery = RXClanQuery
	} else if rx == 2 {
		selectedQuery = APClanQuery
	}

	for i := 0; i < len(r.Clans); i++ {
		var members clanMembersData

		rid := r.Clans[i].ID

		err := md.DB.Select(&members.Members, selectedQuery, rid)

		if err != nil {
			fmt.Println(err)
		}

		members.Code = 200

		if n == "std" {
			sort.Slice(members.Members, func(i, j int) bool {
				return members.Members[i].PpStd > members.Members[j].PpStd
			})

			for idx, u := range members.Members {
				r.Clans[i].ChosenMode.PP = r.Clans[i].ChosenMode.PP + int(float64(u.PpStd)*math.Pow(0.95, float64(idx)))
				r.Clans[i].ChosenMode.RankedScore = r.Clans[i].ChosenMode.RankedScore + u.RankedScoreStd
				r.Clans[i].ChosenMode.TotalScore = r.Clans[i].ChosenMode.TotalScore + u.TotalScoreStd
				r.Clans[i].ChosenMode.PlayCount = r.Clans[i].ChosenMode.PlayCount + u.PlaycountStd
			}

		} else if n == "taiko" {
			sort.Slice(members.Members, func(i, j int) bool {
				return members.Members[i].PpTaiko > members.Members[j].PpTaiko
			})

			for idx, u := range members.Members {
				r.Clans[i].ChosenMode.PP = r.Clans[i].ChosenMode.PP + int(float64(u.PpTaiko)*math.Pow(0.95, float64(idx)))
				r.Clans[i].ChosenMode.RankedScore = r.Clans[i].ChosenMode.RankedScore + u.RankedScoreTaiko
				r.Clans[i].ChosenMode.TotalScore = r.Clans[i].ChosenMode.TotalScore + u.TotalScoreTaiko
				r.Clans[i].ChosenMode.PlayCount = r.Clans[i].ChosenMode.PlayCount + u.PlaycountTaiko
			}

		} else if n == "ctb" {
			sort.Slice(members.Members, func(i, j int) bool {
				return members.Members[i].PpCtb > members.Members[j].PpCtb
			})

			for idx, u := range members.Members {
				r.Clans[i].ChosenMode.PP = r.Clans[i].ChosenMode.PP + int(float64(u.PpCtb)*math.Pow(0.95, float64(idx)))
				r.Clans[i].ChosenMode.RankedScore = r.Clans[i].ChosenMode.RankedScore + u.RankedScoreCtb
				r.Clans[i].ChosenMode.TotalScore = r.Clans[i].ChosenMode.TotalScore + u.TotalScoreCtb
				r.Clans[i].ChosenMode.PlayCount = r.Clans[i].ChosenMode.PlayCount + u.PlaycountCtb
			}

		} else if n == "mania" {
			sort.Slice(members.Members, func(i, j int) bool {
				return members.Members[i].PpMania > members.Members[j].PpMania
			})

			for idx, u := range members.Members {
				r.Clans[i].ChosenMode.PP = r.Clans[i].ChosenMode.PP + int(float64(u.PpMania)*math.Pow(0.95, float64(idx)))
				r.Clans[i].ChosenMode.RankedScore = r.Clans[i].ChosenMode.RankedScore + u.RankedScoreMania
				r.Clans[i].ChosenMode.TotalScore = r.Clans[i].ChosenMode.TotalScore + u.TotalScoreMania
				r.Clans[i].ChosenMode.PlayCount = r.Clans[i].ChosenMode.PlayCount + u.PlaycountMania
			}

		}
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
	row := md.DB.QueryRow("SELECT invite FROM clans_invites WHERE clan = ? LIMIT 1", clan).Scan(&r.Invite)
	if row != nil {
		fmt.Println(row)
	}
	return r
}

// ClanMembersGET retrieves the people who are in a certain clan.
func ClanMembersGET(md common.MethodData) common.CodeMessager {
	i := common.Int(md.Query("id"))
	if i == 0 {
		return ErrMissingField("id")
	}
	r := common.Int(md.Query("r"))
	if r == 0 {
		var members clanMembersData

		err := md.DB.Select(&members.Members, `SELECT users.id, users.username, users.register_datetime, users.privileges,
			latest_activity, users_stats.username_aka,
			
			users.country, users_stats.user_color,
			users_stats.ranked_score_std, users_stats.total_score_std, users_stats.pp_std, users_stats.playcount_std, users_stats.replays_watched_std, users_stats.total_hits_std,
			users_stats.ranked_score_taiko, users_stats.total_score_taiko, users_stats.pp_taiko, users_stats.playcount_taiko, users_stats.replays_watched_taiko, users_stats.total_hits_taiko
			
		FROM user_clans uc
		INNER JOIN users
		ON users.id = uc.user
		INNER JOIN users_stats ON users_stats.id = uc.user
		WHERE clan = ?
		ORDER BY id ASC `, i)

		if err != nil {
			md.Err(err)
			return Err500
		}

		members.Code = 200
		return members
	} else {
		var members clanMembersData

		err := md.DB.Select(&members.Members, `SELECT users.id, users.username, users.register_datetime, users.privileges,
			latest_activity, users_stats.username_aka,
			
			users.country, users_stats.user_color,
			users_stats.ranked_score_std, users_stats.total_score_std, users_stats.pp_std, users_stats.playcount_std, users_stats.replays_watched_std,
			users_stats.ranked_score_taiko, users_stats.total_score_taiko, users_stats.pp_taiko, users_stats.playcount_taiko, users_stats.replays_watched_taiko
			
		FROM user_clans uc
		INNER JOIN users
		ON users.id = uc.user
		INNER JOIN users_stats ON users_stats.id = uc.user
		WHERE clan = ? AND perms = ?
		ORDER BY id ASC `, i, r)

		if err != nil {
			md.Err(err)
			return Err500
		}

		members.Code = 200
		return members
	}
}

// Zunhapan likes this.
