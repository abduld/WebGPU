package bigcode

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/revel/revel"

	. "wb/app/config"
	"wb/app/models"
)

type BigCodeDatabaseEntry struct {
	ProgramId int64  `json:"program_id"`
	Code      string `json:"code"`
	Signature string `json:"signature"`
}
type BigCodeDatabase struct {
	MPNumber int                    `json:"mp_number"`
	Entries  []BigCodeDatabaseEntry `json:"entries"`
}

var databaseCacheMutex *sync.Mutex
var databaseCache map[int]BigCodeDatabase

func LoadOrCreateDatabase(mpNumber int, progs []models.Program) (BigCodeDatabase, error) {
	if val, ok := databaseCache[mpNumber]; ok {
		return val, nil
	}
	dbFile := filepath.Join(MPFileDirectory, strconv.Itoa(mpNumber), "bigcode_database.json")
	revel.TRACE.Println("dbFile = ", dbFile)
	if _, err := os.Stat(dbFile); err == nil {
		db := BigCodeDatabase{}
		bytes, err := ioutil.ReadFile(dbFile)
		if err != nil {
			return db, err
		}
		if err = json.Unmarshal(bytes, &db); err != nil {
			return db, err
		}
		databaseCache[mpNumber] = db
		return db, nil
	}
	if progs == nil {
		return BigCodeDatabase{}, errors.New("The initial programs have not been passed into LoadOrCreateDatabase")
	}
	db, err := CreateDatabase(mpNumber, progs)
	if err == nil {
		databaseCache[mpNumber] = db
		return db, nil
	}
	return db, err
}

func ClosestDatabaseEntry(mpNumber int, sig string) (BigCodeDatabaseEntry, error) {
	min := BigCodeDatabaseEntry{}
	sigVal, err := signatureInteger(sig)
	if err != nil {
		return min, err
	}
	db, err := LoadOrCreateDatabase(mpNumber, nil)
	if err != nil {
		return min, err
	}
	min = db.Entries[0]
	minDist := math.MaxFloat64
	for _, entry := range db.Entries[1:] {
		if entry.Signature != "" {
			if entryVal, err := signatureInteger(entry.Signature); err == nil {
				currDist := tanimotoDistance(entryVal, sigVal)
				if currDist < minDist {
					min = entry
					minDist = currDist
				}
			}
		}
	}
	return min, nil
}

func GetSuggestion(mpNumber int, prog models.Program) ([]BigCodeDatabaseEntry, error) {
	res := []BigCodeDatabaseEntry{}
	db, err := LoadOrCreateDatabase(mpNumber, nil)
	if err != nil {
		revel.TRACE.Println(err)
		return res, err
	}
	sig, err := CreateProgramSignature(mpNumber, prog)
	if err != nil {
		revel.TRACE.Println(err)
		return res, err
	}
	sigVal, err := signatureInteger(sig)
	if err != nil {
		revel.TRACE.Println(err)
		return res, err
	}
	min := db.Entries[0]
	max := db.Entries[0]
	minDist := math.MaxFloat64
	maxDist := -math.MaxFloat64
	found := false
	for _, entry := range db.Entries[1:] {
		if entry.ProgramId == prog.Id {
			found = true
		}
		if entry.ProgramId != prog.Id &&
			entry.Signature != "" &&
			entry.Code != "" &&
			prog.Text != entry.Code {
			if entryVal, err := signatureInteger(entry.Signature); err == nil {
				currDist := tanimotoDistance(entryVal, sigVal)
				if currDist < minDist {
					min = entry
					minDist = currDist
				}
				if currDist > maxDist {
					max = entry
					maxDist = currDist
				}
			}
		}
	}
	if !found {
		databaseCacheMutex.Lock()
		db.Entries = append(db.Entries, BigCodeDatabaseEntry{
			ProgramId: prog.Id,
			Code:      prog.Text,
			Signature: sig,
		})
		databaseCacheMutex.Unlock()
	}
	random := db.Entries[rand.Intn(len(db.Entries))]
	return []BigCodeDatabaseEntry{min, max, random}, nil
}

func getProgramSignatureInDatabase(mpNum int, programId int64) (string, error) {
	db, ok := databaseCache[mpNum]
	if !ok {
		return "", errors.New("the database is not loaded")
	}

	for _, entry := range db.Entries {
		if entry.ProgramId == programId {
			return entry.Signature, nil
		}
	}
	return "", errors.New("cannot find program in the database")
}

func Vote(mpNum int, userProgramId, suggestedProgramId int64, suggestionType string, yesOrNo string) (models.BigcodeVote, error) {

	userProgramSignature, _ := getProgramSignatureInDatabase(mpNum, userProgramId)
	suggestedProgramSignature, _ := getProgramSignatureInDatabase(mpNum, suggestedProgramId)
	voteValue := yesOrNo == "yes"
	vote := models.BigcodeVote{
		UserProgramId:             userProgramId,
		UserProgramSignature:      userProgramSignature,
		SuggestedProgramId:        suggestedProgramId,
		SuggestedProgramSignature: suggestedProgramSignature,
		DistanceMetric:            "tanimoto",
		SuggestionType:            suggestionType,
		Vote:                      voteValue,
	}
	return models.CreateBigcodeVote(vote)
}
