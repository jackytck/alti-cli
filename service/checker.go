package service

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jackytck/alti-cli/config"
	"github.com/jackytck/alti-cli/errors"
	"github.com/jackytck/alti-cli/file"
	"github.com/jackytck/alti-cli/gql"
	"github.com/jackytck/alti-cli/text"
	"github.com/jackytck/alti-cli/web"
)

// CheckFn represents a checker function.
type CheckFn func(LogFn) error

// LogFn represents a logger function. Same signature as log.Printf.
type LogFn func(string, ...interface{})

// QuietLog is a dummy func for logging nothing.
func QuietLog(string, ...interface{}) {
}

// Check checks all of the passed in checker functions.
func Check(logger LogFn, cs ...CheckFn) error {
	if logger == nil {
		logger = log.Printf
	}
	for _, c := range cs {
		err := c(logger)
		if err != nil {
			return err
		}
	}
	return nil
}

// CheckAPIServer checks if API server is in normal mode.
func CheckAPIServer() CheckFn {
	return func(logger LogFn) error {
		mode := gql.ActiveSystemMode()
		if mode != NormalMode {
			logger("API server is in %q mode.\n", mode)
			logger("Nothing could be uploaded at the moment!\n")
			switch mode {
			case ReadOnlyMode:
				return errors.ErrReadOnly
			}
			return errors.ErrOffline
		}
		return nil
	}
}

// CheckAPIServerLite checks if API server is online, possibly in ReadOnly mode.
func CheckAPIServerLite() CheckFn {
	return func(logger LogFn) error {
		mode := gql.ActiveSystemMode()
		if mode != NormalMode && mode != ReadOnlyMode {
			return errors.ErrOffline
		}
		return nil
	}
}

// CheckUploadMethod checks if the supplied upload method is suppored.
// kind is 'image', 'model' or 'meta'
// if skip is true, this check is skipped. This flag is supposed to be given by
// function service.SuggestUploadMethod.
func CheckUploadMethod(kind, method, ip, port string, skip bool) CheckFn {
	return func(logger LogFn) error {
		if skip {
			return nil
		}
		if method == "" {
			logger("No upload method is provided.")
			return errors.ErrUploadMethodInvalid
		}
		supMethods := gql.SupportedCloud("", "", kind)
		method = strings.ToLower(method)

		// check direct upload
		if method == DirectUploadMethod {
			// if ip and port are provided
			if ip != "" && port != "" {
				err := CheckDirectUploadIPPort(ip, port, logger)
				if err != nil {
					return err
				}
				return nil
			}
			// if ip and port are not provided
			err := CheckDirectUpload(false, logger)
			if err != nil {
				logger("Supported upload methods are: %q!", supMethods)
				return err
			}
			return nil
		}

		// check s3 or oss or minio
		if sm := text.BestMatch(supMethods, method, ""); sm == "" {
			logger("Upload method: %q is not supported!\n", method)
			m := len(supMethods)
			switch m {
			case 0:
				logger("No supported mehtod is found! You could only use 'direct' upload!")
			case 1:
				logger("Only %q upload is supported!", supMethods[0])
			default:
				logger("Supported upload methods are: %q!", supMethods)
			}
			return errors.ErrUploadMethodInvalid
		}
		return nil
	}
}

// CheckDirectUpload checks if direct upload is supported.
func CheckDirectUpload(verbose bool, logger LogFn) error {
	if logger == nil {
		logger = log.Printf
	}
	logger("Checking direct upload...")
	pu, _, err := web.PreferredLocalURL(verbose)
	if err != nil {
		logger("Client is invisible. Direct upload is not supported!")
		return err
	}
	logger("Direct upload is supported over %q\n", pu.Hostname())
	return nil
}

// CheckDirectUploadIPPort checks if the given ip and port could be accessed by
// api server.
func CheckDirectUploadIPPort(ip, port string, logger LogFn) error {
	if logger == nil {
		logger = log.Printf
	}
	_, err := web.CheckVisibilityIPPort(ip, port, true)
	if err != nil {
		url := fmt.Sprintf("http://%s:%s", ip, port)
		logger("%q is not accessible!", url)
		return err
	}
	return nil
}

// CheckPID checks if the pid of the right kind (if provided) exists.
// kind is "image", "model" or "meta".
// Provide an empty string of kind if want to check existence of any kind of project.
func CheckPID(kind, pid string) CheckFn {
	return func(logger LogFn) error {
		p, err := gql.SearchProjectID(pid, true)
		if err != nil {
			logger("Project could not be found! Error:", err)
			return err
		}
		notFound := func() error {
			logger("%q project could nont be found!", kind)
			return errors.ErrProjNotFound
		}
		switch kind {
		case "image":
		case "meta":
			if p.IsImported {
				return notFound()
			}
		case "model":
			if !p.IsImported {
				return notFound()
			}
		}
		return nil
	}
}

// CheckNonNegative checks if the number is non negative.
func CheckNonNegative(i float64) CheckFn {
	return func(logger LogFn) error {
		if i < 0 {
			logger("Cloud not be negative: %.2f", i)
			return errors.ErrInvalidInput
		}
		return nil
	}

}

// CheckBalance checks if the current user has at least min number of coins.
func CheckBalance(min float64) CheckFn {
	return func(logger LogFn) error {
		_, myself, err := gql.MySelf()
		if err != nil {
			return err
		}
		if myself.Balance < min {
			logger("Current balance: %.2f is not enough to pay %.2f coins.", myself.Balance, min)
			return errors.ErrInsufficientCoins
		}
		return nil
	}
}

// CheckFile checks if the file exists.
func CheckFile(f string) CheckFn {
	return func(logger LogFn) error {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			logger("Could not found file: %q", f)
			return err
		}
		return nil
	}
}

// CheckDirOrZip checks if the input is a directory or a zip file.
func CheckDirOrZip(p string) CheckFn {
	cd := CheckDir(p)
	cz := CheckZip(p)
	return func(logger LogFn) error {
		e1 := cd(QuietLog)
		e2 := cz(QuietLog)
		if e1 != nil && e2 != nil {
			return errors.ErrFileNotDirOrZip
		}
		return nil
	}
}

// CheckDir checks if the file is a directory.
func CheckDir(d string) CheckFn {
	return func(logger LogFn) error {
		if fi, err := os.Stat(d); err == nil {
			if fi.Mode().IsDir() {
				return nil
			}
		}
		return errors.ErrFileNotDir
	}
}

// CheckZip checks if the file is a zip.
func CheckZip(f string) CheckFn {
	return func(logger LogFn) error {
		isZip, err := file.IsZipFile(f)
		if err != nil {
			return err
		}
		if !isZip {
			logger("Not a zip file: %q", f)
			return errors.ErrFileNotZip
		}
		return nil
	}
}

// CheckFilename checks if the filename matches with the regex.
func CheckFilename(f string, r *regexp.Regexp) CheckFn {
	return func(logger LogFn) error {
		s := filepath.Base(f)
		if !r.MatchString(s) {
			logger("Invalid file name: %q", s)
			return errors.ErrModelFilenameInvalid
		}
		return nil
	}
}

// CheckFilenames checks if filenames are valid.
func CheckFilenames(filePath string, allowed []string) CheckFn {
	return func(logger LogFn) error {
		filename := filepath.Base(filePath)
		_, valid := text.Contains(allowed, filename)
		if !valid {
			logger("Filename: %q is invalid", filename)
			logger("Filename must be one of: [%v]", strings.Join(allowed, ", "))
			return errors.ErrMetaFilenameInvalid
		}
		return nil
	}
}

// GetMetafilePaths gets all valid metafile paths in the root level of a directory.
func GetMetafilePaths(dir string) ([]string, error) {
	var ret []string

	// check if dir exists
	cd := CheckDir(dir)
	err := cd(QuietLog)
	if err != nil {
		return ret, err
	}

	// find valid meta files
	for _, name := range ValidMetafileNames {
		p := filepath.Join(dir, name)
		if file.IsFileExist(p) {
			ret = append(ret, p)
		}
	}

	return ret, nil
}

// CheckIsLogin check if user has logged in.
func CheckIsLogin() CheckFn {
	return func(logger LogFn) error {
		config := config.Load()
		active := config.GetActive()
		logged := active.Token != ""
		if !logged {
			return errors.ErrNotLogin
		}
		return nil
	}
}
