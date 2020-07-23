package merge

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

	"github.com/cloudfoundry-community/merge-bosh-dns/config"
)

type KeyList []string
type RecordPair [2]string

type RecordsConfig struct {
	Keys    KeyList                      `json:"record_keys"`
	Infos   [][]interface{}              `json:"record_infos"`
	Aliases map[string][]AliasDefinition `json:"aliases"`
	Version uint64                       `json:"version"`
	Records []RecordPair                 `json:"records"` // ip -> domain
}

type AliasDefinition struct {
	GroupID            string `json:"group_id"`
	RootDomain         string `json:"root_domain"`
	PlaceholderType    string `json:"placeholder_type"`
	HealthFilter       string `json:"health_filter"`
	InitialHealthCheck string `json:"initial_health_check"`
}

//MergeRecordsConfigs joins the contents of the two RecordsConfig objects and
//outputs a new RecordsConfig object with the given version.
func MergeRecordsConfigs(local, remote *RecordsConfig, version uint64) *RecordsConfig {
	if !local.Keys.Equals(remote.Keys) {
		panic(fmt.Sprintf("Key lists were not the same: local.Keys: %s , remote.Keys: %s", local.Keys, remote.Keys))
	}

	ret := &RecordsConfig{
		Keys:    local.Keys,
		Infos:   append(local.Infos, remote.Infos...),
		Aliases: make(map[string][]AliasDefinition),
		Version: version,
		Records: append(local.Records, remote.Records...),
	}

	for k, v := range remote.Aliases {
		ret.Aliases[k] = v
	}

	for k, v := range local.Aliases {
		ret.Aliases[k] = v
	}

	return ret
}

//Serialize outputs a representation of the records config in a format that
//BOSH DNS can read.
func (r *RecordsConfig) Serialize() ([]byte, error) {
	return json.Marshal(r)
}

//Filter removes entries from the RecordsClient that do not match the given
// filter rules
func (r *RecordsConfig) Filter(filter config.MatchSpec) error {
	r.filterDeployments(filter)

	return nil
}

func (r *RecordsConfig) filterDeployments(filter config.MatchSpec) error {
	if len(filter.Deployments) == 0 {
		return nil
	}

	//Compile all provided regex statements
	regexList := make([]*regexp.Regexp, 0, len(filter.Deployments))
	for _, depName := range filter.Deployments {
		thisRegex, err := regexp.Compile(depName)
		if err != nil {
			return fmt.Errorf("Error compiling regexp: %s", err)
		}

		regexList = append(regexList, thisRegex)
	}

	//Get index of infos deployment name
	const deploymentNameKey = "deployment"
	depNameIdx := r.Keys.indexForKey(deploymentNameKey)
	if depNameIdx < 0 {
		//If there's no deployment name, we can't really filter out anything.
		// This may happen if the file is empty, so we don't want to be too loud
		// about it.
		return nil
	}

	//HACK: This is a hack and a half. Please forgive me. Because the BOSH
	//director appends to both of these lists using the same function, these
	//lists have records that are parallel to each other. If this ever changes,
	//welp, this is gonna break. But, considering the DNS name is built from the
	//parallel struct in Infos, this seems unlikely to change anytime soon.
	if len(r.Infos) != len(r.Records) {
		panic("This whole thing assumes that these two lists are parallel. Apparently they're not?")
	}

	newInfos := [][]interface{}{}
	newRecords := []RecordPair{}
	for i, info := range r.Infos {
		keep := false

		for _, matchWith := range regexList {
			depNameAsString, isString := info[depNameIdx].(string)
			if !isString {
				return fmt.Errorf("deployment name in config was not a string")
			}

			if matchWith.MatchString(depNameAsString) {
				keep = true
				break
			}
		}

		if keep {
			newInfos = append(newInfos, r.Infos[i])
			newRecords = append(newRecords, r.Records[i])
		}
	}

	r.Infos = newInfos
	r.Records = newRecords

	return nil
}

func (k KeyList) indexForKey(key string) int {
	for i := range k {
		if k[i] == key {
			return i
		}
	}
	return -1
}

func RecordsConfigFromFilepath(filepath string) (*RecordsConfig, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	return DecodeRecordsConfig(f)
}

func DecodeRecordsConfig(rd io.Reader) (*RecordsConfig, error) {
	ret := &RecordsConfig{}
	dec := json.NewDecoder(rd)
	err := dec.Decode(&ret)
	return ret, err
}

type RecordsClient struct {
	URL       string
	Client    *http.Client
	BasicAuth *BasicAuthCredentials
}

type BasicAuthCredentials struct {
	Username string
	Password string
}

func (r *RecordsClient) FetchRecordsConfig() (*RecordsConfig, error) {
	if r.URL == "" {
		return nil, fmt.Errorf("No URL was configured to the remote client")
	}

	client := http.DefaultClient
	if r.Client != nil {
		client = r.Client
	}

	req, err := http.NewRequest("GET", r.URL, nil)
	if err != nil {
		return nil, err
	}

	if r.BasicAuth != nil {
		req.SetBasicAuth(r.BasicAuth.Username, r.BasicAuth.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		ioutil.ReadAll(resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("Non-2xx Status Code (%d) received from `%s`", resp.StatusCode, r.URL)
	}

	return DecodeRecordsConfig(resp.Body)
}

//Equals returns true if both KeyList objects have the same length and same
// contents in the same order. Returns false otherwise.
func (k KeyList) Equals(k2 KeyList) bool {
	if len(k) != len(k2) {
		return false
	}

	for i := range k {
		if k[i] != k2[i] {
			return false
		}
	}

	return true
}
