package controllers

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
	. "wb/app/config"
	"wb/app/models"
	"wb/app/routes"
	"wb/app/server"
	"wb/app/stats"

	"github.com/robfig/revel"
	"github.com/russross/blackfriday"
)

type PublicApplication struct {
	*revel.Controller
}

func (c PublicApplication) connected() models.User {
	var user models.User
	if val, ok := c.RenderArgs["user"]; ok && val.(models.User).Id != 0 {
		user = val.(models.User)
	} else if username, ok := c.Session["user"]; ok && username != "" {
		user = c.getUser(username)
		c.RenderArgs["user"] = user
	}
	return user
}

func (c PublicApplication) getUser(userName string) models.User {
	if user, err := models.FindUserByName(userName); err == nil {
		return user
	}
	return models.User{}
}

func (c PublicApplication) SignUp() revel.Result {
	return c.Render()
}

func (c PublicApplication) AddUser() revel.Result {
	if IsMaster {
		if user := c.connected(); user.Id != 0 {
			c.RenderArgs["user"] = user
		}
		stats.Log("App", "StartRequest", c.Action)
	}
	return nil
}

func (c PublicApplication) UserNameExists() revel.Result {
	var name string
	c.Params.Bind(&name, "user_name")
	return c.RenderJson(map[string]interface{}{
		"status": "success",
		"result": models.UserNameExists(name),
	})
}

func (c PublicApplication) CreateUser() revel.Result {
	var user models.User

	c.Params.Bind(&user, "user")

	if models.UserNameExists(user.UserName) {
		c.Flash.Error("Username Already taken")
		return c.Redirect(routes.PublicApplication.CreateUser())
	}

	user.Validate(c.Validation)

	if c.Validation.HasErrors() {
		c.Flash.Error("Cannot signup!")
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(routes.PublicApplication.CreateUser())
	}

	user.Hashed = false

	if err := models.CreateUser(user); err != nil {
		stats.TRACE.Println("Failed to create user")
		return c.Redirect(routes.PublicApplication.CreateUser())
	}

	c.Session["user"] = user.UserName

	c.Flash.Success("Welcome, " + user.UserName)

	stats.Incr("App", "Users")

	return c.Redirect(routes.PublicApplication.Index())
}

func (c PublicApplication) Login() revel.Result {
	return c.Render()
}

func (c PublicApplication) CheckLogin() revel.Result {
	var userName string
	var password string

	c.Params.Bind(&userName, "userName")
	c.Params.Bind(&password, "password")

	if models.ValidUserNamePassword(userName, password) {
		c.Session["user"] = userName
		c.Flash.Success("Welcome, " + userName)
		stats.Incr("App", "Login")
		return c.Redirect(routes.PublicApplication.Index())
	} else {
		c.Flash.Error("Failed to login")
		stats.Incr("App", "LoginFailed")

		return c.Redirect(routes.PublicApplication.Login())

	}
}

type notice struct {
	Time time.Time `json:"time"`
	Text string    `json:"text"`
}

var NoticeCache []notice = nil
var lastReadNotice time.Time

func readNotices() ([]notice, error) {
	if NoticeCache == nil || time.Since(lastReadNotice).Minutes() > 10 {
		var notices []notice

		path := filepath.Join(MPFileDirectory, "notices.json")

		if input, err := ioutil.ReadFile(path); err == nil {
			if err = json.Unmarshal(input, &notices); err == nil {
				NoticeCache = notices
				lastReadNotice = time.Now()
				return notices, nil
			}
		}
		return nil, errors.New("Cannot read notices.")
	} else {
		return NoticeCache, nil
	}
}

func (c PublicApplication) Index() revel.Result {
	if notices, err := readNotices(); err == nil {
		c.RenderArgs["notices"] = notices
	}
	return c.Render()
}

func (c PublicApplication) Help() revel.Result {

	help := func() string {
		path := filepath.Join(MPFileDirectory, "help", "description.markdown")

		if input, err := ioutil.ReadFile(path); err == nil {
			html := string(blackfriday.MarkdownCommon(input))
			return html
		}
		return "Help not found"
	}

	c.RenderArgs["help"] = help()

	return c.Render()
}

func (c PublicApplication) LogPageView() revel.Result {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()

	if IsMaster && GeoIP != nil {
		ip := strings.Split(c.Request.RemoteAddr, ":")[0]
		stats.Incr("User", "IP")
		stats.Incr("User", "PageView")
		if user := c.connected(); user.Id != 0 {
			stats.Log("User", "UserName", user.UserName)
		}
		if loc := GeoIP.GetLocationByIP(ip); loc != nil {
			stats.Log("User", "Address", c.Request.RemoteAddr)
			if js, err := json.Marshal(loc); err == nil {
				stats.Log("User", "GeoLocation", string(js))
			}
		}
	}
	return c.RenderJson(map[string]interface{}{
		"status": "success",
	})
}

func (c PublicApplication) LogMessage() revel.Result {
	var packet stats.Packet

	if err := json.NewDecoder(c.Request.Body).Decode(&packet); err == nil {
		if len(stats.Packets) > stats.MAX_PACKET_HISTORY {
			stats.Packets = stats.Packets[:len(stats.Packets)/2]
		}
		stats.Packets = append([]stats.Packet{packet}, stats.Packets...)
	}
	return c.RenderJson(map[string]interface{}{
		"status": "success",
	})
}

func WorkerIsAlive() bool {
	var alive map[string]bool = map[string]bool{}
	if len(stats.Packets) == 0 {
		return false
	}
	for _, s := range stats.Packets {
		if _, ok := alive[s.Address]; ok {
			continue
		}
		diff := time.Since(s.Time)
		alive[s.Address] = diff.Minutes() < 5
	}
	oneIsAlive := false
	for k, v := range alive {
		if v {
			oneIsAlive = true
		} else {
			server.UnregisterWorkerByAddress(k)
		}
	}
	return oneIsAlive
}

func (c PublicApplication) ResetPasswordForm() revel.Result {
	return c.Render()
}

const resetPasswordTemplate = `
Hello,
You have requested for your WebGPU password to be reset. Please follow the instructions bellow.

If you did not make a request to have your password reset, then ignore this email.

Click on the link bellow to reset your password:

	{{ .ResetLink }}

If clicking does not work, then you should copy the link above into the address bar of your browser.

Once at the site, you will be asked to setup a new password.

Thank you.`

func createPasswordSecret(user models.User) string {
	tohash := fmt.Sprint(user.Password, ApplicationSecret, user.Id)
	hasher := sha1.New()
	hasher.Write([]byte(tohash))
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}

func createPasswordResetLink(userName string) (string, error) {
	user, err := models.FindUserByName(userName)
	if err != nil {
		return "", err
	}
	secret := createPasswordSecret(user)
	secretLink := fmt.Sprint(MasterAddress, "/update_password/", user.Id, "/", url.QueryEscape(secret))
	return secretLink, nil
}

func validPasswordResetLink(user models.User, secret string) bool {
	if s := createPasswordSecret(user); s == secret {
		return true
	}
	revel.TRACE.Println(createPasswordSecret(user))
	revel.TRACE.Println(secret)
	return false
}

func (c PublicApplication) ResetPassword() revel.Result {

	var userName string
	var email string

	c.Params.Bind(&userName, "userName")
	c.Params.Bind(&email, "email")

	user, err := models.FindUserByName(userName)
	if err != nil {
		c.Flash.Error("Cannot find user.")
		return c.Redirect(routes.PublicApplication.Index())
	}

	if user.Email != email {
		c.Flash.Error("User's email did not match.")
		return c.Redirect(routes.PublicApplication.Index())
	}

	resetLink, err := createPasswordResetLink(userName)
	if err != nil {
		c.Flash.Error("Cannot generate reset link:  ", err)
		return c.Redirect(routes.PublicApplication.Index())
	}

	t := template.Must(template.New("resetPasswordTemplate").Parse(resetPasswordTemplate))

	type ResetPasswordTemplateParams struct {
		ResetLink string
	}

	params := ResetPasswordTemplateParams{
		ResetLink: resetLink,
	}

	var buf bytes.Buffer

	if err := t.Execute(&buf, params); err != nil {
		revel.TRACE.Println(err)
		c.Flash.Error("Cannot generate email template.")
		return c.Redirect(routes.PublicApplication.Index())
	}

	if err := SendEmail(ApplicationEmail, email, "Password Reset Request", string(buf.Bytes())); err != nil {
		c.Flash.Error("Cannot send email.")
		return c.Redirect(routes.PublicApplication.Index())
	}

	c.Flash.Success("An email has been sent with the password reset link.")
	return c.Redirect(routes.PublicApplication.Index())
}

func (c PublicApplication) UpdatePasswordForm(userIdString, secret string) revel.Result {
	userId, err := strconv.Atoi(userIdString)
	if err != nil {
		c.Flash.Error("Invalid reset password link.")
		return c.Redirect(routes.PublicApplication.Index())
	}
	user, err := models.FindUser(int64(userId))
	if err != nil {
		c.Flash.Error("User not found.")
		return c.Redirect(routes.PublicApplication.Index())
	}
	if validPasswordResetLink(user, secret) {
		c.RenderArgs["user_id"] = userIdString
		c.RenderArgs["secret"] = secret
		return c.Render()
	}
	c.Flash.Error("Password reset link is not valid.")
	return c.Redirect(routes.PublicApplication.Index())
}

func (c PublicApplication) UpdatePassword(userIdString string, secret string) revel.Result {

	var pass, passConfirm string

	c.Params.Bind(&pass, "password")
	c.Params.Bind(&passConfirm, "password_confirm")

	revel.TRACE.Println(pass)
	revel.TRACE.Println(passConfirm)

	redirect := c.Redirect(routes.PublicApplication.UpdatePasswordForm(userIdString, secret))

	if pass != passConfirm {
		c.Flash.Error("The two passwords provided did not match.")
		return redirect
	}
	if pass == "" {
		c.Flash.Error("The two password provided is blank.")
		return redirect
	}

	userId, err := strconv.Atoi(userIdString)
	if err != nil {
		c.Flash.Error("Invalid reset password link.")
		return redirect
	}
	user, err := models.FindUser(int64(userId))
	if err != nil {
		c.Flash.Error("User not found.")
		return redirect
	}
	if !validPasswordResetLink(user, secret) {
		c.Flash.Error("Password reset link is not valid.")
		return redirect
	}

	if err := models.ResetUserPassword(user, pass); err != nil {
		c.Flash.Error("Was not able to rest password.")
		return redirect
	}
	c.Flash.Success("Password has been reset successfully. Please login with your new password.")
	return c.Redirect(routes.PublicApplication.Login())
}

func (c PublicApplication) SharedAttempt(runId string) revel.Result {
	attempt, err := models.FindAttemptByRunId(runId)
	if err != nil {
		c.Flash.Error("Shared attempt not found.")
		return c.Redirect(routes.PublicApplication.Index())
	}

	mp, err := models.FindMachineProblem(attempt.MachineProblemInstanceId)
	if err != nil {
		c.Flash.Error("Cannot query attempt id")
		return c.Render(routes.PublicApplication.Index())
	}

	attempt_c_data := new(server.InternalCData)

	if err := json.Unmarshal([]byte(attempt.InternalCData), attempt_c_data); err == nil {
		c.RenderArgs["attempt_c_data"] = attempt_c_data
	}

	prog, err := models.FindProgram(attempt.ProgramInstanceId)
	if err != nil {
		c.Flash.Error("Cannot find program.")
		return c.Render(routes.PublicApplication.Index())
	}

	conf, _ := ReadMachineProblemConfig(mp.Number)

	c.RenderArgs["mp"] = mp
	c.RenderArgs["title"] = conf.Name + " Shared Attempt"

	c.RenderArgs["mp_config"] = conf
	c.RenderArgs["attempt"] = attempt
	c.RenderArgs["program"] = prog

	if attemptFailed(attempt) {
		c.RenderArgs["status"] = attemptFailedMessage(attempt)
	} else {
		c.RenderArgs["status"] = "Correct solution for this dataset."
	}

	return c.Render()
}
