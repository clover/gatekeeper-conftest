package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	jsonutil "k8s.io/apimachinery/pkg/util/json"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"os/exec"
)

type ReviewObject struct {
	Object interface{} `json:"object"`
}

type GatekeeperInput struct {
	Parameters interface{}  `json:"parameters"`
	Review     ReviewObject `json:"review"`
}

var parametersFile string
var exitCode = 0
var tmpPluginDir = ".gatekeeper-conftest/"

var rootCmd = &cobra.Command{
	Use:   "gatekeeper [flags] k8s_file.yaml -- [flags to pass to `conftest test`]",
	Short: "conftest plugin to make input compatible with gatekeeper constrint templates",
	Long: `This plugin tranforms input into the format (input.review.object) expected by
gatekeeper's constraint templates allowing you to reuse the policies with conftest`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Error: Missing input.\n\nHelp:")
			cmd.Help()
			os.Exit(5)
		}
		err := runTests(args, parametersFile)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(5)
		}
		// conftest ignores exit codes 1 and 2, so return 3 and 4 for policy failure instead
		if exitCode != 0 {
			exitCode = exitCode + 2
		}
		os.Exit(exitCode)
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.Flags().StringVarP(&parametersFile, "parameters", "p", "", "path to file conatining paramater values used in constraint templates")
}

func createFile(filename string, contents string) {

}

func runTests(args []string, parametersFile string) error {
	inputFile := args[0]

	//create .conftest directory to store temporary files
	err := os.RemoveAll(tmpPluginDir)
	if err != nil {
		return fmt.Errorf("error removing .conftest directory: %w", err)
	}
	err = os.Mkdir(tmpPluginDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating .conftest directory: %w", err)
	}
	defer os.RemoveAll(tmpPluginDir)

	// build arguments to pass to conftest
	var conftestArgs []string
	conftestArgs = append(conftestArgs, "test")
	args[0] = tmpPluginDir
	conftestArgs = append(conftestArgs, args...)

	// Read parameters file and build parameters object
	var parameters interface{}
	if parametersFile != "" {
		parametersFileInfo, err := os.Stat(inputFile)
		if err != nil {
			return fmt.Errorf("get parameters file info: %w", err)
		}
		if parametersFileInfo.IsDir() {
			return fmt.Errorf("Expecting file for parameters. Got directory.")
		}

		parametersYaml, err := ioutil.ReadFile(parametersFile)
		if err != nil {
			return fmt.Errorf("read parameters file: %w", err)
		}
		parametersJson, err := yamlutil.ToJSON(parametersYaml)
		if err != nil {
			return fmt.Errorf("convert parameters yaml to json: %w", err)
		}
		var parametersInterface interface{}
		err = jsonutil.Unmarshal(parametersJson, &parametersInterface)
		if err != nil {
			return fmt.Errorf("unmarshal parameters json: %w", err)
		}
		parametersMap := parametersInterface.(map[string]interface{})
		parameters = parametersMap["parameters"]
	}

	// Read input manifest file
	var inputFileReader *os.File
	if inputFile == "-" {
		inputFileReader = os.Stdin
	} else {
		fileInfo, err := os.Stat(inputFile)
		if err != nil {
			return fmt.Errorf("get file info: %w", err)
		}
		if fileInfo.IsDir() {
			return fmt.Errorf("Expecting file. Got directory.")
		}

		inputFileReader, err = os.Open(inputFile)
		if err != nil {
			return fmt.Errorf("read input file: %w", err)
		}

	}

	yamlReader := yamlutil.NewDocumentDecoder(inputFileReader)
	yamlBuf := make([]byte, 100*1024)

	// build gatekeeper compatible input object for each yaml document found in input manifest
	for {
		var gatekeeperInputString, inputName, inputKind string

		yamlSize, err := yamlReader.Read(yamlBuf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return fmt.Errorf("read yaml document: %w", err)
			}
		}
		json, err := yamlutil.ToJSON(yamlBuf[:yamlSize])
		if err != nil {
			return fmt.Errorf("converting yaml to json: %w", err)
		}
		var inputInterface interface{}
		err = jsonutil.Unmarshal(json, &inputInterface)
		if err != nil {
			return err
		}
		inputMap := inputInterface.(map[string]interface{})
		if inputMetadata, ok := inputMap["metadata"].(map[string]interface{}); ok {
			if inputNameInterface, ok := inputMetadata["name"]; ok {
				inputName = inputNameInterface.(string)
			}

		}
		if inputKindInterface, ok := inputMap["kind"]; ok {
			inputKind = inputKindInterface.(string)
		}

		// build gatekeeper compatible input object and write to a temp file
		reviewObj := ReviewObject{inputInterface}
		gatekeeperInput := GatekeeperInput{parameters, reviewObj}
		gatekeeperInputJson, err := jsonutil.Marshal(gatekeeperInput)
		if err != nil {
			return err
		}
		gatekeeperInputString = string(gatekeeperInputJson)
		tmpFile, err := os.Create(tmpPluginDir + inputKind + "_" + inputName + ".json")
		if err != nil {
			return fmt.Errorf("error creating temporary file: %w", err)
		}
		_, err = tmpFile.WriteString(gatekeeperInputString)
		if err != nil {
			return fmt.Errorf("error writing to temporary file: %w", err)
		}

		// extract pod spec from template if it exists
		switch inputKind {
		case "DaemonSet":
			fallthrough
		case "ReplicaSet":
			fallthrough
		case "Job":
			fallthrough
		case "StatefulSet":
			fallthrough
		case "Deployment":
			spec := inputMap["spec"].(map[string]interface{})
			reviewObj := ReviewObject{spec["template"]}
			gatekeeperInput := GatekeeperInput{parameters, reviewObj}
			gatekeeperInputJson, err := jsonutil.Marshal(gatekeeperInput)
			if err != nil {
				return err
			}
			gatekeeperInputString = string(gatekeeperInputJson)
			tmpFile, err := os.Create(tmpPluginDir + inputKind + "_" + inputName + "_pods.json")
			if err != nil {
				return fmt.Errorf("error creating temporary file: %w", err)
			}
			_, err = tmpFile.WriteString(gatekeeperInputString)
			if err != nil {
				return fmt.Errorf("error writing to temporary file: %w", err)
			}
		case "CronJob":
			spec := inputMap["spec"].(map[string]interface{})
			jobTemplate := spec["jobTemplate"].(map[string]interface{})
			jobSpec := jobTemplate["spec"].(map[string]interface{})
			reviewObj := ReviewObject{jobSpec["template"]}
			gatekeeperInput := GatekeeperInput{parameters, reviewObj}
			gatekeeperInputJson, err := jsonutil.Marshal(gatekeeperInput)
			if err != nil {
				return err
			}
			gatekeeperInputString = string(gatekeeperInputJson)
			tmpFile, err := os.Create(tmpPluginDir + inputKind + "_" + inputName + "_pods.json")
			if err != nil {
				return fmt.Errorf("error creating temporary file: %w", err)
			}
			_, err = tmpFile.WriteString(gatekeeperInputString)
			if err != nil {
				return fmt.Errorf("error writing to temporary file: %w", err)
			}
		}
		continue
	}

	// run conftest test
	cmd := exec.Command("conftest", conftestArgs...)
	stdout, err := cmd.CombinedOutput()
	fmt.Print(string(stdout))
	if err != nil {
		// update the exit code for policy failure
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}
	return nil
}
