package display

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/karrick/gows"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

// ShowPaged displays the `content` in a system dependent pager
func ShowPaged(content string) {
	if isPagerNeeded(content) && (runtime.GOOS == "darwin" || runtime.GOOS == "linux") {
		nixPager(content)
	} else {
		nullPager(content)
	}
}

func isPagerNeeded(content string) bool {
	// only show pager in interactive mode
	if !viper.GetBool(constants.ConfigKeyInteractive) {
		return false
	}

	maxCols, maxRow, _ := gows.GetWinSize()

	// let's scan through it instead of iterating over it fully
	sc := bufio.NewScanner(strings.NewReader(content))

	// explicitly allocate a large bugger for the scanner to use - otherwise we may fail for large rows
	buffSize := 256 * 1024
	buff := make([]byte, buffSize)
	sc.Buffer(buff, buffSize)

	lineCount := 0
	for sc.Scan() {
		line := sc.Text()
		lineCount++
		if lineCount > maxRow {
			return true
		}
		if len(line) > maxCols {
			return true
		}
	}
	return false
}

func nullPager(content string) {
	// just dump the whole thing out
	// we will use this for non-tty environments as well
	fmt.Print(content)
}

func nixPager(content string) {
	if isLessAvailable() {
		execPager(exec.Command("less", "-SRXF"), content)
	} else if isMoreAvailable() {
		execPager(exec.Command("more"), content)
	} else {
		nullPager(content)
	}
}

func isLessAvailable() bool {
	_, err := exec.LookPath("less")
	return err == nil
}

func isMoreAvailable() bool {
	_, err := exec.LookPath("more")
	return err == nil
}

func execPager(cmd *exec.Cmd, content string) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = strings.NewReader(content)
	// run the command - it will block until the pager is exited
	err := cmd.Run()
	if err != nil {
		utils.ShowErrorWithMessage(err, "could not display results")
	}
}
