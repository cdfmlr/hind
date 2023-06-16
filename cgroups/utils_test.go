package cgroups

import (
	"io"
	"os"
	"testing"
)

func Test_appendFile(t *testing.T) {
	testFile := "tmp_Test_appendFile.txt"

	// remove test file first if exists
	if _, err := os.Stat(testFile); err == nil {
		t.Logf("test file exists, remove it. testFile=%v", testFile)
		os.Remove(testFile)
	}

	type args struct {
		filePath string
		line     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"0_noexist_create", args{testFile, "It's safe to delete this file."}, false},
		{"1_append_line1", args{testFile, "line1"}, false},
		{"2_append_line2", args{testFile, "line2"}, false},
		// sb vs code sorts the test cases by name 🤬
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := appendFile(tt.args.filePath, tt.args.line); (err != nil) != tt.wantErr {
				t.Errorf("appendFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// read test file and check content
	t.Run("3_check_content", func(t *testing.T) {
		f, err := os.Open(testFile)
		if err != nil {
			t.Fatalf("open test file failed. testFile=%v, err=%v", testFile, err)
		}
		defer f.Close()

		content, err := io.ReadAll(f)
		if err != nil {
			t.Fatalf("read test file failed. testFile=%v, err=%v", testFile, err)
		}

		expected := "It's safe to delete this file.\nline1\nline2\n"
		if string(content) != expected {
			t.Errorf("test file content not match. testFile=%#v, content=%#v, expected=%#v", testFile, string(content), expected)
		}

	})

	// done, remove test file
	if err := os.Remove(testFile); err != nil {
		t.Fatalf("remove test file failed. testFile=%v, err=%v", testFile, err)
	}

	// the great visual studio code has been failing to fold blocks of code with tailing empty lines (or comments), for YEARS.
}

func Test_overwriteFile(t *testing.T) {
	testFile := "tmp_Test_overwriteFile.txt"

	// remove test file first if exists
	if _, err := os.Stat(testFile); err == nil {
		t.Logf("test file exists, remove it. testFile=%v", testFile)
		os.Remove(testFile)
	}

	type args struct {
		filePath string
		content  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"0_noexist_create", args{"tmp_Test_overwriteFile.txt", "It's safe to delete this file."}, false},
		{"1_overwrite", args{"tmp_Test_overwriteFile.txt", "overwrited"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := overwriteFile(tt.args.filePath, tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("overwriteFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// read test file and check content
	t.Run("2_check_content", func(t *testing.T) {
		f, err := os.Open(testFile)
		if err != nil {
			t.Fatalf("open test file failed. testFile=%v, err=%v", testFile, err)
		}
		defer f.Close()

		content, err := io.ReadAll(f)
		if err != nil {
			t.Fatalf("read test file failed. testFile=%v, err=%v", testFile, err)
		}

		expected := "overwrited"
		if string(content) != expected {
			t.Errorf("test file content not match. testFile=%#v, content=%#v, expected=%#v", testFile, string(content), expected)
		}
	})

	// done, remove test file
	if err := os.Remove(testFile); err != nil {
		t.Fatalf("remove test file failed. testFile=%v, err=%v", testFile, err)
	}
}

func Test_mkdirIfNotExists(t *testing.T) {
	if err := mkdirIfNotExists("tmp_Test_mkdirIfNotExists"); err != nil {
		t.Errorf("mkdirIfNotExists() error = %v", err)
	}

	if err := mkdirIfNotExists("tmp_Test_mkdirIfNotExists"); err != nil {
		t.Errorf("mkdirIfNotExists() error = %v", err)
	}

	if err := mkdirIfNotExists("tmp_Test_mkdirIfNotExists/1/2/3"); err != nil {
		t.Errorf("mkdirIfNotExists() error = %v", err)
	}

	if err := mkdirIfNotExists("tmp_Test_mkdirIfNotExists/1/2/3"); err != nil {
		t.Errorf("mkdirIfNotExists() error = %v", err)
	}

	// done, remove test file
	if err := os.RemoveAll("tmp_Test_mkdirIfNotExists"); err != nil {
		t.Fatalf("remove test file failed. testFile=%v, err=%v", "tmp_Test_mkdirIfNotExists", err)
	}
}

func Test_assertFsType(t *testing.T) {
	type args struct {
		path   string
		fsType string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// XXX: for CentOS 7.9 (Kernel 5.4.246-1.el7.elrepo.x86_64) with cgroup v1
		{"cgroup", args{"/sys/fs/cgroup", "tmpfs"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := assertFsType(tt.args.path, tt.args.fsType); (err != nil) != tt.wantErr {
				t.Errorf("assertFsType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
