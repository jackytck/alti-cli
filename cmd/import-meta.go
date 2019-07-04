package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/jackytck/alti-cli/cloud"
	"github.com/jackytck/alti-cli/gql"
	"github.com/jackytck/alti-cli/service"
	"github.com/spf13/cobra"
)

var meta string
var validNames = []string{"camera.txt", "pose.txt", "group.txt"}

// importMetaCmd represents the meta command
var importMetaCmd = &cobra.Command{
	Use:   "meta",
	Short: "Import meta file to a project",
	Long:  "Import meta files to a project. Recognized filenames are: camera.txt, pose.txt and group.txt.",
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		defer func() {
			if verbose {
				elapsed := time.Since(start)
				log.Println("Took", elapsed)
			}
		}()

		// pre-checks general
		// method = service.SuggestUploadMethod(method, "meta")
		if err := service.Check(
			nil,
			service.CheckAPIServer(),
			// service.CheckUploadMethod("meta", method, ip, port),
			service.CheckPID("meta", id),
			service.CheckFile(meta),
			service.CheckFilenames(meta, validNames),
		); err != nil {
			log.Println(err)
			return
		}

		// get project
		proj, _ := gql.SearchProjectID(id, true)

		// @TODO: direct upload
		var serDone func()

		// set bucket
		method = "s3"
		b, err := service.SuggestBucket(method, bucket, "meta")
		if err != nil {
			log.Println(err)
			return
		}
		bucket = b
		if bucket != "" {
			log.Printf("Bucket %q is chosen", bucket)
		}

		// register + upload + state check
		mru := cloud.MetaFileRegUploader{
			Method:   method,
			PID:      proj.ID,
			MetaPath: meta,
			Filename: filepath.Base(meta),
			Bucket:   bucket,
			Timeout:  timeout,
			Verbose:  verbose,
		}

		// capture and handle ctrl+c
		cc := make(chan os.Signal)
		signal.Notify(cc, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-cc
			fmt.Println()
			if serDone != nil {
				serDone()
			}
			mru.Done()
			log.Println("Bye!")
			os.Exit(1)
		}()

		state, err := mru.Run()
		if err != nil {
			log.Printf(err.Error())
			return
		}

		log.Printf("Successfully registered and uplaoded in state: %q!\n", state)
	},
}

func init() {
	importCmd.AddCommand(importMetaCmd)
	importMetaCmd.Flags().StringVarP(&id, "id", "p", id, "Project id")
	importMetaCmd.Flags().StringVarP(&meta, "file", "f", model, "File path of meta file.")
	importMetaCmd.Flags().StringVarP(&method, "method", "m", method, "Desired method of upload: 'direct' or 's3'")
	importMetaCmd.Flags().IntVarP(&timeout, "timeout", "t", timeout, "Timeout of checking direct upload state in seconds")
	importMetaCmd.Flags().StringVar(&ip, "ip", ip, "IP address of ad-hoc local server for direct upload.")
	importMetaCmd.Flags().StringVar(&port, "port", port, "Port of ad-hoc local server for direct upload.")
	importMetaCmd.Flags().StringVarP(&bucket, "bucket", "b", bucket, "Desired bucket to upload for method: 's3'")
	importMetaCmd.Flags().BoolVarP(&verbose, "verbose", "v", verbose, "Display more info of operation")
	importMetaCmd.MarkFlagRequired("id")
	importMetaCmd.MarkFlagRequired("file")
}
