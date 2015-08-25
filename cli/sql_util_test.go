// Copyright 2015 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License. See the AUTHORS file
// for names of contributors.
//
// Author: Marc Berhault (peter@cockroachlabs.com)

package cli

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/cockroachdb/cockroach/security"
	"github.com/cockroachdb/cockroach/server"
	"github.com/cockroachdb/cockroach/util/leaktest"
)

func makeTestDBClient(t *testing.T, s *server.TestServer) *sql.DB {
	db, err := sql.Open("cockroach", fmt.Sprintf("https://%s@%s?certs=test_certs",
		security.RootUser,
		s.ServingAddr()))
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestRunQuery(t *testing.T) {
	defer leaktest.AfterTest(t)
	s := server.StartTestServer(nil)
	db := makeTestDBClient(t, s)

	defer db.Close()
	defer s.Stop()

	// Override osStdout with our own writer.
	var b bytes.Buffer
	osStdout = &b
	defer func() {
		osStdout = os.Stdout
	}()

	// Non-query statement.
	if err := runQuery(db, `SET DATABASE=system`); err != nil {
		t.Fatal(err)
	}

	expected := `OK
`
	if b.String() != expected {
		t.Fatalf("expected output: %q, got %q", expected, b.String())
	}
	b.Reset()

	// Use system database for sample query/output as they are fairly fixed.
	if err := runQuery(db, `SHOW COLUMNS FROM system.namespace`); err != nil {
		t.Fatal(err)
	}

	expected = `+----------+--------+------+
|  Field   |  Type  | Null |
+----------+--------+------+
| parentID | INT    | true |
| name     | STRING | true |
| id       | INT    | true |
+----------+--------+------+
`
	if b.String() != expected {
		t.Fatalf("expected output: %q, got %q", expected, b.String())
	}
	b.Reset()

	// Test placeholders.
	if err := runQuery(db, `SELECT * FROM system.namespace WHERE name=$1`, "descriptor"); err != nil {
		t.Fatal(err)
	}

	expected = `+----------+------------+----+
| parentID |    name    | id |
+----------+------------+----+
| 1        | descriptor | 3  |
+----------+------------+----+
`
	if b.String() != expected {
		t.Fatalf("expected output: %q, got %q", expected, b.String())
	}
	b.Reset()
}
