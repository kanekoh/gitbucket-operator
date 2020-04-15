package gitbucket

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)


func TestGitbucketController(t *testing.T){
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gitbuckt Controller SUite")
}