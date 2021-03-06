package types

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

// Project represents the gql project type.
type Project struct {
	ID            string
	Name          string
	IsImported    bool
	ImportedState string
	ProjectType   string
	NumImage      int
	GigaPixel     float64
	TaskState     string
	Date          time.Time
	CloudPath     []CloudPath
	Downloads     DownloadsConnection
}

func (p Project) String() string {
	pat := "ID: %s\tName: %s\tIsImported: %v\tProjectType: %s\tNumImage: %d\tGigaPixel: %.2f\tTaskState: %s\tCloud: %v"
	return fmt.Sprintf(pat, p.ID, p.Name, p.IsImported, p.ProjectType, p.NumImage, p.GigaPixel, p.TaskState, strings.Join(p.Cloud(), ", "))
}

// ProjectHeaderString gives a row of string for the table header.
func ProjectHeaderString() []string {
	return []string{
		"ID",
		"Name",
		"Is Imported",
		"Project Type",
		"Num Image",
		"Giga-Pixel",
		"Task State",
		"Cloud",
		"Date",
		"Model Link",
	}
}

// Cloud returns the cloud keys of the project.
func (p Project) Cloud() []string {
	var ret []string
	for _, c := range p.CloudPath {
		ret = append(ret, c.Key)
	}
	return ret
}

// RowString gives a row of string for the table output.
func (p Project) RowString(webDomain string) []string {
	return []string{
		p.ID,
		p.Name,
		fmt.Sprintf("%v", p.IsImported),
		p.ProjectType,
		fmt.Sprintf("%d", p.NumImage),
		fmt.Sprintf("%.2f", p.GigaPixel),
		p.TaskState,
		strings.Join(p.Cloud(), ", "),
		p.Date.Format("2006-01-02 15:04:05"),
		fmt.Sprintf("%s/project-model?pid=%v", webDomain, p.ID),
	}
}

// ProjectsToTable transforms slice of projects into a table.
func ProjectsToTable(ps []Project, webDomain string, w io.Writer) *tablewriter.Table {
	table := tablewriter.NewWriter(w)
	table.SetHeader(ProjectHeaderString())
	for _, p := range ps {
		table.Append(p.RowString(webDomain))
	}
	return table
}
