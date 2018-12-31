// MIT License

// Copyright (c) 2018 Stephen Merrony

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/SMerrony/tello"
	"github.com/simulatedsimian/joystick"
)

var (
	js       joystick.Joystick
	jsConfig joystickConfig
	err      error
)

// Sticks
const (
	axLeftX = iota
	axLeftY
	axRightX
	axRightY
	axL1
	axL2
	axR1
	axR2
)

// Buttons
const (
	btnX = iota
	btnCircle
	btnTriangle
	btnSquare
	btnL1
	btnL2
	btnL3
	btnR1
	btnR2
	btnR3
	btnDL
	btnDR
	btnDU
	btnDD
	btnUnknown
)

// Features
const (
	flipsEnabled = iota
)

const deadZone = 2000

type joystickConfig struct {
	axes     []int
	buttons  []uint
	features []bool
}

var dualShock4Config = joystickConfig{
	axes: []int{
		axLeftX: 0, axLeftY: 1, axRightX: 3, axRightY: 4,
	},
	buttons: []uint{
		btnX: 0, btnCircle: 1, btnTriangle: 2, btnSquare: 3, btnL1: 4,
		btnL2: 6, btnR1: 5, btnR2: 7,
	},
	features: []bool{
		flipsEnabled: false,
	},
}

var eightBitDoSF30Pro = joystickConfig{
	axes: []int{
		axLeftX: 0, axLeftY: 1, axRightX: 2, axRightY: 3,
	},
	// B, A, Y, X, L1, L2, R1, R2
	buttons: []uint{
		btnX: 0, btnCircle: 1, btnTriangle: 3, btnSquare: 2, btnL1: 4,
		btnL2: 6, btnR1: 5, btnR2: 7, btnDL: 13, btnDR: 14, btnDU: 15, btnDD: 16,
	},
	features: []bool{
		flipsEnabled: true,
	},
}

var dualShock4ConfigWin = joystickConfig{
	axes: []int{
		axLeftX: 0, axLeftY: 1, axRightX: 2, axRightY: 3,
	},
	buttons: []uint{
		btnX: 1, btnCircle: 2, btnTriangle: 3, btnSquare: 0, btnL1: 4,
		btnL2: 6, btnR1: 5, btnR2: 7,
	},
	features: []bool{
		flipsEnabled: false,
	},
}

// hotas mapping seems the same on windows and linux
var tflightHotasXConfig = joystickConfig{
	axes: []int{
		axLeftX: 4, axLeftY: 2, axRightX: 0, axRightY: 1,
	},
	buttons: []uint{
		btnR1: 0, btnL1: 1, btnR3: 2, btnL3: 3, btnSquare: 4, btnX: 5,
		btnCircle: 6, btnTriangle: 7, btnR2: 8, btnL2: 9,
	},
	features: []bool{
		flipsEnabled: false,
	},
}

var tflightSteamControllerConfig = joystickConfig{
	axes: []int{
		axLeftX: 0, axLeftY: 1, axRightX: 2, axRightY: 3,
	},
	buttons: []uint{
		btnR1: 7, btnL1: 6, btnR3: 14, btnL3: 13, btnSquare: 4, btnX: 2,
		btnCircle: 3, btnTriangle: 5, btnR2: 9, btnL2: 8,

		btnDL: 19, btnDR: 20, btnDU: 17, btnDD: 18,

		// DTouch = 0
		// R3Touch = 1
		// SELECT = 10
		// START = 11
		// HOME = 12
		// BackL = 15
		// BackR = 16
	},
	features: []bool{
		flipsEnabled: true,
	},
}

func printJoystickHelp() {
	fmt.Print(
		`TelloTerm Joystick Control Mapping

Left Stick   Forward/Backward/Left/Right
Right Stick  Up/Down/Turn

△            Takeoff
╳            Land
○            Take Photo
⌑            Throw takeoff / Palm Land
L1           Slow flight mode
L2           Bounce (on/off)
R1           Fast flight mode
R2           Ultra slow (hold this button for lower sensitivity, does not change flight speed mode)

D-Pad Left    Flip left
D-Pad Right   Flip right
D-Pad Up      Flip forward
D-Pad Down    Flip backward
`)
}

func listJoysticks() {
	for jsid := 0; jsid < 10; jsid++ {
		js, err := joystick.Open(jsid)
		if err != nil {
			if jsid == 0 {
				fmt.Println("No joysticks detected")
			}
			return
		}
		fmt.Printf("Joystick ID: %d: Name: %s, Axes: %d, Buttons: %d\n", jsid, js.Name(), js.AxisCount(), js.ButtonCount())
		js.Close()
	}
}

func setupJoystick(id int) bool {
	if jsTypeFlag == nil || *jsTypeFlag == "" {
		log.Fatalln("No joystick type supplied, please use -jstype option")
	}
	js, err = joystick.Open(id)
	if err != nil {
		log.Fatalf("Could not open specified joystick ID:%d\n", id)
	}
	switch *jsTypeFlag {
	case "DualShock4":
		switch runtime.GOOS {
		case "windows":
			jsConfig = dualShock4ConfigWin
		default:
			jsConfig = dualShock4Config
		}
	case "HotasX":
		jsConfig = tflightHotasXConfig
	case "EightBitDoSF30Pro":
		jsConfig = eightBitDoSF30Pro
	case "SteamController":
		jsConfig = tflightSteamControllerConfig
	default:
		log.Fatalf("Unknown joystick type <%s> supplied\n", *jsTypeFlag)
	}
	// log.Printf("Set up looks good: \n")
	return true
}

func intAbs(x int16) int16 {
	if x < 0 {
		return -x
	}
	return x
}

func readJoystick(test bool) {
	var (
		sm                 tello.StickMessage
		jsState, prevState joystick.State
		err                error
	)

	for {
		jsState, err = js.Read()

		if err != nil {
			log.Printf("Error reading joystick: %v\n", err)
		}

		if jsState.AxisData[jsConfig.axes[axLeftX]] == 32768 {
			sm.Rx = 32767
		} else {
			sm.Rx = int16(jsState.AxisData[jsConfig.axes[axLeftX]])
		}

		if jsState.AxisData[jsConfig.axes[axLeftY]] == 32768 {
			sm.Ry = -32767
		} else {
			sm.Ry = -int16(jsState.AxisData[jsConfig.axes[axLeftY]])
		}

		if jsState.AxisData[jsConfig.axes[axRightX]] == 32768 {
			sm.Lx = 32767
		} else {
			sm.Lx = int16(jsState.AxisData[jsConfig.axes[axRightX]])
		}

		if jsState.AxisData[jsConfig.axes[axRightY]] == 32768 {
			sm.Ly = -32767
		} else {
			sm.Ly = -int16(jsState.AxisData[jsConfig.axes[axRightY]])
		}

		if intAbs(sm.Lx) < deadZone {
			sm.Lx = 0
		}
		if intAbs(sm.Ly) < deadZone {
			sm.Ly = 0
		}
		if intAbs(sm.Rx) < deadZone {
			sm.Rx = 0
		}
		if intAbs(sm.Ry) < deadZone {
			sm.Ry = 0
		}

		if jsState.Buttons&(1<<jsConfig.buttons[btnR2]) != 0 {
			if test && prevState.Buttons&(1<<jsConfig.buttons[btnR2]) == 0 {
				fmt.Println("R2 pressed")
			}

			sm.Lx /= 3
			sm.Ly /= 3
			sm.Rx /= 3
			sm.Ry /= 3
		} else if test && prevState.Buttons&(1<<jsConfig.buttons[btnR2]) != 0 {
			fmt.Println("R2 released")
		}

		if test {
			if sm.Lx != 0 || sm.Ly != 0 || sm.Rx != 0 || sm.Ry != 0 {
				fmt.Printf("JS: Lx: %d, Ly: %d, Rx: %d, Ry: %d\n", sm.Lx, sm.Ly, sm.Rx, sm.Ry)
			}
		} else {
			stickChan <- sm
		}

		if jsState.Buttons&(1<<jsConfig.buttons[btnL1]) != 0 && prevState.Buttons&(1<<jsConfig.buttons[btnL1]) == 0 {
			if test {
				fmt.Println("L1 pressed")
			} else {
				drone.SetSlowMode()
			}
		}
		if jsState.Buttons&(1<<jsConfig.buttons[btnL2]) != 0 && prevState.Buttons&(1<<jsConfig.buttons[btnL2]) == 0 {
			if test {
				fmt.Println("L2 pressed")
			} else {
				drone.Bounce()
			}
		}
		if jsState.Buttons&(1<<jsConfig.buttons[btnR1]) != 0 && prevState.Buttons&(1<<jsConfig.buttons[btnR1]) == 0 {
			if test {
				fmt.Println("R1 pressed")
			} else {
				drone.SetFastMode()
			}
		}

		if jsState.Buttons&(1<<jsConfig.buttons[btnL3]) != 0 && prevState.Buttons&(1<<jsConfig.buttons[btnL3]) == 0 {
			if test {
				fmt.Println("L3 pressed")
			}
		}
		if jsState.Buttons&(1<<jsConfig.buttons[btnR3]) != 0 && prevState.Buttons&(1<<jsConfig.buttons[btnR3]) == 0 {
			if test {
				fmt.Println("R3 pressed")
			}
		}

		if jsState.Buttons&(1<<jsConfig.buttons[btnSquare]) != 0 && prevState.Buttons&(1<<jsConfig.buttons[btnSquare]) == 0 {
			if test {
				fmt.Println("⌑ pressed")
			} else {
				if drone.GetFlightData().Flying {
					drone.PalmLand()
				} else {
					drone.ThrowTakeOff()
				}
			}
		}
		if jsState.Buttons&(1<<jsConfig.buttons[btnTriangle]) != 0 && prevState.Buttons&(1<<jsConfig.buttons[btnTriangle]) == 0 {
			if test {
				fmt.Println("△ pressed")
			} else {
				drone.TakeOff()
			}
		}
		if jsState.Buttons&(1<<jsConfig.buttons[btnCircle]) != 0 && prevState.Buttons&(1<<jsConfig.buttons[btnCircle]) == 0 {
			if test {
				fmt.Println("○ pressed")
			} else {
				drone.TakePicture()
			}
		}
		if jsState.Buttons&(1<<jsConfig.buttons[btnX]) != 0 && prevState.Buttons&(1<<jsConfig.buttons[btnX]) == 0 {
			if test {
				fmt.Println("╳ pressed")
			} else {
				drone.Land()
			}
		}

		// Flip Feature
		if jsConfig.features[flipsEnabled] {
			if jsState.Buttons&(1<<jsConfig.buttons[btnDL]) != 0 && prevState.Buttons&(1<<jsConfig.buttons[btnDL]) == 0 {
				if test {
					fmt.Println("D-Pad Left pressed")
				} else {
					drone.LeftFlip()
				}
			}
			if jsState.Buttons&(1<<jsConfig.buttons[btnDR]) != 0 && prevState.Buttons&(1<<jsConfig.buttons[btnDR]) == 0 {
				if test {
					fmt.Println("D-Pad Right pressed")
				} else {
					drone.RightFlip()
				}
			}
			if jsState.Buttons&(1<<jsConfig.buttons[btnDU]) != 0 && prevState.Buttons&(1<<jsConfig.buttons[btnDU]) == 0 {
				if test {
					fmt.Println("D-Pad Up pressed")
				} else {
					drone.ForwardFlip()
				}
			}
			if jsState.Buttons&(1<<jsConfig.buttons[btnDD]) != 0 && prevState.Buttons&(1<<jsConfig.buttons[btnDD]) == 0 {
				if test {
					fmt.Println("D-Pad Down pressed")
				} else {
					drone.BackFlip()
				}
			}
		}

		prevState = jsState

		if test {
			// Avoid spam of stdout output
			time.Sleep(150 * time.Millisecond)
		} else {
			time.Sleep(updatePeriodMs)
		}
	}
}
