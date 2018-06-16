@test "run sqlite3 test script" {
  run rm -rf test.db

  run $BATS_TEST_DIRNAME/sqlite3/test.sql.exp

  echo "$output" >&3

  [ $status -eq 0 ]
}
