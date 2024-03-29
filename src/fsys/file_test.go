/*
ftu - file transferring utility.
Copyright (C) 2021,2022  Kasyanov Nikolay Alexeyevich (Unbewohnte)

This file is a part of ftu

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package fsys

import (
	"io"
	"testing"
)

func Test_GetFile(t *testing.T) {
	filepath := "../testfiles/testfile.txt"

	file, err := GetFile(filepath)
	if err != nil {
		t.Fatalf("GetFile error: %s", err)
	}

	expectedFilename := "testfile.txt"
	if file.Name != expectedFilename {
		t.Fatalf("GetFile error: filenames do not match: expected filename to be %s; got %s", expectedFilename, file.Name)
	}
}

func Test_GetFileOpen(t *testing.T) {
	filepath := "../testfiles/testfile.txt"

	file, err := GetFile(filepath)
	if err != nil {
		t.Fatalf("GetFile error: %s", err)
	}

	err = file.Open()
	if err != nil {
		t.Fatalf("GetFile error: could not open file: %s", err)
	}

	_, err = io.ReadAll(file.Handler)
	if err != nil {
		t.Fatalf("GetFile error: could not read from file: %s", err)
	}
}
