package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	. "wb/app/config"
	"wb/app/models"
	"wb/app/routes"
	"wb/app/stats"

	"github.com/abduld/oauth"
	"github.com/robfig/revel"
)

type CourseraApplication struct {
	PublicApplication
}

type CourseraIdentity struct {
	Id           string `json:"id"`
	FullName     string `json:"full_name"`
	EmailAddress string `json:"email_address"`
}

type CourseraTrustedIdentity struct {
	Id              string `json:"id"`
	FullName        string `json:"full_name"`
	SubtitlesAccess string `json:"subtitles_access"`
}

var CourseraConsumer *oauth.Consumer

func InitCourseraController() {
	CourseraConsumer = oauth.NewConsumer(
		CourseraOAuthConsumerKey,
		CourseraOAuthConsumerSecret,
		oauth.ServiceProvider{
			RequestTokenUrl:   CourseraRequestTokenAddress,
			AccessTokenUrl:    CourseraAccessTokenAddress,
			AuthorizeTokenUrl: CourseraAuthenticationAddress,
		},
	)
	//CourseraConsumer.Debug(true)
}

func doPostCourseraGrade(user models.User, mp models.MachineProblem,
	grade models.Grade, toPost string, forceQ bool) error {
	type CourseraGrade struct {
		UserId              string `json:"user_id"`
		AssignmentId        string `json:"assignment_part_sid"`
		Score               int64  `json:"score"`
		Feedback            string `json:"feedback"`
		APIKey              string `json:"api_key"`
		CreateNewSubmission int    `json:"create_new_submission"`
	}

	toPost = strings.ToLower(toPost)
	toPostCode := false
	toPostPeer := false
	if toPost == "all" || toPost == "code" {
		toPostCode = true
	}

	if toPost == "all" || toPost == "peer" {
		toPostPeer = true
	}

	conf, _ := ReadMachineProblemConfig(mp.Number)

	getIdentity := func() (string, error) {
		var idty CourseraIdentity
		idtyStr := models.GetUserIdentity(user)
		if idtyStr == "" {
			return "", errors.New("Not connected to cousera")
		}

		if err := json.Unmarshal([]byte(idtyStr), &idty); err != nil {
			return "", err
		}

		return idty.Id, nil
	}

	idty, err := getIdentity()
	if err != nil {
		return err
	}

	postGrade := func(kind string, key string, score int64) error {
		reason := ""
		if kind == "code" {
			reason = grade.Reasons
		}

		t := time.Now().Unix()
		if kind == "code" && time.Since(conf.CodingDeadline).Hours() > conf.GracePeriod {
			t = conf.CodingDeadline.UTC().Unix()
		} else if kind == "peer" && time.Since(conf.PeerReviewDeadline).Hours() > conf.GracePeriod {
			t = conf.PeerReviewDeadline.UTC().Unix()
		}
		resp, err := http.PostForm(CourseraGradeURL,
			url.Values{
				"api_key":             {CourseraGradeAPIKey},
				"user_id":             {idty},
				"score":               {strconv.Itoa(int(score))},
				"assignment_part_sid": {key},
				"feedback":            {reason},
				"submission_time":     {fmt.Sprint(t)},
			})
		if err != nil {
			return err
		}
		models.UpdateCourseraGrade(grade, kind, score)
		defer resp.Body.Close()
		return err
	}

	if toPostPeer && grade.PeerReviewScore != 0 && (forceQ == true || grade.PeerReviewScore >= grade.CourseraPeerReviewGrade) {
		if err := postGrade("peer", conf.CourseraPeerReviewPostKey, grade.PeerReviewScore); err != nil {
			return errors.New("Was not able to post coursera grade. Make sure you are connected to coursera first.")
		}
	}

	if toPostCode && grade.CodeScore != 0 && (forceQ == true || grade.CodeScore >= grade.CourseraCodingGrade) {
		if err := postGrade("code", conf.CourseraCodePostKey, grade.CodeScore); err != nil {
			return errors.New("Was not able to post coursera grade. Make sure you are connected to coursera first.")
		}
	}
	stats.Incr("User", "CourseraPostGrade")

	return nil
}

func (c CourseraApplication) PostGrade(gradeIdString string, toPostString string, forceString string) revel.Result {

	gradeId, err := strconv.Atoi(gradeIdString)
	if err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "Error: Cannot Submit Grade to Coursera.",
			"data":   "Invalid grade Id.",
		})
	}

	user := c.connected()

	grade, err := models.FindGrade(int64(gradeId))
	if err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "Error: Cannot Submit Grade to Coursera.",
			"data":   "Cannot find grade record.",
		})
	}

	mp, err := models.FindMachineProblem(grade.MachineProblemId)
	if err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "Error: Cannot Submit Grade to Coursera.",
			"data":   "Cannot find machine problem record.",
		})
	}

	mpUser, err := models.FindUser(mp.UserInstanceId)
	if err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "Error: Cannot Submit Grade to Coursera.",
			"data":   "User not found.",
		})
	}

	if mpUser.Id != user.Id {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "Error: Cannot Submit Grade to Coursera.",
			"data":   "User did not match.",
		})
	}

	forceQ := forceString == "true"

	if forceQ == false && grade.PeerReviewScore > 0 && grade.PeerReviewScore < grade.CourseraPeerReviewGrade {
		s := fmt.Sprint("Trying to post a code grade of ", grade.PeerReviewScore,
			" to Coursera, while Coursera has a higher score of ", grade.CourseraPeerReviewGrade)
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "Error: PeerReview Grade lower than what has been previously posted to Coursera",
			"data":   s,
		})
	}
	if forceQ == false && grade.CodeScore > 0 && grade.CodeScore < grade.CourseraCodingGrade {
		s := fmt.Sprint("Trying to post a code grade of ", grade.CodeScore,
			" to Coursera, while Coursera has a higher score of ", grade.CourseraCodingGrade)
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"title":  "Error: Code Grade lower than what has been previously posted to Coursera",
			"data":   s,
		})
	}

	if err = doPostCourseraGrade(user, mp, grade, toPostString, forceQ); err != nil {
		return c.RenderJson(map[string]interface{}{
			"status": "error",
			"data":   "Was not able to post coursera grade. Make sure you are connected to coursera first.",
		})
	}

	c.Flash.Success("Grade has been posted to Coursera.")

	return c.RenderJson(map[string]interface{}{
		"status": "success",
		"title":  "Grade has been posted to Coursera.",
		"data":   "Grade has been posted to Coursera.",
	})
}

func (c CourseraApplication) Connect() revel.Result {
	return c.Render()
}

var requestTokenCache map[int64]oauth.RequestToken = map[int64]oauth.RequestToken{}

// See https://instructor-support.desk.com/customer/portal/articles/711559-third-party-authorization-with-oauth
func (c CourseraApplication) Authenticate() revel.Result {
	var oauthToken string
	var oauthVerifier string

	user := c.connected()
	if user.Id == 0 {
		c.Flash.Error("Must log in before connecting user to coursera")
		return c.Redirect(routes.PublicApplication.Login())
	}

	c.Params.Bind(&oauthToken, "oauth_token")
	c.Params.Bind(&oauthVerifier, "oauth_verifier")

	if user.RequestToken == nil {
		revel.TRACE.Println(requestTokenCache[user.Id])
		if val, ok := requestTokenCache[user.Id]; ok {
			user.RequestToken = &val
		}
	}

	if oauthVerifier != "" {
		oauth.TOKEN_SECRET_PARAM = "oauth_token_secret"
		if accessToken, err := CourseraConsumer.AuthorizeToken(user.RequestToken, oauthVerifier); err == nil {
			user.AccessToken = accessToken
			getIdentity := func(addr string) (string, error) {
				resp, err := CourseraConsumer.Get(addr, map[string]string{}, accessToken)
				if err != nil {
					return "", err
				}
				defer resp.Body.Close()
				idty, _ := ioutil.ReadAll(resp.Body)
				return string(idty), nil
			}
			trustedIdentity, err1 := getIdentity(CourseraGetTrustedIdentityAddress)
			regularIdentity, err2 := getIdentity(CourseraGetIdentityAddress)
			if err1 != nil || err2 != nil {
				c.Flash.Error("Cannot get user identity from coursera!")
				stats.Incr("User", "CourseraAuthenticationFailed")
				return c.Redirect(routes.PublicApplication.Index())
			}
			stats.Incr("User", "CourseraAuthentication")
			stats.Log("User", "Coursera", "Connected to coursera with identity: "+regularIdentity)
			models.SetUserCourseraCredentials(user, user.AccessToken, user.RequestToken, trustedIdentity, regularIdentity)
		} else {
			revel.TRACE.Println("Error connecting to coursera:", err)
			c.Flash.Error("Cannot connect to coursera!")
		}
		return c.Redirect(routes.PublicApplication.Index())
	}

	oauth.TOKEN_SECRET_PARAM = "oauth_secret"
	if requestToken, url, err := CourseraConsumer.GetRequestTokenAndUrl(MasterAddress + "/coursera/authenticate"); err == nil {
		user.RequestToken = requestToken
		requestTokenCache[user.Id] = *requestToken
		return c.Redirect(url)
	} else {
		stats.Incr("User", "CourseraAuthenticationFailed")
		c.Flash.Error("Cannot connect account to coursera!")
	}
	return c.Redirect(routes.PublicApplication.Index())
}
