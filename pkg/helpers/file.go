package helpers

func NormalizeExtension(ext string) string {
	switch ext {
	case ".jpe", ".jpeg":
		return ".jpg"
	default:
		return ext
	}
}
