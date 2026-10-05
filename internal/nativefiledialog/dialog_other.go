//go:build !windows && !darwin

package nativefiledialog

func OpenImagePath() (string, error) {
	return "", ErrUnsupported
}

func SaveImagePath() (string, error) {
	return "", ErrUnsupported
}
