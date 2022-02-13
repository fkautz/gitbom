package cmd

import (
	"fmt"
	"github.com/facebookgo/symwalk"
	"github.com/fkautz/gitbom-go"
	"github.com/rwxrob/cmdbox"
	"github.com/rwxrob/cmdbox/util"
	"io/ioutil"
	"log"
	"os"
	"path"
)

func init() {
	x := cmdbox.Add("gitbom")

	x.Method = func(args ...string) error {
		if len(args) == 0 {
			fmt.Println(util.Emph("**NAME**", 0, -1) + `
        gitbom (v0.0.1) - Generate gitboms from files

` + util.Emph("**USAGE**", 0, 01) + `
        gitbom [files]
        gitbom [file] bom [input-files]

        gitbom will create a .gitbom/ directory in the current working
        directory and store generated gitboms in .gitbom/

` + util.Emph("**LEGAL**", 0, 01) + `
        gitbom (v0.0.1) Copyright 2022 gitbom-go contributors
        SPDX-License-Identifier: Apache-2.0
`)
			return nil
		}

		if len(args) > 2 && args[1] == "bom" {
			gb := gitbom.NewGitBom()
			// generate artifact tree
			for i := 2; i < len(args); i++ {
				if err := addPathToGitbom(gb, args[i]); err != nil {
					return err
				}
			}

			// generate target gitbom with artifact tree
			if err := writeObject(".bom", gb); err != nil {
				return err
			}

			gb2 := gitbom.NewGitBom()
			info, err := os.Stat(args[0])
			if err != nil {
				return err
			}
			if err = addFileToGitbom(args[0], info, gb2, gb); err != nil {
				return err
			}

			if err := writeObject(".bom", gb2); err != nil {
				return err
			}

			fmt.Println(gb2.Identity())
		} else {
			gb := gitbom.NewGitBom()
			for i := 0; i < len(args); i++ {
				if err := addPathToGitbom(gb, args[i]); err != nil {
					log.Println(err)
					return err
				}
			}

			// generate target gitbom with artifact tree
			writeObject(".bom", gb)

			fmt.Println(gb.Identity())
		}

		return nil
	}
}

func writeObject(prefix string, gb gitbom.ArtifactTree) error {
	objectDir := path.Join(prefix, "object", gb.Identity()[:2])
	objectPath := path.Join(objectDir, gb.Identity()[2:])
	os.MkdirAll(objectDir, 0755)
	if err := ioutil.WriteFile(objectPath, []byte(gb.String()), 0644); err != nil {
		return err
	}
	return nil
}

func addPathToGitbom(gb gitbom.ArtifactTree, fileName string) error {
	err := symwalk.Walk(fileName, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			err2 := addFileToGitbom(path, info, gb, nil)
			if err2 != nil {
				return err2
			}
		}
		return nil
	})
	return err
}

func addFileToGitbom(path string, info os.FileInfo, gb gitbom.ArtifactTree, identifier gitbom.Identifier) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Printf("error closing %s: %s", path, err)
		}
	}(f)

	if err := gb.AddSha1ReferenceFromReader(f, identifier, info.Size()); err != nil {
		return err
	}
	return nil
}
