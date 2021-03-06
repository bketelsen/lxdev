// Copyright (c) 2019 bketelsen
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"devlx/path"
	lxd "github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

var (
	w bool
	s bool
)

var profiles = []string{"gui", "cli"}

// profileCmd represents the profile command
var profileCmd = &cobra.Command{
	Use:   "profile [name]",
	Short: "create or replace the provisioning profile for devlx",
	Long: `Profile creates or replaces the 'gui' and 'cli' profiles in lxc that allow you
to connect to running containers and possibly display X11 applications on the host. Run with
no arguments to create or update all required profiles.`,
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) > 0 {
			name = args[0]
		}
		err := createProfiles(name)
		if err != nil {
			log.Error("Unable to connect: " + err.Error())
			os.Exit(1)
		}
	},
}

func createProfiles(name string) error {

	log.Running("Creating lxc profiles")
	c, err := lxd.ConnectLXDUnix("/var/snap/lxd/common/lxd/unix.socket", nil)
	if err != nil {
		log.Error("Unable to connect: " + err.Error())
		return err
	}

	profs := make([]string, 0)

	if name == "" {
		profs = make([]string, len(profiles))
		copy(profs, profiles)
	} else {
		profs = append(profs, name)
	}
	for _, p := range profs {
		exists := true
		prof, etag, err := c.GetProfile(p)
		if err != nil {
			exists = false
		}
		if w {
			filename := p + ".yaml"
			fpath := filepath.Join(path.GetConfigPath(), "profiles", filename)
			f, err := os.Open(fpath)
			defer f.Close()
			if err != nil {
				log.Error("Create Profile : " + err.Error())
				return err
			}
			bb, err := ioutil.ReadAll(f)
			if err != nil {
				log.Error("Reading Profile : " + err.Error())
				return err
			}
			if exists {

				log.Running("Updating profile " + p)
				var profile api.ProfilePut
				err = yaml.Unmarshal(bb, &profile)
				if err != nil {
					log.Error("Parsing Profile : " + err.Error())
					return err
				}
				err = c.UpdateProfile(p, profile, etag)
				if err != nil {
					log.Error("Create Profile : " + err.Error())
					return err
				}

				log.Success("Updating profile " + p)
			} else {

				log.Running("Creating profile " + p)
				var profile api.ProfilesPost
				err = yaml.Unmarshal(bb, &profile)
				if err != nil {
					log.Error("Parsing Profile : " + err.Error())
					return err
				}
				profile.Name = p
				err = c.CreateProfile(profile)
				if err != nil {
					log.Error("Create Profile : " + err.Error())
					return err
				}
				log.Success("Creating profile " + p)
			}
		}

		if s {
			fmt.Println(prof, p)
		}
	}
	log.Success("Managing profiles")
	return nil
}
func init() {
	rootCmd.AddCommand(profileCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	profileCmd.PersistentFlags().String("ethernet", "", "the name of your ethernet device e.g. 'enp5s0'")
	viper.BindPFlag("ethernet", profileCmd.PersistentFlags().Lookup("ethernet"))

	profileCmd.PersistentFlags().BoolVarP(&w, "write", "w", true, "Create or update a profile")
	viper.BindPFlag("write", profileCmd.PersistentFlags().Lookup("write"))

	profileCmd.PersistentFlags().BoolVarP(&s, "show", "l", false, "Show a profile")
	viper.BindPFlag("show", profileCmd.PersistentFlags().Lookup("show"))
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// profileCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
