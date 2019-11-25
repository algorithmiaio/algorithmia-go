package algorithmia

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/algorithmiaio/mapstructure"
)

func getUrl(p string) string {
	return "/v1/data/" + p
}

type DirAttributes struct {
	Name string `json:"name" mapstructure:"name"`
}

type DataDirectory struct {
	DataObjectType

	path string
	url  string

	client *Client
}

func NewDataDirectory(client *Client, dataUrl string) *DataDirectory {
	p := strings.TrimSpace(dataUrl)
	if strings.HasPrefix(p, "data://") {
		p = p[len("data://"):]
	} else if strings.HasPrefix(p, "/") {
		p = p[1:]
	}
	return &DataDirectory{
		DataObjectType: Directory,
		client:         client,
		path:           p,
		url:            getUrl(p),
	}
}

func (f *DataDirectory) SetAttributes(attr *DirAttributes) error {
	return nil
}

func (f *DataDirectory) Exists() (bool, error) {
	resp, err := f.client.getHelper(f.url, url.Values{})
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, err
}

func (f *DataDirectory) Name() (string, error) {
	_, name, err := getParentAndBase(f.path)
	return name, err
}

//Creates a directory, optionally include non-nil Acl argument to set permissions
func (f *DataDirectory) Create(acl *Acl) error {
	parent, name, err := getParentAndBase(f.path)
	if err != nil {
		return err
	}
	jso := map[string]interface{}{
		"name": name,
	}
	if acl != nil {
		jso["acl"] = acl.apiParam()
	}

	resp, err := f.client.postJsonHelper(getUrl(parent), jso, nil)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, err := getRaw(resp)
		if err != nil {
			return err
		}
		return errors.New("Directory creation failed: " + string(b))
	}
	return nil
}

func (f *DataDirectory) doDelete(force bool) error {
	url := f.url
	if force {
		url += "?force=true"
	}

	resp, err := f.client.deleteHelper(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := getRaw(resp)
	if err != nil {
		return err
	}

	err = errorFromJsonData(b)
	if err != nil {
		return err
	}

	return nil
}

// Delete directory if it's empty
func (f *DataDirectory) Delete() error {
	return f.doDelete(false)
}

// Forcibly delete directory (even non-empty)
func (f *DataDirectory) ForceDelete() error {
	return f.doDelete(true)
}

func (f *DataDirectory) File(name string) *DataFile {
	return NewDataFile(f.client, PathJoin(f.path, name))
}

func (f *DataDirectory) Dir(name string) *DataDirectory {
	return NewDataDirectory(f.client, PathJoin(f.path, name))
}

/* Returns permissions for this directory or None if it's a special collection such as
   .session or .algo
*/
func (f *DataDirectory) Permissions() (*Acl, error) {
	v := url.Values{}
	v.Add("acl", "true")
	resp, err := f.client.getHelper(f.url, v)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var m map[string]interface{}
	err = getJson(resp, &m)
	if err != nil {
		return nil, err
	}
	if aclr, ok := m["acl"]; ok {
		var aclResp aclResponse
		if err := mapstructure.Decode(aclr, &aclResp); err == nil {
			acl, err := aclFromResponse(&aclResp)
			return acl, err
		} else {
			return nil, err
		}
	}
	return nil, nil
}

func (f *DataDirectory) UpdatePermissions(acl *Acl) error {
	params := map[string]interface{}{
		"acl": acl.apiParam(),
	}
	resp, err := f.client.patchHelper(f.url, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errorFromResponse(resp)
	}
	return nil
}

type contentResponse struct {
	Marker  string            `json:"marker" mapstructure:"marker"`
	Files   []*FileAttributes `json:"files"  mapstructure:"files"`
	Folders []*DirAttributes  `json:"folders"  mapstructure:"folders"`
}

type SubobjectResult struct {
	Object DataObject
	Err    error
}

func (f *DataDirectory) subObjects(filter DataObjectType) <-chan SubobjectResult {
	ch := make(chan SubobjectResult)
	go func() {
		first := true
		marker := ""
		defer close(ch)
		for first || marker != "" {
			func() {
				first = false
				queryParams := url.Values{}
				if marker != "" {
					queryParams.Add("marker", marker)
				}
				resp, err := f.client.getHelper(f.url, queryParams)
				if err != nil {
					ch <- SubobjectResult{nil, err}
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					ch <- SubobjectResult{nil, errorFromResponse(resp)}
					return
				}

				var content contentResponse
				if err := getJson(resp, &content); err != nil {
					ch <- SubobjectResult{nil, err}
					return
				}

				marker = content.Marker

				getFiles := func() {
					if content.Files == nil {
						return
					}
					for _, fa := range content.Files {
						file := NewDataFile(f.client, PathJoin(f.path, fa.FileName))
						file.SetAttributes(fa)
						ch <- SubobjectResult{file, nil}
					}
				}
				getDirs := func() {
					if content.Folders == nil {
						return
					}
					for _, fa := range content.Folders {
						dir := NewDataDirectory(f.client, PathJoin(f.path, fa.Name))
						dir.SetAttributes(fa)
						ch <- SubobjectResult{dir, nil}
					}
				}

				switch filter {
				case File:
					getFiles()
				case Directory:
					getDirs()
				default:
					getDirs()
					getFiles()
				}
			}()
		}
	}()

	return ch
}

func (f *DataDirectory) Files() <-chan SubobjectResult {
	return f.subObjects(File)
}

func (f *DataDirectory) Dirs() <-chan SubobjectResult {
	return f.subObjects(Directory)
}

func (f *DataDirectory) List() <-chan SubobjectResult {
	return f.subObjects(DataObjectNone)
}

func (f *DataDirectory) Path() string {
	return f.path
}

func (f *DataDirectory) Url() string {
	return f.url
}
