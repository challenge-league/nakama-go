/*
Copyright © 2020 Dmitry Kozlov dmitry.f.kozlov@gmail.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package commands

import (
	"bytes"
	"encoding/json"
	"net/url"
	"reflect"
	"sort"
	"text/template"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hako/durafmt"
	"github.com/heroiclabs/nakama-common/api"
	log "github.com/micro/go-micro/v2/logger"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	// https://golang.org/src/time/format.go
	TIME_LAYOUT             = "02 Jan 2006 15:04:05 UTC"
	TIME_START_DEFAULT      = "01 Jan 1970 00:00:00 UTC"
	TIME_END_DEFAULT        = "01 Jan 3000 00:00:00 UTC"
	MAX_LIST_LIMIT          = 100
	DISCORD_BLOCK_CODE_TYPE = "yaml"
	PATCH_NULL_VALUE        = "PATCH_NULL_VALUE"
)

func PatchStructByNewStruct(currentStruct interface{}, patchStruct interface{}) interface{} {
	log.Infof("current %+v", currentStruct)
	log.Infof("patch %+v", patchStruct)
	if reflect.ValueOf(currentStruct).IsNil() {
		log.Infof("result %+v", patchStruct)
		return patchStruct
	}

	patched := false
	cr := reflect.ValueOf(currentStruct).Elem()
	pr := reflect.ValueOf(patchStruct).Elem()
	for i := 0; i < pr.NumField(); i++ {
		crf := cr.Field(i)
		prfv := reflect.Value(pr.Field(i))
		if v, ok := prfv.Interface().(string); ok {
			if v != "" {
				if v == PATCH_NULL_VALUE {
					crf.Set(reflect.ValueOf(""))
				} else {
					crf.Set(prfv)
				}
				patched = true
			}
		}
		if _, ok := prfv.Interface().(int); ok {
			crf.Set(prfv)
			patched = true
		}
	}
	if patched {
		log.Infof("result %+v", currentStruct)
		return currentStruct
	}
	log.Infof("struct was not patched because it has not changed")
	return nil
}

func IsValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme != "http" && u.Scheme != "https" || u.Host == "" {
		return false
	}

	return true
}

func GetKeysFromMap(m interface{}) []string {
	var keys []string
	if rec, ok := m.(map[string]*CaptainsDraftMode); ok {
		for key := range rec {
			keys = append(keys, key)
		}
	}
	/*
		for _, v := range m {
			switch c := v.(type) {
			case CaptainsDraftMode:
				keys = append(keys, c)
			}
		}
	*/
	return keys
}

func IsStringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func Unmarshal(payload string) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	if err := json.Unmarshal([]byte(payload), &params); err != nil {
		return nil, err
	}
	return params, nil
}

func Marshal(v interface{}) []byte {
	resultJSON, err := json.Marshal(v)
	if err != nil {
		log.Error(err)
	}
	return resultJSON
}

func MarshalIndent(v interface{}) string {
	resultJSON, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		log.Error(err)
	}
	return string(resultJSON)
}

func formatTimeAsDate(t time.Time) string {
	return t.Format(TIME_LAYOUT)
}

func formatTimestampAsDate(t timestamp.Timestamp) string {
	tm := time.Unix(t.Seconds, 0)
	return tm.Format(TIME_LAYOUT)
}

func getDurationSinceDate(t time.Time) time.Duration {
	return time.Now().Sub(t)
}

func formatDuraiton(d time.Duration) string {
	return durafmt.Parse(d).LimitFirstN(2).String()
}

func dateIsNotZero(d time.Time) bool {
	return !d.IsZero()
}

func isCaptainsDraft(s string) bool {
	return s == MATCH_TYPE_CAPTAINS_DRAFT
}

func isFloatPositive(f float64) bool {
	return f > 0
}

func isFloatNegative(f float64) bool {
	return f < 0
}

func getCoinsFromWallet(walletString string) float64 {
	log.Infof(walletString)
	var wallet map[string]interface{}
	if err := json.Unmarshal([]byte(walletString), &wallet); err != nil {
		log.Error(err)
	}
	if _, ok := wallet["coins"]; ok {
		return wallet["coins"].(float64)
	}
	return 0
}

func ExecuteTemplate(text string, data interface{}) string {
	fmap := template.FuncMap{
		"formatTimestampAsDate": formatTimestampAsDate,
		"formatTimeAsDate":      formatTimeAsDate,
		"formatDuration":        formatDuraiton,
		"getDurationSinceDate":  getDurationSinceDate,
		"dateIsNotZero":         dateIsNotZero,
		"isFloatPositive":       isFloatPositive,
		"isFloatNegative":       isFloatNegative,
		"getCoinsFromWallet":    getCoinsFromWallet,
		"isCaptainsDraft":       isCaptainsDraft,
	}
	tmpl, err := template.New("").Funcs(fmap).Parse(text)
	if err != nil {
		log.Errorf("%+v %+v, %+v", text, tmpl, err)
		return ""
	}
	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		log.Error("%+v %+v", tmpl, err)
		return ""
	}
	return buffer.String()
}

func writeUserStorageObect(cmdBuilder *commandsBuilder, collection string, key string, value string, version string) error {
	acks, err := cmdBuilder.nakamaCtx.Client.WriteStorageObjects(cmdBuilder.nakamaCtx.Ctx, &api.WriteStorageObjectsRequest{
		Objects: []*api.WriteStorageObject{
			&api.WriteStorageObject{
				Collection: collection,
				Key:        key,
				Value:      value,
				Version:    version,
			},
		},
	})
	if err != nil {
		log.Error(err)
		return err
	}

	if len(acks.Acks) != 1 {
		log.Infof("Invocation failed. Return result not expected: ", len(acks.Acks))
		return err
	}

	return nil
}

func readUserStorageObjects(cmdBuilder *commandsBuilder, collection string, key string, userID string) ([]*api.StorageObject, error) {
	log.Infof("%v ", cmdBuilder)
	log.Infof("%v ", cmdBuilder.nakamaCtx)
	log.Infof("%v ", cmdBuilder.nakamaCtx.Ctx)
	log.Infof("%v %v %v", collection, key, userID)
	storageObjectLists, err := cmdBuilder.nakamaCtx.Client.ReadStorageObjects(cmdBuilder.nakamaCtx.Ctx, &api.ReadStorageObjectsRequest{
		ObjectIds: []*api.ReadStorageObjectId{
			&api.ReadStorageObjectId{
				Collection: collection,
				Key:        key,
				UserId:     userID,
			},
		},
	})

	if err != nil {
		log.Error(err)
		return nil, err
	}
	if len(storageObjectLists.Objects) == 0 {
		log.Infof("storageObjectList is empty")
		return nil, nil
	}

	f := func(storageObject *api.StorageObject) int64 {
		return storageObject.CreateTime.Seconds
	}

	objects := storageObjectLists.Objects
	sort.Slice(objects[:], func(i, j int) bool {
		return f(objects[i]) > f(objects[j])
	})

	return objects, nil

}

func listUserStorageObjects(cmdBuilder *commandsBuilder, collection string, userID string, cursor string) ([]*api.StorageObject, error) {
	storageObjectLists, err := cmdBuilder.nakamaCtx.Client.ListStorageObjects(cmdBuilder.nakamaCtx.Ctx, &api.ListStorageObjectsRequest{
		Collection: collection,
		UserId:     userID,
		Limit:      &wrapperspb.Int32Value{Value: MAX_LIST_LIMIT},
		Cursor:     cursor,
	})

	if err != nil {
		log.Error(err)
		return nil, err
	}
	if len(storageObjectLists.Objects) == 0 {
		log.Infof("storageObjectList is empty")
		return nil, nil
	}

	f := func(storageObject *api.StorageObject) int64 {
		return storageObject.CreateTime.Seconds
	}

	objects := storageObjectLists.Objects
	sort.Slice(objects[:], func(i, j int) bool {
		return f(objects[i]) > f(objects[j])
	})

	return objects, nil
}

type iterableSlice struct {
	i int
	s []string
}

func (s *iterableSlice) Next() (value string) {
	s.i++
	if s.i >= len(s.s) {
		s.i = 0
	}
	return s.s[s.i]
}

func NewSliceIterator(i int, s []string) *iterableSlice {
	return &iterableSlice{i, s}
}
