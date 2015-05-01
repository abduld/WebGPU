package config

const CourseraAuthTokenAddress = "https://accounts.coursera.org/oauth2/v1/auth"
const CourseraTokenAddress = "https://accounts.coursera.org/oauth2/v1/token"
const CourseraGetIdentityAddress = "https://api.coursera.org/api/externalBasicProfiles.v1?q=me"

var (
	CourseraOAuthClientKey    string
	CourseraOAuthClientSecret string
	CourseraGradeAPIKey       string
	CourseraGradeURL          string
	CourseraOAuthClientId     string
)

func InitCourseraConfig() {
	CourseraOAuthClientKey, _ = NestedRevelConfig.String("coursera.oauth_client_key")
	CourseraOAuthClientSecret, _ = NestedRevelConfig.String("coursera.oauth_client_secret")
	CourseraGradeAPIKey, _ = NestedRevelConfig.String("coursera.grade.api_key")
	CourseraGradeURL, _ = NestedRevelConfig.String("coursera.grade.url")
	CourseraOAuthClientId, _ = NestedRevelConfig.String("coursera.oauth_client_id")
}
