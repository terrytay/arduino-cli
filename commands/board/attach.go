/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package board

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/sketches"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/rpc"
	discovery "github.com/arduino/board-discovery"
	paths "github.com/arduino/go-paths-helper"
)

func BoardAttach(ctx context.Context, req *rpc.BoardAttachReq) (*rpc.BoardAttachResp, error) {

	pm := commands.GetPackageManager(req)
	if pm == nil {
		return nil, errors.New("invalid instance")
	}
	var sketchPath *paths.Path
	if req.GetSketchPath() != "" {
		sketchPath = paths.New(req.GetSketchPath())
	}
	sketch, err := sketches.NewSketchFromPath(sketchPath)
	if err != nil {
		return nil, fmt.Errorf("opening sketch: %s", err)
	}
	if sketch.Metadata == nil {
		formatter.Print("sketch errrorrrerereererer")
	}
	boardURI := req.GetBoardURI()
	fqbn, err := cores.ParseFQBN(boardURI)
	if err != nil && !strings.HasPrefix(boardURI, "serial") {
		boardURI = "serial://" + boardURI
	}

	if fqbn != nil {
		sketch.Metadata.CPU = sketches.BoardMetadata{
			Fqbn: fqbn.String(),
		}
	} else {
		deviceURI, err := url.Parse(boardURI)
		if err != nil {
			return nil, fmt.Errorf("invalid Device URL format: %s", err)
		}

		var findBoardFunc func(*packagemanager.PackageManager, *discovery.Monitor, *url.URL) *cores.Board
		switch deviceURI.Scheme {
		case "serial", "tty":
			findBoardFunc = findSerialConnectedBoard
		case "http", "https", "tcp", "udp":
			findBoardFunc = findNetworkConnectedBoard
		default:
			return nil, fmt.Errorf("invalid device port type provided")
		}

		duration, err := time.ParseDuration(req.GetSearchTimeout())
		if err != nil {
			//logrus.WithError(err).Warnf("Invalid interval `%s` provided, using default (5s).", req.GetSearchTimeout())
			duration = time.Second * 5
		}

		monitor := discovery.New(time.Second)
		monitor.Start()

		time.Sleep(duration)

		// TODO: Handle the case when no board is found.
		board := findBoardFunc(pm, monitor, deviceURI)
		if board == nil {
			return nil, fmt.Errorf("no supported board found at %s", deviceURI.String())
		}
		formatter.Print("Board found: " + board.Name())

		sketch.Metadata.CPU = sketches.BoardMetadata{
			Fqbn: board.FQBN(),
			Name: board.Name(),
		}
	}

	err = sketch.ExportMetadata()
	if err != nil {
<<<<<<< HEAD
		return nil, fmt.Errorf("cannot export sketch metadata: %s", err)
=======
		formatter.PrintError(err, "Cannot export sketch metadata.")
		os.Exit(commands.ErrGeneric)
>>>>>>> 5358b8ed08945f8db0242c77427e79e44f03c7d9
	}
	formatter.PrintResult("Selected fqbn: " + sketch.Metadata.CPU.Fqbn)
	return &rpc.BoardAttachResp{}, nil
}

// FIXME: Those should probably go in a "BoardManager" pkg or something
// findSerialConnectedBoard find the board which is connected to the specified URI via serial port, using a monitor and a set of Boards
// for the matching.
func findSerialConnectedBoard(pm *packagemanager.PackageManager, monitor *discovery.Monitor, deviceURI *url.URL) *cores.Board {
	found := false
	location := deviceURI.Path
	var serialDevice discovery.SerialDevice
	for _, device := range monitor.Serial() {
		if device.Port == location {
			// Found the device !
			found = true
			serialDevice = *device
		}
	}
	if !found {
		return nil
	}

	boards := pm.FindBoardsWithVidPid(serialDevice.VendorID, serialDevice.ProductID)
	if len(boards) == 0 {
		return nil
	}

	return boards[0]
}

// findNetworkConnectedBoard find the board which is connected to the specified URI on the network, using a monitor and a set of Boards
// for the matching.
func findNetworkConnectedBoard(pm *packagemanager.PackageManager, monitor *discovery.Monitor, deviceURI *url.URL) *cores.Board {
	found := false

	var networkDevice discovery.NetworkDevice

	for _, device := range monitor.Network() {
		if device.Address == deviceURI.Host &&
			fmt.Sprint(device.Port) == deviceURI.Port() {
			// Found the device !
			found = true
			networkDevice = *device
		}
	}
	if !found {
		return nil
	}

	boards := pm.FindBoardsWithID(networkDevice.Name)
	if len(boards) == 0 {
		return nil
	}

	return boards[0]
}
