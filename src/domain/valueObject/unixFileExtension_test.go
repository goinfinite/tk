package tkValueObject

import "testing"

func TestUnixFileExtension(t *testing.T) {
	t.Run("ValidUnixFileExtension", func(t *testing.T) {
		validUnixFileExtensions := []any{
			".png", "png", ".c", "c", ".ecelp4800", ".n-gage", ".application",
			".fe_launch", ".cdbcmsg",
		}

		for _, extension := range validUnixFileExtensions {
			_, err := NewUnixFileExtension(extension)
			if err != nil {
				t.Errorf(
					"ExpectingNoErrorButGot: %s [%s]", err.Error(), extension,
				)
			}
		}
	})

	t.Run("InvalidUnixFileExtension", func(t *testing.T) {
		invalidUnixFileExtensions := []any{
			"", "file.php?blabla", "@<php52.sandbox.ntorga.com>.php", "../file.php",
			"hello10/info.php",
		}

		for _, extension := range invalidUnixFileExtensions {
			_, err := NewUnixFileExtension(extension)
			if err == nil {
				t.Errorf("ExpectingErrorButDidNotGetFor: %v", extension)
			}
		}
	})
}
