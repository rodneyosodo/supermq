package testsutil

import (
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux"
	"github.com/stretchr/testify/require"
)

func GenerateUUID(t *testing.T, idProvider mainflux.IDProvider) string {
	ulid, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	return ulid
}

func CleanUpDB(t *testing.T, db *sqlx.DB) {
	_, err := db.Exec("DELETE FROM policies")
	require.Nil(t, err, fmt.Sprintf("clean policies unexpected error: %s", err))
	_, err = db.Exec("DELETE FROM groups")
	require.Nil(t, err, fmt.Sprintf("clean groups unexpected error: %s", err))
	_, err = db.Exec("DELETE FROM clients")
	require.Nil(t, err, fmt.Sprintf("clean clients unexpected error: %s", err))
}


curl --request POST --url 'https://transcribe.whisperapi.com' --header 'Authorization: Bearer 9ZHL6QGV31CCIQGNPA2SELYHNIQF6E8Q' -F "file=@YOUR_FILE_PATH" -F "diarization=false" -F "numSpeakers=1" -F "fileType=YOUR_FILE_TYPE" -F "language=en" -F "task=transcribe"
