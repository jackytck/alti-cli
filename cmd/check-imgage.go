package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/jackytck/alti-cli/errors"
	"github.com/jackytck/alti-cli/file"
	"github.com/jackytck/alti-cli/gql"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var dir string
var skip string
var verbose bool
var printTable bool
var thread = -1

// checkImageCmd represents the checkImage command
var checkImageCmd = &cobra.Command{
	Use:   "image",
	Short: "Check images of given directory recursively",
	Long: `Compute checksum, find duplicates and compute total giga-pixel
of all images of a given directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		defer func() {
			if verbose {
				elapsed := time.Since(start)
				log.Println("Took", elapsed)
			}
		}()

		log.Printf("Checking %s...\n", dir)

		var totalGP float64
		var totalImg int
		var totalByte datasize.ByteSize

		done := make(chan struct{})
		defer close(done)

		paths, errc := file.WalkFiles(done, dir, skip)
		result := make(chan file.ImageDigest)

		digester := file.ImageDigester{
			Root:   dir,
			Done:   done,
			Paths:  paths,
			Result: result,
		}
		threads := digester.Run(thread)
		if verbose {
			log.Printf("Working in %d thread(s)...", threads)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Filename", "Dimension", "GP", "Size (MB)", "Checksum"})

		for r := range result {
			if r.Error != nil {
				log.Printf("Invalid image: %q, Reason: %v", r.Path, r.Error)
				continue
			}

			mb := file.BytesToMB(r.Filesize)
			if verbose {
				log.Printf("Path: %q, URL: %q, Filename: %q, Dimension: %d x %d, GP: %.2f, Type: %s, Size: %.2f MB, Checksum: %s\n",
					r.Path, r.URL, r.Filename, r.Width, r.Height, r.GP, r.Filetype, mb, r.SHA1)
			}

			if printTable {
				r := []string{
					fmt.Sprintf("%q", r.Filename),
					fmt.Sprintf("%d x %d", r.Width, r.Height),
					fmt.Sprintf("%.2f", r.GP),
					fmt.Sprintf("%.2f", mb),
					r.SHA1,
				}
				table.Append(r)
			}

			totalGP += r.GP
			totalImg++
			totalByte += datasize.ByteSize(r.Filesize)
		}

		// check whether the Walk failed
		if err := <-errc; err != nil {
			panic(err)
		}

		usd, err := gql.CoinsToMoney(totalGP, "USD")
		if err != nil {
			panic(err)
		}
		if totalImg > 0 {
			log.Printf("Found %d images, total %.2f GP, %s, USD $%.2f", totalImg, totalGP, totalByte.HumanReadable(), usd)
		} else {
			log.Println("No image is found!")
		}

		if printTable {
			table.SetFooter([]string{fmt.Sprintf("%d image(s)", totalImg), fmt.Sprintf("USD $%.2f", usd), fmt.Sprintf("%.2f GP", totalGP), totalByte.HumanReadable(), `\ (•◡•) /`})
			table.Render()
		}
	},
}

func init() {
	checkCmd.AddCommand(checkImageCmd)
	checkImageCmd.Flags().StringVarP(&dir, "dir", "d", dir, "Directory path")
	checkImageCmd.Flags().StringVarP(&skip, "skip", "s", skip, "Regular expression to skip paths")
	checkImageCmd.Flags().BoolVarP(&verbose, "verbose", "v", verbose, "Display individual image info")
	checkImageCmd.Flags().BoolVarP(&printTable, "table", "t", printTable, "Output all of the found images in table format")
	checkImageCmd.Flags().IntVarP(&thread, "thread", "n", thread, "Number of threads to process, default is number of cores x 4")
	errors.Must(checkImageCmd.MarkFlagRequired("dir"))
}
