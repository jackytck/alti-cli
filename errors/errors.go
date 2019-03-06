package errors

const (
	// ErrNoConfig is returned when config file is not found.
	ErrNoConfig AppError = "app: no config"
	// ErrNotLogin is returned when user is not login.
	ErrNotLogin AppError = "app: not login"
	// ErrProfileNotFound is returned when the queried profile is not found.
	ErrProfileNotFound ConfigError = "config: profile not found"
	// ErrProfileNotRemovable is returned when the default profile is chosen to be removed.
	ErrProfileNotRemovable ConfigError = "config: default profile not removable"
	// ErrClientInvisible is returned when the client is invisible to the api server.
	ErrClientInvisible ConfigError = "client: invisible"
	// ErrOffline is returned when the server is offline.
	ErrOffline ServerError = "server: offline"
	// ErrProjCreate is returned when a new project could not be created.
	ErrProjCreate ProjectError = "project: create"
	// ErrProjRemove is returned when a project could not be removed.
	ErrProjRemove ProjectError = "project: remove"
	// ErrProjNotFound is returned when a project is not found.
	ErrProjNotFound ProjectError = "project: not found"
	// ErrFileNotImage is returned when a file is not a supported image.
	ErrFileNotImage FileError = "file: not image"
	// ErrFilesize is returned when the filesize of a file could not be determined.
	ErrFilesize FileError = "file: unknown filesize"
	// ErrFileImageDim is returned when the dimension of an image could not be determined.
	ErrFileImageDim FileError = "file: unknown image dimension"
	// ErrFileChecksum is returned when the checksum of a file could not be computed.
	ErrFileChecksum FileError = "file: unknown checksum"
)

// AppError is the application specific error.
type AppError string

func (e AppError) Error() string {
	return string(e)
}

// ConfigError is the config specific error.
type ConfigError string

func (e ConfigError) Error() string {
	return string(e)
}

// ServerError is the server specific error.
type ServerError string

func (e ServerError) Error() string {
	return string(e)
}

// ProjectError is the project related error.
type ProjectError string

func (e ProjectError) Error() string {
	return string(e)
}

// FileError is the file related error.
type FileError string

func (e FileError) Error() string {
	return string(e)
}
