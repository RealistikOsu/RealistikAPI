package v1

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/RealistikOsu/RealistikAPI/common"
)

// TODO: Allow users to disable comments (through settings and admin panel)
// TODO: Let admin bypass disable_comments
// TODO: Let cm remove comments
// TODO? Profanity check

const (
	MAX_LENGTH = 380
	MIN_LENGTH = 3
)

type comment struct {
	Op       int    `json:"op"`
	UserID   int    `json:"userid"`
	Username string `json:"username"`
	Message  string `json:"message"`
	PostedAt int64  `json:"posted_at"`
}

type comments struct {
	common.ResponseBase
	Comments []comment `json:"comments"`
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
	var userExists bool

	res := comments{}
	userid := common.Int(md.Query("id"))
	err := md.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE privileges & 1 = 1 AND id = ?);", userid).Scan(&userExists)

	if err != nil && err != sql.ErrNoRows {
		md.Err(err)
		return Err500
	}

	err = md.DB.QueryRow("SELECT disabled_comments FROM users WHERE id = ?;", userid).Scan(&commentsAllowed)
	if err != nil && err != sql.ErrNoRows {
		md.Err(err)
		return Err500
	}

	if !userExists || commentsAllowed == 1 {
		return common.SimpleResponse(404, "Profile not found/comments are disabled.")
	}

	cquery := `
		SELECT
			user_comments.op, user_comments.prof,
			user_comments.msg, user_comments.comment_date,
			users.username
		FROM user_comments
		INNER JOIN users ON users.id = user_comments.op
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
			&cmt.Op, &cmt.UserID,
			&cmt.Message, &cmt.PostedAt,
			&cmt.Username,
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

func CommentDELETE() {
	panic("unimplemented")
}
