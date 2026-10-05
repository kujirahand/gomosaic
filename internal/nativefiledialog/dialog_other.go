//go:build !windows && !darwin

package nativefiledialog

func OpenImagePath() (string, error) {
	return "", ErrUnsupported
}

func SaveImagePath(defaultName string) (string, error) {
	return "", ErrUnsupported
}
