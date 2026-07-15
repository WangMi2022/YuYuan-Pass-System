package upload

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUploadContentTypeForExcel(t *testing.T) {
	require.Equal(t,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		uploadContentType(".xlsx"),
	)
	require.Equal(t, "application/vnd.ms-excel", uploadContentType(".XLS"))
}
