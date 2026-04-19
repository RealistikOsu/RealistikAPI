package v1

import (
	"math/rand"
	"time"

	"github.com/RealistikOsu/RealistikAPI/common"
)

var rn = rand.New(rand.NewSource(time.Now().UnixNano()))

var kaomojis = [...]string{
	"o7",
	"",
}

var randomSentences = [...]string{
	"i dint even lknow in netherlands? - Schydude",
	"so i was shopping for a new car",
	"“Yes sir” - Ryoui 2020",
	"babe, i'm breaking up with you. it's not you, you were poggers. it's me, i'm omegalul. im sorry if this is pepehands but it has to be done, i've just been feeling pepega and our relationship has been weirdchamp for months, it's time to end it, no kappa.",
	"woman.png will defend and save us!",
	"flimbo",
	"the earth is flat",
	"creeper?",
	"im dry out of ideas please help....",
	"Please visit c.ussr.pl for more info!",
	"piano note",
	"What is your in game (RealistikOsu) username? * dont play it",
	"Imagine HOI4 with RTX",
	"Tartarus looks like an incest accident with bloodlust and yatagarasu",
	"i am learning about heart disease on valentines day",
	"People who have airpods can't afford the wire",
}

func surpriseMe() string {
	n := int(time.Now().UnixNano())
	return randomSentences[n%len(randomSentences)] + " " + kaomojis[n%len(kaomojis)]
}

type pingResponse struct {
	common.ResponseBase
	ID              int                   `json:"user_id"`
	Privileges      common.Privileges     `json:"privileges"`
	UserPrivileges  common.UserPrivileges `json:"user_privileges"`
	PrivilegesS     string                `json:"privileges_string"`
	UserPrivilegesS string                `json:"user_privileges_string"`
}

// PingGET is a message to check with the API that we are logged in, and know what are our privileges.
func PingGET(md common.MethodData) common.CodeMessager {
	var r pingResponse
	r.Code = 200

	if md.ID() == 0 {
		r.Message = "You have not given us a token, so we don't know who you are! But you can still login with POST /tokens " + kaomojis[rn.Intn(len(kaomojis))]
	} else {
		r.Message = surpriseMe()
	}

	r.ID = md.ID()
	r.Privileges = md.User.TokenPrivileges
	r.UserPrivileges = md.User.UserPrivileges
	r.PrivilegesS = md.User.TokenPrivileges.String()
	r.UserPrivilegesS = md.User.UserPrivileges.String()

	return r
}

type surpriseMeResponse struct {
	common.ResponseBase
	Cats [100]string `json:"cats"`
}

// SurpriseMeGET generates cute cats.
//
// ... Yes.
func SurpriseMeGET(md common.MethodData) common.CodeMessager {
	var r surpriseMeResponse
	r.Code = 200
	for i := 0; i < 100; i++ {
		r.Cats[i] = surpriseMe()
	}
	return r
}
