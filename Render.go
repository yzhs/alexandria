// Alexandria
//
// Copyright (C) 2015  Colin Benner
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package alexandria

import (
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type errorHandler struct {
	err error
}

func latexToPdf(id Id) error {
	msg, err := exec.Command("xelatex", "-interaction", "nonstopmode",
		"-output-directory", Config.TempDirectory,
		Config.TempDirectory+string(id)).CombinedOutput()
	if err != nil {
		log.Fatal(string(msg))
		return err
	}
	return nil
}

func pdfToPng(i Id) error {
	id := string(i)
	return exec.Command("convert", "-trim",
		"-quality", strconv.Itoa(Config.Quality),
		"-density", strconv.Itoa(Config.Dpi),
		Config.TempDirectory+id+".pdf", Config.CacheDirectory+id+".png").Run()
}

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

func render(id Id) (string, error) {
	doc, err := readTemplate("header")
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	scroll, err := readScroll(id)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	tags := ParseTags(scroll)

	temp, err := readTemplate(DocumentType(tags) + "_header")
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	doc += temp + scroll

	temp, err = readTemplate(DocumentType(tags) + "_footer")
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	doc += temp

	temp, err = readTemplate("footer")
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	doc += temp

	return doc, nil
}

func processScroll(id Id) error {
	if isUpToDate(id) {
		return nil
	}

	temp, err := render(id)
	if err != nil {
		return err
	}

	err = writeTemp(id, temp)
	if err != nil {
		return err
	}

	err = latexToPdf(id)
	if err != nil {
		return err
	}

	err = pdfToPng(id)
	if err != nil {
		return err
	}

	return nil
}

func processAllScrolls(ids []Id) {
	for _, id := range ids {
		err := processScroll(id)
		if err != nil {
			log.Panic("An error ocurred when processing scroll ", id, ": ", err)
		}
	}
}

func FindScrolls(query []string) ([]Id, error) {
	ids, err := searchSwish(query)
	if err != nil {
		return nil, err
	}
	processAllScrolls(ids)

	return ids, nil
}
