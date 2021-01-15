package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sort"
	"strings"
)

const FileName = "source.txt"

func main() {
	well(FileName)
}

func well(fileName string) error {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	importContents, beforeImportContents, afterImportContents := extractImportContents(string(file))
	if len(importContents) == 0 {
		return fmt.Errorf("there is no import in %s", fileName)
	}

	builtInPackages, externalPackages := categorizePackages(
		normalizeImportLines(importContents),
	)

	builtInPackages = sortPackages(builtInPackages)
	externalPackages = sortPackages(externalPackages)

	importContents = makeUpImportContents(builtInPackages, externalPackages)

	if err := writeTo(fileName, []string{
		beforeImportContents,
		importContents,
		afterImportContents,
	}); err != nil {
		return err
	}

	return nil
}

func sortPackages(packages []string) []string {
	o := make(map[string]string, 0)
	for _, packageName := range packages {
		if isAliased(packageName) {
			alias, packageName := extractAliasedPackage(packageName)
			o[packageName] = alias
		} else {
			o[packageName] = packageName
		}
	}

	keys := make([]string, 0, len(o))
	for key, _ := range o {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	temp := make([]string, 0)
	for _, k := range keys {
		if k == o[k] {
			temp = append(temp, k)
			continue
		}
		temp = append(temp, o[k] + " " + k)
	}

	return temp
}

func writeTo(fileName string, contents []string) error {
	f, err := os.OpenFile(fileName, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	var output []byte

	for _, content := range contents {
		output = append(output, content...)
	}

	ioutil.WriteFile(fileName, output, 0666)

	return nil
}

func makeUpImportContents(builtInPackages, externalPackages []string) string {
	return "import (\n" + makeUpImportLines(builtInPackages) + "\n" + makeUpImportLines(externalPackages) + ")"
}

func makeUpImportLines(packageNames []string) (output string) {
	for _, line := range packageNames {
		output = output + "    " + line + "\n"
	}
	return
}

func categorizePackages(importLines []string) (builtInPackages, externalPackages []string) {
	for _, packageName := range importLines {
		var aliasName string
		if isAliased(packageName) {
			aliasName, packageName = extractAliasedPackage(packageName)
		}

		if strings.Contains(packageName, "/") {
			if isACorrectDomainName(packageName) {
				externalPackages = append(
					externalPackages,
					makeFinalPackageName(packageName, aliasName),
				)
			} else {
				builtInPackages = append(
					builtInPackages,
					makeFinalPackageName(packageName, aliasName),
				)
			}
		} else {
			builtInPackages = append(
				builtInPackages,
				makeFinalPackageName(packageName, aliasName),
			)
		}
	}

	return
}

func isACorrectDomainName(packageName string) bool {
	_, err := net.LookupHost(strings.Split(packageName, "/")[0])
	if err != nil {
		return false
	}
	return true
}

func makeFinalPackageName(packageName string, aliasName string) string {
	packageName = "\"" + packageName + "\""
	if len(aliasName) != 0 {
		packageName = aliasName + " " + packageName
	}
	return packageName
}

func extractAliasedPackage(name string) (alias, packageName string) {
	explodedByWhiteSpace := strings.Split(name, " ")
	return explodedByWhiteSpace[0], explodedByWhiteSpace[1]
}

func isAliased(packageName string) bool {
	return strings.Contains(packageName, " ")
}

func normalizeImportLines(importContent string) []string {
	importLines := strings.Split(importContent, "\n")

	normalizedImportLines := make([]string, 0)

	for _, packageName := range importLines {
		packageName = strings.TrimSpace(packageName)
		packageName = strings.ReplaceAll(packageName, "\"", "")

		if packageName == "" {
			continue
		}

		normalizedImportLines = append(normalizedImportLines, packageName)
	}

	return normalizedImportLines
}

func extractImportContents(content string) (importContent, beforeImportContent, afterImportContent string) {
	startsWith := "import ("
	endsWith := ")"

	startOfImport := strings.Index(content, startsWith)
	endOfImport := strings.Index(content, endsWith)
	if startOfImport < 0 || endOfImport < 0 {
		return
	}

	beforeImportContent = content[0:int64(startOfImport)]
	afterImportContent = content[endOfImport+1:]
	importContent = content[startOfImport+len(startsWith): endOfImport]

	return
}
