package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	. "wb/app/config"
	"wb/app/models"
	"wb/app/routes"
	"wb/app/stats"

	"github.com/revel/revel"
	"golang.org/x/oauth2"
)

type CourseraApplication struct {
	PublicApplication
}

var OAuthEndPoint oauth2.Endpoint

func InitCourseraController() {
	OAuthEndPoint = oauth2.Endpoint{
		AuthURL:  CourseraAuthTokenAddress,
		TokenURL: CourseraTokenAddress,
	}
}

var requestTokenCache map[int64]oauth2.Token = map[int64]oauth2.Token{}

// See https://tech.coursera.org/app-platform/oauth2/
func (c CourseraApplication) Authenticate() revel.Result {
	var code string
	var state string

	user := c.connected()
	if user.Id == 0 {
		c.Flash.Error("Must log in before connecting user to coursera")
		return c.Redirect(routes.PublicApplication.Login())
	}

	c.Params.Bind(&code, "code")
	c.Params.Bind(&state, "state")

	if user.RequestToken == nil {
		revel.TRACE.Println(requestTokenCache[user.Id])
		if val, ok := requestTokenCache[user.Id]; ok {
			user.RequestToken = &val
		}
	}

	if code == "" {

		OAuthConfig := &oauth2.Config{
			ClientID:    CourseraOAuthClientKey,
			Scopes:      []string{"view_profile"},
			Endpoint:    OAuthEndPoint,
			RedirectURL: "http://www.webgpu.com/coursera/ltiview",
		}
		url := OAuthConfig.AuthCodeURL("CRF-WEBGPU", oauth2.AccessTypeOffline)
		revel.ERROR.Println(url)
		models.SetUserCourseraIdentity(user, user.RequestToken, "")
		return c.Redirect(url)
	}

	return c.Redirect(routes.PublicApplication.Index())
}

func (c CourseraApplication) LTIViewAuthentication() revel.Result {
	return c.Render()
}

func (c CourseraApplication) LTIAuthenticate() revel.Result {
	var values string
	var form string
	var userid string

	user := c.connected()
	if user.Id == 0 {
		c.Flash.Error("Must log in before connecting user to coursera")
		return c.Redirect(routes.PublicApplication.Login())
	}

	c.Params.Bind(&values, "Values")
	c.Params.Bind(&form, "Form")
	c.Params.Bind(&userid, "user_id")

	if userid == "" {
		c.Flash.Error("Cannot get user identity from coursera!")
		stats.Incr("User", "CourseraAuthenticationFailed")
		return c.Redirect(routes.PublicApplication.Index())
	}
	revel.TRACE.Println("identity = ", userid)
	stats.Log("User", "Coursera", "Connected to coursera with identity: "+userid)
	models.SetUserCourseraCredentials(user, form, values, userid)

	c.Flash.Success("Connected to coursera!")
	return c.Redirect(routes.PublicApplication.Index())

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

	idty := models.GetUserIdentity(user)
	// Post grade to coursera
	postGrade := func(kind string, key string, score int64) error {
		reason := ""
		if kind == "code" {
			reason = grade.Reasons
		}

		if key == "NONE" {
			return errors.New("You are not graded on this MP")
		}

		t := time.Now().Unix()
		if kind == "code" && time.Since(conf.CodingDeadline).Hours() > conf.GracePeriod {
			t = conf.CodingDeadline.UTC().Unix()
		} else if kind == "peer" && time.Since(conf.PeerReviewDeadline).Hours() > conf.GracePeriod {
			t = conf.PeerReviewDeadline.UTC().Unix()
		}
		vals := url.Values{
			"api_key":             {CourseraGradeAPIKey},
			"user_id":             {idty},
			"score":               {strconv.Itoa(int(score))},
			"assignment_part_sid": {key},
			"feedback":            {reason},
			"submission_time":     {fmt.Sprint(t)},
		}
		resp, err := http.PostForm(CourseraGradeURL,
			vals)
		revel.TRACE.Println("Posting grade for ", idty, "  ", CourseraGradeURL, " with key ", key, vals)
		if err != nil {
			return err
		}
		models.UpdateCourseraGrade(grade, kind, score)
		defer resp.Body.Close()
		return err
	}

	if toPostPeer && grade.PeerReviewScore != 0 && (forceQ == true || grade.PeerReviewScore >= grade.CourseraPeerReviewGrade) {
		if err := postGrade("peer", conf.CourseraPeerReviewPostKey, grade.PeerReviewScore); err != nil {
			revel.ERROR.Println(err)
			return errors.New("Was not able to post coursera grade. Make sure you are connected to coursera first and/or this MP is graded.")
		}
	}

	if toPostCode && grade.CodeScore != 0 && (forceQ == true || grade.CodeScore >= grade.CourseraCodingGrade) {
		if err := postGrade("code", conf.CourseraCodePostKey, grade.CodeScore); err != nil {
			return errors.New("Was not able to post coursera grade. Make sure you are connected to coursera first and/or this MP is graded.")
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
