package v1

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"

	redis "gopkg.in/redis.v5"

	"github.com/RealistikOsu/RealistikAPI/common"
	"zxq.co/ripple/ocl"
)

type leaderboardUser struct {
	userData
	ChosenMode    modeData `json:"chosen_mode"`
	PlayStyle     int      `json:"play_style"`
	FavouriteMode int      `json:"favourite_mode"`
}

type leaderboardResponse struct {
	common.ResponseBase
	Users []leaderboardUser `json:"users"`
}

const rxUserQuery = `
		SELECT
			users.id, users.username, users.register_datetime, users.privileges, users.latest_activity, users.coins,

			users_stats.username_aka, users.country,
			users_stats.play_style, users_stats.favourite_mode,

			rx_stats.ranked_score_%[1]s, rx_stats.total_score_%[1]s, rx_stats.playcount_%[1]s,
			users_stats.replays_watched_%[1]s, users_stats.total_hits_%[1]s,
			rx_stats.avg_accuracy_%[1]s, rx_stats.pp_%[1]s
		FROM users
		INNER JOIN rx_stats ON rx_stats.id = users.id
		INNER JOIN users_stats ON users_stats.id = users.id `

const apUserQuery = `
		SELECT
			users.id, users.username, users.register_datetime, users.privileges, users.latest_activity, users.coins,

			users_stats.username_aka, users.country,
			users_stats.play_style, users_stats.favourite_mode,

			ap_stats.ranked_score_%[1]s, ap_stats.total_score_%[1]s, ap_stats.playcount_%[1]s,
			users_stats.replays_watched_%[1]s, users_stats.total_hits_%[1]s,
			ap_stats.avg_accuracy_%[1]s, ap_stats.pp_%[1]s
		FROM users
		INNER JOIN ap_stats ON ap_stats.id = users.id
		INNER JOIN users_stats ON users_stats.id = users.id `

const lbUserQuery = `
		SELECT
			users.id, users.username, users.register_datetime, users.privileges, users.latest_activity, users.coins,

			users_stats.username_aka, users.country,
			users_stats.play_style, users_stats.favourite_mode,

			users_stats.ranked_score_%[1]s, users_stats.total_score_%[1]s, users_stats.playcount_%[1]s,
			users_stats.replays_watched_%[1]s, users_stats.total_hits_%[1]s,
			users_stats.avg_accuracy_%[1]s, users_stats.pp_%[1]s
		FROM users
		INNER JOIN users_stats ON users_stats.id = users.id `

const lbCoinsUserQuery = `
		SELECT
			users.id, users.username, users.register_datetime, users.privileges, users.latest_activity, users.coins,

			users_stats.username_aka, users.country,
			users_stats.play_style, users_stats.favourite_mode

		FROM users
		INNER JOIN users_stats ON users_stats.id = users.id `

func getScoreLb(m string, rx int, p int, l int, country string, sorted string, md *common.MethodData) []leaderboardUser {
	var query, order, whereClause string

	switch rx {
	case 1:
		order = " ORDER BY rx_stats.ranked_score_%[1]s DESC, rx_stats.pp_%[1]s DESC"
		if country != "" {
			whereClause = " AND users.country = ?"
		}
		query = fmt.Sprintf(rxUserQuery+"WHERE (users.privileges & 3) >= 3"+whereClause+order+" LIMIT %d, %d", m, p*l, l)
	case 2:
		order = " ORDER BY ap_stats.ranked_score_%[1]s DESC, ap_stats.pp_%[1]s DESC"
		if country != "" {
			whereClause = " AND users.country = ?"
		}
		query = fmt.Sprintf(apUserQuery+"WHERE (users.privileges & 3) >= 3"+whereClause+order+" LIMIT %d, %d", m, p*l, l)
	default:
		order = " ORDER BY users_stats.ranked_score_%[1]s DESC, users_stats.pp_%[1]s DESC"
		if country != "" {
			whereClause = " AND users.country = ?"
		}
		query = fmt.Sprintf(lbUserQuery+"WHERE (users.privileges & 3) >= 3"+whereClause+order+" LIMIT %d, %d", m, p*l, l)
	}

	var rows *sql.Rows
	var err error
	if country != "" {
		rows, err = md.DB.Query(query, country)
	} else {
		rows, err = md.DB.Query(query)
	}
	if err != nil {
		md.Err(err)
		return make([]leaderboardUser, 0)
	}
	defer rows.Close()
	var users []leaderboardUser
	for i := 1; rows.Next(); i++ {
		u := leaderboardUser{}
		err := rows.Scan(
			&u.ID, &u.Username, &u.RegisteredOn, &u.Privileges, &u.LatestActivity, &u.Coins,

			&u.UsernameAKA, &u.Country, &u.PlayStyle, &u.FavouriteMode,

			&u.ChosenMode.RankedScore, &u.ChosenMode.TotalScore, &u.ChosenMode.PlayCount,
			&u.ChosenMode.ReplaysWatched, &u.ChosenMode.TotalHits,
			&u.ChosenMode.Accuracy, &u.ChosenMode.PP,
		)
		if err != nil {
			md.Err(err)
			continue
		}
		u.ChosenMode.Level = ocl.GetLevelPrecise(int64(u.ChosenMode.TotalScore))
		users = append(users, u)
	}
	return users

}

func getCoinLb(p int, l int, country string, sorted string, md *common.MethodData) []leaderboardUser {
	var query, whereClause string

	if country != "" {
		whereClause = " AND users.country = ?"
	}
	query = fmt.Sprintf(lbCoinsUserQuery+"WHERE (users.privileges & 3) >= 3"+whereClause+" ORDER BY users.coins DESC LIMIT %d, %d", p*l, l)

	var rows *sql.Rows
	var err error
	if country != "" {
		rows, err = md.DB.Query(query, country)
	} else {
		rows, err = md.DB.Query(query)
	}
	if err != nil {
		md.Err(err)
		return make([]leaderboardUser, 0)
	}
	defer rows.Close()
	var users []leaderboardUser
	for i := 1; rows.Next(); i++ {
		u := leaderboardUser{}
		err := rows.Scan(
			&u.ID, &u.Username, &u.RegisteredOn, &u.Privileges, &u.LatestActivity, &u.Coins,

			&u.UsernameAKA, &u.Country, &u.PlayStyle, &u.FavouriteMode,
		)
		if err != nil {
			md.Err(err)
			continue
		}
		users = append(users, u)
	}
	return users

}

// LeaderboardGET gets the leaderboard.
func LeaderboardGET(md common.MethodData) common.CodeMessager {
	m := getMode(md.Query("mode"))

	// md.Query.Country
	p := common.Int(md.Query("p")) - 1
	if p < 0 {
		p = 0
	}
	l := common.InString(1, md.Query("l"), 500, 50)
	sorted := md.Query("sort")
	rx := common.Int(md.Query("rx"))
	country := md.Query("country")

	key := "ripple:leaderboard:" + m
	if common.Int(md.Query("rx")) == 1 {
		key = "ripple:leaderboard_relax:" + m
	}
	if common.Int(md.Query("rx")) == 2 {
		key = "ripple:leaderboard_ap:" + m
	}
	if country != "" {
		key += ":" + strings.ToLower(country)
	}
	if sorted == "" {
		sorted = "pp"
	}

	if sorted == "score" {
		response := leaderboardResponse{Users: getScoreLb(m, rx, p, l, country, sorted, &md)}
		response.Code = 200
		return response
	}
	if sorted == "coins" {
		response := leaderboardResponse{Users: getCoinLb(p, l, country, sorted, &md)}
		response.Code = 200
		return response
	}

	results, err := md.R.ZRevRange(key, int64(p*l), int64(p*l+l-1)).Result()
	if err != nil {
		md.Err(err)
		return Err500
	}

	var resp leaderboardResponse
	resp.Code = 200

	if len(results) == 0 {
		return resp
	}

	query := fmt.Sprintf(lbUserQuery+`WHERE users.id IN (?) ORDER BY users_stats.pp_%[1]s DESC, users_stats.ranked_score_%[1]s DESC`, m)
	if common.Int(md.Query("rx")) == 1 {
		query = fmt.Sprintf(rxUserQuery+`WHERE users.id IN (?) ORDER BY rx_stats.pp_%[1]s DESC, rx_stats.ranked_score_%[1]s DESC`, m)
	}
	if common.Int(md.Query("rx")) == 2 {
		query = fmt.Sprintf(apUserQuery+`WHERE users.id IN (?) ORDER BY ap_stats.pp_%[1]s DESC, ap_stats.ranked_score_%[1]s DESC`, m)
	}
	query, params, _ := sqlx.In(query, results)
	rows, err := md.DB.Query(query, params...)
	if err != nil {
		md.Err(err)
		return Err500
	}
	for rows.Next() {
		var u leaderboardUser
		err := rows.Scan(
			&u.ID, &u.Username, &u.RegisteredOn, &u.Privileges, &u.LatestActivity, &u.Coins,

			&u.UsernameAKA, &u.Country, &u.PlayStyle, &u.FavouriteMode,

			&u.ChosenMode.RankedScore, &u.ChosenMode.TotalScore, &u.ChosenMode.PlayCount,
			&u.ChosenMode.ReplaysWatched, &u.ChosenMode.TotalHits,
			&u.ChosenMode.Accuracy, &u.ChosenMode.PP,
		)
		if err != nil {
			md.Err(err)
			continue
		}
		u.ChosenMode.Level = ocl.GetLevelPrecise(int64(u.ChosenMode.TotalScore))
		if common.Int(md.Query("rx")) == 1 {
			if i := relaxboardPosition(md.R, m, u.ID); i != nil {
				u.ChosenMode.GlobalLeaderboardRank = i
			}
			if i := rxcountryPosition(md.R, m, u.ID, u.Country); i != nil {
				u.ChosenMode.CountryLeaderboardRank = i
			}
		}
		if common.Int(md.Query("rx")) == 2 {
			if i := autoPosition(md.R, m, u.ID); i != nil {
				u.ChosenMode.GlobalLeaderboardRank = i
			}
			if i := apcountryPosition(md.R, m, u.ID, u.Country); i != nil {
				u.ChosenMode.CountryLeaderboardRank = i
			}
		} else {
			if i := leaderboardPosition(md.R, m, u.ID); i != nil {
				u.ChosenMode.GlobalLeaderboardRank = i
			}
			if i := countryPosition(md.R, m, u.ID, u.Country); i != nil {
				u.ChosenMode.CountryLeaderboardRank = i
			}
		}
		resp.Users = append(resp.Users, u)
	}
	return resp
}

func leaderboardPosition(r *redis.Client, mode string, user int) *int {
	return _position(r, "ripple:leaderboard:"+mode, user)
}

func countryPosition(r *redis.Client, mode string, user int, country string) *int {
	return _position(r, "ripple:leaderboard:"+mode+":"+strings.ToLower(country), user)
}

func relaxboardPosition(r *redis.Client, mode string, user int) *int {
	return _position(r, "ripple:leaderboard_relax:"+mode, user)
}

func autoPosition(r *redis.Client, mode string, user int) *int {
	return _position(r, "ripple:leaderboard_ap:"+mode, user)
}

func rxcountryPosition(r *redis.Client, mode string, user int, country string) *int {
	return _position(r, "ripple:leaderboard_relax:"+mode+":"+strings.ToLower(country), user)
}

func apcountryPosition(r *redis.Client, mode string, user int, country string) *int {
	return _position(r, "ripple:leaderboard_ap:"+mode+":"+strings.ToLower(country), user)
}

func _position(r *redis.Client, key string, user int) *int {
	res := r.ZRevRank(key, strconv.Itoa(user))
	if res.Err() == redis.Nil {
		return nil
	}
	x := int(res.Val()) + 1
	return &x
}
