package display

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/bhpcv252/portmap/internal/iohelp"
)

type Row struct {
	Port    string `json:"port"`
	PID     int    `json:"pid"`
	Process string `json:"process"`
	CWD     string `json:"cwd"`
}

func RenderTable(w io.Writer, rows []Row, noColor bool) error {
	ew := &iohelp.ErrWriter{W: w}

	if len(rows) == 0 {
		ew.Println("no active ports")
		return ew.Err
	}

	// compute column widths from data
	portW, pidW, procW := 4, 3, 7
	for _, r := range rows {
		if len(r.Port) > portW {
			portW = len(r.Port)
		}
		if w := len(fmt.Sprintf("%d", r.PID)); w > pidW {
			pidW = w
		}
		if len(r.Process) > procW {
			procW = len(r.Process)
		}
	}

	ew.Printf("%-*s  %-*s  %-*s  %s\n", portW, "PORT", pidW, "PID", procW, "PROCESS", "CWD")
	for _, r := range rows {
		ew.Printf("%-*s  %-*d  %-*s  %s\n",
			portW, r.Port,
			pidW, r.PID,
			procW, r.Process,
			r.CWD,
		)
	}
	return ew.Err
}

func RenderJSON(w io.Writer, rows []Row) error {
	ew := &iohelp.ErrWriter{W: w}
	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	ew.Printf("%s\n", data)
	return ew.Err
}
