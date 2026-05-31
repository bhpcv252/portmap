package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bhpcv252/portmap/internal/iohelp"
)

var (
	suggestFrom  int
	suggestTo    int
	suggestCount int
)

var suggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "Suggest an available port in a range",
	Args:  cobra.NoArgs,
	RunE:  runSuggest,
}

func init() {
	suggestCmd.Flags().IntVar(&suggestFrom, "from", 3000, "Start of range")
	suggestCmd.Flags().IntVar(&suggestTo, "to", 9999, "End of range")
	suggestCmd.Flags().IntVar(&suggestCount, "count", 1, "Number of ports to suggest")
}

func runSuggest(cmd *cobra.Command, args []string) error {
	activePorts, err := getDetector().ActivePorts()
	if err != nil {
		return err
	}
	activeSet := make(map[int]bool, len(activePorts))
	for _, ap := range activePorts {
		activeSet[ap.Port] = true
	}

	var suggested []int
	for port := suggestFrom; port <= suggestTo && len(suggested) < suggestCount; port++ {
		if !activeSet[port] {
			suggested = append(suggested, port)
		}
	}

	ew := &iohelp.ErrWriter{W: out}

	if len(suggested) == 0 {
		printError(
			fmt.Sprintf("no available port in range %d-%d", suggestFrom, suggestTo),
			"try a different range with --from and --to",
		)
		return errSilent
	}

	if suggestCount == 1 {
		ew.Printf("suggested port: %d\n", suggested[0])
	} else {
		parts := make([]string, len(suggested))
		for i, p := range suggested {
			parts[i] = strconv.Itoa(p)
		}
		ew.Printf("suggested ports: %s\n", strings.Join(parts, ", "))
		if len(suggested) < suggestCount {
			ew.Printf("note: only %d port(s) available in range %d-%d\n",
				len(suggested), suggestFrom, suggestTo)
		}
	}

	return ew.Err
}
