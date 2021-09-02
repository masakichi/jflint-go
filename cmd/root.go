/*
Copyright © 2021 Yuanji <self@gimo.me>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type CrumbHeader struct {
	Key   string
	Value string
}

type BasicAuth struct {
	Username string
	Password string
}

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "jflint-go Jenkinsfile",
	Short: "A helper tool to lint a Declarative Jenkinsfile",
	Long: `jflint-go helps to lint a Declarative Jenkinsfile.

This tool itself does not lint a Jenkinsfile,
but sends a request to Jenkins in the same way
as curl approach and displays the result.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		jenkinsUrl := viper.GetString("jenkinsUrl")
		if jenkinsUrl == "" {
			log.Fatal("Error: jenkins URL must be specified")
		}

		basicAuth := BasicAuth{viper.GetString("username"), viper.GetString("password")}
		// Read Jenkinsfile content
		data, err := os.ReadFile(args[0])
		if err != nil {
			log.Fatal(err)
		}
		jenkinsFile := string(data)

		jenkinsCrumbUrl := jenkinsUrl + "/crumbIssuer/api/xml?xpath=concat(//crumbRequestField,\":\",//crumb)"
		jenkinsValidateUrl := jenkinsUrl + "/pipeline-model-converter/validate"

		client := &http.Client{}

		crumbHeader := CrumbHeader{}
		csrfDisabled := viper.GetBool("csrfDisabled")
		if !csrfDisabled {
			crumbHeader, err = fetchCrumbHeader(client, "GET", basicAuth, jenkinsCrumbUrl)
			if err != nil {
				log.Fatal(err)
			}
		}
		result := validate(client, "POST", basicAuth, jenkinsValidateUrl, jenkinsFile, crumbHeader)
		fmt.Println(result)
	},
}

func fetchCrumbHeader(client *http.Client, method string, basicAuth BasicAuth, jenkinsCrumbUrl string) (CrumbHeader, error) {
	req, err := http.NewRequest(method, jenkinsCrumbUrl, nil)
	if err != nil {
		log.Fatal(err)
	}
	if (basicAuth != BasicAuth{}) {
		req.SetBasicAuth(basicAuth.Username, basicAuth.Password)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	crumbString := string(body)
	crumbSlice := strings.Split(crumbString, ":")
	if len(crumbSlice) != 2 {
		return CrumbHeader{}, errors.New("failed parse jenkins crumb header")
	}
	return CrumbHeader{crumbSlice[0], crumbSlice[1]}, nil
}

func validate(client *http.Client, method string, basicAuth BasicAuth, jenkinsValidateUrl string, jenkinsFile string, crumbHeader CrumbHeader) string {
	data := url.Values{
		"jenkinsfile": []string{jenkinsFile},
	}
	reqBody := strings.NewReader(data.Encode())

	req, err := http.NewRequest(method, jenkinsValidateUrl, reqBody)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if (crumbHeader != CrumbHeader{}) {
		req.Header.Set(crumbHeader.Key, crumbHeader.Value)
	}
	if (basicAuth != BasicAuth{}) {
		req.SetBasicAuth(basicAuth.Username, basicAuth.Password)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(body)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.jflintrc)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringP("username", "u", "", "Specify username on Jenkins")
	rootCmd.Flags().StringP("password", "p", "", "Specify password/API token on Jenkins")
	rootCmd.Flags().StringP("jenkins-url", "j", "", "Specify Jenkins URL")
	rootCmd.Flags().Bool("csrf-disabled", false, "Specify when CSRF security setting is disabled on Jenkins.")

	viper.BindPFlag("username", rootCmd.Flags().Lookup("username"))
	viper.BindPFlag("password", rootCmd.Flags().Lookup("password"))
	viper.BindPFlag("jenkinsUrl", rootCmd.Flags().Lookup("jenkins-url"))
	viper.BindPFlag("csrfDisabled", rootCmd.Flags().Lookup("csrf-disabled"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".jflintrc" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("json")
		viper.SetConfigName(".jflintrc")
	}

	// env variable should start with JFLINTGO, eg: JFLINTGO_USERNAME
	viper.SetEnvPrefix("jflintgo")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
