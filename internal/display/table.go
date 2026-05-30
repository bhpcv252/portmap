package display

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/bhpcv252/portmap/internal/iohelp"
)

type Status string

const (
	StatusRunning   Status = "running"
	StatusStopped   Status = "stopped"
	StatusConflict  Status = "conflict"
	StatusUnclaimed Status = "unclaimed"
	StatusFree      Status = "free"
)

type Row struct {
	Port        string
	Project     string
	Service     string
	Status      Status
	Description string
	PID         int // 0 when not running
}

func RenderTable(w io.Writer, rows []Row, unclaimed []Row, noColor bool) error {
	ew := &iohelp.ErrWriter{W: w}

	if len(rows) == 0 && len(unclaimed) == 0 {
		ew.Println("no ports registered")
		return ew.Err
	}

	sortByPort(rows)
	sortByPort(unclaimed)

	// compute column widths dynamically so the table fits any content
	colPort, colProj, colSvc := 4, 7, 7 // minimums match header label lengths
	for _, r := range rows {
		colPort = max(colPort, len(r.Port))
		colProj = max(colProj, len(r.Project))
		colSvc = max(colSvc, len(r.Service))
	}
	// "● running (unclaimed)" is the longest status at 21 display chars
	const colStatus = 21

	hdr := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s  %s",
		colPort, "PORT", colProj, "PROJECT", colSvc, "SERVICE", colStatus, "STATUS", "DESCRIPTION")
	if !noColor {
		hdr = styleHeader.Render(hdr)
	}
	ew.Println(hdr)

	for _, r := range rows {
		plain := plainStatus(r.Status)
		colored := renderStatus(r.Status, noColor)
		pad := strings.Repeat(" ", max(0, colStatus-len(plain)))
		ew.Printf("%-*s  %-*s  %-*s  %s%s  %s\n",
			colPort, r.Port, colProj, r.Project, colSvc, r.Service,
			colored, pad, r.Description)
	}

	if len(unclaimed) > 0 {
		ew.Println("")
		ew.Println("Unclaimed active ports:")
		for _, r := range unclaimed {
			plain := plainStatus(StatusUnclaimed)
			colored := renderStatus(StatusUnclaimed, noColor)
			pad := strings.Repeat(" ", max(0, colStatus-len(plain)))
			ew.Printf("%-*s  %-*s  %-*s  %s%s  %s\n",
				colPort, r.Port, colProj, "-", colSvc, "-",
				colored, pad, "(no claim registered)")
		}
	}

	return ew.Err
}

type jsonRow struct {
	Port        string `json:"port"`
	Project     string `json:"project"`
	Service     string `json:"service"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

func RenderJSON(w io.Writer, rows []Row, unclaimed []Row) error {
	all := make([]jsonRow, 0, len(rows)+len(unclaimed))
	for _, r := range rows {
		all = append(all, jsonRow{
			Port:        r.Port,
			Project:     r.Project,
			Service:     r.Service,
			Status:      string(r.Status),
			Description: r.Description,
		})
	}
	for _, r := range unclaimed {
		all = append(all, jsonRow{
			Port:   r.Port,
			Status: string(StatusRunning),
		})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(all)
}

func sortByPort(rows []Row) {
	sort.Slice(rows, func(i, j int) bool {
		pi, _ := strconv.Atoi(rows[i].Port)
		pj, _ := strconv.Atoi(rows[j].Port)
		return pi < pj
	})
}
