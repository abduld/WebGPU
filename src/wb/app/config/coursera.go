package config

const CourseraRequestTokenAddress = "https://authentication.coursera.org/auth/oauth/api/request_token"
const CourseraAccessTokenAddress = "https://authentication.coursera.org/auth/oauth/api/access_token"
const CourseraAuthenticationAddress = "https://authentication.coursera.org/auth//oauth/login/index.php"
const CourseraGetIdentityAddress = "https://authentication.coursera.org/auth/oauth/api/get_identity"
const CourseraGetTrustedIdentityAddress = "https://authentication.coursera.org/auth/oauth/api/get_trusted_identity"

var (
	CourseraOAuthConsumerKey    string
	CourseraOAuthConsumerSecret string
	CourseraGradeAPIKey         string
	CourseraGradeURL            string
)

func InitCourseraConfig() {
	CourseraOAuthConsumerKey, _ = NestedRevelConfig.String("coursera.oauth_consumer_key")
	CourseraOAuthConsumerSecret, _ = NestedRevelConfig.String("coursera.oauth_consumer_secret")
	CourseraGradeAPIKey, _ = NestedRevelConfig.String("coursera.grade.api_key")
	CourseraGradeURL, _ = NestedRevelConfig.String("coursera.grade.url")
}
