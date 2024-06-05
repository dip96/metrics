package utils_test

import (
	"bytes"
	"compress/gzip"
	"github.com/dip96/metrics/internal/utils"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipCompress(t *testing.T) {
	t.Run("compress and decompress data", func(t *testing.T) {
		// Исходные данные для сжатия
		data := []byte("Test")

		// Сжимаем данные
		compressed, err := utils.GzipCompress(data)
		require.NoError(t, err)

		// Проверяем, что сжатые данные не пустые
		assert.NotEmpty(t, compressed)

		// Распаковываем сжатые данные
		r, err := gzip.NewReader(bytes.NewReader(compressed))
		require.NoError(t, err)
		defer r.Close()

		decompressed, err := ioutil.ReadAll(r)
		require.NoError(t, err)

		// Проверяем, что распакованные данные соответствуют исходным
		assert.Equal(t, data, decompressed)
	})

	t.Run("compress empty data", func(t *testing.T) {
		// Пустые данные для сжатия
		data := []byte("")

		// Сжимаем данные
		compressed, err := utils.GzipCompress(data)
		require.NoError(t, err)

		// Проверяем, что сжатые данные не пустые (содержат заголовок gzip)
		assert.NotEmpty(t, compressed)

		// Распаковываем сжатые данные
		r, err := gzip.NewReader(bytes.NewReader(compressed))
		require.NoError(t, err)
		defer r.Close()

		decompressed, err := ioutil.ReadAll(r)
		require.NoError(t, err)

		// Проверяем, что распакованные данные пустые
		assert.Empty(t, decompressed)
	})
}
