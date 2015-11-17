package workloads

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/runner"
	"github.com/cloudfoundry-incubator/pat/context"
	"github.com/nu7hatch/gouuid"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega/gexec"
	. "github.com/pivotal-cf-experimental/cf-test-helpers/cf"
)

//Todo(simon) Remove, for dev testing only
func random(min, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	r := min + rand.Intn(max-min)
	return r
}

func Dummy(ctx context.Context) error {
	guid, _ := uuid.NewV4()
	appName := "pats-" + guid.String()
	appNames, _ := ctx.GetString("appNames")

	if appNames != "" {
		appNames += fmt.Sprintf(",%s", appName)
	} else {
		appNames = appName
	}
	ctx.PutString("appNames", appNames)

	time.Sleep(time.Duration(random(1, 5)) * time.Second)
	return nil
}

func DummyDelete(ctx context.Context) error {
	time.Sleep(time.Duration(random(1, 5)) * time.Second)
	return nil
}

func DummyWithErrors(ctx context.Context) error {
	Dummy(ctx)
	if random(0, 10) > 8 {
		return errors.New("Random (dummy) error")
	}
	return nil
}

func Push(ctx context.Context) error {
	guid, _ := uuid.NewV4()
	pathToApp, _ := ctx.GetString("app")
	pathToManifest, _ := ctx.GetString("app:manifest")
	appName := "pats-" + guid.String()
	appNames, _ := ctx.GetString("appNames")

	if appNames != "" {
		appNames += fmt.Sprintf(",%s", appName)
	} else {
		appNames = appName
	}
	ctx.PutString("appNames", appNames)

	cfhome := cfhomeForWorker(ctx)
	if pathToManifest == "" {
		return expectCfToSay("App started", cfhome, "push", appName, "-m", "64M", "-p", pathToApp)
	} else {
		return expectCfToSay("App started", cfhome, "push", appName, "-p", pathToApp, "-f", pathToManifest)
	}
}

func Delete(ctx context.Context) error {
	appNames, _ := ctx.GetString("appNames")
	if appNames == "" {
		return errors.New("No app to delete")
	}
	appNamesArray := strings.Split(appNames, ",")
	appNameToDelete := appNamesArray[len(appNamesArray)-1]

	appNames = strings.Replace(appNames, ","+appNameToDelete, "", -1)
	appNames = strings.Replace(appNames, appNameToDelete, "", -1)
	ctx.PutString("appNames", appNames)
	cfhome := cfhomeForWorker(ctx)
	return expectCfToSay("Deleting app", cfhome, "delete", appNameToDelete, "-f")
}

func CopyAndReplaceText(srcDir string, dstDir string, searchText string, replaceText string) error {
	return filepath.Walk(srcDir, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		pathTail := strings.SplitAfter(file, srcDir)[1]
		if info.IsDir() {
			err = os.Mkdir(path.Join(dstDir, pathTail), 0777)
			if err != nil {
				return err
			}
		} else if info.Mode().IsRegular() {
			input, err := ioutil.ReadFile(file)
			if err != nil {
				return err
			}
			inputString := strings.Replace(string(input), searchText, replaceText, -1)
			input = []byte(inputString)
			output, err := os.Create(path.Join(dstDir, pathTail))
			if err != nil {
				return err
			}
			defer output.Close()
			output.Write(input)
		}
		return err
	})
}

func GenerateAndPush(ctx context.Context) error {
	pathToApp, _ := ctx.GetString("app")
	pathToManifest, _ := ctx.GetString("app:manifest")

	guid, _ := uuid.NewV4()
	rand.Seed(time.Now().UTC().UnixNano())
	salt := strconv.FormatInt(rand.Int63(), 10)

	dstDir := path.Join(os.TempDir(), salt)
	defer os.RemoveAll(dstDir)

	err := CopyAndReplaceText(pathToApp, dstDir, "$RANDOM_TEXT", salt)
	if err != nil {
		return err
	}

	cfhome := cfhomeForWorker(ctx)
	if pathToManifest == "" {
		return expectCfToSay("App started", cfhome, "push", "pats-"+guid.String(), "-m", "64M", "-p", pathToApp)
	} else {
		return expectCfToSay("App started", cfhome, "push", "pats-"+guid.String(), "-p", pathToApp, "-f", pathToManifest)
	}
}

type ExpectToSay func(expect string, home string, args ...string) error

var expectCfToSay ExpectToSay

func NewExpectCFToSay(expectfunc ExpectToSay) {
	expectCfToSay = expectfunc
}

func init() {
	NewExpectCFToSay(func(expect string, cfhome string, args ...string) error {
		var outBuffer bytes.Buffer
		oldWriter := ginkgo.GinkgoWriter
		ginkgo.GinkgoWriter = bufio.NewWriter(&outBuffer)
		if cfhome == "" {
			cfOutBuffer := Cf(args...).Wait(10 * time.Minute).Out
			cfContents := cfOutBuffer.Contents()
			ginkgo.GinkgoWriter = oldWriter
			if strings.Contains(string(cfContents), expect) {
				return nil
			} else {
				return errors.New(string(cfContents))
			}
		} else {
			cfOutBuffer := envCf(cfhome, args...).Wait(10 * time.Minute).Out
			cfContents := cfOutBuffer.Contents()
			ginkgo.GinkgoWriter = oldWriter
			if strings.Contains(string(cfContents), expect) {
				return nil
			} else {
				return errors.New(string(cfContents))
			}
		}
	})
}

func envCf(cfhome string, args ...string) *Session {
	env := fmt.Sprintf("CF_HOME=%s", cfhome)
	args = append([]string{env, "cf"}, args...)
	return runner.Run("env", args...)
}

func cfhomeForWorker(ctx context.Context) string {
	iterationIndex, exist := ctx.GetInt("iterationIndex")
	if !exist {
		return ""
	}
	cfhomes, exist := ctx.GetString("cfhomes")
	if !exist || len(cfhomes) == 0 {
		return ""
	}
	var cfhomeList = strings.Split(cfhomes, ",")
	return cfhomeList[iterationIndex%len(cfhomeList)]
}
