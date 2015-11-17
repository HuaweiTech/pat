package workloads_test

import (
	//"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/cloudfoundry-incubator/pat/context"

	. "github.com/cloudfoundry-incubator/pat/workloads"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("cf Workloads", func() {
	var (
		srcDir string
		dstDir string
	)
	BeforeEach(func() {
		srcDir = path.Join(os.TempDir(), "src")
		dstDir = path.Join(os.TempDir(), "dst")
	})
	AfterEach(func() {
		os.RemoveAll(srcDir)
		os.RemoveAll(dstDir)
	})

	Describe("Generating and Pushing an app", func() {
		Context("CopyAndReplaceText", func() {
			It("Copies the directory structure", func() {
				os.MkdirAll(path.Join(srcDir, "subdir"), 0777)
				CopyAndReplaceText(srcDir, dstDir, "", "")

				info, err := os.Lstat(dstDir)
				subInfo, err2 := os.Lstat(path.Join(dstDir, "subdir"))

				Ω(err).ShouldNot(HaveOccurred())
				Ω(err2).ShouldNot(HaveOccurred())
				Ω(info.IsDir()).Should(Equal(true))
				Ω(subInfo.IsDir()).Should(Equal(true))
			})

			It("Copies any files contained the source directory or subdirectories", func() {
				os.MkdirAll(path.Join(srcDir, "subdir"), 0777)
				file, _ := os.Create(path.Join(srcDir, "test.txt"))
				file.WriteString("abc123")
				file.Close()
				subFile, _ := os.Create(path.Join(srcDir, "subdir", "subfile.txt"))
				subFile.WriteString("foobar")
				subFile.Close()

				CopyAndReplaceText(srcDir, dstDir, "", "")
				dstFile, err := ioutil.ReadFile(path.Join(dstDir, "test.txt"))
				dstSubfile, err2 := ioutil.ReadFile(path.Join(dstDir, "subdir", "subfile.txt"))

				Ω(err).ShouldNot(HaveOccurred())
				Ω(err2).ShouldNot(HaveOccurred())
				Ω(string(dstFile)).Should(Equal("abc123"))
				Ω(string(dstSubfile)).Should(Equal("foobar"))
			})

			It("Replaces the target text in any copied files", func() {
				os.MkdirAll(path.Join(srcDir, "subdir"), 0777)
				file, _ := os.Create(path.Join(srcDir, "test.txt"))
				file.WriteString("abc123")
				file.WriteString("$RANDOM_TEXT")
				file.Close()
				subFile, _ := os.Create(path.Join(srcDir, "subdir", "subfile.txt"))
				subFile.WriteString("foobar")
				subFile.Close()

				CopyAndReplaceText(srcDir, dstDir, "$RANDOM_TEXT", "qwerty")

				dstFile, err := ioutil.ReadFile(path.Join(dstDir, "test.txt"))
				dstSubfile, err2 := ioutil.ReadFile(path.Join(dstDir, "subdir", "subfile.txt"))

				Ω(err).ShouldNot(HaveOccurred())
				Ω(err2).ShouldNot(HaveOccurred())
				Ω(strings.Contains(string(dstFile), "qwerty")).Should(Equal(true))
				Ω(strings.Contains(string(dstSubfile), "qwerty")).Should(Equal(false))
			})
		})
	})

	Describe("Mulit user", func() {
		var (
			replies    map[string]string
			cli        *dummyCfCli
			cliContext context.Context
		)

		BeforeEach(func() {
			cliContext = context.New()
			cliContext.PutInt("iterationIndex", 0)
			replies = make(map[string]string)
			cli = &dummyCfCli{"", make([]string, 0), replies}
			NewExpectCFToSay(cli.expectCFToSay)
		})
		Context("when cfhomes is not set", func() {
			BeforeEach(func() {
				replies["push"] = "App started"
				cliContext.PutString("app", "someapp")
			})

			It("cfhome should be empty string", func() {
				err := Push(cliContext)
				Ω(err).ShouldNot(HaveOccurred())
				cli.ShouldHaveBeenCalledWith("", "push")

			})
		})

		Context("when cfhomes is set", func() {
			BeforeEach(func() {
				replies["home1push"] = "App started"
				replies["home2push"] = "App started"
				replies["home3push"] = "App started"
				cliContext.PutString("app", "someapp")
				cliContext.PutString("cfhomes", "home1,home2,home3")
			})

			It("cfhome should be different from iterationIndex", func() {
				cliContext.PutInt("iterationIndex", 0)
				err := Push(cliContext)
				Ω(err).ShouldNot(HaveOccurred())
				cli.ShouldHaveBeenCalledWith("home1", "push")
				cliContext.PutInt("iterationIndex", 1)
				err = Push(cliContext)
				Ω(err).ShouldNot(HaveOccurred())
				cli.ShouldHaveBeenCalledWith("home2", "push")
				cliContext.PutInt("iterationIndex", 2)
				err = Push(cliContext)
				Ω(err).ShouldNot(HaveOccurred())
				cli.ShouldHaveBeenCalledWith("home3", "push")
				cliContext.PutInt("iterationIndex", 3)
				err = Push(cliContext)
				Ω(err).ShouldNot(HaveOccurred())
				cli.ShouldHaveBeenCalledWith("home1", "push")

			})
		})

	})

})

type dummyCfCli struct {
	cfhome  string
	args    []string
	replies map[string]string
}

func (cli *dummyCfCli) ShouldHaveBeenCalledWith(cfhome, method string, args ...string) {
	Ω(cli.cfhome).Should(Equal(cfhome))
	Ω(cli.args[0]).Should(Equal(method))
}

func (cli *dummyCfCli) expectCFToSay(expect, cfhome string, args ...string) error {
	cli.cfhome = cfhome
	for _, arg := range args {
		cli.args = append(cli.args, arg)
	}

	if len(cli.args) == 0 {
		return errors.New("no method")
	}

	if cli.replies[cli.cfhome+cli.args[0]] != expect {
		return errors.New(fmt.Sprintf("expect :%s, but :%s", expect, cli.replies[cli.cfhome+cli.args[0]]))
	}
	return nil
}
