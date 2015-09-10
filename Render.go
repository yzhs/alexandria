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

package main

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

func latexToPdf(id string) error {
	return run("rubber", "--module", "xelatex", "--force", "--into",
		config.tempDirectory, config.tempDirectory+id+".tex")
}

func pdfToPng(id string) error {
	return run("convert", "-quality", strconv.Itoa(config.quality),
		"-density", strconv.Itoa(config.dpi), config.tempDirectory+id+".pdf",
		config.cacheDirectory+id+".png")
}

func searchSwish(query []string) (int, []string, error) {
	tmp := append([]string{"-c", config.swishConfig}, query...)
	cmd := exec.Command("search++", tmp...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, nil, err
	}
	cmd.Start()

	buffer := make([]byte, 65536)
	bytesRead, _ := io.ReadFull(stdout, buffer)
	cmd.Wait()
	stdout.Close()

	output := strings.Split(string(buffer[:bytesRead]), "\n")
	num, _ := strconv.Atoi(strings.TrimPrefix(output[0], "# results: "))
	output = output[1:]

	result := make([]string, 100)
	i := 0
	for _, line := range output {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		result[i] = strings.TrimSuffix(fields[len(fields)-1], ".tex")
		i++
	}
	result = result[:i]

	return num, result, nil
}

func render(id string) (string, error) {
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
	tags := parseTags(scroll)

	temp, err := readTemplate(documentType(tags) + "_header")
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	doc += temp + scroll

	temp, err = readTemplate(documentType(tags) + "_footer")
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

func processScroll(id string) error {
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

func processAllScrolls(ids []string) {
	for _, id := range ids {
		err := processScroll(id)
		if err != nil {
			log.Panic("An error ocurred when processing scroll ", id, ": ", err)
		}
	}
}

func findScrolls(query []string) (int, []string, error) {
	num, ids, err := searchSwish(query)
	if err != nil {
		return 0, nil, err
	}
	processAllScrolls(ids)

	return num, ids, nil
}
