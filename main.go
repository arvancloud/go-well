package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

const FileExtension = "go"
const SearchIn = "."

func main() {
	files, err := getFilesIn(SearchIn, FileExtension)
	if err != nil {
		fmt.Printf("failed loading files duo to %s\n", err.Error())
		os.Exit(1)
	}

	wg := sync.WaitGroup{}

	for _, fileName := range files {
		wg.Add(1)

		go func(fileName string) {
			defer wg.Done()
			if err := well(fileName); err != nil {
				fmt.Printf("failed welling %s duo to %s\n", fileName, err.Error())
			}
		}(fileName)
	}

	wg.Wait()
}

func getFilesIn(path, extension string) ([]string, error) {
	libRegEx, err := regexp.Compile("^.+\\.(" + extension + ")$")
	if err != nil {
		return nil, err
	}

	files := make([]string, 0)
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if libRegEx.MatchString(info.Name()) {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func well(fileName string) error {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	importContents, beforeImportContents, afterImportContents := extractImportContents(string(file))
	if len(importContents) == 0 {
		return nil
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
	matched, err := regexp.MatchString("\\b([a-z0-9]+(-[a-z0-9]+)*\\.)+[a-z]{2,}\\b", packageName)
	if err != nil || !matched {
		return false
	}
	return true
}

func extractAliasedPackage(name string) (alias, packageName string) {
	explodedByWhiteSpace := strings.Split(name, " ")
	return explodedByWhiteSpace[0], explodedByWhiteSpace[1]
}

func isAliased(packageName string) bool {
	return strings.Contains(packageName, " ")
}

func sortPackages(packages []string) []string {
	extracted := make(map[string]string, 0)
	for _, packageName := range packages {
		if isAliased(packageName) {
			alias, packageName := extractAliasedPackage(packageName)
			extracted[packageName] = alias
		} else {
			extracted[packageName] = packageName
		}
	}

	keys := make([]string, 0, len(extracted))
	for key := range extracted {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	output := make([]string, 0)
	for _, k := range keys {
		if k == extracted[k] {
			output = append(output, k)
			continue
		}
		output = append(output, extracted[k] + " " + k)
	}

	return output
}

func makeUpImportContents(builtInPackages, externalPackages []string) string {
	output := "import (\n"

	if len(builtInPackages) != 0 {
		output = output + makeUpImportLines(builtInPackages) + "\n"
	}

	if len(externalPackages) != 0 {
		if len(builtInPackages) != 0 {
			output = output + "\n"
		}

		output = output + makeUpImportLines(externalPackages) + "\n"
	}

	output = output + ")"

	return output
}

func makeUpImportLines(packageNames []string) string {
	output := make([]string, 0)
	for _, line := range packageNames {
		output = append(output, "	" + line)
	}

	return strings.Join(output, "\n")
}

func makeFinalPackageName(packageName string, aliasName string) string {
	packageName = "\"" + packageName + "\""
	if len(aliasName) != 0 {
		packageName = aliasName + " " + packageName
	}
	return packageName
}

func writeTo(fileName string, contents []string) error {
	f, err := os.OpenFile(fileName, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}

	var output []byte

	for _, content := range contents {
		output = append(output, content...)
	}

	if err := ioutil.WriteFile(fileName, output, 0666); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}
