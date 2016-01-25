// Alexandria
//
// Copyright (C) 2015  Colin Benner
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package alexandria

import (
	"io"
	"os/exec"
	"strconv"
	"strings"
)

// Run index++ to generate a (new) swish++ index file.
func GenerateIndex() error {
	return exec.Command("index++", "-c", Config.SwishConfig, Config.KnowledgeDirectory).Run()
}

// Search the swish index for a given query.
func searchSwish(query []string) ([]Id, error) {
	tmp := append([]string{"-c", Config.SwishConfig, "--max-results=" + strconv.Itoa(Config.MaxResults)}, query...)
	cmd := exec.Command("search++", tmp...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Start()

	buffer := make([]byte, 1048576)
	bytesRead, _ := io.ReadFull(stdout, buffer)
	cmd.Wait()
	stdout.Close()

	output := strings.Split(string(buffer[:bytesRead]), "\n")
	//num, _ := strconv.Atoi(strings.TrimPrefix(output[0], "# results: "))

	result := make([]Id, Config.MaxResults)
	i := 0
	for _, line := range output {
		if line == "" || strings.HasPrefix(line, "# ") {
			continue
		}
		fields := strings.Fields(line)
		result[i] = Id(strings.TrimSuffix(fields[len(fields)-1], ".tex"))
		i++
	}
	result = result[:i]

	return result, nil
}

// Get a list of scrolls matching the query.
func FindScrolls(query []string) ([]Id, error) {
	ids, err := searchSwish(query)
	if err != nil {
		return nil, err
	}
	ProcessScrolls(ids)

	return ids, nil
}
