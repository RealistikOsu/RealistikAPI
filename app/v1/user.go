// Package v1 implements the first version of the Ripple API.
package v1

import (
	"database/sql"
	"strconv"
	"strings"
	"unicode"

	"github.com/jmoiron/sqlx"
	"zxq.co/ripple/ocl"
	"github.com/RealistikOsu/RealistikAPI/common"
)

type userData struct {
	ID             int                  `json:"id"`
	Username       string               `json:"username"`
	UsernameAKA    string               `json:"username_aka"`
	RegisteredOn   common.UnixTimestamp `json:"registered_on"`
	Privileges     uint64               `json:"privileges"`
	LatestActivity common.UnixTimestamp `json:"latest_activity"`
	Country        string               `json:"country"`
}

const userFields = `SELECT users.id, users.username, register_datetime, users.privileges,
	latest_activity, users_stats.username_aka,
	users_stats.country
FROM users
INNER JOIN users_stats
ON users.id=users_stats.id
`

// UsersGET is the API handler for GET /users
func UsersGET(md common.MethodData) common.CodeMessager {
	shouldRet, whereClause, param := whereClauseUser(md, "users")
	if shouldRet != nil {
		return userPutsMulti(md)
	}

	query := userFields + `
WHERE ` + whereClause + ` AND ` + md.User.OnlyUserPublic(true) + `
LIMIT 1`
	return userPutsSingle(md, md.DB.QueryRowx(query, param))
}

type userPutsSingleUserData struct {
	common.ResponseBase
	userData
}

func userPutsSingle(md common.MethodData, row *sqlx.Row) common.CodeMessager {
	var err error
	var user userPutsSingleUserData

	err = row.StructScan(&user.userData)
	switch {
	case err == sql.ErrNoRows:
		return common.SimpleResponse(404, "No such user was found!")
	case err != nil:
		md.Err(err)
		return Err500
	}

	user.Code = 200
	return user
}

type userPutsMultiUserData struct {
	common.ResponseBase
	Users []userData `json:"users"`
}

func userPutsMulti(md common.MethodData) common.CodeMessager {
	pm := md.Ctx.Request.URI().QueryArgs().PeekMulti
	// query composition
	wh := common.
		Where("users.username_safe = ?", common.SafeUsername(md.Query("nname"))).
		Where("users.id = ?", md.Query("iid")).
		Where("users.privileges = ?", md.Query("privileges")).
		Where("users.privileges & ? > 0", md.Query("has_privileges")).
		Where("users.privileges & ? = 0", md.Query("has_not_privileges")).
		Where("users_stats.country = ?", md.Query("country")).
		Where("users_stats.username_aka = ?", md.Query("name_aka")).
		Where("privileges_groups.name = ?", md.Query("privilege_group")).
		In("users.id", pm("ids")...).
		In("users.username_safe", safeUsernameBulk(pm("names"))...).
		In("users_stats.username_aka", pm("names_aka")...).
		In("users_stats.country", pm("countries")...)

	var extraJoin string
	if md.Query("privilege_group") != "" {
		extraJoin = " LEFT JOIN privileges_groups ON users.privileges & privileges_groups.privileges = privileges_groups.privileges "
	}

	query := userFields + extraJoin + wh.ClauseSafe() + " AND " + md.User.OnlyUserPublic(true) +
		" " + common.Sort(md, common.SortConfiguration{
		Allowed: []string{
			"id",
			"username",
			"privileges",
			"donor_expire",
			"latest_activity",
			"silence_end",
		},
		Default: "id ASC",
		Table:   "users",
	}) +
		" " + common.Paginate(md.Query("p"), md.Query("l"), 100)

	// query execution
	rows, err := md.DB.Queryx(query, wh.Params...)
	if err != nil {
		md.Err(err)
		return Err500
	}
	var r userPutsMultiUserData
	for rows.Next() {
		var u userData
		err := rows.StructScan(&u)
		if err != nil {
			md.Err(err)
			continue
		}
		r.Users = append(r.Users, u)
	}
	r.Code = 200
	return r
}

// UserSelfGET is a shortcut for /users/id/self. (/users/self)
func UserSelfGET(md common.MethodData) common.CodeMessager {
	md.Ctx.Request.URI().SetQueryString("id=self")
	return UsersGET(md)
}

func safeUsernameBulk(us [][]byte) [][]byte {
	for _, u := range us {
		for idx, v := range u {
			if v == ' ' {
				u[idx] = '_'
				continue
			}
			u[idx] = byte(unicode.ToLower(rune(v)))
		}
	}
	return us
}

type whatIDResponse struct {
	common.ResponseBase
	ID int `json:"id"`
}

// UserWhatsTheIDGET is an API request that only returns an user's ID.
func UserWhatsTheIDGET(md common.MethodData) common.CodeMessager {
	var (
		r          whatIDResponse
		privileges uint64
	)
	err := md.DB.QueryRow("SELECT id, privileges FROM users WHERE username_safe = ? LIMIT 1", common.SafeUsername(md.Query("name"))).Scan(&r.ID, &privileges)
	if err != nil || ((privileges&uint64(common.UserPrivilegePublic)) == 0 &&
		(md.User.UserPrivileges&common.AdminPrivilegeManageUsers == 0)) {
		return common.SimpleResponse(404, "That user could not be found!")
	}
	r.Code = 200
	return r
}

var modesToReadable = [...]string{
	"std",
	"taiko",
	"ctb",
	"mania",
}

type modeData struct {
	RankedScore            uint64  `json:"ranked_score"`
	TotalScore             uint64  `json:"total_score"`
	PlayCount              int     `json:"playcount"`
	PlayTime               int     `json:"play_time"`
	ReplaysWatched         int     `json:"replays_watched"`
	TotalHits              int     `json:"total_hits"`
	Level                  float64 `json:"level"`
	Accuracy               float64 `json:"accuracy"`
	PP                     int     `json:"pp"`
	MaxCombo               int     `json:"max_combo"`
	GlobalLeaderboardRank  *int    `json:"global_leaderboard_rank"`
	CountryLeaderboardRank *int    `json:"country_leaderboard_rank"`
}
type userStats struct {
	STD   modeData `json:"std"`
	Taiko modeData `json:"taiko"`
	CTB   modeData `json:"ctb"`
	Mania modeData `json:"mania"`
}

type generalStats struct {
	Vanilla userStats `json:"vn"`
	Relax   userStats `json:"rx"`
	Autopilot userStats `json:"ap"`
}

type userFullResponse struct {
	common.ResponseBase
	userData
	Stats         generalStats          `json:"stats"`
	PlayStyle     int                   `json:"play_style"`
	FavouriteMode int                   `json:"favourite_mode"`
	Badges        []singleBadge         `json:"badges"`
	Clan          singleClan            `json:"clan"`
	CustomBadge   *singleBadge          `json:"custom_badge"`
	SilenceInfo   silenceInfo           `json:"silence_info"`
	CMNotes       *string               `json:"cm_notes,omitempty"`
	BanDate       *common.UnixTimestamp `json:"ban_date,omitempty"`
	Email         string                `json:"email,omitempty"`
}
type silenceInfo struct {
	Reason string               `json:"reason"`
	End    common.UnixTimestamp `json:"end"`
}
type userNotFullResponse struct {
	Id             int                  `json:"id"`
	Username       string               `json:"username"`
	UsernameAKA    string               `json:"username_aka"`
	RegisteredOn   common.UnixTimestamp `json:"registered_on"`
	Privileges     uint64               `json:"privileges"`
	LatestActivity common.UnixTimestamp `json:"latest_activity"`
	Country        string               `json:"country"`
	UserColor        string               `json:"user_color"`
	RankedScoreStd            uint64  `json:"ranked_score_std"`
	TotalScoreStd             uint64  `json:"total_score_std"`
	PlaycountStd              int     `json:"playcount_std"`
	ReplaysWatchedStd         int     `json:"replays_watched_std"`
	TotalHitsStd              int     `json:"total_hits_std"`
	PpStd                     int     `json:"pp_std"`
	RankedScoreTaiko            uint64  `json:"ranked_score_taiko"`
	TotalScoreTaiko             uint64  `json:"total_score_taiko"`
	PlaycountTaiko              int     `json:"playcount_taiko"`
	ReplaysWatchedTaiko         int     `json:"replays_watched_taiko"`
	TotalHitsTaiko              int     `json:"total_hits_taiko"`
	PpTaiko                     int     `json:"pp_taiko"`
	RankedScoreCtb            uint64  `json:"ranked_score_ctb"`
	TotalScoreCtb            uint64  `json:"total_score_ctb"`
	PlaycountCtb              int     `json:"playcount_ctb"`
	ReplaysWatchedCtb         int     `json:"replays_watched_ctb"`
	TotalHitsCtb              int     `json:"total_hits_ctb"`
	PpCtb                     int     `json:"pp_ctb"`
	RankedScoreMania            uint64  `json:"ranked_score_mania"`
	TotalScoreMania             uint64  `json:"total_score_mania"`
	PlaycountMania              int     `json:"playcount_mania"`
	ReplaysWatchedMania         int     `json:"replays_watched_mania"`
	TotalHitsMania              int     `json:"total_hits_mania"`
	PpMania                     int     `json:"pp_mania"`
	// STD       clappedModeData  `json:"std"`
	// Taiko     clappedModeData  `json:"taiko"`
	// CTB       clappedModeData  `json:"ctb"`
	// Mania     clappedModeData  `json:"mania"`
}

// UserFullGET gets all of an user's information, with one exception: their userpage.
func UserFullGET(md common.MethodData) common.CodeMessager {
	shouldRet, whereClause, param := whereClauseUser(md, "users")
	if shouldRet != nil {
		return *shouldRet
	}

	// Hellest query I've ever done.
	query := `
SELECT
	users.id, users.username, users.register_datetime, users.privileges, users.latest_activity,
	users_stats.username_aka, users_stats.country, users_stats.play_style, users_stats.favourite_mode,
	users_stats.custom_badge_icon, users_stats.custom_badge_name, users_stats.can_custom_badge,
	users_stats.show_custom_badge,

	users_stats.ranked_score_std, users_stats.total_score_std, users_stats.playcount_std, users_stats.playtime_std,
	users_stats.replays_watched_std, users_stats.total_hits_std,
	users_stats.avg_accuracy_std, users_stats.pp_std, users_stats.max_combo_std,
	users_stats.ranked_score_taiko, users_stats.total_score_taiko, users_stats.playcount_taiko, users_stats.playtime_taiko,
	users_stats.replays_watched_taiko, users_stats.total_hits_taiko,
	users_stats.avg_accuracy_taiko, users_stats.pp_taiko, users_stats.max_combo_taiko,
	users_stats.ranked_score_ctb, users_stats.total_score_ctb, users_stats.playcount_ctb, users_stats.playtime_ctb,
	users_stats.replays_watched_ctb, users_stats.total_hits_ctb,
	users_stats.avg_accuracy_ctb, users_stats.pp_ctb, users_stats.max_combo_ctb,
	users_stats.ranked_score_mania, users_stats.total_score_mania, users_stats.playcount_mania, users_stats.playtime_mania,
	users_stats.replays_watched_mania, users_stats.total_hits_mania,
	users_stats.avg_accuracy_mania, users_stats.pp_mania, users_stats.max_combo_mania,
	
	rx_stats.ranked_score_std, rx_stats.total_score_std, rx_stats.playcount_std, users_stats.playtime_std,
	rx_stats.replays_watched_std, rx_stats.total_hits_std,
	rx_stats.avg_accuracy_std, rx_stats.pp_std, rx_stats.max_combo_std,
	rx_stats.ranked_score_taiko, rx_stats.total_score_taiko, rx_stats.playcount_taiko, users_stats.playtime_taiko,
	rx_stats.replays_watched_taiko, rx_stats.total_hits_taiko,
	rx_stats.avg_accuracy_taiko, rx_stats.pp_taiko, rx_stats.max_combo_taiko,
	rx_stats.ranked_score_ctb, rx_stats.total_score_ctb, rx_stats.playcount_ctb, users_stats.playtime_ctb,
	rx_stats.replays_watched_ctb, rx_stats.total_hits_ctb,
	rx_stats.avg_accuracy_ctb, rx_stats.pp_ctb, rx_stats.max_combo_ctb,
	rx_stats.ranked_score_mania, rx_stats.total_score_mania, rx_stats.playcount_mania, users_stats.playtime_mania,
	rx_stats.replays_watched_mania, rx_stats.total_hits_mania,
	rx_stats.avg_accuracy_mania, rx_stats.pp_mania, rx_stats.max_combo_mania,

	ap_stats.ranked_score_std, ap_stats.total_score_std, ap_stats.playcount_std, users_stats.playtime_std,
	ap_stats.replays_watched_std, ap_stats.total_hits_std,
	ap_stats.avg_accuracy_std, ap_stats.pp_std, ap_stats.max_combo_std,
	ap_stats.ranked_score_taiko, ap_stats.total_score_taiko, ap_stats.playcount_taiko, users_stats.playtime_taiko,
	ap_stats.replays_watched_taiko, ap_stats.total_hits_taiko,
	ap_stats.avg_accuracy_taiko, ap_stats.pp_taiko, ap_stats.max_combo_taiko,
	ap_stats.ranked_score_ctb, ap_stats.total_score_ctb, ap_stats.playcount_ctb, users_stats.playtime_ctb,
	ap_stats.replays_watched_ctb, ap_stats.total_hits_ctb,
	ap_stats.avg_accuracy_ctb, ap_stats.pp_ctb, ap_stats.max_combo_ctb,
	ap_stats.ranked_score_mania, ap_stats.total_score_mania, ap_stats.playcount_mania, users_stats.playtime_mania,
	ap_stats.replays_watched_mania, ap_stats.total_hits_mania,
	ap_stats.avg_accuracy_mania, ap_stats.pp_mania, ap_stats.max_combo_mania,

	users.silence_reason, users.silence_end,
	users.notes, users.ban_datetime, users.email
FROM users
LEFT JOIN users_stats
ON users.id=users_stats.id
LEFT JOIN rx_stats
ON users.id=rx_stats.id
LEFT JOIN ap_stats
ON users.id=ap_stats.id
WHERE ` + whereClause + ` AND ` + md.User.OnlyUserPublic(true) + `
LIMIT 1
`

	// Fuck.
	r := userFullResponse{}
	var (
		b    singleBadge
		can  bool
		show bool
	)
	err := md.DB.QueryRow(query, param).Scan(
		&r.ID, &r.Username, &r.RegisteredOn, &r.Privileges, &r.LatestActivity,

		&r.UsernameAKA, &r.Country,
		&r.PlayStyle, &r.FavouriteMode,

		&b.Icon, &b.Name, &can, &show,

		&r.Stats.Vanilla.STD.RankedScore, &r.Stats.Vanilla.STD.TotalScore, &r.Stats.Vanilla.STD.PlayCount, &r.Stats.Vanilla.STD.PlayTime,
		&r.Stats.Vanilla.STD.ReplaysWatched, &r.Stats.Vanilla.STD.TotalHits,
		&r.Stats.Vanilla.STD.Accuracy, &r.Stats.Vanilla.STD.PP, &r.Stats.Vanilla.STD.MaxCombo,

		&r.Stats.Vanilla.Taiko.RankedScore, &r.Stats.Vanilla.Taiko.TotalScore, &r.Stats.Vanilla.Taiko.PlayCount, &r.Stats.Vanilla.Taiko.PlayTime,
		&r.Stats.Vanilla.Taiko.ReplaysWatched, &r.Stats.Vanilla.Taiko.TotalHits,
		&r.Stats.Vanilla.Taiko.Accuracy, &r.Stats.Vanilla.Taiko.PP, &r.Stats.Vanilla.Taiko.MaxCombo,

		&r.Stats.Vanilla.CTB.RankedScore, &r.Stats.Vanilla.CTB.TotalScore, &r.Stats.Vanilla.CTB.PlayCount, &r.Stats.Vanilla.CTB.PlayTime,
		&r.Stats.Vanilla.CTB.ReplaysWatched, &r.Stats.Vanilla.CTB.TotalHits,
		&r.Stats.Vanilla.CTB.Accuracy, &r.Stats.Vanilla.CTB.PP, &r.Stats.Vanilla.CTB.MaxCombo,

		&r.Stats.Vanilla.Mania.RankedScore, &r.Stats.Vanilla.Mania.TotalScore, &r.Stats.Vanilla.Mania.PlayCount, &r.Stats.Vanilla.Mania.PlayTime,
		&r.Stats.Vanilla.Mania.ReplaysWatched, &r.Stats.Vanilla.Mania.TotalHits,
		&r.Stats.Vanilla.Mania.Accuracy, &r.Stats.Vanilla.Mania.PP, &r.Stats.Vanilla.Mania.MaxCombo,

		&r.Stats.Relax.STD.RankedScore, &r.Stats.Relax.STD.TotalScore, &r.Stats.Relax.STD.PlayCount, &r.Stats.Relax.STD.PlayTime,
		&r.Stats.Relax.STD.ReplaysWatched, &r.Stats.Relax.STD.TotalHits,
		&r.Stats.Relax.STD.Accuracy, &r.Stats.Relax.STD.PP, &r.Stats.Relax.STD.MaxCombo,

		&r.Stats.Relax.Taiko.RankedScore, &r.Stats.Relax.Taiko.TotalScore, &r.Stats.Relax.Taiko.PlayCount, &r.Stats.Relax.Taiko.PlayTime,
		&r.Stats.Relax.Taiko.ReplaysWatched, &r.Stats.Relax.Taiko.TotalHits,
		&r.Stats.Relax.Taiko.Accuracy, &r.Stats.Relax.Taiko.PP, &r.Stats.Relax.Taiko.MaxCombo,

		&r.Stats.Relax.CTB.RankedScore, &r.Stats.Relax.CTB.TotalScore, &r.Stats.Relax.CTB.PlayCount, &r.Stats.Relax.CTB.PlayTime,
		&r.Stats.Relax.CTB.ReplaysWatched, &r.Stats.Relax.CTB.TotalHits,
		&r.Stats.Relax.CTB.Accuracy, &r.Stats.Relax.CTB.PP, &r.Stats.Relax.CTB.MaxCombo,

		&r.Stats.Relax.Mania.RankedScore, &r.Stats.Relax.Mania.TotalScore, &r.Stats.Relax.Mania.PlayCount, &r.Stats.Relax.Mania.PlayTime,
		&r.Stats.Relax.Mania.ReplaysWatched, &r.Stats.Relax.Mania.TotalHits,
		&r.Stats.Relax.Mania.Accuracy, &r.Stats.Relax.Mania.PP, &r.Stats.Relax.Mania.MaxCombo,


		&r.Stats.Autopilot.STD.RankedScore, &r.Stats.Autopilot.STD.TotalScore, &r.Stats.Autopilot.STD.PlayCount, &r.Stats.Autopilot.STD.PlayTime,
		&r.Stats.Autopilot.STD.ReplaysWatched, &r.Stats.Autopilot.STD.TotalHits,
		&r.Stats.Autopilot.STD.Accuracy, &r.Stats.Autopilot.STD.PP, &r.Stats.Autopilot.STD.MaxCombo,

		&r.Stats.Autopilot.Taiko.RankedScore, &r.Stats.Autopilot.Taiko.TotalScore, &r.Stats.Autopilot.Taiko.PlayCount, &r.Stats.Autopilot.Taiko.PlayTime,
		&r.Stats.Autopilot.Taiko.ReplaysWatched, &r.Stats.Autopilot.Taiko.TotalHits,
		&r.Stats.Autopilot.Taiko.Accuracy, &r.Stats.Autopilot.Taiko.PP, &r.Stats.Autopilot.Taiko.MaxCombo,

		&r.Stats.Autopilot.CTB.RankedScore, &r.Stats.Autopilot.CTB.TotalScore, &r.Stats.Autopilot.CTB.PlayCount, &r.Stats.Autopilot.CTB.PlayTime,
		&r.Stats.Autopilot.CTB.ReplaysWatched, &r.Stats.Autopilot.CTB.TotalHits,
		&r.Stats.Autopilot.CTB.Accuracy, &r.Stats.Autopilot.CTB.PP, &r.Stats.Autopilot.CTB.MaxCombo,

		&r.Stats.Autopilot.Mania.RankedScore, &r.Stats.Autopilot.Mania.TotalScore, &r.Stats.Autopilot.Mania.PlayCount, &r.Stats.Autopilot.Mania.PlayTime,
		&r.Stats.Autopilot.Mania.ReplaysWatched, &r.Stats.Autopilot.Mania.TotalHits,
		&r.Stats.Autopilot.Mania.Accuracy, &r.Stats.Autopilot.Mania.PP, &r.Stats.Autopilot.Mania.MaxCombo,

		&r.SilenceInfo.Reason, &r.SilenceInfo.End,
		&r.CMNotes, &r.BanDate, &r.Email,
	)
	switch {
	case err == sql.ErrNoRows:
		return common.SimpleResponse(404, "That user could not be found!")
	case err != nil:
		md.Err(err)
		return Err500
	}

	can = can && show && common.UserPrivileges(r.Privileges)&common.UserPrivilegeDonor > 0
	if can && (b.Name != "" || b.Icon != "") {
		r.CustomBadge = &b
	}

	for modeID, m := range [...]*modeData{&r.Stats.Vanilla.STD, &r.Stats.Vanilla.Taiko, &r.Stats.Vanilla.CTB, &r.Stats.Vanilla.Mania} {
		m.Level = ocl.GetLevelPrecise(int64(m.TotalScore))

		if i := leaderboardPosition(md.R, modesToReadable[modeID], r.ID); i != nil {
			m.GlobalLeaderboardRank = i
		}
		if i := countryPosition(md.R, modesToReadable[modeID], r.ID, r.Country); i != nil {
			m.CountryLeaderboardRank = i
		}
	}
	// I'm sorry for this horribleness but ripple and past mistakes have forced my hand
	for modeID, m := range [...]*modeData{&r.Stats.Relax.STD, &r.Stats.Relax.Taiko, &r.Stats.Relax.CTB, &r.Stats.Relax.Mania} {
		m.Level = ocl.GetLevelPrecise(int64(m.TotalScore))

		if i := relaxboardPosition(md.R, modesToReadable[modeID], r.ID); i != nil {
			m.GlobalLeaderboardRank = i
		}
		if i := rxcountryPosition(md.R, modesToReadable[modeID], r.ID, r.Country); i != nil {
			m.CountryLeaderboardRank = i
		}
	}
	// I'm sorry for this horribleness but ripple and past mistakes have forced my hand
	for modeID, m := range [...]*modeData{&r.Stats.Autopilot.STD, &r.Stats.Autopilot.Taiko, &r.Stats.Autopilot.CTB, &r.Stats.Autopilot.Mania} {
		m.Level = ocl.GetLevelPrecise(int64(m.TotalScore))

		if i := autoPosition(md.R, modesToReadable[modeID], r.ID); i != nil {
			m.GlobalLeaderboardRank = i
		}
		if i := apcountryPosition(md.R, modesToReadable[modeID], r.ID, r.Country); i != nil {
			m.CountryLeaderboardRank = i
		}
	}

	rows, err := md.DB.Query("SELECT b.id, b.name, b.icon FROM user_badges ub "+
		"LEFT JOIN badges b ON ub.badge = b.id WHERE user = ?", r.ID)
	if err != nil {
		md.Err(err)
	}

	for rows.Next() {
		var badge singleBadge
		err := rows.Scan(&badge.ID, &badge.Name, &badge.Icon)
		if err != nil {
			md.Err(err)
			continue
		}
		r.Badges = append(r.Badges, badge)
	}

	if md.User.TokenPrivileges&common.PrivilegeManageUser == 0 {
		r.CMNotes = nil
		r.BanDate = nil
		r.Email = ""
	}

	rows, err = md.DB.Query("SELECT c.id, c.name, c.description, c.tag, c.icon FROM user_clans uc "+
		"LEFT JOIN clans c ON uc.clan = c.id WHERE user = ?", r.ID)
	if err != nil {
		md.Err(err)
	}

	for rows.Next() {
		var clan singleClan
		err = rows.Scan(&clan.ID, &clan.Name, &clan.Description, &clan.Tag, &clan.Icon)
		if err != nil {
			md.Err(err)
			continue
		}
		r.Clan = clan
	}

	r.Code = 200
	return r
}

type userpageResponse struct {
	common.ResponseBase
	Userpage *string `json:"userpage"`
}

// UserUserpageGET gets an user's userpage, as in the customisable thing.
func UserUserpageGET(md common.MethodData) common.CodeMessager {
	shouldRet, whereClause, param := whereClauseUser(md, "users_stats")
	if shouldRet != nil {
		return *shouldRet
	}
	var r userpageResponse
	err := md.DB.QueryRow("SELECT userpage_content FROM users_stats WHERE "+whereClause+" LIMIT 1", param).Scan(&r.Userpage)
	switch {
	case err == sql.ErrNoRows:
		return common.SimpleResponse(404, "No such user!")
	case err != nil:
		md.Err(err)
		return Err500
	}
	if r.Userpage == nil {
		r.Userpage = new(string)
	}
	r.Code = 200
	return r
}

// UserSelfUserpagePOST allows to change the current user's userpage.
func UserSelfUserpagePOST(md common.MethodData) common.CodeMessager {
	var d struct {
		Data *string `json:"data"`
	}
	md.Unmarshal(&d)
	if d.Data == nil {
		return ErrMissingField("data")
	}
	cont := common.SanitiseString(*d.Data)
	_, err := md.DB.Exec("UPDATE users_stats SET userpage_content = ? WHERE id = ? LIMIT 1", cont, md.ID())
	if err != nil {
		md.Err(err)
	}
	md.Ctx.URI().SetQueryString("id=self")
	return UserUserpageGET(md)
}

func whereClauseUser(md common.MethodData, tableName string) (*common.CodeMessager, string, interface{}) {
	switch {
	case md.Query("id") == "self":
		return nil, tableName + ".id = ?", md.ID()
	case md.Query("id") != "":
		id, err := strconv.Atoi(md.Query("id"))
		if err != nil {
			a := common.SimpleResponse(400, "please pass a valid user ID")
			return &a, "", nil
		}
		return nil, tableName + ".id = ?", id
	case md.Query("name") != "":
		return nil, tableName + ".username_safe = ?", common.SafeUsername(md.Query("name"))
	}
	a := common.SimpleResponse(400, "you need to pass either querystring parameters name or id")
	return &a, "", nil
}

type userLookupResponse struct {
	common.ResponseBase
	Users []lookupUser `json:"users"`
}
type lookupUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

// UserLookupGET does a quick lookup of users beginning with the passed
// querystring value name.
func UserLookupGET(md common.MethodData) common.CodeMessager {
	name := common.SafeUsername(md.Query("name"))
	name = strings.NewReplacer(
		"%", "\\%",
		"_", "\\_",
		"\\", "\\\\",
	).Replace(name)
	if name == "" {
		return common.SimpleResponse(400, "please provide an username to start searching")
	}
	name = "%" + name + "%"

	var email string
	if md.User.TokenPrivileges&common.PrivilegeManageUser != 0 &&
		strings.Contains(md.Query("name"), "@") {
		email = md.Query("name")
	}

	rows, err := md.DB.Query("SELECT users.id, users.username FROM users WHERE "+
		"(username_safe LIKE ? OR email = ?) AND "+
		md.User.OnlyUserPublic(true)+" LIMIT 25", name, email)
	if err != nil {
		md.Err(err)
		return Err500
	}

	var r userLookupResponse
	for rows.Next() {
		var l lookupUser
		err := rows.Scan(&l.ID, &l.Username)
		if err != nil {
			continue // can't be bothered to handle properly
		}
		r.Users = append(r.Users, l)
	}

	r.Code = 200
	return r
}
