package test

import (
	"os"
	"testing"

	"github.com/algorithmiaio/algorithmia-go"
)

var client2 = algorithmia.NewClient(os.Getenv("ALGORITHMIA_DEFAULT_API_KEY"), "")

type some interface{}

func checkDirExists(dd *algorithmia.DataDirectory, onError func(err some)) {
	if exists, err := dd.Exists(); err != nil {
		onError(some(err))
	} else if !exists {
		onError(some("directory must exist"))
	}
}

func checkDirNotExists(dd *algorithmia.DataDirectory, onError func(err some)) {
	if exists, err := dd.Exists(); err != nil {
		onError(some(err))
	} else if exists {
		onError(some("directory must not exist"))
	}
}

func checkFileExists(f *algorithmia.DataFile, onError func(err some)) {
	if exists, err := f.Exists(); err != nil {
		onError(some(err))
	} else if !exists {
		onError(some("file must exist"))
	}
}

func checkFileNotExists(f *algorithmia.DataFile, onError func(err some)) {
	if exists, err := f.Exists(); err != nil {
		onError(some(err))
	} else if exists {
		onError(some("file must not exist"))
	}
}

func TestAcl(t *testing.T) {
	const myPath = "data://.my/privatePermissions"

	dd := client2.Dir(myPath)
	defer dd.ForceDelete()

	if exists, err := dd.Exists(); err != nil {
		t.Fatal(err)
	} else if exists {
		if err := dd.ForceDelete(); err != nil {
			t.Fatal(err)
		}
	}

	if err := dd.Create(algorithmia.ReadAclPrivate); err != nil {
		t.Fatal(err)
	}

	perms, err := client2.Dir(myPath).Permissions()
	if err != nil {
		t.Fatal(err)
	}

	if perms.ReadAcl() != algorithmia.AclTypePrivate {
		t.Fatal("private permissions expected")
	}

	err = dd.UpdatePermissions(algorithmia.ReadAclPublic)
	if err != nil {
		t.Fatal(err)
	}

	perms, err = client2.Dir(myPath).Permissions()
	if err != nil {
		t.Fatal(err)
	}

	if perms.ReadAcl() != algorithmia.AclTypePublic {
		t.Fatal("public permissions expected, got", perms.ReadAcl())
	}
}

func TestDirName(t *testing.T) {
	dd := client2.Dir("data://.my/this/is/a/long/path")
	n, err := dd.Name()
	if err != nil {
		t.Fatal(err)
	}
	if n != "path" {
		t.Fatal("'path' expected for directory name, got", n)
	}
}

func TestDirDoesNotExist(t *testing.T) {
	dd := client2.Dir("data://.my/this_should_never_be_created")

	checkDirNotExists(dd, func(err some) {
		t.Fatal(err)
	})
}

func TestEmptyDirectoryCreationAndDeletion(t *testing.T) {
	dd := client2.Dir("data://.my/empty_test_directory")

	if exists, err := dd.Exists(); err != nil {
		t.Fatal(err)
	} else if exists {
		if err := dd.Delete(); err != nil {
			t.Fatal(err)
		}
	}

	checkDirNotExists(dd, func(err some) {
		t.Fatal(err)
	})

	err := dd.Create(nil)
	if err != nil {
		t.Fatal(err)
	}

	checkDirExists(dd, func(err some) {
		t.Fatal(err)
	})

	if err := dd.Delete(); err != nil {
		t.Fatal(err)
	}

	checkDirNotExists(dd, func(err some) {
		t.Fatal(err)
	})
}

func TestNonemptyDirectoryCreationAndDeletion(t *testing.T) {
	dd := client2.Dir("data://.my/nonempty_test_directory")

	if exists, err := dd.Exists(); err != nil {
		t.Fatal(err)
	} else if exists {
		if err := dd.Delete(); err != nil {
			t.Fatal(err)
		}
	}

	checkDirNotExists(dd, func(err some) {
		t.Fatal(err)
	})

	err := dd.Create(nil)
	if err != nil {
		t.Fatal(err)
	}

	checkDirExists(dd, func(err some) {
		t.Fatal(err)
	})

	f := dd.File("one")
	checkFileNotExists(f, func(err some) {
		t.Fatal(err)
	})

	err = f.Put([]byte("data"))
	if err != nil {
		t.Fatal(err)
	}

	checkFileExists(f, func(err some) {
		t.Fatal(err)
	})

	if err := dd.Delete(); err == nil {
		t.Fatal("removing non-empty directory should fail")
	}

	checkFileExists(f, func(err some) {
		t.Fatal(err)
	})

	checkDirExists(dd, func(err some) {
		t.Fatal(err)
	})

	if err := dd.ForceDelete(); err != nil {
		t.Fatal(err)
	}

	checkDirNotExists(dd, func(err some) {
		t.Fatal(err)
	})
}

func listFilesSmall(t *testing.T, collectionName string) {
	dd := client2.Dir(collectionName)

	if exists, err := dd.Exists(); err != nil {
		t.Fatal(err)
	} else if exists {
		if err := dd.Delete(); err != nil {
			t.Fatal(err)
		}
	}

	err := dd.Create(nil)
	if err != nil {
		t.Fatal(err)
	}

	f1 := dd.File("a")
	err = f1.Put([]byte("data"))
	if err != nil {
		t.Fatal(err)
	}

	f2 := dd.File("b")
	err = f2.Put([]byte("data"))
	if err != nil {
		t.Fatal(err)
	}

	size := 0
	allFiles := map[string]bool{}

	for f := range dd.Files() {
		if f.Err != nil {
			t.Fatal(f.Err)
		}
		size += 1
		allFiles[f.Object.(*algorithmia.DataFile).Path()] = true
	}

	if size != 2 {
		t.Fatal("number of files listed should be 2")
	}

	if _, ok := allFiles[f1.Path()]; !ok {
		t.Fatal("file 'a' not found in collection")
	}

	if _, ok := allFiles[f2.Path()]; !ok {
		t.Fatal("file 'b' not found in collection")
	}

	if err := dd.ForceDelete(); err != nil {
		t.Fatal(err)
	}
}

func TestListFilesSmallWithoutTrailingSlash(t *testing.T) {
	listFilesSmall(t, "data://.my/test_list_files_small")
}

func TestListFilesSmallWithTrailingSlash(t *testing.T) {
	listFilesSmall(t, "data://.my/test_list_files_small/")
}

func TestListFolders(t *testing.T) {
	dd := client2.Dir("data://.my/")
	dirName := ".my/test_list_directory"
	testDir := client2.Dir("data://" + dirName)

	if exists, err := testDir.Exists(); err != nil {
		t.Fatal(err)
	} else if exists {
		if err := testDir.Delete(); err != nil {
			t.Fatal(err)
		}
	}
	allFolders := map[string]bool{}

	for f := range dd.Dirs() {
		if f.Err != nil {
			t.Fatal(f.Err)
		}
		allFolders[f.Object.(*algorithmia.DataDirectory).Path()] = true
	}

	if _, ok := allFolders[dirName]; ok {
		t.Fatal("directory '" + dirName + "' should not be found in collection")
	}

	err := testDir.Create(nil)
	if err != nil {
		t.Fatal(err)
	}

	for f := range dd.Dirs() {
		if f.Err != nil {
			t.Fatal(f.Err)
		}
		allFolders[f.Object.(*algorithmia.DataDirectory).Path()] = true
	}

	if _, ok := allFolders[dirName]; !ok {
		t.Fatal("directory '" + dirName + "' not found in collection")
	}

	err = testDir.ForceDelete()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDataObject(t *testing.T) {
	d := client2.Dir("data://foo")
	if !d.IsDir() {
		t.Fatal("object expected to be a directory")
	}
	if d.IsFile() {
		t.Fatal("object is not expected to be a file")
	}
	if d.Type() != algorithmia.Directory {
		t.Fatal("object expected to be a directory")
	}
}
