package config

var (
	GithubUser       string
	GithubToken      string
	GithubRepository string
)

func InitGitHubConfig() {

	GithubUser, _ = NestedRevelConfig.String("github.user")
	GithubToken, _ = NestedRevelConfig.String("github.token")
	GithubRepository, _ = NestedRevelConfig.String("github.repository")
}
