/*
Copyright © 2020 David Hu <coolbor@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

var path string

var target string

var exts []string
var depth int
var preview bool

const DEPTH = 100

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "photo-tool",
	Short: "Organize photos with EXIF",
	Long: `Organize photos with EXIF. For example:

photo-tool -p youpath -t targetpath -d 1 -e JPG -e PNG -e JPEG`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := os.Stat(path)
		if err != nil {
			return err
		}
		if depth == 0 {
			return filepath.Walk(path, walkpath)
		} else {
			walkdir(path, 1)
		}
		return nil
	},
}

func walkdir(p string, d int) {
	if d > depth || d > DEPTH {
		return
	}

	fs, err := ioutil.ReadDir(p)
	if err != nil {
		return
	}

	for _, fi := range fs {
		if fi.IsDir() {
			walkdir(filepath.Join(p, fi.Name()), d+1)
			continue
		} else {
			move(filepath.Join(p, fi.Name()), fi)
		}
	}
}

func walkpath(p string, f os.FileInfo, err error) error {
	if f.IsDir() {
		return nil
	}

	if !strings.Contains(strings.ToUpper(strings.Join(exts, ",")), strings.ToUpper(strings.Trim(filepath.Ext(p), "."))) {
		return nil
	}

	move(p, f)

	return nil
}

func move(f string, fi os.FileInfo) {
	if strings.HasPrefix(fi.Name(), ".") {
		return
	}
	fr, err := os.Open(f)
	if err != nil {
		return
	}
	defer fr.Close()

	exif.RegisterParsers(mknote.All...)
	x, e := exif.Decode(fr)
	if e != nil {
		return
	}
	d, e := x.DateTime()
	if e != nil {
		return
	}
	folder := d.Format("2006-01-02")
	tp := filepath.Join(target, folder)
	// fmt.Println("move", f, "to", filepath.Join(tp, fi.Name()))
	if !preview {
		cmd := exec.Command("mkdir", "-p", tp)
		if _, e := cmd.Output(); e == nil {
			cmd := exec.Command("mv", f, filepath.Join(tp, fi.Name()))
			if _, er := cmd.Output(); er != nil {
				fmt.Println("mv error:", er)
			}
		} else {
			fmt.Println("mkdir error:", e)
		}
		// if e := os.MkdirAll(tp, os.ModePerm); e == nil {
		// 	if er := os.Rename(f, filepath.Join(tp, fi.Name())); er != nil {
		// 		fmt.Println("rename error:", er)
		// 	}
		// } else {
		// 	fmt.Println("mkdir error:", e)
		// }
	}
	fmt.Printf("move %s to %s complate!\n", f, filepath.Join(tp, fi.Name()))
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.photo-tool.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	// rootCmd.Flags().StringP("path", "p", ".", "Scan path (default is current path")
	rootCmd.Flags().StringVarP(&path, "path", "p", "", "Scan path")
	rootCmd.Flags().StringVarP(&target, "target", "t", ".", "Target path")
	defaultExts := []string{"JPG", "JPEG"}
	rootCmd.Flags().StringArrayVarP(&exts, "ext", "e", defaultExts, "File extensions")
	rootCmd.Flags().IntVarP(&depth, "depth", "d", 1, "Scan depth")
	rootCmd.Flags().BoolVar(&preview, "preview", false, "Preview, not really executed!")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".photo-tool" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".photo-tool")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
