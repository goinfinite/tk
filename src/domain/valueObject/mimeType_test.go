package tkValueObject

import "testing"

func TestMimeType(t *testing.T) {
	t.Run("ValidMimeType", func(t *testing.T) {
		validMimeTypes := []any{
			"directory", "generic", "application/cdmi-object", "application/cdmi-queue",
			"application/cu-seeme", "application/davmount+xml",
			"application/dssc+der", "application/dssc+xml",
			"application/vnd.ms-excel.sheet.macroenabled.12",
			"application/vnd.ms-excel.template.macroenabled.12",
			"video/vnd.ms-playready.media.pyv",
			"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		}

		for _, mimeType := range validMimeTypes {
			_, err := NewMimeType(mimeType)
			if err != nil {
				t.Errorf("ExpectingNoErrorButGot: %s [%s]", err.Error(), mimeType)
			}
		}
	})

	t.Run("InvalidMimeType", func(t *testing.T) {
		invalidMimeTypes := []any{
			"", ".", "..", "blabla", "application+blabla/vnd.ms~excel", "csv",
		}

		for _, mimeType := range invalidMimeTypes {
			_, err := NewMimeType(mimeType)
			if err == nil {
				t.Errorf("ExpectingErrorButDidNotGetFor: %s", mimeType)
			}
		}
	})
}
