package config

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"time"

	"github.com/russross/blackfriday"
)

type MachineProblemKeywordConfig struct {
	Score float32 `json:"score"`
	Data  string  `json:"data"`
}

type MachineProblemDatasetConfig struct {
	Id          int      `json:"id"`
	Score       float64  `json:"score"`
	Description string   `json:"description"`
	Input       []string `json:"input"`
	Output      string   `json:"output"`
}

type MachineProblemConfig struct {
	Number                    int                           `json:"number"`
	Name                      string                        `json:"name"`
	Week                      int                           `json:"week"`
	CompileScore              float32                       `json:"compile_score"`
	QuestionsScore            float32                       `json:"questions_score"`
	CodeScore                 float32                       `json:"code_score"`
	PeerReviewScore           float32                       `json:"peer_review_score"`
	Keywords                  []MachineProblemKeywordConfig `json:"keyword"`
	Datasets                  []MachineProblemDatasetConfig `json:"data"`
	Language                  string                        `json:"language"`
	InputType                 string                        `json:"input_type"`
	OutputType                string                        `json:"output_type"`
	PeerReviewDeadlineString  string                        `json:"peer_review_deadline"`
	PeerReviewDeadline        time.Time                     `json:"peer_review_deadline_time"`
	CodingDeadlineString      string                        `json:"coding_deadline"`
	CodingDeadline            time.Time                     `json:"coding_deadline_time"`
	CourseraCodePostKey       string                        `json:"coursera_code_post_key"`
	CourseraPeerReviewPostKey string                        `json:"coursera_peer_review_post_key"`
	Description               string                        `json:"-"`
	Questions                 []string                      `json:"-"`
	CodeTemplate              string                        `json:"-"`
	Directory                 string                        `json:"-"`
	GracePeriod               float64                       `json:"-"`
}

var CommonMachineProblemDescription string = ""

var machineProblemConfigCache map[int]*MachineProblemConfig = map[int]*MachineProblemConfig{}

func markdownFile(path string) (string, error) {
	if input, err := ioutil.ReadFile(path); err == nil {
		htmlFlags := 0
		htmlFlags |= blackfriday.HTML_USE_XHTML
		htmlFlags |= blackfriday.HTML_USE_SMARTYPANTS
		htmlFlags |= blackfriday.HTML_SMARTYPANTS_FRACTIONS
		htmlFlags |= blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
		renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")

		// set up the parser
		extensions := 0
		extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
		extensions |= blackfriday.EXTENSION_TABLES
		extensions |= blackfriday.EXTENSION_FENCED_CODE
		extensions |= blackfriday.EXTENSION_AUTOLINK
		extensions |= blackfriday.EXTENSION_STRIKETHROUGH
		extensions |= blackfriday.EXTENSION_SPACE_HEADERS
		extensions |= blackfriday.EXTENSION_FOOTNOTES

		data := blackfriday.Markdown(input, renderer, extensions)
		return string(data), nil
	} else {
		return "", err
	}
}

func readDescription(mp *MachineProblemConfig) {

	path := filepath.Join(mp.Directory, "description.markdown")

	html, err := markdownFile(path)
	if err != nil {
		return
	}

	mp.Description = html

	if CommonMachineProblemDescription == "" {
		path = filepath.Join(MPFileDirectory, "common.markdown")

		if html, err = markdownFile(path); err == nil {
			CommonMachineProblemDescription = html
		}
	}

	mp.Description += CommonMachineProblemDescription
}

func readQuestions(mp *MachineProblemConfig) {

	type data struct {
		Questions []string `json:"questions"`
	}

	var t data

	path := filepath.Join(mp.Directory, "questions.json")

	if input, err := ioutil.ReadFile(path); err == nil {
		if err = json.Unmarshal(input, &t); err == nil {
			mp.Questions = t.Questions
		}
	}
}

func readCodeTemplate(mp *MachineProblemConfig) {
	path := filepath.Join(mp.Directory, "template.cu")

	if txt, err := ioutil.ReadFile(path); err == nil {
		mp.CodeTemplate = string(txt)
	}
}

func populateMachineProblemConfig(mp *MachineProblemConfig) {
	readDescription(mp)
	readQuestions(mp)
	readCodeTemplate(mp)
}

func ReadMachineProblemConfig(mpNum int) (*MachineProblemConfig, error) {

	if mp, ok := machineProblemConfigCache[mpNum]; ok {
		return mp, nil
	}

	var res = new(MachineProblemConfig)

	mpNumString := strconv.Itoa(mpNum)

	path := filepath.Join(MPFileDirectory, mpNumString, "config.json")

	input, err := ioutil.ReadFile(path)
	if err != nil {
		return res, err
	}

	res.Number = mpNum
	res.Directory = filepath.Join(MPFileDirectory, mpNumString)

	res.GracePeriod = 2.0

	if err = json.Unmarshal(input, &res); err == nil {
		for i := range res.Datasets {
			res.Datasets[i].Id = i
		}

		res.CodingDeadline, _ = time.Parse(time.RFC3339, res.CodingDeadlineString)
		res.PeerReviewDeadline, _ = time.Parse(time.RFC3339, res.PeerReviewDeadlineString)
		machineProblemConfigCache[mpNum] = res

		populateMachineProblemConfig(res)
	}

	return res, err
}

func MachineProblemCodingDeadlineExpiredQ(mp *MachineProblemConfig) bool {
	dur := time.Since(mp.CodingDeadline)
	return dur.Hours() > mp.GracePeriod
}

func MachineProblemPeerReviewDeadlineExpiredQ(mp *MachineProblemConfig) bool {
	dur := time.Since(mp.PeerReviewDeadline)
	return dur.Hours() > mp.GracePeriod
}
