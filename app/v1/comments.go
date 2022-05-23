package v1

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/RealistikOsu/RealistikAPI/common"
)

// TODO: Allow users to disable comments (through settings and admin panel)
// TODO? Profanity check

const (
	MAX_LENGTH = 380
	MIN_LENGTH = 3
)

func CommentPOST(md common.MethodData) common.CodeMessager {
	var commentDate int64 = time.Now().Unix()
	var userExists bool
	var doIExist bool
	var canComment int

	res := common.ResponseBase{}
	userid := common.Int(md.Query("id"))
	comment := string(md.Ctx.Request.Body())
	op := md.User.UserID

	// is user restricted?
	err := md.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE privileges & 1 = 1 AND id = ?);", userid).Scan(&userExists)
	if err != nil && err != sql.ErrNoRows {
		md.Err(err)
		return Err500
	}

	// am i restricted?
	err = md.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE privileges & 1 = 1 AND id = ?);", op).Scan(&doIExist)
	if err != nil && err != sql.ErrNoRows {
		md.Err(err)
		return Err500
	}

	// does user allow comments?
	err = md.DB.QueryRow("SELECT disabled_comments FROM users WHERE id = ?;", userid).Scan(&canComment)
	if err != nil && err != sql.ErrNoRows {
		md.Err(err)
		return Err500
	}

	if !userExists || !doIExist {
		res.Code = 403
		res.Message = "You don't have the permissions to carry out this action!"
		return res
	}

	if canComment == 1 {
		res.Code = 403
		res.Message = "This user has disabled comments on their profile."
		return res
	}

	if len(comment) > MAX_LENGTH || len(comment) < MIN_LENGTH {
		res.Code = 400
		res.Message = fmt.Sprintf("Invalid comment! Comment must be between %d and %d in length.", MIN_LENGTH, MAX_LENGTH)
		return res
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

func CommentGET() {
	panic("unimplemented")
}

func CommentDELETE() {
	panic("unimplemented")
}
