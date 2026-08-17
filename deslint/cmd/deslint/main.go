// Command deslint validates design-system source and evidence against a policy.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aprilgom/AnslDes/deslint/internal/lint"
	"github.com/aprilgom/AnslDes/deslint/internal/policy"
	"github.com/aprilgom/AnslDes/deslint/internal/report"
	"github.com/aprilgom/AnslDes/deslint/internal/source"
	"github.com/aprilgom/AnslDes/deslint/internal/source/treesitter"
)

var version = "0.1.0-dev"

var errLintFailed = errors.New("lint failed")

func main() {
	err := run(os.Args[1:])
	if err != nil {
		if !errors.Is(err, errLintFailed) {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	os.Exit(exitCode(err))
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, errLintFailed) {
		return 2
	}
	return 1
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: deslint <version|parse|lint> [arguments]")
	}

	switch args[0] {
	case "version":
		fmt.Println(version)
		return nil
	case "parse":
		if len(args) != 2 {
			return errors.New("usage: deslint parse <file.ts|file.tsx>")
		}
		return parseFile(args[1])
	case "lint":
		return lintFiles(args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

type pathsFlag []string

func (p *pathsFlag) String() string { return fmt.Sprintf("%v", []string(*p)) }
func (p *pathsFlag) Set(value string) error {
	*p = append(*p, value)
	return nil
}

func lintFiles(args []string) error {
	flags := flag.NewFlagSet("lint", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	definitionPath := flags.String("definition", "", "product definition JSON")
	policyPath := flags.String("policy", "", "product lint policy JSON")
	pencilPath := flags.String("pencil", "", "Pencil document JSON")
	layoutPath := flags.String("layout", "", "computed-layout report JSON")
	conformancePath := flags.String("conformance", "", "consumer conformance evidence JSON")
	designContextPath := flags.String("design-context", "", "generated design context JSON")
	formatName := flags.String("format", "text", "text, json, or sarif")
	outputPath := flags.String("out", "-", "output path or - for stdout")
	var sourcePaths pathsFlag
	var visualDetailPaths pathsFlag
	var typographyPaths pathsFlag
	var colorPaths pathsFlag
	var layoutDetailPaths pathsFlag
	var motionPaths pathsFlag
	var copyPaths pathsFlag
	var imageryPaths pathsFlag
	var runtimePaths pathsFlag
	var nativeSourceConformancePaths pathsFlag
	var nativeRuntimeConformancePaths pathsFlag
	var webProviderPaths pathsFlag
	var stageExecutionPaths pathsFlag
	flags.Var(&sourcePaths, "source", "TypeScript or TSX source path; repeatable")
	flags.Var(&visualDetailPaths, "visual-detail", "visual detail evidence JSON; repeatable")
	flags.Var(&typographyPaths, "typography", "typography evidence JSON; repeatable")
	flags.Var(&colorPaths, "color", "theme color evidence JSON; repeatable")
	flags.Var(&layoutDetailPaths, "layout-detail", "semantic layout evidence JSON; repeatable")
	flags.Var(&motionPaths, "motion", "motion and reduced-motion evidence JSON; repeatable")
	flags.Var(&copyPaths, "copy", "locale-aware copy evidence JSON; repeatable")
	flags.Var(&imageryPaths, "imagery", "imagery and asset evidence JSON; repeatable")
	flags.Var(&runtimePaths, "runtime", "Web or native runtime quality evidence JSON; repeatable")
	flags.Var(&nativeSourceConformancePaths, "native-source-conformance", "React Native source conformance evidence JSON; repeatable")
	flags.Var(&nativeRuntimeConformancePaths, "native-runtime-conformance", "simulator, emulator, or device conformance evidence JSON; repeatable")
	flags.Var(&webProviderPaths, "web-provider", "source, static HTML, browser, or visual-contrast provider evidence JSON; repeatable")
	flags.Var(&stageExecutionPaths, "stage-execution", "provider process execution evidence JSON; repeatable")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse lint flags: %w", err)
	}
	if flags.NArg() != 0 || *definitionPath == "" || *policyPath == "" {
		return errors.New("usage: deslint lint --definition FILE --policy FILE [--source FILE] [--pencil FILE] [--layout FILE] [--conformance FILE] [--design-context FILE] [--visual-detail FILE] [--typography FILE] [--color FILE] [--layout-detail FILE] [--motion FILE] [--copy FILE] [--imagery FILE] [--runtime FILE] [--native-source-conformance FILE] [--native-runtime-conformance FILE] [--web-provider FILE] [--format text|json|sarif] [--out FILE]")
	}

	definition, err := readInput(*definitionPath)
	if err != nil {
		return err
	}
	policyContents, err := readFile(*policyPath)
	if err != nil {
		return fmt.Errorf("read policy %s: %w", *policyPath, err)
	}
	productPolicy, err := policy.Parse(policyContents)
	if err != nil {
		return err
	}
	sources := make([]lint.Input, 0, len(sourcePaths))
	for _, path := range sourcePaths {
		sourceInput, readErr := readInput(path)
		if readErr != nil {
			return readErr
		}
		sources = append(sources, sourceInput)
	}
	pencil, err := optionalInput(*pencilPath)
	if err != nil {
		return err
	}
	layoutInput, err := optionalInput(*layoutPath)
	if err != nil {
		return err
	}
	conformanceInput, err := optionalInput(*conformancePath)
	if err != nil {
		return err
	}
	designContextInput, err := optionalInput(*designContextPath)
	if err != nil {
		return err
	}
	visualDetails := make([]lint.Input, 0, len(visualDetailPaths))
	for _, path := range visualDetailPaths {
		visualInput, readErr := readInput(path)
		if readErr != nil {
			return readErr
		}
		visualDetails = append(visualDetails, visualInput)
	}
	typographies := make([]lint.Input, 0, len(typographyPaths))
	for _, path := range typographyPaths {
		typeInput, readErr := readInput(path)
		if readErr != nil {
			return readErr
		}
		typographies = append(typographies, typeInput)
	}
	colors := make([]lint.Input, 0, len(colorPaths))
	for _, path := range colorPaths {
		colorInput, readErr := readInput(path)
		if readErr != nil {
			return readErr
		}
		colors = append(colors, colorInput)
	}
	layoutDetails := make([]lint.Input, 0, len(layoutDetailPaths))
	for _, path := range layoutDetailPaths {
		layoutInput, readErr := readInput(path)
		if readErr != nil {
			return readErr
		}
		layoutDetails = append(layoutDetails, layoutInput)
	}
	motions := make([]lint.Input, 0, len(motionPaths))
	for _, path := range motionPaths {
		motionInput, readErr := readInput(path)
		if readErr != nil {
			return readErr
		}
		motions = append(motions, motionInput)
	}
	copies := make([]lint.Input, 0, len(copyPaths))
	for _, path := range copyPaths {
		copyInput, readErr := readInput(path)
		if readErr != nil {
			return readErr
		}
		copies = append(copies, copyInput)
	}
	imagery := make([]lint.Input, 0, len(imageryPaths))
	for _, path := range imageryPaths {
		imageryInput, readErr := readInput(path)
		if readErr != nil {
			return readErr
		}
		imagery = append(imagery, imageryInput)
	}
	runtimes := make([]lint.Input, 0, len(runtimePaths))
	for _, path := range runtimePaths {
		runtimeInput, readErr := readInput(path)
		if readErr != nil {
			return readErr
		}
		runtimes = append(runtimes, runtimeInput)
	}
	nativeSources := make([]lint.Input, 0, len(nativeSourceConformancePaths))
	for _, path := range nativeSourceConformancePaths {
		nativeInput, readErr := readInput(path)
		if readErr != nil {
			return readErr
		}
		nativeSources = append(nativeSources, nativeInput)
	}
	nativeRuntimes := make([]lint.Input, 0, len(nativeRuntimeConformancePaths))
	for _, path := range nativeRuntimeConformancePaths {
		nativeInput, readErr := readInput(path)
		if readErr != nil {
			return readErr
		}
		nativeRuntimes = append(nativeRuntimes, nativeInput)
	}
	webProviders := make([]lint.Input, 0, len(webProviderPaths))
	for _, path := range webProviderPaths {
		providerInput, readErr := readInput(path)
		if readErr != nil {
			return readErr
		}
		webProviders = append(webProviders, providerInput)
	}
	stageExecutions := make([]lint.Input, 0, len(stageExecutionPaths))
	for _, path := range stageExecutionPaths {
		executionInput, readErr := readInput(path)
		if readErr != nil {
			return readErr
		}
		stageExecutions = append(stageExecutions, executionInput)
	}

	runner := lint.Runner{SourceAnalyzer: treesitter.NewAnalyzer()}
	result, err := runner.Run(lint.Request{
		Definition:      definition,
		Policy:          productPolicy,
		Sources:         sources,
		Pencil:          pencil,
		Layout:          layoutInput,
		Conformance:     conformanceInput,
		DesignContext:   designContextInput,
		VisualDetails:   visualDetails,
		Typographies:    typographies,
		Colors:          colors,
		LayoutDetails:   layoutDetails,
		Motions:         motions,
		Copies:          copies,
		Imagery:         imagery,
		Runtimes:        runtimes,
		NativeSources:   nativeSources,
		NativeRuntimes:  nativeRuntimes,
		WebProviders:    webProviders,
		StageExecutions: stageExecutions,
		Now:             time.Now().UTC(),
	})
	if err != nil {
		return err
	}
	var output bytes.Buffer
	if err := report.Write(&output, result, report.Format(*formatName)); err != nil {
		return err
	}
	if err := writeOutput(*outputPath, output.Bytes()); err != nil {
		return err
	}
	if result.Status == "fail" {
		return errLintFailed
	}
	return nil
}

func readInput(path string) (lint.Input, error) {
	contents, err := readFile(path)
	if err != nil {
		return lint.Input{}, fmt.Errorf("read %s: %w", path, err)
	}
	return lint.Input{Path: filepath.ToSlash(filepath.Clean(path)), Contents: contents}, nil
}

func readFile(path string) ([]byte, error) {
	// #nosec G304,G703 -- Reading caller-selected input paths is the purpose of this CLI.
	return os.ReadFile(path)
}

func optionalInput(path string) (*lint.Input, error) {
	if path == "" {
		return nil, nil
	}
	input, err := readInput(path)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func writeOutput(path string, contents []byte) error {
	if path == "-" {
		_, err := os.Stdout.Write(contents)
		return err
	}
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".deslint-report-*")
	if err != nil {
		return fmt.Errorf("create report: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() {
		_ = os.Remove(temporaryPath)
	}()
	if _, err := temporary.Write(contents); err != nil {
		if closeErr := temporary.Close(); closeErr != nil {
			return errors.Join(
				fmt.Errorf("write report: %w", err),
				fmt.Errorf("close temporary report: %w", closeErr),
			)
		}
		return fmt.Errorf("write report: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close report: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace report: %w", err)
	}
	return nil
}

func parseFile(path string) error {
	contents, err := readFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	language, err := languageForPath(path)
	if err != nil {
		return err
	}

	analyzer := treesitter.NewAnalyzer()
	summary, err := analyzer.Analyze(path, contents, language)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(summary)
}

func languageForPath(path string) (source.Language, error) {
	switch filepath.Ext(path) {
	case ".tsx":
		return source.LanguageTSX, nil
	case ".ts":
		return source.LanguageTypeScript, nil
	default:
		return "", fmt.Errorf("unsupported source extension for %s", path)
	}
}
