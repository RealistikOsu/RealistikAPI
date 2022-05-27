package v1

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/RealistikOsu/RealistikAPI/common"
)

const (
	MAX_LENGTH = 380
	MIN_LENGTH = 3
)

type comment struct {
	ID       int    `json:"id"`
	Op       int    `json:"op"`
	Username string `json:"username"`
	Message  string `json:"message"`
	PostedAt int64  `json:"posted_at"`
}

type comments struct {
	common.ResponseBase
	Comments []comment `json:"comments"`
}

type commentInfo struct {
	common.ResponseBase
	Total    int  `json:"total"`
	Disabled bool `json:"disabled"`
}

func CommentPOST(md common.MethodData) common.CodeMessager {
	var commentDate int64 = time.Now().Unix()
	var userExists bool
	var doIExist bool
	var canComment int

	res := common.ResponseBase{}
	userid := common.Int(md.Query("id"))
	comment := string(md.Ctx.Request.Body())
	op := md.User.UserID

	// is user restricted? am i restricted? do they allow comments?
	err := md.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE privileges & 1 = 1 AND id = ?);", userid).Scan(&userExists)
	if err != nil && err != sql.ErrNoRows {
		md.Err(err)
		return Err500
	}

	err = md.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE privileges & 1 = 1 AND id = ?);", op).Scan(&doIExist)
	if err != nil && err != sql.ErrNoRows {
		md.Err(err)
		return Err500
	}

	err = md.DB.QueryRow("SELECT disabled_comments FROM users WHERE id = ?;", userid).Scan(&canComment)
	if err != nil && err != sql.ErrNoRows {
		md.Err(err)
		return Err500
	}

	if !userExists || !doIExist {
		return common.SimpleResponse(403, "You don't have the permissions to carry out this action!")
	} else if canComment == 1 {
		return common.SimpleResponse(403, "This user has disabled comments on their profile.")
	} else if len(comment) > MAX_LENGTH || len(comment) < MIN_LENGTH {
		return common.SimpleResponse(400, fmt.Sprintf("Invalid comment! Comment must be between %d and %d in length.", MIN_LENGTH, MAX_LENGTH))
	}

	_, err = md.DB.Exec("INSERT INTO user_comments (op, prof, msg, comment_date) VALUES (?, ?, ?, ?)", op, userid, comment, commentDate)
	if err != nil {
		md.Err(err)
		return Err500
	}

	res.Code = 200
	res.Message = "success!"
	return res
}

func CommentGET(md common.MethodData) common.CodeMessager {
	var commentsAllowed int
	var profileNotFound bool

	res := comments{}
	userid := common.Int(md.Query("id"))
	canComment := fmt.Sprintf("SELECT disabled_comments FROM users WHERE id = ? AND %s;", md.User.OnlyUserPublic(true))

	if err := md.DB.QueryRow(canComment, userid).Scan(&commentsAllowed); err != nil {
		if err == sql.ErrNoRows {
			profileNotFound = true
		} else {
			md.Err(err)
			return Err500
		}
	}

	if (commentsAllowed == 1 && !strings.Contains(md.User.UserPrivileges.String(), "AdminManageUsers")) || profileNotFound {
		return common.SimpleResponse(404, "Profile not found/comments are disabled.")
	}

	cquery := `
		SELECT
			user_comments.op, user_comments.msg,
			user_comments.comment_date,
			users.username, user_comments.id
		FROM user_comments
		JOIN users ON users.id = user_comments.op
		WHERE user_comments.prof = ? AND users.privileges & 1 = 1
		ORDER BY user_comments.comment_date DESC
	` + common.Paginate(md.Query("p"), md.Query("l"), 5)

	rows, err := md.DB.Query(cquery, userid)

	if err != nil {
		md.Err(err)
		return Err500
	}

	for rows.Next() {
		cmt := comment{}
		err = rows.Scan(
			&cmt.Op,
			&cmt.Message, &cmt.PostedAt,
			&cmt.Username, &cmt.ID,
		)

		if err != nil {
			md.Err(err)
			return Err500
		}

		res.Comments = append(res.Comments, cmt)
	}

	res.Code = 200
	return res
}

func CommentDELETE(md common.MethodData) common.CodeMessager {
	var op int
	var prof int

	res := common.ResponseBase{}
	id := common.Int(md.Query("id"))

	if err := md.DB.QueryRow("SELECT op, prof FROM user_comments WHERE user_comments.id = ?", id).Scan(&op, &prof); err != nil && err != sql.ErrNoRows {
		md.Err(err)
		return Err500
	}

	if (op == md.User.UserID || strings.Contains(md.User.UserPrivileges.String(), "AdminManageUsers") || prof == md.User.UserID) && op != 0 {
		_, err := md.DB.Exec("DELETE FROM user_comments WHERE id = ?", id)
		if err != nil {
			md.Err(err)
			return Err500
		}

		res.Code = 200
		res.Message = "success!"
		return res
	}

	return common.SimpleResponse(403, "You cannot delete this!")
}

func CommentInfoGET(md common.MethodData) common.CodeMessager {
	res := commentInfo{}
	id := common.Int(md.Query("id"))
	total := "SELECT COUNT(user_comments.id) FROM user_comments JOIN users ON users.id = user_comments.op WHERE prof = ? AND users.privileges & 1 = 1;"

	if err := md.DB.QueryRow(total, id).Scan(&res.Total); err != nil && err != sql.ErrNoRows {
		md.Err(err)
		return Err500
	}

	if err := md.DB.QueryRow("SELECT disabled_comments FROM users WHERE users.id = ? AND "+md.User.OnlyUserPublic(true), id).Scan(&res.Disabled); err != nil {
		if err == sql.ErrNoRows {
			return common.SimpleResponse(404, "User not found!")
		}

		md.Err(err)
		return Err500
	}

	res.Code = 200
	return res
}
