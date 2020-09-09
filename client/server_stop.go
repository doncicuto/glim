//
// Copyright © 2020 Muultipla Devops <devops@muultipla.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
//

package client

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a Glim server. Windows systems are not supported.",
	Run: func(cmd *cobra.Command, args []string) {
		// SIGTERM cannot be used with Go in Windows Ref: https://golang.org/pkg/os/#Signal
		if runtime.GOOS == "windows" {
			fmt.Printf("%s [Glim] ⇨ stop command is not supported for Windows platform as it doesn't support the SIGTERM signal. You should terminate Glim process by hand (Ctrl-C). \n", time.Now().Format(time.RFC3339))
			os.Exit(1)
		}

		// Try to read glim.pid file in order to get server's PID
		pidFile := fmt.Sprintf("%s\\glim.pid", os.TempDir())
		data, err := ioutil.ReadFile(pidFile)
		if err != nil {
			fmt.Printf("%s [Glim] ⇨ could not find process file: %s. You should terminate Glim process by hand. \n", pidFile, time.Now().Format(time.RFC3339))
			os.Exit(1)
		}

		pid, err := strconv.Atoi(string(data))
		if err != nil {
			fmt.Printf("%s [Glim] ⇨ could not read PID from %s. You should terminate Glim process by hand. \n", pidFile, time.Now().Format(time.RFC3339))
			os.Exit(1)
		}

		p, err := os.FindProcess(pid)
		if err != nil {
			fmt.Printf("%s [Glim] ⇨ could not find PID in process list. You should terminate Glim process by hand. \n", time.Now().Format(time.RFC3339))
			os.Exit(1)
		}

		err = p.Signal(syscall.SIGTERM)
		if err != nil {
			fmt.Printf("%s [Glim] ⇨ could not send SIGTERM signal to Glim. You should terminate Glim process by hand. \n", time.Now().Format(time.RFC3339))
			os.Exit(1)
		}
	},
}
