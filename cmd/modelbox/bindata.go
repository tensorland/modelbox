// Code generated for package command by go-bindata DO NOT EDIT. (@generated)
// sources:
// assets/modelbox_client.toml
// assets/modelbox_server.toml
package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _assetsModelbox_clientToml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x2a\x4e\x2d\x2a\x4b\x2d\x8a\x4f\x4c\x49\x29\x52\xb0\x55\x50\xb2\xb2\x30\xb0\x30\x55\xe2\x02\x04\x00\x00\xff\xff\x39\x18\x96\xde\x16\x00\x00\x00")

func assetsModelbox_clientTomlBytes() ([]byte, error) {
	return bindataRead(
		_assetsModelbox_clientToml,
		"assets/modelbox_client.toml",
	)
}

func assetsModelbox_clientToml() (*asset, error) {
	bytes, err := assetsModelbox_clientTomlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/modelbox_client.toml", size: 22, mode: os.FileMode(436), modTime: time.Unix(1652483967, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsModelbox_serverToml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x84\x92\x41\x6f\xdb\x30\x0c\x85\xef\xfa\x15\x84\x7b\x1d\xd2\x34\x59\xd7\x62\xc0\x0e\x3b\x0e\xd8\x61\x43\x77\x1b\x02\x83\x96\x68\x5b\xa8\x2d\x7a\x22\xb3\xac\xff\x7e\xa0\x52\x37\x46\x13\x20\x3e\x59\x34\xf9\xbe\xc7\x67\xdd\xc0\xaf\x9e\x40\x94\x33\x76\x04\xf2\x22\x4a\x23\xec\x85\x02\xb4\x9c\x4b\x3d\xa6\x0e\x34\x63\x4c\xf6\x82\x59\x63\x8b\x5e\xc5\xd9\x60\x14\x88\x02\x3c\x69\xe4\x84\x03\xc4\x16\x46\x0e\x34\x34\xfc\xcf\xea\xa2\x98\x95\x02\xa0\x00\xc2\x48\x8a\x01\x15\x41\x28\xff\x8d\x9e\x5c\x33\x70\x53\xcf\xdc\x2f\x50\xb5\x71\xa0\x23\xbe\x72\xee\x06\x9e\xae\x38\x7a\xd3\xcb\x34\xa0\x51\x94\x2f\xbb\x94\xbd\xef\xcd\x42\x71\x26\x80\x29\x80\xef\xc9\x3f\x4f\x1c\x93\xca\x87\x52\x18\x62\x22\x83\x71\xbb\x18\x9d\x09\x4b\x93\x31\x29\x75\xd9\x78\xc5\xe4\xb7\xa4\x94\x5b\xf4\x04\x9c\xe0\xd0\x47\xdf\x83\xf6\x74\xb6\xab\x85\x31\x44\x51\x2a\xe6\x6c\x8b\x44\x7a\xe0\xfc\x0c\x9e\x53\x22\x6f\xf1\x89\x3b\x76\xd4\x18\x42\x36\xd4\xe7\xc7\xf5\xe3\x7d\xa1\x7c\x0d\x21\xbe\x26\xec\x39\xb5\xb1\xdb\x67\xb4\x73\x51\x3a\xc5\x06\x0d\x5a\x48\xcb\x5c\xdd\xef\xe5\xa9\x3e\xf5\xee\x9c\x35\xd7\x21\x16\xd4\xad\x8e\xd3\xed\xfc\xe7\x6c\x42\xae\x73\x4f\x49\x2c\xd6\x9d\xa9\xef\xa3\xab\x4f\xdd\x3b\x37\xa1\xf6\x67\xd4\x55\x40\xbd\x0e\x7d\x4f\x82\x98\xe0\x07\x8b\x76\x99\x9e\x7e\x7e\xbf\xc0\x9d\x8e\x1f\x65\xe7\x7a\x16\x35\xea\xdd\xc3\x66\x75\xf7\xb0\x5a\xaf\x36\x95\x9b\x38\x5b\xed\xfe\xe3\x76\xe3\xf6\x42\x25\x8b\x79\xa2\x72\xee\x82\xde\xf8\x22\x7f\x86\x57\x31\x7b\x2e\x09\x1e\xeb\xdb\xed\xfa\x53\x11\x4d\x38\x96\xab\x93\x99\xb5\x72\x13\x8a\x1c\x38\x87\x72\xe3\x99\x2b\x17\x9a\xd2\x50\xa4\xe6\x30\x2a\xf7\x3f\x00\x00\xff\xff\x95\xac\x4c\xed\x96\x03\x00\x00")

func assetsModelbox_serverTomlBytes() ([]byte, error) {
	return bindataRead(
		_assetsModelbox_serverToml,
		"assets/modelbox_server.toml",
	)
}

func assetsModelbox_serverToml() (*asset, error) {
	bytes, err := assetsModelbox_serverTomlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/modelbox_server.toml", size: 918, mode: os.FileMode(436), modTime: time.Unix(1652492063, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"assets/modelbox_client.toml": assetsModelbox_clientToml,
	"assets/modelbox_server.toml": assetsModelbox_serverToml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"assets": &bintree{nil, map[string]*bintree{
		"modelbox_client.toml": &bintree{assetsModelbox_clientToml, map[string]*bintree{}},
		"modelbox_server.toml": &bintree{assetsModelbox_serverToml, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
