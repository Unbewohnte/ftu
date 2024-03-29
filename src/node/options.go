/*
ftu - file transferring utility.
Copyright (C) 2021,2022  Kasyanov Nikolay Alexeyevich (Unbewohnte)

This file is a part of ftu

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package node

type SenderNodeOptions struct {
	ServingPath string
	Recursive   bool
}

type ReceiverNodeOptions struct {
	ConnectionAddr      string
	DownloadsFolderPath string
}

// Options to configure the node
type NodeOptions struct {
	IsSending     bool
	WorkingPort   uint
	VerboseOutput bool
	SenderSide    *SenderNodeOptions
	ReceiverSide  *ReceiverNodeOptions
}
